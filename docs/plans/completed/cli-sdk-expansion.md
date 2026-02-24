# CLI SDK Expansion — Completion Summary

**Completed:** 21 February 2026
**Module:** `forge.lthn.ai/core/go/pkg/cli` (later migrated to `forge.lthn.ai/core/cli`)
**Status:** Complete — all TUI primitives shipped, then extracted to core/cli

## What Was Built

Extended `pkg/cli` with charmbracelet TUI primitives so domain repos only
import `core/cli` for all CLI concerns. Charmbracelet dependencies (bubbletea,
bubbles, lipgloss) are encapsulated behind our own types.

### Components added

| Component | File | Purpose |
|-----------|------|---------|
| RunTUI | `runtui.go` | Escape hatch with `Model`/`Msg`/`Cmd`/`KeyMsg` types |
| Spinner | `spinner.go` | Async handle with `Update()`, `Done()`, `Fail()` |
| ProgressBar | `progressbar.go` | `Increment()`, `Set()`, `SetMessage()`, `Done()` |
| InteractiveList | `list.go` | Keyboard navigation with terminal fallback |
| TextInput | `textinput.go` | Placeholder, masking, validation |
| Viewport | `viewport.go` | Scrollable content for logs, diffs, docs |
| Form (stub) | `form.go` | Interface defined, bufio fallback |
| FilePicker (stub) | `filepicker.go` | Interface defined, bufio fallback |
| Tabs (stub) | `tabs.go` | Interface defined, simple fallback |

### Subsequent migration

On 22 February 2026, `pkg/cli` was extracted from `core/go` into its own
module at `forge.lthn.ai/core/cli` and all imports were updated. The TUI
primitives now live in the standalone CLI module.

### Frame upgrade (follow-on)

The Frame layout system was upgraded to implement `tea.Model` directly on
22 February 2026 (in `core/cli`), adding bubbletea lifecycle, `KeyMap` for
configurable bindings, `Navigate()`/`Back()` for panel switching, and
lipgloss-based HLCRF rendering. This was a separate plan
(`frame-bubbletea`) that built on the SDK expansion.
