#!/usr/bin/env python3
"""
Phase 7: Upload deliverables to Google Drive (Off-Market R8+ Properties)

Uploads:
  - XLSX as Google Sheet: "Off-Market R8+ Scan — {DATE}"
  - PDF report:           "Off-Market R8+ Report — {DATE}.pdf"

Input:  /tmp/off_market_scan/Off_Market_R8plus_Scan_*.xlsx
        /tmp/off_market_scan/report.pdf
"""

import json
import subprocess
import sys
import os
import glob
from datetime import datetime

DATE = datetime.now().strftime("%Y-%m-%d")


def upload_file(path: str, name: str, convert: bool = False) -> dict | None:
    """Upload a file to Google Drive."""
    if not os.path.exists(path):
        print(f"  File not found: {path}", file=sys.stderr)
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
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
    except subprocess.TimeoutExpired:
        print(f"  Upload timed out after 120s: {name}", file=sys.stderr)
        return None

    if result.returncode != 0:
        print(f"  Upload failed: {result.stderr[:200]}", file=sys.stderr)
        return None

    try:
        data = json.loads(result.stdout)
        file_id = data.get("id", "")
        print(f"  Uploaded: https://drive.google.com/file/d/{file_id}/view", file=sys.stderr)
        return data
    except json.JSONDecodeError:
        print(f"  Raw output: {result.stdout[:300]}", file=sys.stderr)
        return {"raw": result.stdout}


def main():
    links: dict = {}

    xlsx_files = glob.glob("/tmp/off_market_scan/Off_Market_R8plus_Scan_*.xlsx")
    pdf_files = glob.glob("/tmp/off_market_scan/report.pdf")

    # 1. Upload XLSX as Google Sheet
    if xlsx_files:
        xlsx_result = upload_file(
            xlsx_files[0],
            f"Off-Market R8+ Scan \u2014 {DATE}",
            convert=True
        )
        if xlsx_result:
            file_id = xlsx_result.get("id", "")
            links["spreadsheet"] = f"https://docs.google.com/spreadsheets/d/{file_id}/edit"

    # 2. Upload PDF Report
    if pdf_files:
        pdf_result = upload_file(
            pdf_files[0],
            f"Off-Market R8+ Report \u2014 {DATE}.pdf",
        )
        if pdf_result:
            file_id = pdf_result.get("id", "")
            links["report"] = f"https://drive.google.com/file/d/{file_id}/view"

    print("\n=== Upload Summary ===", file=sys.stderr)
    for k, v in links.items():
        print(f"  {k}: {v}", file=sys.stderr)

    files_expected = bool(xlsx_files or pdf_files)
    if files_expected and not links:
        print("  ERROR: Files were found but all uploads failed.", file=sys.stderr)
        print(json.dumps(links))
        sys.exit(1)

    print(json.dumps(links))


if __name__ == "__main__":
    main()
