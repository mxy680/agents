#!/bin/bash
#
# Run the full NYC assemblage intelligence pipeline.
# Each phase is independent — failures are logged but don't stop the pipeline.
# Phases that completed today (tracked in Supabase) are automatically skipped.
#
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/nyc_assemblage"
mkdir -p "$OUT_DIR"

DATE=$(date +%Y-%m-%d)
FAILED=0
SKIPPED=0
SUCCEEDED=0

echo "━━━ NYC Assemblage Intelligence Pipeline ━━━"
echo "Date: $DATE"
echo "Output: $OUT_DIR"
echo ""

run_phase() {
  local num=$1
  local name=$2
  local script=$3
  local phase_id=$4

  echo "━━━ Phase $num: $name ━━━"

  # Check if phase already completed today
  python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import is_phase_done
if is_phase_done('real-estate', 'weekly-scan', '$phase_id'):
    print('  [SKIP] Already completed today')
    sys.exit(42)
" 2>/dev/null
  if [ $? -eq 42 ]; then
    SKIPPED=$((SKIPPED + 1))
    echo ""
    return 0
  fi

  # Run the phase
  python3 "$SCRIPT_DIR/$script" 2>&1
  local exit_code=$?

  if [ $exit_code -eq 0 ]; then
    # Mark as completed
    python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_done
mark_phase_done('real-estate', 'weekly-scan', '$phase_id')
" 2>/dev/null
    SUCCEEDED=$((SUCCEEDED + 1))
  else
    echo "  [ERROR] Phase $num failed (exit code $exit_code)"
    python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_failed
mark_phase_failed('real-estate', 'weekly-scan', '$phase_id')
" 2>/dev/null
    FAILED=$((FAILED + 1))
  fi
  echo ""
}

run_phase 1  "Zillow Search (from Supabase)"     phase1_zillow_search.py  phase1_zillow
run_phase 2  "PLUTO Geocoding + Zoning Filter"    phase2_pluto.py          phase2_pluto
run_phase 3  "Signal Checks"                       phase3_signals.py        phase3_signals
run_phase 4  "StreetEasy Enrichment"               phase4_streeteasy.py     phase4_streeteasy
run_phase 5  "Cluster Detection"                    phase5_clusters.py       phase5_clusters
run_phase 6  "Data Verification"                    phase6_verify.py         phase6_verify
run_phase 7  "XLSX Spreadsheet"                     phase7_xlsx.py           phase7_xlsx
run_phase 8  "LaTeX PDF Report"                     phase9_pdf.py            phase8_pdf
run_phase 9  "Upload to Google Drive"               phase10_upload.py        phase9_upload

echo "━━━ Pipeline Complete ━━━"
echo "  Succeeded: $SUCCEEDED | Failed: $FAILED | Skipped: $SKIPPED"

if [ $FAILED -gt 0 ]; then
  exit 1
fi
