#!/usr/bin/env python3
"""AX-7 naming guard — every public symbol in core/go production code
must have all three Test*_Good, Test*_Bad, Test*_Ugly variants.

Exit 0 with "AX-7 clean" when every public symbol has the triplet.
Exit 1 with the gap list otherwise.
"""

import os
import re
import sys
from collections import defaultdict

# ---------- regex parsers ----------

# Top-level func: "func Name" (public OR private) optionally followed by [TypeParams](
# Private (lowercase-starting) symbols ARE in scope — the triplet exposes branch
# gaps in helpers, not just the public surface.
_TOP_LEVEL = re.compile(r"^func ([A-Za-z][A-Za-z0-9_]*)\s*[\[(]")

# Method: "func (recv [*]Type[generic]?) Method(" — captures Type and Method.
# Both Type and Method may be public OR private.
_METHOD = re.compile(
    r"^func \([^)]*?\*?([A-Za-z][A-Za-z0-9_]*)(?:\[[^\]]+\])?\) ([A-Za-z][A-Za-z0-9_]*)\s*[\[(]"
)


def collect_symbols():
    """Return list of (symbol, source_label) tuples for every public symbol."""
    symbols = set()
    for fname in sorted(os.listdir(".")):
        if not fname.endswith(".go"):
            continue
        if fname.endswith("_test.go") or fname.endswith("_example_test.go") or fname.endswith("_fuzz_test.go"):
            continue
        label = fname[:-3]  # strip .go
        try:
            with open(fname, encoding="utf-8") as fh:
                for line in fh:
                    m = _TOP_LEVEL.match(line)
                    if m:
                        symbols.add((m.group(1), label))
                        continue
                    m = _METHOD.match(line)
                    if m:
                        symbols.add((f"{m.group(1)}_{m.group(2)}", label))
        except OSError:
            pass
    return sorted(symbols)


def collect_tests():
    """Return set of test function names (Test*_X_Y_Variant) defined in *_test.go."""
    test_re = re.compile(r"^func (Test[A-Za-z0-9_]+)\s*\(")
    tests = set()
    for fname in sorted(os.listdir(".")):
        if not fname.endswith("_test.go"):
            continue
        try:
            with open(fname, encoding="utf-8") as fh:
                for line in fh:
                    m = test_re.match(line)
                    if m:
                        tests.add(m.group(1))
        except OSError:
            pass
    return tests


def main():
    symbols = collect_symbols()
    tests = collect_tests()

    # Build a quick suffix index: map "_Symbol_Variant" → True for any test ending with it
    # We need to test whether ANY Test{Anything}_{Symbol}_{Variant} exists.
    test_suffixes = set()
    for name in tests:
        # Extract the suffix after the first "_": Test{Label}_{rest}
        # We want to flag membership of each "{Symbol}_{Variant}" suffix anywhere.
        idx = name.find("_")
        if idx >= 0:
            test_suffixes.add(name[idx + 1:])  # everything after first "_"

    missing_lines = []
    total = len(symbols)
    for sym, label in symbols:
        gaps = []
        for variant in ("Good", "Bad", "Ugly"):
            wanted = f"{sym}_{variant}"
            # Match if any test name's suffix equals or ends-with "_{wanted}"
            # Test{Label}_{Symbol}_{Variant} → suffix "{Label}_{Symbol}_{Variant}"
            # We accept if the suffix ENDS with "_{Symbol}_{Variant}" (any label)
            # OR if it equals "{Symbol}_{Variant}" exactly (when Label and Symbol overlap).
            target = "_" + wanted
            hit = any(s == wanted or s.endswith(target) for s in test_suffixes)
            if not hit:
                gaps.append(variant)
        if gaps:
            missing_lines.append(f"{label}.go: {sym} — missing: {' '.join(gaps)}")

    if missing_lines:
        for line in missing_lines:
            print(line)
        print()
        print(f"{len(missing_lines)} of {total} symbols missing one or more Test*_{{Good,Bad,Ugly}} variants.")
        print("AX-7: every symbol (public AND private) gets the triplet; gaps mean the surface")
        print("is too narrow or the edges aren't understood. Fill in the missing variants.")
        sys.exit(1)

    print(f"AX-7 clean: all {total} symbols have Test*_{{Good,Bad,Ugly}} triplets.")


if __name__ == "__main__":
    main()
