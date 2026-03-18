FROM python:3.12-slim

# Install system deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Anthropic Agent SDK
RUN pip install --no-cache-dir anthropic[agent-sdk]

# Create workspace
RUN mkdir -p /agent/workspace

WORKDIR /agent/workspace

# Default entrypoint (overridden by pod spec)
CMD ["python", "entrypoint.py"]
