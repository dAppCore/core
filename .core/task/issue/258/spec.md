# Issue 258: Smart Test Detection

## Original Issue
https://github.com/host-uk/core/issues/258

## Summary
Make `core test` smart - detect changed files and run only relevant tests.

## Commands
```bash
core test                    # Run tests for changed files only
core test --all              # Run all tests  
core test --filter UserTest  # Run specific test
core test --coverage         # With coverage report
```

## Acceptance Criteria
- [ ] Detect changed `.go` files via `git diff --name-only`
- [ ] Map source files to test files (`foo.go` → `foo_test.go`)
- [ ] Run only relevant tests via `go test`
- [ ] Support `--all` flag to run all tests
- [ ] Support `--filter` flag for pattern matching
- [ ] Support `--coverage` flag for coverage reports

## Technical Context
- Go CLI using Cobra
- Commands in `cmd/core/cmd/`
- Follow existing patterns in `dev_*.go` files

