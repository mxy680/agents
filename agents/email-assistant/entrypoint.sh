#!/bin/sh
set -e

# Source integration credentials (written by init container)
if [ -f /tmp/creds/env.sh ]; then
    . /tmp/creds/env.sh
fi

# Build system prompt from workspace docs
SYSTEM_PROMPT=""
if [ -f /agent/workspace/role.md ]; then
    ROLE_CONTENT="$(cat /agent/workspace/role.md)"
    SYSTEM_PROMPT="${ROLE_CONTENT}"
fi
if [ -f /agent/workspace/CLAUDE.md ]; then
    CLAUDE_CONTENT="$(cat /agent/workspace/CLAUDE.md)"
    SYSTEM_PROMPT="${SYSTEM_PROMPT}

${CLAUDE_CONTENT}"
fi

if [ -z "$SYSTEM_PROMPT" ]; then
    SYSTEM_PROMPT="You are a helpful email assistant."
fi

# Run Claude Code in non-interactive print mode.
# CLAUDE_SESSION_TOKEN is injected via K8s secret.
# The integrations CLI binary must be in PATH.
exec claude --print \
    --system-prompt "$SYSTEM_PROMPT" \
    --allowedTools "Bash" \
    "List my recent unread emails using: integrations gmail messages list --query=is:unread --since=24h --json. Then summarize them."
