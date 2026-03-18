FROM node:22-slim

# Install system deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Create workspace
RUN mkdir -p /agent/workspace

WORKDIR /agent/workspace

# Default entrypoint (overridden by pod spec)
CMD ["/bin/sh", "entrypoint.sh"]
