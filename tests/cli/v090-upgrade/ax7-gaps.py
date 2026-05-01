#!/usr/bin/env python3
"""ax7-gaps.py — file-aware AX-7 triplet coverage check.

For each public symbol in a production `<source>.go` file, the audit demands a
TestSource_Symbol_{Good,Bad,Ugly} triplet living in the matching test file
(`<source>_test.go` or `<source>_example_test.go`). Tests living anywhere else
do NOT count toward the symbol's coverage — the source-file prefix is
load-bearing because it ties tests to their subject and fails any test placed
in a monolith or versioned-suffix file.

Naming convention (per core/go reference):
  - `result.go` symbol `Or`   → `TestResult_Or_{Good,Bad,Ugly}` in `result_test.go`
  - `foo_bar.go` symbol `Run` → `TestFooBar_Run_{Good,Bad,Ugly}` in `foo_bar_test.go`
  - method `(r *Result) Or`  → counts as compound symbol `Result_Or` in result.go
                              → `TestResult_Result_Or_{...}` (rare) OR the
                                method's receiver-name shorthand
                                `TestResult_Or_{...}` (preferred — file prefix
                                already matches receiver, no need to repeat)

The check accepts both shapes for methods so existing core/go-style tests pass.

Output (single line, mirrors prior contract):
  <gap-count> public-symbols missing one or more of Good/Bad/Ugly  (of <total>)

Exit 0 when zero gaps, 1 otherwise. `--list` prints one line per gap with the
expected test function name and the file it should live in.
"""

import os
import re
import sys

TOP_LEVEL = re.compile(r"^func ([A-Z][A-Za-z0-9_]*)\s*[\[(]")
METHOD = re.compile(
    r"^func \([^)]*?\*?([A-Z][A-Za-z0-9_]*)(?:\[[^\]]+\])?\) ([A-Z][A-Za-z0-9_]*)\s*[\[(]"
)
TEST_NAME = re.compile(r"^func (Test[A-Za-z0-9_]+)\s*\(")

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache", "external"}


def file_prefix(filename: str) -> str:
    """Convert a source filename like 'foo_bar.go' to its test-prefix 'FooBar'."""
    stem = filename
    if stem.endswith(".go"):
        stem = stem[:-3]
    parts = [p for p in stem.split("_") if p]
    return "".join(p[:1].upper() + p[1:] for p in parts)


def walk_packages(root: str):
    """Yield (pkg_dir, [(prod_path, prefix)], {prefix: [test_paths]})."""
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        prod = []
        tests_by_prefix = {}
        for f in filenames:
            if not f.endswith(".go"):
                continue
            full = os.path.join(dirpath, f)
            if f.endswith("_test.go"):
                # Strip the trailing _test.go (or _example_test.go, _fuzz_test.go) to find
                # the source basename, then PascalCase that to the expected test prefix.
                stem = f[:-len("_test.go")]
                for tail in ("_example", "_fuzz"):
                    if stem.endswith(tail):
                        stem = stem[:-len(tail)]
                        break
                prefix = file_prefix(stem + ".go")
                tests_by_prefix.setdefault(prefix, []).append(full)
            else:
                prod.append((full, file_prefix(f)))
        if prod:
            yield dirpath, prod, tests_by_prefix


def collect_symbols(prod_files):
    """Return list of (symbol, file_prefix) for each public symbol per source file."""
    symbols = []
    for path, prefix in prod_files:
        try:
            with open(path, encoding="utf-8") as fh:
                for line in fh:
                    m = TOP_LEVEL.match(line)
                    if m:
                        symbols.append((m.group(1), prefix, path))
                        continue
                    m = METHOD.match(line)
                    if m:
                        symbols.append((f"{m.group(1)}_{m.group(2)}", prefix, path))
        except OSError:
            pass
    return symbols


def collect_test_names(test_paths):
    """Return set of Test* function names defined in the given files."""
    names = set()
    for path in test_paths:
        try:
            with open(path, encoding="utf-8") as fh:
                for line in fh:
                    m = TEST_NAME.match(line)
                    if m:
                        names.add(m.group(1))
        except OSError:
            pass
    return names


def has_triplet(symbol: str, prefix: str, test_names: set, variant: str) -> bool:
    """True iff a Test name in the symbol's file scope matches the expected shape."""
    # Preferred shape: Test<Prefix>_<Symbol>_<Variant>
    if f"Test{prefix}_{symbol}_{variant}" in test_names:
        return True
    # Method shorthand: when symbol is "Receiver_Method" and prefix matches the
    # receiver, accept Test<Prefix>_<Method>_<Variant> (the receiver doubles as
    # the file prefix in core/go convention — no need to repeat).
    if "_" in symbol:
        receiver, _, method = symbol.partition("_")
        if receiver == prefix and f"Test{prefix}_{method}_{variant}" in test_names:
            return True
    return False


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    total_gaps = 0
    total_symbols = 0
    gap_lines = []

    for pkg_dir, prod, tests_by_prefix in walk_packages(root):
        symbols = collect_symbols(prod)
        for symbol, prefix, src_path in symbols:
            total_symbols += 1
            test_paths = tests_by_prefix.get(prefix, [])
            test_names = collect_test_names(test_paths)
            missing = [v for v in ("Good", "Bad", "Ugly")
                       if not has_triplet(symbol, prefix, test_names, v)]
            if missing:
                total_gaps += 1
                rel = os.path.relpath(src_path, root)
                expect = f"Test{prefix}_{symbol}_{{{','.join(missing)}}}"
                gap_lines.append(f"{rel}: {expect}")

    if "--list" in sys.argv:
        for line in gap_lines:
            print(line)

    print(f"{total_gaps} public-symbols missing one or more of Good/Bad/Ugly  (of {total_symbols})")
    sys.exit(0 if total_gaps == 0 else 1)


if __name__ == "__main__":
    main()
