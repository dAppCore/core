#!/usr/bin/env python3
"""identical-triplets.py — flag Test_<Symbol>_{Good,Bad,Ugly} triplets whose
three function bodies are byte-identical (modulo the function name and
any leading/trailing whitespace).

A real Good/Bad/Ugly triplet exercises three different cases: happy path,
explicit error path, edge case. If all three function bodies hash the
same, it's not a triplet — it's three copies of one test, satisfying the
audit's coverage check while testing nothing different.

Detection:
  - Walk every `*_test.go` file
  - Pair up Test*_<Symbol>_Good / _Bad / _Ugly within the same file
  - Hash each function body (after the opening brace, before the matching
    close brace) with the function name stripped
  - If two or more bodies in a triplet hash the same, count once per
    duplicated body

Output (single line):
  <duplicate-count> Test triplets with identical Good/Bad/Ugly bodies (of <total>)

Exit 0 when zero duplicates, 1 otherwise. `--list` prints each offender.
"""

import hashlib
import os
import re
import sys

EXCLUDED_DIRS = {".git", "node_modules", "vendor", "third_party", ".scannerwork", ".tmp", "gomodcache"}

TRIPLET_HEADER = re.compile(
    r"^func (Test[A-Za-z0-9]+_([A-Za-z][A-Za-z0-9_]*)_(Good|Bad|Ugly))\s*\("
)


def extract_function_bodies(path: str):
    """Yield (full_func_name, symbol_key, variant, body_bytes) for each
    Test*_<Symbol>_{Good,Bad,Ugly} function in path."""
    try:
        with open(path, encoding="utf-8") as fh:
            lines = fh.readlines()
    except OSError:
        return
    i = 0
    while i < len(lines):
        m = TRIPLET_HEADER.match(lines[i])
        if not m:
            i += 1
            continue
        full_name, symbol_key, variant = m.group(1), m.group(2), m.group(3)
        depth = 0
        started = False
        body_lines = []
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
        # Strip the header line so the func name doesn't poison the hash.
        body_text = "".join(body_lines[1:])
        # Normalise trailing/leading whitespace.
        normalised = "\n".join(line.rstrip() for line in body_text.splitlines() if line.strip())
        digest = hashlib.sha1(normalised.encode("utf-8")).hexdigest()
        yield path, full_name, symbol_key, variant, digest
        i = j + 1


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else "."

    # Group by (file, symbol_key) so we only compare the SAME triplet's
    # three variants. Different symbols may legitimately have similar
    # bodies — only intra-triplet identity is gaming.
    groups: dict = {}
    total_triplets = 0
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDED_DIRS and not d.startswith(".")]
        for f in filenames:
            if not f.endswith("_test.go"):
                continue
            for path, name, symbol_key, variant, digest in extract_function_bodies(
                os.path.join(dirpath, f)
            ):
                key = (path, symbol_key)
                groups.setdefault(key, {})[variant] = (name, digest)

    duplicate_bodies = 0
    offenders = []
    for (path, symbol_key), variants in groups.items():
        if len(variants) < 2:
            continue
        digests = [v[1] for v in variants.values()]
        unique = set(digests)
        if len(unique) < len(digests):
            # at least two bodies match
            extra = len(digests) - len(unique)
            duplicate_bodies += extra
            total_triplets += 1
            rel = os.path.relpath(path, root)
            offenders.append(f"{rel}: {symbol_key} — {len(unique)} unique body across {len(digests)} variants")

    if "--list" in sys.argv:
        for line in offenders:
            print(line)

    print(f"{duplicate_bodies} Test triplets with identical Good/Bad/Ugly bodies (of {len(groups)})")
    sys.exit(0 if duplicate_bodies == 0 else 1)


if __name__ == "__main__":
    main()
