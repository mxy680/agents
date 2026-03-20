FROM node:22-slim

# Install system deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI and Agent SDK
RUN npm install -g @anthropic-ai/claude-code @anthropic-ai/claude-agent-sdk

# Create non-root user (Claude Code refuses bypassPermissions as root)
RUN useradd -m -s /bin/bash agent

# Create workspace and entrypoint dir
RUN mkdir -p /agent/workspace && chown -R agent:agent /agent
RUN mkdir -p /app

# Copy integrations CLI binary
COPY docker/agent-base/bin/integrations /usr/local/bin/integrations

# Copy entrypoint (lives outside workspace so bind mounts don't overwrite it)
COPY docker/agent-base/entrypoint.mjs /app/entrypoint.mjs

# ESM ignores NODE_PATH — symlink global packages so they resolve from /app
RUN ln -s /usr/local/lib/node_modules /app/node_modules

USER agent
WORKDIR /agent/workspace

ENTRYPOINT ["node", "/app/entrypoint.mjs"]
