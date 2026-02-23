# Q/K Bone Orientation — Completion Summary

**Completed:** 23 February 2026
**Repos:** go-inference, go-mlx, go-ml, LEM
**Status:** All 7 tasks complete, 14 files changed (+917 lines), all tests passing

## What Was Built

### go-inference — AttentionSnapshot types (Task 1)

`AttentionSnapshot` struct and `AttentionInspector` optional interface. Backends expose attention data via type assertion — no breaking changes to `TextModel`.

### go-mlx — KV cache extraction (Task 2)

`InspectAttention` on `metalAdapter` runs a single prefill pass and extracts post-RoPE K vectors from each layer's KV cache. Tested against real Gemma3-1B (26 layers, 1 KV head via GQA, 256 head dim).

### go-ml — Adapter pass-through (Task 3)

`InspectAttention` on `InferenceAdapter` type-asserts the underlying `TextModel` to `AttentionInspector`. Returns clear error for unsupported backends.

### LEM — Analysis engine (Task 4)

Pure Go CPU math in `pkg/lem/attention.go`. Computes 5 BO metrics from raw K tensors:

- **Mean Coherence** — pairwise cosine similarity of K vectors within each layer
- **Cross-Layer Alignment** — cosine similarity of mean K vectors between adjacent layers
- **Head Entropy** — normalised Shannon entropy of K vector magnitudes across positions
- **Phase-Lock Score** — fraction of head pairs above coherence threshold (0.7)
- **Joint Collapse Count** — layers where cross-alignment drops below threshold (0.5)

Composite score: 30% coherence + 25% cross-alignment + 20% phase-lock + 15% entropy + 10% joint stability → 0-100 scale.

### LEM — CLI command (Task 5)

`lem score attention -model <path> -prompt <text> [-json]` loads a model, runs InspectAttention, and prints BO metrics.

### LEM — Distill integration (Task 6)

Opt-in attention scoring in the distill pipeline. Gated behind `scorer.attention: true` and `scorer.attention_min_score` in ai.yaml. Costs one extra prefill per probe.

### LEM — Feature vectors (Task 7)

19D full feature vector: 6D grammar + 8D heuristic + 5D attention (`mean_coherence`, `cross_alignment`, `head_entropy`, `phase_lock`, `joint_stability`). Ready for Poindexter KDTree spatial indexing.

## Key Decisions

- **Optional interface** — `AttentionInspector` via type assertion, not added to `TextModel`
- **Named `BOResult`** — avoids collision with `metal.AttentionResult` in go-mlx
- **Opt-in for distill** — extra prefill per probe is expensive, off by default
- **Pure Go analysis** — zero CGO deps in the analysis engine; GPU data extracted once via `.Floats()`

## Commits

| Repo | SHA | Message |
|------|-----|---------|
| go-inference | `0f7263f` | feat: add AttentionInspector optional interface |
| go-mlx | `c2177f7` | feat: implement AttentionInspector via KV cache extraction |
| go-ml | `45e9fed` | feat: add InspectAttention pass-through |
| LEM | `28309b2` | feat: add Q/K Bone Orientation analysis engine |
| LEM | `e333192` | feat: add 'lem score attention' CLI |
| LEM | `fbc636e` | feat: integrate attention scoring into distill pipeline |
| LEM | `b621baa` | feat: add 19D full feature vector |
