#!/bin/bash
# agent-runner.sh — Clotho-Verified Queue Runner for AgentCI.
# Deployed to agent machines, triggered by cron every 5 minutes.
#
# Usage: */5 * * * * ~/ai-work/agent-runner.sh >> ~/ai-work/logs/runner.log 2>&1
set -euo pipefail

WORK_DIR="${HOME}/ai-work"
QUEUE_DIR="${WORK_DIR}/queue"
ACTIVE_DIR="${WORK_DIR}/active"
DONE_DIR="${WORK_DIR}/done"
LOG_DIR="${WORK_DIR}/logs"
LOCK_FILE="${WORK_DIR}/.runner.lock"

# Ensure directories exist.
mkdir -p "$QUEUE_DIR" "$ACTIVE_DIR" "$DONE_DIR" "$LOG_DIR"

# --- 1. Check lock (is another run active?) ---
if [ -f "$LOCK_FILE" ]; then
    PID=$(cat "$LOCK_FILE" 2>/dev/null || echo "")
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        echo "$(date -Iseconds) Runner already active (PID $PID), exiting."
        exit 0
    fi
    echo "$(date -Iseconds) Removing stale lock (PID $PID)."
    rm -f "$LOCK_FILE"
fi

# --- 2. Pick oldest ticket ---
TICKET=$(find "$QUEUE_DIR" -name 'ticket-*.json' -type f 2>/dev/null | sort | head -1)
if [ -z "$TICKET" ]; then
    exit 0  # No work
fi

TICKET_BASENAME=$(basename "$TICKET")
echo "$(date -Iseconds) Processing ticket: $TICKET_BASENAME"

# --- 3. Lock ---
echo $$ > "$LOCK_FILE"
cleanup() {
    rm -f "$LOCK_FILE"
    # Secure cleanup of env file if it still exists.
    if [ -n "${ENV_FILE:-}" ] && [ -f "$ENV_FILE" ]; then
        rm -f "$ENV_FILE"
    fi
    echo "$(date -Iseconds) Lock released."
}
trap cleanup EXIT

# --- 4. Move to active ---
mv "$TICKET" "$ACTIVE_DIR/"
TICKET_FILE="$ACTIVE_DIR/$TICKET_BASENAME"

# --- 5. Extract ticket data ---
ID=$(jq -r .id "$TICKET_FILE")
REPO_OWNER=$(jq -r .repo_owner "$TICKET_FILE")
REPO_NAME=$(jq -r .repo_name "$TICKET_FILE")
ISSUE_NUM=$(jq -r .issue_number "$TICKET_FILE")
ISSUE_TITLE=$(jq -r .issue_title "$TICKET_FILE")
ISSUE_BODY=$(jq -r .issue_body "$TICKET_FILE")
TARGET_BRANCH=$(jq -r .target_branch "$TICKET_FILE")
FORGE_URL=$(jq -r .forge_url "$TICKET_FILE")
DUAL_RUN=$(jq -r '.dual_run // false' "$TICKET_FILE")
MODEL=$(jq -r '.model // "sonnet"' "$TICKET_FILE")
RUNNER=$(jq -r '.runner // "claude"' "$TICKET_FILE")
VERIFY_MODEL=$(jq -r '.verify_model // ""' "$TICKET_FILE")

echo "$(date -Iseconds) Issue: ${REPO_OWNER}/${REPO_NAME}#${ISSUE_NUM} - ${ISSUE_TITLE}"

# --- 6. Load secure token from .env file ---
ENV_FILE="$QUEUE_DIR/.env.$ID"
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
    rm -f "$ENV_FILE"  # Delete immediately after sourcing
else
    echo "$(date -Iseconds) ERROR: Token file not found for ticket $ID"
    mv "$TICKET_FILE" "$DONE_DIR/"
    exit 1
fi

if [ -z "${FORGE_TOKEN:-}" ]; then
    echo "$(date -Iseconds) ERROR: FORGE_TOKEN missing from env file."
    mv "$TICKET_FILE" "$DONE_DIR/"
    exit 1
fi

# --- 7. Clone or update repo ---
JOB_DIR="$WORK_DIR/jobs/${REPO_OWNER}-${REPO_NAME}-${ISSUE_NUM}"
REPO_DIR="$JOB_DIR/$REPO_NAME"
mkdir -p "$JOB_DIR"

FORGEJO_USER=$(jq -r '.forgejo_user // empty' "$TICKET_FILE")
if [ -z "$FORGEJO_USER" ]; then
    FORGEJO_USER="$(hostname -s)-$(whoami)"
fi
# TODO: Replace token-in-URL with git credential helper or SSH clone via charmbracelet/keygen.
CLONE_URL="https://${FORGEJO_USER}:${FORGE_TOKEN}@${FORGE_URL#https://}/${REPO_OWNER}/${REPO_NAME}.git"

if [ -d "$REPO_DIR/.git" ]; then
    echo "$(date -Iseconds) Updating existing clone..."
    cd "$REPO_DIR"
    git fetch origin
    git checkout "$TARGET_BRANCH" 2>/dev/null || git checkout -b "$TARGET_BRANCH" "origin/$TARGET_BRANCH"
    git pull origin "$TARGET_BRANCH"
else
    echo "$(date -Iseconds) Cloning repo..."
    git clone -b "$TARGET_BRANCH" "$CLONE_URL" "$REPO_DIR"
    cd "$REPO_DIR"
fi

# --- 8. Agent execution function ---
run_agent() {
    local model="$1"
    local log_suffix="$2"
    local prompt="You are working on issue #${ISSUE_NUM} in ${REPO_OWNER}/${REPO_NAME}.

Title: ${ISSUE_TITLE}

Description:
${ISSUE_BODY}

The repo is cloned at the current directory on branch '${TARGET_BRANCH}'.
Create a feature branch from '${TARGET_BRANCH}', make minimal targeted changes, commit referencing #${ISSUE_NUM}, and push.
Then create a PR targeting '${TARGET_BRANCH}' using the forgejo MCP tools or git push."

    local log_file="$LOG_DIR/${ID}-${log_suffix}.log"
    echo "$(date -Iseconds) Running ${RUNNER} (model: ${model}, suffix: ${log_suffix})..."

    case "$RUNNER" in
        codex)
            codex exec --full-auto "$prompt" > "$log_file" 2>&1
            ;;
        gemini)
            local model_flag=""
            if [ -n "$model" ] && [ "$model" != "sonnet" ]; then
                model_flag="-m $model"
            fi
            echo "$prompt" | gemini -p - -y $model_flag > "$log_file" 2>&1
            ;;
        *)
            echo "$prompt" | claude -p \
                --model "$model" \
                --dangerously-skip-permissions \
                --output-format text \
                > "$log_file" 2>&1
            ;;
    esac
    return $?
}

# --- 9. Execute ---
run_agent "$MODEL" "primary"
EXIT_CODE_A=$?

FINAL_EXIT=$EXIT_CODE_A
COMMENT=""

if [ "$DUAL_RUN" = "true" ] && [ -n "$VERIFY_MODEL" ]; then
    echo "$(date -Iseconds) Clotho Dual Run: resetting for verifier..."
    HASH_A=$(git rev-parse HEAD)
    git checkout "$TARGET_BRANCH" 2>/dev/null || true

    run_agent "$VERIFY_MODEL" "verifier"
    EXIT_CODE_B=$?
    HASH_B=$(git rev-parse HEAD)

    # Compare the two runs.
    echo "$(date -Iseconds) Comparing threads..."
    DIFF_COUNT=$(git diff --shortstat "$HASH_A" "$HASH_B" 2>/dev/null | wc -l || echo "1")

    if [ "$DIFF_COUNT" -eq 0 ] && [ "$EXIT_CODE_A" -eq 0 ] && [ "$EXIT_CODE_B" -eq 0 ]; then
        echo "$(date -Iseconds) Clotho Verification: Threads converged."
        FINAL_EXIT=0
        git checkout "$HASH_A" 2>/dev/null
        git push origin "HEAD:refs/heads/feat/issue-${ISSUE_NUM}"
    else
        echo "$(date -Iseconds) Clotho Verification: Divergence detected."
        FINAL_EXIT=1
        COMMENT="**Clotho Verification Failed**\n\nPrimary ($MODEL) and Verifier ($VERIFY_MODEL) produced divergent results.\nPrimary Exit: $EXIT_CODE_A | Verifier Exit: $EXIT_CODE_B"
    fi
else
    # Standard single run — push if successful.
    if [ $FINAL_EXIT -eq 0 ]; then
        git push origin "HEAD:refs/heads/feat/issue-${ISSUE_NUM}" 2>/dev/null || true
    fi
fi

# --- 10. Move to done ---
mv "$TICKET_FILE" "$DONE_DIR/"

# --- 11. Report result back to Forgejo ---
if [ $FINAL_EXIT -eq 0 ] && [ -z "$COMMENT" ]; then
    COMMENT="Agent completed work on #${ISSUE_NUM}. Exit code: 0."
elif [ -z "$COMMENT" ]; then
    COMMENT="Agent failed on #${ISSUE_NUM} (exit code: ${FINAL_EXIT}). Check logs on agent machine."
fi

curl -s -X POST "${FORGE_URL}/api/v1/repos/${REPO_OWNER}/${REPO_NAME}/issues/${ISSUE_NUM}/comments" \
    -H "Authorization: token $FORGE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$(jq -n --arg body "$COMMENT" '{body: $body}')" \
    > /dev/null 2>&1 || true

echo "$(date -Iseconds) Done: $TICKET_BASENAME (exit: $FINAL_EXIT)"
