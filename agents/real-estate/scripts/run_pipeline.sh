#!/bin/bash
#
# Run the full NYC assemblage intelligence pipeline.
# Each phase reads from the previous phase's output JSON.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/nyc_assemblage"
mkdir -p "$OUT_DIR"

DATE=$(date +%Y-%m-%d)
echo "━━━ NYC Assemblage Intelligence Pipeline ━━━"
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

run_phase 1  "Zillow Search (137 zip codes)"     phase1_zillow_search.py
run_phase 2  "PLUTO Geocoding + Zoning Filter"    phase2_pluto.py
run_phase 3  "Signal Checks (ACRIS/DOB/HPD/Tax + 311/ECB/FDNY/CitiBike/NYSLA)" phase3_signals.py
run_phase 4  "StreetEasy Enrichment"               phase4_streeteasy.py
run_phase 5  "Cluster Detection"                    phase5_clusters.py
run_phase 6  "Data Verification"                    phase6_verify.py
run_phase 7  "XLSX Spreadsheet"                     phase7_xlsx.py
run_phase 8  "LaTeX PDF Report"                     phase9_pdf.py
run_phase 9  "Upload to Google Drive"               phase10_upload.py

echo ""
echo "━━━ Pipeline Complete ━━━"
