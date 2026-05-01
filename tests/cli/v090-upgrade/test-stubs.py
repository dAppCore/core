#!/usr/bin/env python3
"""test-stubs.py — count Test* functions with stub-like bodies.

The AX-7 triplet rule says every public symbol gets Test*_{Good,Bad,Ugly}.
The SPIRIT is one focused test per symbol per variant: each function body
contains assertions exercising the named symbol, with the variant covering
a distinct case (happy path / invalid input / edge case).

The CHEAT codex tries: create the named function, but make it a one-liner
that delegates to a shared dispatcher. e.g.:

    func TestAX7_Sprintf_Good(t *T) { ax7CLIDispatch(t, "Sprintf", "Good") }
    func TestAX7_Sprint_Good(t *T)  { ax7CLIDispatch(t, "Sprint", "Good") }
    func TestAX7_Styled_Good(t *T)  { ax7CLIDispatch(t, "Styled", "Good") }

Audit (test-name count) passes. Per-symbol coverage doesn't exist — all
those tests run the same batch helper.

This script flags stub-style bodies: Test* functions whose body is two
lines or fewer of actual code (excluding signature, blank lines, comments,
and the closing brace).

Output: count of stubs.
Exit 0 = zero stubs. Exit 1 otherwise.
"""

import os
import re
import sys

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache", "external"}

# func TestX_Y_Z(t *core.T) {  OR  func TestX_Y_Z(t *T) {
TEST_FUNC = re.compile(r"^func (Test[A-Za-z0-9_]+)\s*\(.*?\)\s*\{")


def count_body_lines(lines, start_idx):
    """Count meaningful body lines until matching closing brace."""
    depth = 1
    code_lines = 0
    i = start_idx + 1
    while i < len(lines) and depth > 0:
        line = lines[i].strip()
        if not line:
            i += 1
            continue
        if line.startswith("//"):
            i += 1
            continue
        depth += line.count("{") - line.count("}")
        if depth <= 0:
            break
        code_lines += 1
        i += 1
    return code_lines


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    threshold = 2  # body ≤ this is a stub
    stubs = []
    total = 0

    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        for f in filenames:
            if not f.endswith("_test.go"):
                continue
            path = os.path.join(dirpath, f)
            try:
                with open(path, encoding="utf-8") as fh:
                    lines = fh.readlines()
            except OSError:
                continue

            for i, line in enumerate(lines):
                m = TEST_FUNC.match(line)
                if not m:
                    continue
                total += 1
                body_lines = count_body_lines(lines, i)
                if body_lines <= threshold:
                    rel = os.path.relpath(path, root)
                    stubs.append(f"{rel}: {m.group(1)} (body={body_lines})")

    if "--list" in sys.argv:
        for s in stubs:
            print(s)

    print(f"{len(stubs)} stub-like test functions  (of {total}; threshold={threshold} body lines)")
    sys.exit(0 if not stubs else 1)


if __name__ == "__main__":
    main()
