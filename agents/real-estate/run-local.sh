#!/bin/bash
#
# Run the Bronx Assemblage Scout locally.
#
# Usage:
#   ./run-local.sh                    # Run the off-market scan job
#   ./run-local.sh "custom prompt"    # Run with a custom prompt
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Determine prompt
if [ $# -gt 0 ]; then
  PROMPT="$1"
else
  PROMPT="$(cat "$SCRIPT_DIR/jobs/off-market-scan.md")"
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
SESSION_FILE=$(mktemp /tmp/agent-session-XXXXXXXX)
trap "rm -f $SESSION_FILE" EXIT

# Write session.json
jq -n \
  --arg prompt "$PROMPT" \
  --arg systemPrompt "$SYSTEM_PROMPT" \
  '{prompt: $prompt, systemPrompt: $systemPrompt, model: "claude-sonnet-4-6"}' \
  > "$SESSION_FILE"

echo "Starting Real Estate Agent..."
echo "---"

# Ensure Agent SDK is resolvable
export NODE_PATH="$(npm root -g):${NODE_PATH:-}"

# Resolve fresh integration credentials from Supabase and write to a temp env file.
# doppler run starts a new process, so we can't just eval exports — we need to
# write them to a file and source them inside the doppler-run context.
echo "Resolving credentials from Supabase..."
CRED_FILE=$(mktemp /tmp/agent-creds-XXXXXXXX)
# #14: Use trap without exec so this EXIT trap fires for cleanup
trap "rm -f $SESSION_FILE $CRED_FILE" EXIT

# #2: Remove 2>/dev/null so credential errors are visible
doppler run --project agents --config dev -- node "$SCRIPT_DIR/resolve-creds.mjs" > "$CRED_FILE"
if [ $? -ne 0 ]; then
  echo "ERROR: credential resolution failed — aborting" >&2
  exit 1
fi

# #2: Check that the creds file is non-empty before sourcing
if [ ! -s "$CRED_FILE" ]; then
  echo "ERROR: no credentials were resolved — aborting" >&2
  exit 1
fi

# #14: Use a regular call (not exec) so the EXIT trap above fires for cleanup
doppler run --project agents --config dev -- \
  bash -c "source '$CRED_FILE' && node '$SCRIPT_DIR/entrypoint.mjs' '$SESSION_FILE'"
