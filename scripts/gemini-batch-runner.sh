#!/bin/bash
# gemini-batch-runner.sh — Rate-limit-aware tiered Gemini analysis pipeline.
#
# Uses cheap models to prep work for expensive models, respecting TPM limits.
# Designed for Tier 1 (1M TPM) with 80% safety margin (800K effective).
#
# Usage: ./scripts/gemini-batch-runner.sh <batch-number> <pkg1> <pkg2> ...
# Example: ./scripts/gemini-batch-runner.sh 1 log config io crypt auth
set -euo pipefail

BATCH_NUM="${1:?Usage: gemini-batch-runner.sh <batch-num> <pkg1> [pkg2] ...}"
shift
PACKAGES=("$@")

if [ ${#PACKAGES[@]} -eq 0 ]; then
    echo "Error: No packages specified" >&2
    exit 1
fi

# --- Config ---
API_KEY="${GEMINI_API_KEY:?Set GEMINI_API_KEY}"
API_BASE="https://generativelanguage.googleapis.com/v1beta/models"
TPM_LIMIT=800000  # 80% of 1M Tier 1 limit
OUTPUT_DIR="${OUTPUT_DIR:-docs}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

# Models (cheapest → most expensive)
MODEL_LITE="gemini-2.5-flash-lite"
MODEL_FLASH="gemini-3-flash-preview"
MODEL_PRO="gemini-3-pro-preview"

# --- Helpers ---
log() { echo "$(date -Iseconds) $*"; }

api_call() {
    local model="$1" prompt_file="$2" max_tokens="${3:-4096}"
    local tmpfile
    tmpfile=$(mktemp /tmp/gemini-payload-XXXXXX.json)
    trap "rm -f '$tmpfile'" RETURN

    # Read prompt from file to avoid argument length limits.
    jq -n --rawfile text "$prompt_file" --argjson max "$max_tokens" \
        '{contents: [{parts: [{text: $text}]}], generationConfig: {maxOutputTokens: $max}}' \
        > "$tmpfile"

    local response
    response=$(curl -s "${API_BASE}/${model}:generateContent?key=${API_KEY}" \
        -H 'Content-Type: application/json' \
        -d "@${tmpfile}")

    # Check for errors
    local error
    error=$(echo "$response" | jq -r '.error.message // empty')
    if [ -n "$error" ]; then
        log "ERROR from $model: $error"
        # Rate limited — wait and retry once
        if echo "$error" | grep -qi "rate\|quota\|resource_exhausted"; then
            log "Rate limited. Waiting 60s..."
            sleep 60
            response=$(curl -s "${API_BASE}/${model}:generateContent?key=${API_KEY}" \
                -H 'Content-Type: application/json' \
                -d "@${tmpfile}")
        else
            echo "$response"
            return 1
        fi
    fi

    echo "$response"
}

extract_text() {
    jq -r '.candidates[0].content.parts[0].text // "ERROR: no output"'
}

extract_tokens() {
    jq -r '.usageMetadata.totalTokenCount // 0'
}

# --- 1. Build context bundle ---
log "Building context for batch ${BATCH_NUM}: ${PACKAGES[*]}"

CONTEXT_FILE=$(mktemp /tmp/gemini-context-XXXXXX.txt)
trap "rm -f '$CONTEXT_FILE' /tmp/gemini-prompt-*.txt" EXIT

TOTAL_LINES=0
for pkg in "${PACKAGES[@]}"; do
    PKG_DIR="${REPO_ROOT}/pkg/${pkg}"
    if [ ! -d "$PKG_DIR" ]; then
        log "WARN: pkg/${pkg} not found, skipping"
        continue
    fi
    echo "=== Package: pkg/${pkg} ===" >> "$CONTEXT_FILE"
    while IFS= read -r -d '' f; do
        echo "--- $(basename "$f") ---" >> "$CONTEXT_FILE"
        cat "$f" >> "$CONTEXT_FILE"
        echo "" >> "$CONTEXT_FILE"
        TOTAL_LINES=$((TOTAL_LINES + $(wc -l < "$f")))
    done < <(find "$PKG_DIR" -maxdepth 1 -name '*.go' ! -name '*_test.go' -type f -print0 | sort -z)
done

EST_TOKENS=$((TOTAL_LINES * 4))
log "Context: ${TOTAL_LINES} lines (~${EST_TOKENS} tokens estimated)"

if [ "$EST_TOKENS" -gt "$TPM_LIMIT" ]; then
    log "WARNING: Estimated tokens (${EST_TOKENS}) exceeds TPM budget (${TPM_LIMIT})"
    log "Consider splitting this batch further."
    exit 1
fi

# Helper: write prompt to temp file (prefix + context)
write_prompt() {
    local outfile="$1" prefix="$2"
    echo "$prefix" > "$outfile"
    echo "" >> "$outfile"
    cat "$CONTEXT_FILE" >> "$outfile"
}

# --- 2. Flash Lite: quick scan (verify batch is reasonable) ---
log "Step 1/3: Flash Lite scan..."
LITE_FILE=$(mktemp /tmp/gemini-prompt-XXXXXX.txt)
write_prompt "$LITE_FILE" "For each Go package below, give a one-line description and list the exported types. Be very concise."

LITE_RESP=$(api_call "$MODEL_LITE" "$LITE_FILE" 2048)
LITE_TOKENS=$(echo "$LITE_RESP" | extract_tokens)
log "Flash Lite used ${LITE_TOKENS} tokens"

# --- 3. Flash: structured prep ---
log "Step 2/3: Gemini 3 Flash prep..."
FLASH_FILE=$(mktemp /tmp/gemini-prompt-XXXXXX.txt)
write_prompt "$FLASH_FILE" "You are analyzing Go packages for documentation. For each package below, produce:
1. A one-line description
2. Key exported types and functions (names + one-line purpose)
3. Dependencies on other packages in this codebase (pkg/* imports only)
4. Complexity rating (simple/moderate/complex)

Output as structured markdown. Be concise."

FLASH_RESP=$(api_call "$MODEL_FLASH" "$FLASH_FILE" 4096)
FLASH_TEXT=$(echo "$FLASH_RESP" | extract_text)
FLASH_TOKENS=$(echo "$FLASH_RESP" | extract_tokens)
log "Gemini 3 Flash used ${FLASH_TOKENS} tokens"

# Check cumulative TPM before hitting Pro
CUMULATIVE=$((LITE_TOKENS + FLASH_TOKENS))
if [ "$CUMULATIVE" -gt "$((TPM_LIMIT / 2))" ]; then
    log "Cumulative tokens high (${CUMULATIVE}). Pausing 60s before Pro call..."
    sleep 60
fi

# --- 4. Pro: deep analysis ---
log "Step 3/3: Gemini 3 Pro deep analysis..."
PRO_FILE=$(mktemp /tmp/gemini-prompt-XXXXXX.txt)
write_prompt "$PRO_FILE" "You are a senior Go engineer documenting a framework. Analyze each package below and produce a detailed markdown document with:

For each package:
1. **Overview**: 2-3 sentence description of purpose and design philosophy
2. **Public API**: All exported types, functions, methods with type signatures and brief purpose
3. **Internal Design**: Key patterns used (interfaces, generics, dependency injection, etc.)
4. **Dependencies**: What pkg/* packages it imports and why
5. **Test Coverage Notes**: What would need testing based on the API surface
6. **Integration Points**: How other packages would use this package

Output as a single structured markdown document."

PRO_RESP=$(api_call "$MODEL_PRO" "$PRO_FILE" 8192)
PRO_TEXT=$(echo "$PRO_RESP" | extract_text)
PRO_TOKENS=$(echo "$PRO_RESP" | extract_tokens)
log "Gemini 3 Pro used ${PRO_TOKENS} tokens"

TOTAL_TOKENS=$((LITE_TOKENS + FLASH_TOKENS + PRO_TOKENS))
log "Total tokens for batch ${BATCH_NUM}: ${TOTAL_TOKENS}"

# --- 5. Save output ---
mkdir -p "${REPO_ROOT}/${OUTPUT_DIR}"
OUTPUT_FILE="${REPO_ROOT}/${OUTPUT_DIR}/pkg-batch${BATCH_NUM}-analysis.md"

cat > "$OUTPUT_FILE" << HEADER
# Package Analysis — Batch ${BATCH_NUM}

Generated by: gemini-batch-runner.sh
Models: ${MODEL_LITE} → ${MODEL_FLASH} → ${MODEL_PRO}
Date: $(date -I)
Packages: ${PACKAGES[*]}
Total tokens: ${TOTAL_TOKENS}

---

HEADER

echo "$PRO_TEXT" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" << FOOTER

---

## Quick Reference (Flash Summary)

${FLASH_TEXT}
FOOTER

log "Output saved to ${OUTPUT_FILE}"
log "Done: Batch ${BATCH_NUM} (${#PACKAGES[@]} packages, ${TOTAL_TOKENS} tokens)"
