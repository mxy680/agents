FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the export-creds binary (built externally)
COPY bin/export-creds /usr/local/bin/export-creds

# Copy resolve script
COPY scripts/resolve-creds.sh /usr/local/bin/resolve-creds.sh
RUN chmod +x /usr/local/bin/resolve-creds.sh

ENTRYPOINT ["/usr/local/bin/resolve-creds.sh"]
