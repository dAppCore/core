#!/usr/bin/env bash
# AX-7 naming guard — every public symbol in core/go production code
# must have all three Test*_Good, Test*_Bad, Test*_Ugly variants.
#
# The triplet drives natural coverage: if you can't write a Bad case,
# the function surface is too narrow; if you can't write an Ugly case,
# the edges aren't understood. Easy to find gaps mechanically.
#
# Output is the gap list — one symbol per line with which variants
# are missing. Exit 0 if zero gaps; exit 1 otherwise.

set -uo pipefail

cd "$(dirname "$0")/../../.."

# ---------- collect public symbols from production .go ----------

PRODUCTION_GO=$(ls *.go 2>/dev/null \
  | grep -v -E '_test\.go$|_example_test\.go$|_fuzz_test\.go$' \
  || true)

if [ -z "$PRODUCTION_GO" ]; then
  echo "no production .go files found"
  exit 1
fi

SYMBOLS_FILE=$(mktemp)
trap 'rm -f "$SYMBOLS_FILE"' EXIT

# Top-level funcs:        func PublicName[T any](...)  → PublicName
# Method on type T:       func (recv *T) Method(...)   → T_Method
for f in $PRODUCTION_GO; do
  label="${f%.go}"
  # Top-level public functions (no receiver). Capture name before [ or (
  grep -E '^func [A-Z][A-Za-z0-9_]*\s*[\[(]' "$f" 2>/dev/null \
    | sed -E 's/^func ([A-Z][A-Za-z0-9_]*).*/\1/' \
    | awk -v lbl="$label" '{print $0 "\t" lbl}' >> "$SYMBOLS_FILE"

  # Methods: func (recv *Type) Method  or  func (recv Type) Method
  grep -E '^func \([^)]+\) [A-Z][A-Za-z0-9_]*\s*[\[(]' "$f" 2>/dev/null \
    | sed -E 's/^func \([^)]* \*?([A-Za-z0-9_]+)\) ([A-Z][A-Za-z0-9_]*).*/\1_\2/' \
    | awk -v lbl="$label" '{print $0 "\t" lbl}' >> "$SYMBOLS_FILE"
done

# Dedupe + strip generic-method noise (receiver like Array[T]) — those
# methods are tested under the non-parameterised test name, e.g.
# TestArray_Add_Good. The script extracts the type-parameter as the
# receiver letter; skip those rows so they don't show as false gaps.
sort -u "$SYMBOLS_FILE" -o "$SYMBOLS_FILE"
grep -vE '^[A-Z]_' "$SYMBOLS_FILE" > "$SYMBOLS_FILE.clean" && mv "$SYMBOLS_FILE.clean" "$SYMBOLS_FILE"

# ---------- collect existing Test* names from _test.go ----------

TESTS_FILE=$(mktemp)
trap 'rm -f "$SYMBOLS_FILE" "$TESTS_FILE"' EXIT

grep -hE '^func Test[A-Z][A-Za-z0-9_]*_(Good|Bad|Ugly)' *_test.go 2>/dev/null \
  | sed -E 's/^func (Test[A-Za-z0-9_]+)\(.*/\1/' \
  | sort -u > "$TESTS_FILE"

# ---------- compare ----------

gaps=0
total=0
missing_lines=()

while IFS=$'\t' read -r sym label; do
  [ -z "$sym" ] && continue
  total=$((total + 1))
  missing=""
  for variant in Good Bad Ugly; do
    # Match Test{anything}_{sym}_{variant} anywhere in the test corpus
    if ! grep -qE "^Test[A-Za-z0-9_]+_${sym}_${variant}\$" "$TESTS_FILE"; then
      missing="${missing} ${variant}"
    fi
  done
  if [ -n "$missing" ]; then
    missing_lines+=("${label}.go: ${sym} — missing:${missing}")
    gaps=$((gaps + 1))
  fi
done < "$SYMBOLS_FILE"

if [ "$gaps" -gt 0 ]; then
  printf '%s\n' "${missing_lines[@]}"
  echo
  echo "$gaps of $total public symbols missing one or more Test*_{Good,Bad,Ugly} variants."
  echo "AX-7: every public symbol gets the triplet; gaps mean the surface is too narrow"
  echo "or the edges aren't understood. Fill in the missing variants."
  exit 1
fi

echo "AX-7 clean: all $total public symbols have Test*_{Good,Bad,Ugly} triplets."
