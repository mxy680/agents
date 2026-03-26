#!/bin/bash
set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="/tmp/llc_monitor"
mkdir -p "$OUT_DIR"
DATE=$(date +%Y-%m-%d)
echo "━━━ Daily LLC Entity Monitor ━━━"
echo "Date: $DATE"
echo ""

# Check if already completed today
python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import is_phase_done
if is_phase_done('real-estate', 'llc-scan', 'scan'):
    print('[SKIP] Already completed today')
    sys.exit(42)
" 2>/dev/null
if [ $? -eq 42 ]; then
  echo "━━━ Scan Complete (skipped) ━━━"
  exit 0
fi

python3 "$SCRIPT_DIR/scan_llcs.py" 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_done
mark_phase_done('real-estate', 'llc-scan', 'scan')
" 2>/dev/null
else
  echo "[ERROR] Scan failed (exit code $EXIT_CODE)"
  python3 -c "
import sys; sys.path.insert(0, '$SCRIPT_DIR/../..')
from shared.checkpoint import mark_phase_failed
mark_phase_failed('real-estate', 'llc-scan', 'scan')
" 2>/dev/null
fi

echo ""
echo "━━━ Scan Complete ━━━"
exit $EXIT_CODE
