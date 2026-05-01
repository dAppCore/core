#!/usr/bin/env bash
# Re-audit every Go repo locally — find drift in repos that previously closed
# v0.9.0 tickets. Outputs a CSV of repo,verdict,total_findings.
set -uo pipefail

OUT=/tmp/audit-sweep.csv
echo "repo,verdict,total" > "$OUT"

REPOS=(
  agent api app BugSETI cli config docs go-ai go-ansible go-api
  go-blockchain go-build go-cache go-cgo go-config go-container
  go-crypt go-devops go-forge go-git go-html go-i18n go-inference
  go-infra go-io go-lns go-log go-miner go-ml go-mlx go-netops
  go-p2p go-pool go-process go-proxy go-rag go-ratelimit go-rocm
  go-scm go-session go-store go-stream go-tenant go-update
  go-webview go-ws gui ide lem lint mcp php play
)

for r in "${REPOS[@]}"; do
  d="$HOME/Code/core/$r"
  [ -d "$d/.git" ] || { echo "$r,MISSING,0" >> "$OUT"; continue; }
  ( cd "$d" && git fetch homelab dev --quiet 2>/dev/null && git reset --hard homelab/dev --quiet 2>/dev/null ) || true
  audit_out=$(cd "$d" && bash $HOME/Code/core/go/tests/cli/v090-upgrade/audit.sh . 2>&1 || true)
  clean=$(echo "$audit_out" | sed 's/\x1b\[[0-9;]*m//g')
  if echo "$clean" | grep -q 'verdict: COMPLIANT'; then
    verdict=COMPLIANT
    total=0
  else
    verdict=NON-COMPLIANT
    total=$(echo "$clean" | grep -oE 'NON-COMPLIANT — [0-9]+ findings' | grep -oE '[0-9]+' | head -1)
    total="${total:-?}"
  fi
  echo "$r,$verdict,$total" >> "$OUT"
done

echo "=== sweep complete ==="
column -t -s, "$OUT"
