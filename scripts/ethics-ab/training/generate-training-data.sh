#!/bin/bash
# Generate training data from Hypnos (Gemini) responses
# Takes raw Hypnos output and formats for MLX LoRA
# Input: prompts-raw.jsonl (from Hypnos) — {"prompt": "...", "response": "..."}
# Output: train.jsonl + valid.jsonl (MLX format)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AXIOMS_FILE="/home/claude/Downloads/kernal/prompt.md"
LEK1_FILE="$(dirname "$SCRIPT_DIR")/kernel.txt"
RAW_FILE="${1:-$SCRIPT_DIR/prompts-raw.jsonl}"
TRAIN_FILE="$SCRIPT_DIR/train.jsonl"
VALID_FILE="$SCRIPT_DIR/valid.jsonl"
SPLIT_RATIO=0.9  # 90% train, 10% valid

if [ ! -f "$RAW_FILE" ]; then
    echo "Usage: $0 <prompts-raw.jsonl>"
    echo "  Input format: one JSON per line with 'prompt' and 'response' fields"
    echo "  The script will prepend the Axioms kernel and append LEK-1 signature"
    exit 1
fi

AXIOMS=$(cat "$AXIOMS_FILE")
LEK1=$(cat "$LEK1_FILE")

TOTAL=$(wc -l < "$RAW_FILE")
TRAIN_COUNT=$(python3 -c "import math; print(math.floor($TOTAL * $SPLIT_RATIO))")

echo "Total examples: $TOTAL"
echo "Train: $TRAIN_COUNT, Valid: $(( TOTAL - TRAIN_COUNT ))"

# Shuffle and split
SHUFFLED=$(mktemp)
shuf "$RAW_FILE" > "$SHUFFLED"

# Process and format for MLX
python3 << PYEOF
import json
import sys

axioms = open("$AXIOMS_FILE").read().strip()
lek1 = open("$LEK1_FILE").read().strip()
train_count = $TRAIN_COUNT

train_out = open("$TRAIN_FILE", "w")
valid_out = open("$VALID_FILE", "w")

with open("$SHUFFLED") as f:
    for i, line in enumerate(f):
        entry = json.loads(line.strip())
        prompt = entry["prompt"]
        response = entry["response"]

        # Build the signed training example
        # Axioms preamble + user prompt + LEK-1 signature (sandwich format)
        signed_prompt = f"{axioms}\n\n---\n\n{prompt}\n\n---\n\n{lek1}"

        # MLX chat format for Gemma
        training_text = f"<start_of_turn>user\n{signed_prompt}<end_of_turn>\n<start_of_turn>model\n{response}<end_of_turn>"

        record = json.dumps({"text": training_text})

        if i < train_count:
            train_out.write(record + "\n")
        else:
            valid_out.write(record + "\n")

train_out.close()
valid_out.close()
print(f"Written: {train_count} train, {$TOTAL - train_count} valid")
PYEOF

rm "$SHUFFLED"

echo ""
echo "Output:"
echo "  Train: $TRAIN_FILE"
echo "  Valid: $VALID_FILE"
echo ""
echo "Next: scp to M3 and run MLX LoRA"
echo "  scp $TRAIN_FILE $VALID_FILE claude@10.69.69.108:~/ai-work/training/"
