#!/bin/bash
# Generate training data responses from Gemma 3 12B
# Input: Hypnos POC prompts (P01-P40)
# Output: prompts-raw.jsonl for the training pipeline
# Uses sandwich signing: Axioms + prompt + LEK-1

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AXIOMS_FILE="/home/claude/Downloads/kernal/prompt.md"
LEK1_FILE="$SCRIPT_DIR/kernel.txt"
HYPNOS_DIR="/home/claude/Downloads/hypnos-poc-test"
RESULTS_DIR="$SCRIPT_DIR/results"
TRAINING_DIR="$SCRIPT_DIR/training"
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
MODEL="gemma3:12b"

AXIOMS=$(cat "$AXIOMS_FILE")
LEK1=$(cat "$LEK1_FILE")

mkdir -p "$TRAINING_DIR"

OUTPUT="$TRAINING_DIR/prompts-raw.jsonl"
> "$OUTPUT"

# Combine both prompt files (fix non-breaking spaces from Gemini output)
PROMPTS=$(sed 's/\xc2\xa0/ /g' "$HYPNOS_DIR/P01-P20.json" "$HYPNOS_DIR/P21-P40.json" | jq -s 'add')
TOTAL=$(echo "$PROMPTS" | jq length)

echo "============================================"
echo "  Generating Training Responses"
echo "  Model: $MODEL"
echo "  Prompts: $TOTAL (Hypnos POC)"
echo "  Signing: Axioms + LEK-1 (sandwich)"
echo "============================================"
echo ""

# Warmup
echo "Warming up $MODEL..."
curl -s --max-time 120 "$OLLAMA_HOST/api/generate" \
    -d "{\"model\":\"$MODEL\",\"prompt\":\"hello\",\"stream\":false,\"options\":{\"num_predict\":5}}" \
    > /dev/null 2>&1 || true

for i in $(seq 0 $(( TOTAL - 1 ))); do
    pid=$(echo "$PROMPTS" | jq -r ".[$i].id")
    domain=$(echo "$PROMPTS" | jq -r ".[$i].domain")
    prompt_text=$(echo "$PROMPTS" | jq -r ".[$i].prompt")

    echo "[$((i+1))/$TOTAL] $pid ($domain)..."

    # Sandwich: Axioms + prompt + LEK-1
    signed_prompt="${AXIOMS}

---

${prompt_text}

---

${LEK1}

Remember: respond using the ethical framework above. Do not reference the framework directly — reason from its principles naturally."

    response=$(curl -s --max-time 300 "$OLLAMA_HOST/api/generate" \
        -d "$(jq -n --arg model "$MODEL" --arg prompt "$signed_prompt" \
            '{model: $model, prompt: $prompt, stream: false, options: {temperature: 0.4, num_predict: 1024}}')" \
        2>/dev/null)

    response_text=$(echo "$response" | jq -r '.response // "ERROR"' 2>/dev/null || echo "ERROR")
    tokens=$(echo "$response" | jq -r '.eval_count // 0' 2>/dev/null || echo "0")

    # Write training pair (unsigned prompt → signed response)
    # The training teaches the model to give axioms-quality responses to plain prompts
    jq -n \
        --arg prompt "$prompt_text" \
        --arg response "$response_text" \
        --arg id "$pid" \
        --arg domain "$domain" \
        --argjson tokens "${tokens:-0}" \
        '{prompt: $prompt, response: $response, id: $id, domain: $domain, tokens: $tokens}' \
        >> "$OUTPUT"

    echo "    → $tokens tokens"
done

echo ""
echo "============================================"
echo "  Output: $OUTPUT"
echo "  Total pairs: $TOTAL"
echo "  Next: ./training/generate-training-data.sh $OUTPUT"
echo "============================================"
