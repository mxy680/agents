#!/bin/bash
#
# Run the Bronx Assemblage Scout locally.
#
# Usage:
#   ./run-local.sh                    # Run the weekly scan job
#   ./run-local.sh "custom prompt"    # Run with a custom prompt
#
# Prerequisites:
#   - doppler CLI configured (project: agents, config: dev)
#   - integrations binary built (make build from repo root)
#   - @anthropic-ai/claude-agent-sdk installed globally or in node_modules
#   - ANTHROPIC_API_KEY set in doppler
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

# Create temp workspace
WORKSPACE=$(mktemp -d)
trap "rm -rf $WORKSPACE" EXIT

# Write session.json
cat > "$WORKSPACE/session.json" <<EOF
{
  "prompt": $(echo "$PROMPT" | jq -Rs .),
  "model": "claude-sonnet-4-6"
}
EOF

# Write system prompt files
echo "$SYSTEM_PROMPT" > "$WORKSPACE/role.md"

echo "Starting Bronx Assemblage Scout..."
echo "Workspace: $WORKSPACE"
echo "---"

# Run the agent with doppler env vars (includes ANTHROPIC_API_KEY + integration credentials)
exec doppler run --project agents --config dev -- \
  node -e "
    import { readFileSync } from 'fs';
    import { query } from '@anthropic-ai/claude-agent-sdk';

    const session = JSON.parse(readFileSync('$WORKSPACE/session.json', 'utf-8'));
    const systemPrompt = readFileSync('$WORKSPACE/role.md', 'utf-8');

    const conversation = query({
      prompt: session.prompt,
      options: {
        cwd: '$WORKSPACE',
        permissionMode: 'bypassPermissions',
        allowDangerouslySkipPermissions: true,
        systemPrompt,
        maxTurns: 30,
        model: session.model || 'claude-sonnet-4-6',
      },
    });

    for await (const event of conversation) {
      if (event.type === 'assistant' && event.message?.content) {
        for (const block of event.message.content) {
          if (block.type === 'text') {
            process.stderr.write(block.text);
          }
        }
      } else if (event.type === 'result') {
        process.stderr.write('\n---\nAgent finished.\n');
      }
    }
  "
