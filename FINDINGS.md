# FINDINGS.md — Core Go Research

## go-i18n (forge.lthn.ai/core/go-i18n)

**Explored**: 2026-02-19
**Location**: `/Users/snider/Code/host-uk/go-i18n`
**Module**: `forge.lthn.ai/core/go-i18n`
**State**: 20 commits on main, clean, all tests pass
**Lines**: ~5,800 across 32 files (14 test files)
**Deps**: only `golang.org/x/text`

### What It Is

A **grammar engine** — not a translation file manager. Provides:

1. **Forward composition**: `PastTense()`, `Gerund()`, `Pluralize()`, `Article()`, handlers
2. **Reverse grammar**: Tokeniser reads grammar tables backwards to extract structure
3. **GrammarImprint**: Feature vector projection (content → grammar fingerprint, lossy)
4. **Multiplier**: Deterministic training data augmentation (no LLM)

Consumers (core/cli, apps) bring their own translation files. go-i18n provides the grammar primitives.

### Current Capabilities

| Feature | Status | Notes |
|---------|--------|-------|
| Grammar primitives (past/gerund/plural/article) | Working | 100 irregular verbs, 40 irregular nouns |
| Magic namespace handlers (i18n.label/progress/count/done/fail/numeric) | Working | 6 handler types |
| Service + message lookup | Working | Thread-safe, fallback chain |
| Subject builder (S()) | Working | Fluent API with count/gender/location/formality |
| Plural categories (CLDR) | Working | 7+ languages |
| RTL/LTR detection | Working | 12+ RTL languages |
| Number formatting | Working | Locale-specific separators |
| Reversal tokeniser | Working | 3-tier: JSON → irregular → regular morphology |
| GrammarImprint similarity | Working | Weighted cosine (verbs 30%, tense 20%, nouns 25%) |
| Multiplier expand | Working | Tense + number flipping, dedup, round-trip verify |

### What's Missing / Incomplete

| Gap | Priority | Notes |
|-----|----------|-------|
| Reference distribution builder | High | Process scored seeds → calibrate imprints |
| Non-English grammar tables | Medium | Only en.json exists, reversal needs gram.* per language |
| Ambiguity resolution | Medium | "run", "file", "test" are both verb and noun |
| Domain vocabulary expansion | Low | 150+ words, needs legal/medical/financial |
| Poindexter integration | Deferred | Awaiting Poindexter library |
| TIM container image | Deferred | Distroless Go binary for confidential compute |

### Key Architecture Decisions

- **Bijective grammar tables**: Forward and reverse use same JSON → reversal is deterministic
- **Lossy projection**: GrammarImprint intentionally loses content, preserves only structure
- **No LLM dependency**: Multiplier generates variants purely from morphological rules
- **Consumer translations are external**: go-i18n doesn't ship or manage app-specific locale files
- **gram.* keys are sacred**: Agents MUST NOT flatten — grammar engine depends on nested structure

### pkg/i18n in core/go

- Full i18n framework with 34 locale files — but locale data is bad/stale
- Only imported by `pkg/cli/` which has been extracted to `core/cli`
- Effectively orphaned in core/go
- Can be removed once core/cli imports go-i18n directly
- The locale files need full rework, not migration

---

## CoreDeno (PR #9 — merged)

**Explored**: 2026-02-19

Deno sidecar for core-gui JS runtime. Go↔Deno bidirectional bridge:
- Go→Deno: JSON-RPC over Unix socket (module lifecycle)
- Deno→Go: gRPC over Unix socket (file I/O, store, manifest)
- Each module in isolated Deno Worker with declared permissions
- Marketplace: git clone + ed25519 manifest verification + SQLite registry

10 security/correctness issues found and fixed in review.
