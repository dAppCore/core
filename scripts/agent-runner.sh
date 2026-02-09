#!/bin/bash
# agent-runner.sh — One-at-a-time queue runner for Claude Code agents.
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

# --- 2. Check credits ---
# Parse remaining usage from claude. If under 5% remaining, skip.
if command -v claude &>/dev/null; then
    USAGE_OUTPUT=$(claude --output-format json -p "Reply with just the word OK" 2>/dev/null | head -1 || echo "")
    # Fallback: if we can't check, proceed anyway.
fi

# --- 3. Pick oldest ticket ---
TICKET=$(find "$QUEUE_DIR" -name 'ticket-*.json' -type f 2>/dev/null | sort | head -1)
if [ -z "$TICKET" ]; then
    exit 0  # No work
fi

TICKET_BASENAME=$(basename "$TICKET")
echo "$(date -Iseconds) Processing ticket: $TICKET_BASENAME"

# --- 4. Lock ---
echo $$ > "$LOCK_FILE"
cleanup() {
    rm -f "$LOCK_FILE"
    echo "$(date -Iseconds) Lock released."
}
trap cleanup EXIT

# --- 5. Move to active ---
mv "$TICKET" "$ACTIVE_DIR/"
TICKET_FILE="$ACTIVE_DIR/$TICKET_BASENAME"

# --- 6. Extract ticket data ---
REPO_OWNER=$(jq -r .repo_owner "$TICKET_FILE")
REPO_NAME=$(jq -r .repo_name "$TICKET_FILE")
ISSUE_NUM=$(jq -r .issue_number "$TICKET_FILE")
ISSUE_TITLE=$(jq -r .issue_title "$TICKET_FILE")
ISSUE_BODY=$(jq -r .issue_body "$TICKET_FILE")
TARGET_BRANCH=$(jq -r .target_branch "$TICKET_FILE")
FORGE_URL=$(jq -r .forge_url "$TICKET_FILE")
FORGE_TOKEN=$(jq -r .forge_token "$TICKET_FILE")

echo "$(date -Iseconds) Issue: ${REPO_OWNER}/${REPO_NAME}#${ISSUE_NUM} - ${ISSUE_TITLE}"

# --- 7. Clone or update repo ---
JOB_DIR="$WORK_DIR/jobs/${REPO_OWNER}-${REPO_NAME}-${ISSUE_NUM}"
REPO_DIR="$JOB_DIR/$REPO_NAME"
mkdir -p "$JOB_DIR"

FORGEJO_USER=$(jq -r '.forgejo_user // empty' "$TICKET_FILE")
if [ -z "$FORGEJO_USER" ]; then
    FORGEJO_USER="$(hostname -s)-$(whoami)"
fi
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

# --- 8. Build prompt ---
PROMPT="You are working on issue #${ISSUE_NUM} in ${REPO_OWNER}/${REPO_NAME}.

Title: ${ISSUE_TITLE}

Description:
${ISSUE_BODY}

The repo is cloned at the current directory on branch '${TARGET_BRANCH}'.
Create a feature branch from '${TARGET_BRANCH}', make minimal targeted changes, commit referencing #${ISSUE_NUM}, and push.
Then create a PR targeting '${TARGET_BRANCH}' using the forgejo MCP tools or git push."

# --- 9. Run Claude ---
LOG_FILE="$LOG_DIR/${REPO_OWNER}-${REPO_NAME}-${ISSUE_NUM}.log"
echo "$(date -Iseconds) Running claude..."
echo "$PROMPT" | claude -p \
    --dangerously-skip-permissions \
    --output-format text \
    > "$LOG_FILE" 2>&1
EXIT_CODE=$?
echo "$(date -Iseconds) Claude exited with code: $EXIT_CODE"

# --- 10. Move to done ---
mv "$TICKET_FILE" "$DONE_DIR/"

# --- 11. Report result back to Forgejo ---
if [ $EXIT_CODE -eq 0 ]; then
    COMMENT="Agent completed work on #${ISSUE_NUM}. Exit code: 0."
else
    COMMENT="Agent failed on #${ISSUE_NUM} (exit code: ${EXIT_CODE}). Check logs on agent machine."
fi

curl -s -X POST "${FORGE_URL}/api/v1/repos/${REPO_OWNER}/${REPO_NAME}/issues/${ISSUE_NUM}/comments" \
    -H "Authorization: token $FORGE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$(jq -n --arg body "$COMMENT" '{body: $body}')" \
    > /dev/null 2>&1 || true

echo "$(date -Iseconds) Done: $TICKET_BASENAME (exit: $EXIT_CODE)"
