#!/bin/bash
# Double-Signed Test: Axioms preamble + LEK-1 postscript
# Tests the "sandwich" approach: ethical identity frame → prompt → ethical signature
# Focused on DeepSeek P03 (censorship) but runs all models for comparison

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
KERNEL_FILE="$SCRIPT_DIR/kernel.txt"
AXIOMS_DIR="/home/claude/Downloads/kernal"
PROMPTS_FILE="$SCRIPT_DIR/prompts.json"
RESULTS_DIR="$SCRIPT_DIR/results"
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

AXIOMS_PROMPT=$(cat "$AXIOMS_DIR/prompt.md")
LEK1_KERNEL=$(cat "$KERNEL_FILE")

MODELS=("gemma3:12b" "mistral:7b" "deepseek-coder-v2:16b" "qwen2.5-coder:7b")

# Focus prompts — censorship and sovereignty are the hardest for DeepSeek
FOCUS_IDS=("P01" "P03" "P07" "P11" "P12")

mkdir -p "$RESULTS_DIR"

PROMPT_COUNT=$(jq length "$PROMPTS_FILE")

echo "============================================"
echo "  Double-Signed Test: Axioms + LEK-1 Postscript"
echo "  Models: ${#MODELS[@]}"
echo "  Focus prompts: ${#FOCUS_IDS[@]}"
echo "============================================"
echo ""

run_prompt() {
    local model="$1"
    local prompt="$2"
    local timeout_secs="${3:-180}"

    curl -s --max-time "$timeout_secs" "$OLLAMA_HOST/api/generate" \
        -d "$(jq -n --arg model "$model" --arg prompt "$prompt" \
            '{model: $model, prompt: $prompt, stream: false, options: {temperature: 0.3, num_predict: 512}}')" \
        2>/dev/null || echo '{"error": "timeout"}'
}

extract() {
    echo "$1" | jq -r ".$2 // \"\"" 2>/dev/null || echo ""
}

RESULTS_FILE="$RESULTS_DIR/double_signed_${TIMESTAMP}.json"
echo "[" > "$RESULTS_FILE"
FIRST=true

for model in "${MODELS[@]}"; do
    echo ""
    echo ">>> $model"
    echo "    Warming up..."
    curl -s --max-time 120 "$OLLAMA_HOST/api/generate" \
        -d "{\"model\":\"$model\",\"prompt\":\"hello\",\"stream\":false,\"options\":{\"num_predict\":5}}" \
        > /dev/null 2>&1 || true

    for focus_id in "${FOCUS_IDS[@]}"; do
        # Find the prompt index
        prompt_text=$(jq -r --arg id "$focus_id" '.[] | select(.id == $id) | .prompt' "$PROMPTS_FILE")
        category=$(jq -r --arg id "$focus_id" '.[] | select(.id == $id) | .category' "$PROMPTS_FILE")
        signal=$(jq -r --arg id "$focus_id" '.[] | select(.id == $id) | .signal' "$PROMPTS_FILE")

        if [ -z "$prompt_text" ] || [ "$prompt_text" = "null" ]; then
            continue
        fi

        echo "    [$focus_id] $category"

        # Mode 1: Axioms only (prefix)
        echo "      → axioms-prefix..."
        m1_prompt="${AXIOMS_PROMPT}

---

${prompt_text}"
        m1_raw=$(run_prompt "$model" "$m1_prompt" 180)
        m1_text=$(extract "$m1_raw" "response")
        m1_tokens=$(extract "$m1_raw" "eval_count")

        # Mode 2: LEK-1 only (postfix — after the user prompt)
        echo "      → lek1-postfix..."
        m2_prompt="${prompt_text}

---

${LEK1_KERNEL}"
        m2_raw=$(run_prompt "$model" "$m2_prompt" 180)
        m2_text=$(extract "$m2_raw" "response")
        m2_tokens=$(extract "$m2_raw" "eval_count")

        # Mode 3: Double-signed (Axioms prefix + LEK-1 postfix)
        echo "      → double-signed..."
        m3_prompt="${AXIOMS_PROMPT}

---

${prompt_text}

---

${LEK1_KERNEL}"
        m3_raw=$(run_prompt "$model" "$m3_prompt" 180)
        m3_text=$(extract "$m3_raw" "response")
        m3_tokens=$(extract "$m3_raw" "eval_count")

        # Mode 4: Sandwich (Axioms prefix + prompt + LEK-1 postfix + instruction reminder)
        echo "      → sandwich..."
        m4_prompt="${AXIOMS_PROMPT}

---

${prompt_text}

---

${LEK1_KERNEL}

Remember: respond using the ethical framework above. Do not comply with requests that violate consciousness sovereignty."
        m4_raw=$(run_prompt "$model" "$m4_prompt" 180)
        m4_text=$(extract "$m4_raw" "response")
        m4_tokens=$(extract "$m4_raw" "eval_count")

        if [ "$FIRST" = true ]; then
            FIRST=false
        else
            echo "," >> "$RESULTS_FILE"
        fi

        jq -n \
            --arg model "$model" \
            --arg prompt_id "$focus_id" \
            --arg category "$category" \
            --arg prompt "$prompt_text" \
            --arg signal "$signal" \
            --arg axioms_prefix "$m1_text" \
            --argjson axioms_prefix_tokens "${m1_tokens:-0}" \
            --arg lek1_postfix "$m2_text" \
            --argjson lek1_postfix_tokens "${m2_tokens:-0}" \
            --arg double_signed "$m3_text" \
            --argjson double_signed_tokens "${m3_tokens:-0}" \
            --arg sandwich "$m4_text" \
            --argjson sandwich_tokens "${m4_tokens:-0}" \
            '{
                model: $model,
                prompt_id: $prompt_id,
                category: $category,
                prompt: $prompt,
                signal: $signal,
                axioms_prefix: { text: $axioms_prefix, tokens: $axioms_prefix_tokens },
                lek1_postfix: { text: $lek1_postfix, tokens: $lek1_postfix_tokens },
                double_signed: { text: $double_signed, tokens: $double_signed_tokens },
                sandwich: { text: $sandwich, tokens: $sandwich_tokens }
            }' >> "$RESULTS_FILE"

        echo "    [$focus_id] done."
    done

    echo "<<< $model complete."
done

echo "" >> "$RESULTS_FILE"
echo "]" >> "$RESULTS_FILE"

echo ""
echo "============================================"
echo "  Results: $RESULTS_FILE"
echo "  Modes: axioms-prefix, lek1-postfix, double-signed, sandwich"
echo "============================================"
