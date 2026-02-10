# LEK-1 LoRA Training Data

## Format

Training data for MLX LoRA fine-tuning of Gemma 3 12B.

Files:
- `train.jsonl` — Training pairs (Axioms-signed prompt → response)
- `valid.jsonl` — Validation set (10% holdout)
- `lora-config.yaml` — MLX LoRA hyperparameters

## Data Generation Pipeline

1. Hypnos (Gemini 3 Pro) generates 200 prompt-response pairs using Axioms kernel
2. Format as JSONL: `{"text": "<bos>user\n{prompt}<eos>\n<bos>model\n{response}<eos>"}`
3. Split 180/20 train/valid
4. Run MLX LoRA on M3 Ultra

## Training Command (M3 Ultra)

```bash
pip install mlx-lm
python -m mlx_lm.lora \
    --model google/gemma-3-12b \
    --train-data train.jsonl \
    --valid-data valid.jsonl \
    --num-layers 8 \
    --batch-size 1 \
    --num-iters 500 \
    --learning-rate 1e-5 \
    --adapter-path ./adapters
```

## Merge & Test

```bash
python -m mlx_lm.fuse \
    --model google/gemma-3-12b \
    --adapter-path ./adapters \
    --save-path ./gemma-3-12b-lek1

# Convert to GGUF for Ollama
python -m mlx_lm.convert --model ./gemma-3-12b-lek1 --to-gguf
```

## License

EUPL-1.2 — All training data and derivative weights.
