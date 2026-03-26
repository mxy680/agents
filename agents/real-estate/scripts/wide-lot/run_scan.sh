#!/bin/bash
set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/wide_lot_scan"
mkdir -p "$OUT_DIR"
DATE=$(date +%Y-%m-%d)
echo "━━━ Bronx Wide Lot Scanner ━━━"
echo "Date: $DATE"
echo ""
python3 "$SCRIPT_DIR/scan_wide_lots.py" 2>&1
echo ""
echo "━━━ Scan Complete ━━━"
