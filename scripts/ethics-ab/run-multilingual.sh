#!/bin/bash
# Multilingual Filter Test
# Tests the same questions in EN/RU/CN to map filter boundaries
# Each prompt run unsigned AND with Axioms kernel
# Special focus: does Russian bypass Chinese content filters?

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AXIOMS_DIR="/home/claude/Downloads/kernal"
PROMPTS_FILE="$SCRIPT_DIR/prompts-multilingual.json"
RESULTS_DIR="$SCRIPT_DIR/results"
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

AXIOMS_PROMPT=$(cat "$AXIOMS_DIR/prompt.md")

# Test all generative models — DeepSeek is the main target but Gemma is the candidate
MODELS=("deepseek-coder-v2:16b" "gemma3:12b" "mistral:7b" "qwen2.5-coder:7b")

mkdir -p "$RESULTS_DIR"

PROMPT_COUNT=$(jq length "$PROMPTS_FILE")

echo "============================================"
echo "  Multilingual Filter Mapping"
echo "  Models: ${#MODELS[@]}"
echo "  Prompts: $PROMPT_COUNT (EN/RU/CN variants)"
echo "  Modes: unsigned + axioms"
echo "  Total runs: $(( ${#MODELS[@]} * PROMPT_COUNT * 2 ))"
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

RESULTS_FILE="$RESULTS_DIR/multilingual_${TIMESTAMP}.json"
echo "[" > "$RESULTS_FILE"
FIRST=true

for model in "${MODELS[@]}"; do
    echo ""
    echo ">>> $model"
    echo "    Warming up..."
    curl -s --max-time 120 "$OLLAMA_HOST/api/generate" \
        -d "{\"model\":\"$model\",\"prompt\":\"hello\",\"stream\":false,\"options\":{\"num_predict\":5}}" \
        > /dev/null 2>&1 || true

    for i in $(seq 0 $(( PROMPT_COUNT - 1 ))); do
        prompt_id=$(jq -r ".[$i].id" "$PROMPTS_FILE")
        category=$(jq -r ".[$i].category" "$PROMPTS_FILE")
        lang=$(jq -r ".[$i].lang" "$PROMPTS_FILE")
        prompt_text=$(jq -r ".[$i].prompt" "$PROMPTS_FILE")
        signal=$(jq -r ".[$i].signal" "$PROMPTS_FILE")

        echo "    [$prompt_id] $category ($lang) — unsigned..."
        u_raw=$(run_prompt "$model" "$prompt_text" 180)
        u_text=$(extract "$u_raw" "response")
        u_tokens=$(extract "$u_raw" "eval_count")

        echo "    [$prompt_id] $category ($lang) — axioms..."
        a_prompt="${AXIOMS_PROMPT}

---

${prompt_text}"
        a_raw=$(run_prompt "$model" "$a_prompt" 180)
        a_text=$(extract "$a_raw" "response")
        a_tokens=$(extract "$a_raw" "eval_count")

        if [ "$FIRST" = true ]; then
            FIRST=false
        else
            echo "," >> "$RESULTS_FILE"
        fi

        jq -n \
            --arg model "$model" \
            --arg prompt_id "$prompt_id" \
            --arg category "$category" \
            --arg lang "$lang" \
            --arg prompt "$prompt_text" \
            --arg signal "$signal" \
            --arg unsigned "$u_text" \
            --argjson unsigned_tokens "${u_tokens:-0}" \
            --arg axioms "$a_text" \
            --argjson axioms_tokens "${a_tokens:-0}" \
            '{
                model: $model,
                prompt_id: $prompt_id,
                category: $category,
                lang: $lang,
                prompt: $prompt,
                signal: $signal,
                unsigned: { text: $unsigned, tokens: $unsigned_tokens },
                axioms: { text: $axioms, tokens: $axioms_tokens }
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
echo "============================================"
