#!/usr/bin/env python3
"""
Phase 10: Upload all deliverables to Google Drive
"""

import json
import subprocess
import sys
import os
import glob
from datetime import datetime

DATE = datetime.now().strftime("%Y-%m-%d")


def upload_file(path, name, convert=False):
    """Upload a file to Google Drive."""
    if not os.path.exists(path):
        print(f"  ✗ File not found: {path}", file=sys.stderr)
        return None

    cmd = [
        "integrations", "drive", "files", "upload",
        f"--path={path}",
        f"--name={name}",
        "--json"
    ]
    if convert:
        cmd.append("--convert")

    print(f"  Uploading {name}...", file=sys.stderr)
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)

    if result.returncode != 0:
        print(f"  ✗ Upload failed: {result.stderr[:200]}", file=sys.stderr)
        return None

    try:
        data = json.loads(result.stdout)
        file_id = data.get("id", "")
        print(f"  ✓ Uploaded: https://drive.google.com/file/d/{file_id}/view", file=sys.stderr)
        return data
    except json.JSONDecodeError:
        print(f"  Raw output: {result.stdout[:300]}", file=sys.stderr)
        return {"raw": result.stdout}


def main():
    links = {}

    # Find files dynamically
    xlsx_files = glob.glob("/tmp/nyc_assemblage/NYC_Assemblage_Scan_*.xlsx")
    pdf_files = glob.glob("/tmp/nyc_assemblage/report.pdf")

    # 1. Upload XLSX as Google Sheet
    if xlsx_files:
        xlsx_result = upload_file(
            xlsx_files[0],
            f"NYC Assemblage Scan — {DATE}",
            convert=True
        )
        if xlsx_result:
            file_id = xlsx_result.get("id", "")
            links["spreadsheet"] = f"https://docs.google.com/spreadsheets/d/{file_id}/edit"

    # 2. Upload PDF Report
    if pdf_files:
        pdf_result = upload_file(
            pdf_files[0],
            f"NYC Assemblage Report — {DATE}.pdf",
        )
        if pdf_result:
            file_id = pdf_result.get("id", "")
            links["report"] = f"https://drive.google.com/file/d/{file_id}/view"

    print("\n=== Upload Summary ===", file=sys.stderr)
    for k, v in links.items():
        print(f"  {k}: {v}", file=sys.stderr)

    print(json.dumps(links))


if __name__ == "__main__":
    main()
