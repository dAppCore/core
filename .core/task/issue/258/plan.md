# Implementation Plan: Issue 258

## Phase 1: Command Structure
1. Create `cmd/core/cmd/test.go`
2. Register `test` command with Cobra
3. Add flags: `--all`, `--filter`, `--coverage`

## Phase 2: Change Detection  
1. Run `git diff --name-only HEAD` to get changed files
2. Filter for `.go` files (exclude `_test.go`)
3. Map each file to its test file:
   - `internal/foo/bar.go` → `internal/foo/bar_test.go`
   - Skip if test file does not exist

## Phase 3: Test Execution
1. Build `go test` command with detected test files
2. Pass through `--coverage` flag as `-cover`
3. Pass through `--filter` as `-run` pattern
4. Stream output to terminal

## Phase 4: Edge Cases
- No changed files → inform user, optionally run all
- No matching test files → inform user
- `--all` flag → skip detection, run `go test ./...`

## Files to Create/Modify
- `cmd/core/cmd/test.go` (new)
- `cmd/core/cmd/root.go` (register command)

## Testing
- Add `cmd/core/cmd/test_test.go` with unit tests
- Manual testing with actual git changes

