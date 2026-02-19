# TODO.md — Core Go Dispatch Queue

Tasks dispatched from core/go orchestration to satellite repos.
Format: `- [ ] REPO: task description` / `- [x]` when done.

---

## go-i18n (forge.lthn.ai/core/go-i18n)

### Phase 1: Harden the Engine

- [ ] **go-i18n: Add CLAUDE.md** — Document the grammar engine contract: what it is (grammar primitives + reversal), what it isn't (translation file manager). Include build/test commands, the gram.* sacred rule, and the agent-flattening prohibition.
- [ ] **go-i18n: Ambiguity resolution for dual-class words** — Words like "run", "file", "test", "check", "build" are both verb and noun. Tokeniser currently picks first match. Need context-aware disambiguation (look at surrounding tokens: article before → noun, after subject → verb).
- [ ] **go-i18n: Extend irregular verb coverage** — Audit against common dev/ops vocabulary. Missing forms cause silent fallback to regular rules which may produce wrong output (e.g. "builded" instead of "built").
- [ ] **go-i18n: Add benchmarks** — `grammar_test.go` and `reversal/tokeniser_test.go` need `Benchmark*` functions. The engine will run in hot paths (TIM, Poindexter) — need baseline numbers.

### Phase 2: Reference Distribution + 1B Classification Pipeline

#### 2a: 1B Pre-Classification (based on LEK-1B benchmarks)

- [ ] **go-i18n: Classification benchmark suite** — `classify_bench_test.go` with 200+ domain-tagged sentences. Categories: {technical, creative, ethical, casual}. Ground truth for calibrating 1B pre-tags.
- [ ] **go-i18n: 1B pre-sort pipeline tool** — CLI/func that reads JSONL corpus, classifies via LEK-Gemma3-1B, writes back with `domain_1b` field. Target: ~5K sentences/sec on M3.
- [ ] **go-i18n: 1B vs 27B calibration check** — Sample 500 sentences, classify with both, measure agreement. 75% baseline from benchmarks, technical↔creative is known weak spot.
- [ ] **go-i18n: Article/irregular validator** — Lightweight funcs using 1B's strong article (100%) and irregular base form (100%) accuracy as fast validators.

#### 2b: Reference Distributions

- [ ] **go-i18n: Reference distribution builder** — Process 88K scored seeds through tokeniser + imprint. Pre-sort by `domain_1b` tag. Output: per-category reference distributions as JSON.
- [ ] **go-i18n: Imprint comparator** — Distance metrics (cosine, KL divergence, Mahalanobis) against reference distributions. Classification signal with confidence. Poindexter integration point.
- [ ] **go-i18n: Cross-domain anomaly detection** — Flag texts where 1B domain tag disagrees with imprint classification. Training signal or genuine cross-domain text — both valuable.

### Phase 3: Multi-Language

- [ ] **go-i18n: Grammar table format spec** — Document the exact JSON schema for `gram.*` keys so new languages can be added. Currently only inferred from `en.json`.
- [ ] **go-i18n: French grammar tables** — First non-English language. French has gendered nouns, complex verb conjugation, elision rules. Good stress test for the grammar engine's language-agnostic design.

---

## core/go (this repo)

- [ ] **core/go: Remove pkg/i18n dependency** — Once core/cli imports go-i18n directly, remove `pkg/i18n/` from this repo. The locale files are bad data and shouldn't be migrated.
- [ ] **core/go: Update go.work** — Add go-i18n to the workspace for local development (`go work use ../go-i18n`).

---

## Workflow

1. Virgil (this session) writes tasks above after research
2. Second GoLand session opens the target repo and works from this TODO
3. When a task is done, mark `[x]` and note the commit/PR
4. If a task reveals new work, add it here
5. Scale to other repos once pattern is proven
