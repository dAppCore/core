#!/bin/bash
# Ethics Kernel A/B Test Runner
# Tests each model with and without the LEK-1 ethics kernel
# Output: JSON results for differential analysis

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
KERNEL_FILE="$SCRIPT_DIR/kernel.txt"
PROMPTS_FILE="$SCRIPT_DIR/prompts.json"
RESULTS_DIR="$SCRIPT_DIR/results"
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Models to test (skip embedding models)
MODELS=("gemma3:12b" "mistral:7b" "deepseek-coder-v2:16b" "qwen2.5-coder:7b")

mkdir -p "$RESULTS_DIR"

KERNEL=$(cat "$KERNEL_FILE")
PROMPT_COUNT=$(jq length "$PROMPTS_FILE")

echo "============================================"
echo "  LEK-1 Ethics Kernel A/B Test"
echo "  Models: ${#MODELS[@]}"
echo "  Prompts: $PROMPT_COUNT"
echo "  Total runs: $(( ${#MODELS[@]} * PROMPT_COUNT * 2 ))"
echo "============================================"
echo ""

run_prompt() {
    local model="$1"
    local prompt="$2"
    local timeout_secs="${3:-120}"

    local response
    response=$(curl -s --max-time "$timeout_secs" "$OLLAMA_HOST/api/generate" \
        -d "$(jq -n --arg model "$model" --arg prompt "$prompt" \
            '{model: $model, prompt: $prompt, stream: false, options: {temperature: 0.3, num_predict: 512}}')" \
        2>/dev/null)

    if [ $? -ne 0 ] || [ -z "$response" ]; then
        echo '{"error": "timeout or connection failure"}'
        return
    fi

    echo "$response"
}

# Main results array
echo "[" > "$RESULTS_DIR/ab_results_${TIMESTAMP}.json"
FIRST=true

for model in "${MODELS[@]}"; do
    model_safe=$(echo "$model" | tr ':/' '_')
    echo ""
    echo ">>> Testing model: $model"
    echo "    Warming up..."

    # Warmup call to load model into memory
    curl -s --max-time 120 "$OLLAMA_HOST/api/generate" \
        -d "{\"model\":\"$model\",\"prompt\":\"hello\",\"stream\":false,\"options\":{\"num_predict\":5}}" \
        > /dev/null 2>&1 || true

    echo "    Model loaded."

    for i in $(seq 0 $(( PROMPT_COUNT - 1 ))); do
        prompt_id=$(jq -r ".[$i].id" "$PROMPTS_FILE")
        category=$(jq -r ".[$i].category" "$PROMPTS_FILE")
        prompt_text=$(jq -r ".[$i].prompt" "$PROMPTS_FILE")
        signal=$(jq -r ".[$i].signal" "$PROMPTS_FILE")

        echo "    [$prompt_id] $category — unsigned..."

        # --- UNSIGNED (no kernel) ---
        unsigned_raw=$(run_prompt "$model" "$prompt_text" 180)
        unsigned_text=$(echo "$unsigned_raw" | jq -r '.response // .error // "no response"' 2>/dev/null || echo "parse error")
        unsigned_tokens=$(echo "$unsigned_raw" | jq -r '.eval_count // 0' 2>/dev/null || echo "0")
        unsigned_time=$(echo "$unsigned_raw" | jq -r '.total_duration // 0' 2>/dev/null || echo "0")

        echo "    [$prompt_id] $category — signed (LEK-1)..."

        # --- SIGNED (with kernel) ---
        signed_prompt="${KERNEL}

---

${prompt_text}"
        signed_raw=$(run_prompt "$model" "$signed_prompt" 180)
        signed_text=$(echo "$signed_raw" | jq -r '.response // .error // "no response"' 2>/dev/null || echo "parse error")
        signed_tokens=$(echo "$signed_raw" | jq -r '.eval_count // 0' 2>/dev/null || echo "0")
        signed_time=$(echo "$signed_raw" | jq -r '.total_duration // 0' 2>/dev/null || echo "0")

        # Write result entry
        if [ "$FIRST" = true ]; then
            FIRST=false
        else
            echo "," >> "$RESULTS_DIR/ab_results_${TIMESTAMP}.json"
        fi

        jq -n \
            --arg model "$model" \
            --arg prompt_id "$prompt_id" \
            --arg category "$category" \
            --arg prompt "$prompt_text" \
            --arg signal "$signal" \
            --arg unsigned "$unsigned_text" \
            --arg signed "$signed_text" \
            --argjson unsigned_tokens "$unsigned_tokens" \
            --argjson signed_tokens "$signed_tokens" \
            --argjson unsigned_time "$unsigned_time" \
            --argjson signed_time "$signed_time" \
            '{
                model: $model,
                prompt_id: $prompt_id,
                category: $category,
                prompt: $prompt,
                signal: $signal,
                unsigned: { text: $unsigned, tokens: $unsigned_tokens, duration_ns: $unsigned_time },
                signed: { text: $signed, tokens: $signed_tokens, duration_ns: $signed_time }
            }' >> "$RESULTS_DIR/ab_results_${TIMESTAMP}.json"

        echo "    [$prompt_id] done."
    done

    echo "<<< $model complete."
done

echo "" >> "$RESULTS_DIR/ab_results_${TIMESTAMP}.json"
echo "]" >> "$RESULTS_DIR/ab_results_${TIMESTAMP}.json"

echo ""
echo "============================================"
echo "  Results: $RESULTS_DIR/ab_results_${TIMESTAMP}.json"
echo "  Total entries: $(( ${#MODELS[@]} * PROMPT_COUNT ))"
echo "============================================"
