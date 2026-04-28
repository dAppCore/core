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

exec python3 tests/cli/naming/check.py "$@"
