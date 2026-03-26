#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/off_market_scan"
mkdir -p "$OUT_DIR"
DATE=$(date +%Y-%m-%d)
echo "━━━ Off-Market R8+ Scanner ━━━"
echo "Date: $DATE"
echo "Output: $OUT_DIR"
echo ""

run_phase() {
  local num=$1
  local name=$2
  local script=$3
  echo "━━━ Phase $num: $name ━━━"
  python3 "$SCRIPT_DIR/$script" 2>&1
  echo ""
}

run_phase 1  "PLUTO Query (R8+ small residential)"  phase1_pluto_query.py
run_phase 2  "Signal Checks"                         phase2_signals.py
run_phase 3  "Cluster Detection"                      phase3_clusters.py
run_phase 4  "Verification"                            phase4_verify.py
run_phase 5  "XLSX Spreadsheet"                        phase5_xlsx.py
run_phase 6  "PDF Report"                              phase6_pdf.py
run_phase 7  "Upload to Google Drive"                  phase7_upload.py

echo ""
echo "━━━ Pipeline Complete ━━━"
