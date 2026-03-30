#!/bin/bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/off_market_scan"
mkdir -p "$OUT_DIR"

DATE=$(date +%Y-%m-%d)
FAILED=0
SKIPPED=0
SUCCEEDED=0

echo "━━━ Off-Market R8+ Scanner ━━━"
echo "Date: $DATE"
echo "Output: $OUT_DIR"
echo ""

run_phase() {
  local num=$1
  local name=$2
  local script=$3
  local phase_id=$4

  echo "━━━ Phase $num: $name ━━━"

  python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import is_phase_done
if is_phase_done('real-estate', 'off-market-scan', '$phase_id'):
    print('  [SKIP] Already completed today')
    sys.exit(42)
" 2>/dev/null
  if [ $? -eq 42 ]; then
    SKIPPED=$((SKIPPED + 1))
    echo ""
    return 0
  fi

  python3 "$SCRIPT_DIR/$script" 2>&1
  local exit_code=$?

  if [ $exit_code -eq 0 ]; then
    python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_done
mark_phase_done('real-estate', 'off-market-scan', '$phase_id')
" 2>/dev/null
    SUCCEEDED=$((SUCCEEDED + 1))
  else
    echo "  [ERROR] Phase $num failed (exit code $exit_code)"
    python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_failed
mark_phase_failed('real-estate', 'off-market-scan', '$phase_id')
" 2>/dev/null
    FAILED=$((FAILED + 1))
  fi
  echo ""
}

run_phase 1  "PLUTO Query (R8+ small residential)"  phase1_pluto_query.py  phase1_pluto
run_phase 2  "Signal Checks"                         phase2_signals.py      phase2_signals
run_phase 3  "Cluster Detection"                      phase3_clusters.py     phase3_clusters
run_phase 4  "Verification"                            phase4_verify.py       phase4_verify
run_phase 5  "XLSX Spreadsheet"                        phase5_xlsx.py         phase5_xlsx
run_phase 6  "PDF Report"                              phase6_pdf.py          phase6_pdf
run_phase 7  "Upload to Google Drive"                  phase7_upload.py       phase7_upload

echo ""
echo "━━━ Pipeline Complete ━━━"
echo "  Succeeded: $SUCCEEDED | Failed: $FAILED | Skipped: $SKIPPED"

if [ $FAILED -gt 0 ]; then
  exit 1
fi
