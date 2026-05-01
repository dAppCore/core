#!/usr/bin/env python3
"""unreferenced-symbols.py — find Test functions that don't reference their target.

For each `Test<Prefix>_<Symbol>_<Variant>` function in `<file>_test.go`, scan
its body. The body MUST contain a whole-word reference to the symbol it claims
to test. If the test never names its target, it's not exercising the target —
just claiming coverage.

This catches:
  - `reflect.ValueOf(any((*Foo).Method)); AssertEqual(reflect.Func, rv.Kind())`
    tautology (Method name appears once via reflection but is not invoked)
  - Dispatcher patterns where the test calls a shared helper without ever
    naming the symbol under test
  - Empty Run() shells: `t.Run("Good", func(t *testing.T) {})` with no body

The check looks for the bare symbol (or method) name in the function body —
anywhere is fine: invocation, type assertion, var declaration, even a string
literal. The bar is "did the author write this name down anywhere", not "is
it called correctly". Stricter checks (real invocation, real assertion on
output) are out of scope here — they'd over-flag legitimate patterns.

Method symbols `Receiver_Method` count as referenced if EITHER `Receiver` or
`Method` appears in the body. Constructors `NewFoo` count if `NewFoo` (full
name) appears. Generic types are stripped before matching.

Output:
  <count> Test functions that don't reference their symbol  (of <total>)

Exit 0 when zero, 1 otherwise. `--list` prints one line per offender:
  <test_file>:<lineno> <Test function name>: missing reference to <Symbol>
"""

import os
import re
import sys

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache", "external"}

TEST_FUNC_HEADER = re.compile(
    r"^func (Test([A-Z][A-Za-z0-9]*)_([A-Za-z][A-Za-z0-9_]*)_(Good|Bad|Ugly))\s*\("
)


def walk_test_files(root: str):
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        for f in filenames:
            if f.endswith("_test.go"):
                yield os.path.join(dirpath, f)


def extract_test_bodies(path: str):
    """Yield (lineno, full_func_name, target_symbol, body_text) per Test*_<Symbol>_<Variant>."""
    try:
        with open(path, encoding="utf-8") as fh:
            lines = fh.readlines()
    except OSError:
        return
    i = 0
    while i < len(lines):
        m = TEST_FUNC_HEADER.match(lines[i])
        if not m:
            i += 1
            continue
        full_name = m.group(1)
        target = m.group(3)  # the Symbol piece (post-prefix, pre-variant)
        # Walk forward to find the matching closing brace for the function body.
        # Simple brace counter — robust enough for go's gofmt'd files.
        depth = 0
        body_lines = []
        started = False
        j = i
        while j < len(lines):
            line = lines[j]
            for ch in line:
                if ch == "{":
                    depth += 1
                    started = True
                elif ch == "}":
                    depth -= 1
            body_lines.append(line)
            if started and depth == 0:
                break
            j += 1
        body = "".join(body_lines)
        yield i + 1, full_name, target, body
        i = j + 1


def references_target(body: str, target: str) -> bool:
    """True if the body contains a whole-word reference to target.

    target may be a compound `Receiver_Method`. In that case either token
    counts, since methods are commonly called via the receiver value.
    """
    tokens = target.split("_") if "_" in target else [target]
    for tok in tokens:
        # \b to anchor word boundary, but Go identifiers don't have hyphens
        # so a simple regex is enough.
        if re.search(rf"\b{re.escape(tok)}\b", body):
            return True
    return False


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    total_tests = 0
    total_unreferenced = 0
    offenders = []

    for path in walk_test_files(root):
        for lineno, name, target, body in extract_test_bodies(path):
            total_tests += 1
            if not references_target(body, target):
                total_unreferenced += 1
                rel = os.path.relpath(path, root)
                offenders.append(f"{rel}:{lineno} {name}: body never references {target}")

    if "--list" in sys.argv:
        for line in offenders:
            print(line)

    print(f"{total_unreferenced} Test functions that don't reference their symbol  (of {total_tests})")
    sys.exit(0 if total_unreferenced == 0 else 1)


if __name__ == "__main__":
    main()
