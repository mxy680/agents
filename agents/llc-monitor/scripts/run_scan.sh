#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/llc_monitor"
mkdir -p "$OUT_DIR"
DATE=$(date +%Y-%m-%d)
echo "━━━ Daily LLC Entity Monitor ━━━"
echo "Date: $DATE"
echo ""
python3 "$SCRIPT_DIR/scan_llcs.py" 2>&1
echo ""
echo "━━━ Scan Complete ━━━"
