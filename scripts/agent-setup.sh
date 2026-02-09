#!/bin/bash
# agent-setup.sh — Bootstrap an AgentCI agent machine via SSH.
#
# Usage: agent-setup.sh <user@host>
#
# Creates work directories, copies agent-runner.sh, installs cron,
# and verifies prerequisites.
set -euo pipefail

HOST="${1:?Usage: agent-setup.sh <user@host>}"
SSH_OPTS="-o StrictHostKeyChecking=accept-new -o ConnectTimeout=10"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RUNNER_SCRIPT="${SCRIPT_DIR}/agent-runner.sh"

if [ ! -f "$RUNNER_SCRIPT" ]; then
    echo "ERROR: agent-runner.sh not found at $RUNNER_SCRIPT"
    exit 1
fi

echo "=== AgentCI Setup: $HOST ==="

# --- 1. Test SSH ---
echo -n "Testing SSH... "
if ! ssh $SSH_OPTS "$HOST" "echo ok" >/dev/null 2>&1; then
    echo "FAILED — cannot reach $HOST"
    exit 1
fi
echo "OK"

# --- 2. Create directories ---
echo -n "Creating directories... "
ssh $SSH_OPTS "$HOST" "mkdir -p ~/ai-work/{queue,active,done,logs,jobs}"
echo "OK"

# --- 3. Copy runner script ---
echo -n "Copying agent-runner.sh... "
scp $SSH_OPTS "$RUNNER_SCRIPT" "${HOST}:~/ai-work/agent-runner.sh"
ssh $SSH_OPTS "$HOST" "chmod +x ~/ai-work/agent-runner.sh"
echo "OK"

# --- 4. Install cron (idempotent) ---
echo -n "Installing cron... "
CRON_LINE="*/5 * * * * ~/ai-work/agent-runner.sh >> ~/ai-work/logs/runner.log 2>&1"
ssh $SSH_OPTS "$HOST" "
    if crontab -l 2>/dev/null | grep -qF 'agent-runner.sh'; then
        echo 'already installed'
    else
        (crontab -l 2>/dev/null; echo '$CRON_LINE') | crontab -
        echo 'installed'
    fi
"

# --- 5. Verify prerequisites ---
echo "Checking prerequisites..."
MISSING=""
for tool in jq git claude; do
    if ssh $SSH_OPTS "$HOST" "command -v $tool" >/dev/null 2>&1; then
        echo "  $tool: OK"
    else
        echo "  $tool: MISSING"
        MISSING="$MISSING $tool"
    fi
done

if [ -n "$MISSING" ]; then
    echo ""
    echo "WARNING: Missing tools:$MISSING"
    echo "Install them before the agent can process tickets."
fi

# --- 6. Round-trip test ---
echo -n "Round-trip test... "
TEST_FILE="queue/test-setup-$(date +%s).json"
ssh $SSH_OPTS "$HOST" "echo '{\"test\":true}' > ~/ai-work/$TEST_FILE"
RESULT=$(ssh $SSH_OPTS "$HOST" "cat ~/ai-work/$TEST_FILE && rm ~/ai-work/$TEST_FILE")
if [ "$RESULT" = '{"test":true}' ]; then
    echo "OK"
else
    echo "FAILED"
    exit 1
fi

echo ""
echo "=== Setup complete ==="
echo "Agent queue: $HOST:~/ai-work/queue/"
echo "Runner log:  $HOST:~/ai-work/logs/runner.log"
