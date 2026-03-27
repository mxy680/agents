#!/bin/bash
# Deploy to Fly.io with rolling strategy (doesn't kill running jobs)
set -euo pipefail

ANON_KEY=$(doppler run --project agents --config prd -- bash -c 'echo $NEXT_PUBLIC_SUPABASE_ANON_KEY')
SUPABASE_URL=$(doppler run --project agents --config prd -- bash -c 'echo $NEXT_PUBLIC_SUPABASE_URL')

flyctl deploy \
  --build-arg NEXT_PUBLIC_SUPABASE_URL="$SUPABASE_URL" \
  --build-arg NEXT_PUBLIC_SUPABASE_ANON_KEY="$ANON_KEY" \
  --remote-only \
  --strategy rolling
