#!/usr/bin/env bash
# SPOR ownership check — fails when any stdlib package is imported by
# more than one production .go file in core/go's root.
#
# Production = *.go that is NOT *_test.go, *_example_test.go, or
# *_fuzz_test.go. Imports are extracted from `import "..."` blocks.

set -euo pipefail

cd "$(dirname "$0")/../../.."

# Stdlib packages to track (every package the codebase currently uses).
# Keep this list aligned with AGENTS.md's SPOR ownership table.
PACKAGES=(
  bufio bytes cmp "compress/gzip" context
  "crypto/hkdf" "crypto/hmac" "crypto/rand" "crypto/sha256" "crypto/sha3" "crypto/sha512"
  "database/sql" embed
  "encoding/base64" "encoding/binary" "encoding/hex" "encoding/json"
  errors fmt
  "go/ast" "go/parser" "go/token"
  hash html "html/template" io "io/fs" iter
  maps math "math/big" "math/bits" "math/rand/v2"
  "mime/multipart" net "net/http" "net/http/httptest" "net/url"
  os "os/exec" "os/user" "path/filepath"
  reflect regexp runtime "runtime/debug"
  slices sort strconv strings sync "sync/atomic"
  testing "text/tabwriter" "text/template"
  time "unicode/utf8"
)

violations=0
for pkg in "${PACKAGES[@]}"; do
  # Match either tab-indented "pkg" inside an import block, or a single-line import "pkg"
  files=$(grep -lE "^	\"${pkg}\"\$|^import \"${pkg}\"\$" *.go 2>/dev/null \
            | grep -v -e "_test\.go\$" -e "_example_test\.go\$" -e "_fuzz_test\.go\$" \
            || true)
  count=$(printf '%s\n' "$files" | grep -c . || true)
  if [ "$count" -gt 1 ]; then
    echo "SPOR violation: $pkg imported by $count files:"
    printf '  %s\n' $files
    violations=$((violations + 1))
  fi
done

if [ "$violations" -gt 0 ]; then
  echo
  echo "$violations stdlib package(s) imported by more than one production owner."
  echo "Move shared usage to the canonical owner file (see AGENTS.md SPOR table)."
  exit 1
fi

echo "SPOR clean: every stdlib package has exactly one production owner."
