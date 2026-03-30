#!/bin/sh
# resolve-creds.sh — init container script that resolves credentials
# and writes them as shell exports to /tmp/creds/env.sh
#
# Environment:
#   RESOLVE_PROVIDERS     — comma-separated list of providers (e.g. "google,github")
#   SUPABASE_DB_URL       — database connection string
#   ENCRYPTION_MASTER_KEY — decryption key
#   GOOGLE_CLIENT_ID      — for token refresh
#   GOOGLE_CLIENT_SECRET  — for token refresh

set -e

OUTPUT_FILE="/tmp/creds/env.sh"
: > "$OUTPUT_FILE"

IFS=',' read -ra PROVIDERS <<< "$RESOLVE_PROVIDERS"
for provider in "${PROVIDERS[@]}"; do
    echo "Resolving credentials for $provider..."
    export-creds --provider "$provider" >> "$OUTPUT_FILE"
done

echo "Credentials resolved for ${#PROVIDERS[@]} provider(s)"
chmod 400 "$OUTPUT_FILE"
