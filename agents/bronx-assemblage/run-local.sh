#!/bin/bash
#
# Run the Bronx Assemblage Scout locally.
#
# Usage:
#   ./run-local.sh                    # Run the weekly scan job
#   ./run-local.sh "custom prompt"    # Run with a custom prompt
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Determine prompt
if [ $# -gt 0 ]; then
  PROMPT="$1"
else
  PROMPT="$(cat "$SCRIPT_DIR/jobs/weekly-scan.md")"
fi

# Build system prompt from role.md + CLAUDE.md
SYSTEM_PROMPT=""
if [ -f "$SCRIPT_DIR/role.md" ]; then
  SYSTEM_PROMPT="$(cat "$SCRIPT_DIR/role.md")"
fi
if [ -f "$SCRIPT_DIR/CLAUDE.md" ]; then
  SYSTEM_PROMPT="$SYSTEM_PROMPT

$(cat "$SCRIPT_DIR/CLAUDE.md")"
fi

# Ensure integrations binary is built
if [ ! -f "$REPO_ROOT/bin/integrations" ]; then
  echo "Building integrations CLI..."
  (cd "$REPO_ROOT" && make build)
fi

# Add integrations binary to PATH
export PATH="$REPO_ROOT/bin:$PATH"

# Create temp session file
SESSION_FILE=$(mktemp /tmp/agent-session-XXXXXX.json)
trap "rm -f $SESSION_FILE" EXIT

# Write session.json
jq -n \
  --arg prompt "$PROMPT" \
  --arg systemPrompt "$SYSTEM_PROMPT" \
  '{prompt: $prompt, systemPrompt: $systemPrompt, model: "claude-sonnet-4-6"}' \
  > "$SESSION_FILE"

echo "Starting Bronx Assemblage Scout..."
echo "---"

# Ensure Agent SDK is resolvable
export NODE_PATH="$(npm root -g):${NODE_PATH:-}"

# Run with doppler env vars (includes ANTHROPIC_API_KEY + integration credentials)
exec doppler run --project agents --config dev -- \
  node "$SCRIPT_DIR/entrypoint.mjs" "$SESSION_FILE"
