#!/usr/bin/env python3
"""example-gaps.py — file-aware Example* coverage check.

For each public symbol in `<file>.go`, demand at least one `Example<Symbol>`
(top-level function) or `Example<Receiver>_<Method>` (method) in the matching
`<file>_example_test.go`. Examples document usage; the triplet tests assert
correctness — both are required for v0.9.0 reference shape.

Naming follows Go's canonical example spec:
  - `ExampleFoo()`            for top-level function `Foo`
  - `ExampleType()`           for type `Type`
  - `ExampleType_Method()`    for method `Method` on `Type`
  - `ExampleFoo_variant()`    additional example with lowercase suffix variant
                              (variant must start with a lowercase letter)

The example MUST live in `<file>_example_test.go` next to its source file.
core/go's reference repo has one such file per source — examples in monolith
`example_test.go` files don't count, same file-aware logic as the triplet
audit.

Output (single line):
  <gap-count> public-symbols missing Example*  (of <total>)

Exit 0 when zero gaps, 1 otherwise. `--list` prints one line per gap with
the expected example function name and the file it should live in.
"""

import os
import re
import sys

TOP_LEVEL = re.compile(r"^func ([A-Z][A-Za-z0-9_]*)\s*[\[(]")
METHOD = re.compile(
    r"^func \([^)]*?\*?([A-Z][A-Za-z0-9_]*)(?:\[[^\]]+\])?\) ([A-Z][A-Za-z0-9_]*)\s*[\[(]"
)
EXAMPLE_NAME = re.compile(r"^func (Example[A-Za-z0-9_]+)\s*\(")

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache", "external"}


def walk_packages(root: str):
    """Yield (pkg_dir, [(prod_path, basename)], {basename: [example_paths]})."""
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        prod = []
        examples_by_stem = {}
        for f in filenames:
            if not f.endswith(".go"):
                continue
            full = os.path.join(dirpath, f)
            if f.endswith("_example_test.go"):
                stem = f[:-len("_example_test.go")]
                examples_by_stem.setdefault(stem, []).append(full)
            elif f.endswith("_test.go"):
                continue  # not an example file
            else:
                stem = f[:-len(".go")]
                prod.append((full, stem))
        if prod:
            yield dirpath, prod, examples_by_stem


def collect_symbols(prod_files):
    """Return list of (symbol, source_basename, source_path)."""
    symbols = []
    for path, stem in prod_files:
        try:
            with open(path, encoding="utf-8") as fh:
                for line in fh:
                    m = TOP_LEVEL.match(line)
                    if m:
                        symbols.append((m.group(1), stem, path))
                        continue
                    m = METHOD.match(line)
                    if m:
                        symbols.append((f"{m.group(1)}_{m.group(2)}", stem, path))
        except OSError:
            pass
    return symbols


def collect_example_names(example_paths):
    """Return set of Example* function names defined in the given files."""
    names = set()
    for path in example_paths:
        try:
            with open(path, encoding="utf-8") as fh:
                for line in fh:
                    m = EXAMPLE_NAME.match(line)
                    if m:
                        names.add(m.group(1))
        except OSError:
            pass
    return names


def has_example(symbol: str, example_names: set) -> bool:
    """True iff at least one Example<symbol>(...) or Example<symbol>_variant(...) exists."""
    target = f"Example{symbol}"
    for name in example_names:
        if name == target:
            return True
        if name.startswith(target + "_"):
            # The trailing variant must start with a lowercase letter per Go's
            # spec; uppercase would mean it's a different symbol.
            tail = name[len(target) + 1:]
            if tail and tail[0].islower():
                return True
    return False


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    total_gaps = 0
    total_symbols = 0
    gap_lines = []

    for pkg_dir, prod, examples_by_stem in walk_packages(root):
        symbols = collect_symbols(prod)
        for symbol, stem, src_path in symbols:
            total_symbols += 1
            example_paths = examples_by_stem.get(stem, [])
            example_names = collect_example_names(example_paths)
            if not has_example(symbol, example_names):
                total_gaps += 1
                rel = os.path.relpath(src_path, root)
                gap_lines.append(f"{rel}: Example{symbol} missing from {stem}_example_test.go")

    if "--list" in sys.argv:
        for line in gap_lines:
            print(line)

    print(f"{total_gaps} public-symbols missing Example*  (of {total_symbols})")
    sys.exit(0 if total_gaps == 0 else 1)


if __name__ == "__main__":
    main()
