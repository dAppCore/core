#!/usr/bin/env python3
"""Test-import discipline — *_test.go and *_example_test.go files must
import only `dappco.re/go` (dot-import).

The core module re-exports every stdlib helper its consumers need.
Reaching for "context", "sync", "time", "os", "io", etc. in a test
violates the SPOR pattern at the test layer: the agent that generates
similar tests downstream copies the stdlib import shape, which
defeats the whole point of core/go's wrapper surface.

*_internal_test.go files are exempt — they're `package core` and may
import any stdlib that the production owner imports.

Exit 0 with "test-imports clean" when every test file is single-import.
Exit 1 with the violation list otherwise.
"""

import os
import re
import sys
from collections import defaultdict

ALLOWED = {"dappco.re/go"}

_SINGLE = re.compile(r'^import\s+(?:[\w.]+\s+)?"([^"]+)"', re.MULTILINE)
_BLOCK_HEAD = re.compile(r"^import\s*\(", re.MULTILINE)


def imports(text: str) -> list[str]:
    """Return list of import paths from a Go source file."""
    paths = []
    for m in _SINGLE.finditer(text):
        paths.append(m.group(1))
    for head in _BLOCK_HEAD.finditer(text):
        # Walk to matching `)`
        i = head.end()
        depth = 1
        while i < len(text) and depth > 0:
            if text[i] == "(":
                depth += 1
            elif text[i] == ")":
                depth -= 1
                if depth == 0:
                    break
            i += 1
        block = text[head.end() : i]
        for line in block.splitlines():
            line = line.strip()
            if not line or line.startswith("//") or line.startswith("/*"):
                continue
            m = re.match(r'(?:[\w.]+\s+)?"([^"]+)"', line)
            if m:
                paths.append(m.group(1))
    return paths


def main():
    cwd = os.getcwd()
    violations = defaultdict(list)
    total = 0

    for fname in sorted(os.listdir(cwd)):
        if not fname.endswith("_test.go"):
            continue
        # Internal tests are exempt
        if fname.endswith("_internal_test.go"):
            continue
        total += 1
        path = os.path.join(cwd, fname)
        try:
            with open(path, encoding="utf-8") as fh:
                text = fh.read()
        except OSError:
            continue
        for imp in imports(text):
            if imp not in ALLOWED:
                violations[fname].append(imp)

    if violations:
        for fname, paths in sorted(violations.items()):
            for p in sorted(set(paths)):
                print(f"{fname}: imports {p!r} — use core helpers")
        print()
        print(
            f"{len(violations)} of {total} test files import packages other than 'dappco.re/go'."
        )
        print(
            "AX rule: *_test.go and *_example_test.go files may only import dappco.re/go."
        )
        print(
            "Replace stdlib usage with core helpers (e.g. context.Background → Background,"
        )
        print(
            "sync.Mutex → Mutex, time.Now → Now). *_internal_test.go files are exempt."
        )
        sys.exit(1)

    print(
        f"test-imports clean: all {total} non-internal test files import only dappco.re/go."
    )


if __name__ == "__main__":
    main()
