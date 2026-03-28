#!/bin/bash
# Deploy to Fly.io with rolling strategy
set -euo pipefail

# Pre-deploy check: abort if jobs are running
DB_URL=$(doppler run --project agents --config prd -- bash -c 'echo $SUPABASE_DB_URL')
RUNNING=$(/opt/homebrew/opt/postgresql@17/bin/psql "$DB_URL" -t -A -c "SELECT count(*) FROM local_job_runs WHERE status = 'running';" 2>/dev/null || echo "0")

if [ "$RUNNING" -gt 0 ]; then
  echo "ABORT: $RUNNING job(s) still running. Wait for them to finish or cancel them first."
  /opt/homebrew/opt/postgresql@17/bin/psql "$DB_URL" -c "SELECT id, job_slug, started_at FROM local_job_runs WHERE status = 'running';" 2>/dev/null
  exit 1
fi

ANON_KEY=$(doppler run --project agents --config prd -- bash -c 'echo $NEXT_PUBLIC_SUPABASE_ANON_KEY')
SUPABASE_URL=$(doppler run --project agents --config prd -- bash -c 'echo $NEXT_PUBLIC_SUPABASE_URL')

flyctl deploy \
  --build-arg NEXT_PUBLIC_SUPABASE_URL="$SUPABASE_URL" \
  --build-arg NEXT_PUBLIC_SUPABASE_ANON_KEY="$ANON_KEY" \
  --remote-only \
  --strategy rolling
