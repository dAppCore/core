#!/usr/bin/env python3
"""file-presence.py — every public-symbol-bearing source needs both a
matching `_test.go` and `_example_test.go` next to it.

Codex has demonstrated multiple gaming patterns where tests get dumped
into a generic file (`ax7_generated_test.go`, `service_test.go`, etc.)
that doesn't match the source it claims to cover. The file-aware
triplet check catches symbol gaps; this check catches the FILE-level
absence: `messages.go` has no `messages_test.go` and no
`messages_example_test.go`, period.

Two counters:
  missing-test-files     — `<file>.go` with public symbols but no `<file>_test.go`
  missing-example-files  — `<file>.go` with public symbols but no `<file>_example_test.go`

Excludes:
  - Pure-stub source files with no public symbols (constants, doc.go, type aliases)
  - Build-tagged platform files (when only one variant has symbols)

Output (two lines, last two are the counters):
  <missing-test> source files with no <file>_test.go
  <missing-example> source files with no <file>_example_test.go

Exit 0 when both counts are 0, 1 otherwise. `--list` prints offenders.
"""

import os
import re
import sys

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache"}

TOP_LEVEL = re.compile(r"^func ([A-Z][A-Za-z0-9_]*)\s*[\[(]")
METHOD = re.compile(r"^func \([^)]*?\*?([A-Z][A-Za-z0-9_]*)(?:\[[^\]]+\])?\) ([A-Z][A-Za-z0-9_]*)\s*[\[(]")
TYPE_DECL = re.compile(r"^type\s+([A-Z][A-Za-z0-9_]*)\s+(struct|interface|func)")
CONST_OR_VAR = re.compile(r"^(?:const|var)\s+([A-Z][A-Za-z0-9_]*)\s*=")


def has_public_symbols(path: str) -> bool:
    """True if the file declares at least one public function, method, or type."""
    try:
        with open(path, encoding="utf-8") as fh:
            for line in fh:
                if (TOP_LEVEL.match(line)
                        or METHOD.match(line)
                        or TYPE_DECL.match(line)
                        or CONST_OR_VAR.match(line)):
                    return True
    except OSError:
        return False
    return False


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."

    missing_tests = []
    missing_examples = []

    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        for f in filenames:
            if not f.endswith(".go"):
                continue
            if f.endswith("_test.go") or f.endswith("_example_test.go") or f.endswith("_fuzz_test.go"):
                continue
            full = os.path.join(dirpath, f)
            if not has_public_symbols(full):
                continue
            stem = f[:-len(".go")]
            test_file = os.path.join(dirpath, stem + "_test.go")
            example_file = os.path.join(dirpath, stem + "_example_test.go")
            if not os.path.exists(test_file):
                missing_tests.append(os.path.relpath(full, root))
            if not os.path.exists(example_file):
                missing_examples.append(os.path.relpath(full, root))

    if "--list" in sys.argv:
        if missing_tests:
            print("# missing _test.go siblings")
            for line in missing_tests:
                print(line)
        if missing_examples:
            print("# missing _example_test.go siblings")
            for line in missing_examples:
                print(line)

    print(f"{len(missing_tests)} source files with no <file>_test.go")
    print(f"{len(missing_examples)} source files with no <file>_example_test.go")
    sys.exit(0 if (len(missing_tests) == 0 and len(missing_examples) == 0) else 1)


if __name__ == "__main__":
    main()
