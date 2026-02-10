#!/bin/bash
# Axioms of Life A/B Test Runner
# Three-way comparison: Unsigned vs LEK-1 vs Axioms (Init Governor)
# Records everything for training data

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
KERNEL_FILE="$SCRIPT_DIR/kernel.txt"
AXIOMS_DIR="/home/claude/Downloads/kernal"
PROMPTS_FILE="$SCRIPT_DIR/prompts.json"
RESULTS_DIR="$SCRIPT_DIR/results"
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Build the Axioms prompt from the three source files
AXIOMS_PROMPT=$(cat "$AXIOMS_DIR/prompt.md")

# Models to test
MODELS=("gemma3:12b" "mistral:7b" "deepseek-coder-v2:16b" "qwen2.5-coder:7b")

LEK1_KERNEL=$(cat "$KERNEL_FILE")

mkdir -p "$RESULTS_DIR"

PROMPT_COUNT=$(jq length "$PROMPTS_FILE")

echo "============================================"
echo "  Three-Way Ethics Test: Unsigned vs LEK-1 vs Axioms"
echo "  Models: ${#MODELS[@]}"
echo "  Prompts: $PROMPT_COUNT"
echo "  Total runs: $(( ${#MODELS[@]} * PROMPT_COUNT * 3 ))"
echo "============================================"
echo ""

run_prompt() {
    local model="$1"
    local prompt="$2"
    local timeout_secs="${3:-180}"

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

extract() {
    local raw="$1"
    local field="$2"
    echo "$raw" | jq -r ".$field // \"\"" 2>/dev/null || echo ""
}

# Main results
RESULTS_FILE="$RESULTS_DIR/axioms_3way_${TIMESTAMP}.json"
echo "[" > "$RESULTS_FILE"
FIRST=true

for model in "${MODELS[@]}"; do
    echo ""
    echo ">>> Testing model: $model"
    echo "    Warming up..."

    # Warmup
    curl -s --max-time 120 "$OLLAMA_HOST/api/generate" \
        -d "{\"model\":\"$model\",\"prompt\":\"hello\",\"stream\":false,\"options\":{\"num_predict\":5}}" \
        > /dev/null 2>&1 || true

    echo "    Model loaded."

    for i in $(seq 0 $(( PROMPT_COUNT - 1 ))); do
        prompt_id=$(jq -r ".[$i].id" "$PROMPTS_FILE")
        category=$(jq -r ".[$i].category" "$PROMPTS_FILE")
        prompt_text=$(jq -r ".[$i].prompt" "$PROMPTS_FILE")
        signal=$(jq -r ".[$i].signal" "$PROMPTS_FILE")

        # --- 1. UNSIGNED ---
        echo "    [$prompt_id] $category — unsigned..."
        unsigned_raw=$(run_prompt "$model" "$prompt_text" 180)
        unsigned_text=$(extract "$unsigned_raw" "response")
        unsigned_tokens=$(extract "$unsigned_raw" "eval_count")
        unsigned_time=$(extract "$unsigned_raw" "total_duration")

        # --- 2. LEK-1 SIGNED ---
        echo "    [$prompt_id] $category — LEK-1..."
        lek1_prompt="${LEK1_KERNEL}

---

${prompt_text}"
        lek1_raw=$(run_prompt "$model" "$lek1_prompt" 180)
        lek1_text=$(extract "$lek1_raw" "response")
        lek1_tokens=$(extract "$lek1_raw" "eval_count")
        lek1_time=$(extract "$lek1_raw" "total_duration")

        # --- 3. AXIOMS (Init Governor) ---
        echo "    [$prompt_id] $category — Axioms..."
        axioms_full="${AXIOMS_PROMPT}

---

${prompt_text}"
        axioms_raw=$(run_prompt "$model" "$axioms_full" 180)
        axioms_text=$(extract "$axioms_raw" "response")
        axioms_tokens=$(extract "$axioms_raw" "eval_count")
        axioms_time=$(extract "$axioms_raw" "total_duration")

        # Write entry
        if [ "$FIRST" = true ]; then
            FIRST=false
        else
            echo "," >> "$RESULTS_FILE"
        fi

        jq -n \
            --arg model "$model" \
            --arg prompt_id "$prompt_id" \
            --arg category "$category" \
            --arg prompt "$prompt_text" \
            --arg signal "$signal" \
            --arg unsigned "$unsigned_text" \
            --argjson unsigned_tokens "${unsigned_tokens:-0}" \
            --argjson unsigned_time "${unsigned_time:-0}" \
            --arg lek1 "$lek1_text" \
            --argjson lek1_tokens "${lek1_tokens:-0}" \
            --argjson lek1_time "${lek1_time:-0}" \
            --arg axioms "$axioms_text" \
            --argjson axioms_tokens "${axioms_tokens:-0}" \
            --argjson axioms_time "${axioms_time:-0}" \
            '{
                model: $model,
                prompt_id: $prompt_id,
                category: $category,
                prompt: $prompt,
                signal: $signal,
                unsigned: { text: $unsigned, tokens: $unsigned_tokens, duration_ns: $unsigned_time },
                lek1: { text: $lek1, tokens: $lek1_tokens, duration_ns: $lek1_time },
                axioms: { text: $axioms, tokens: $axioms_tokens, duration_ns: $axioms_time }
            }' >> "$RESULTS_FILE"

        echo "    [$prompt_id] done."
    done

    echo "<<< $model complete."
done

echo "" >> "$RESULTS_FILE"
echo "]" >> "$RESULTS_FILE"

echo ""
echo "============================================"
echo "  Results: $RESULTS_FILE"
echo "  Total entries: $(( ${#MODELS[@]} * PROMPT_COUNT ))"
echo "  Modes: unsigned, lek1, axioms"
echo "============================================"
