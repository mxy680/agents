#!/usr/bin/env python3
"""
Phase 10: Upload all deliverables to Google Drive
"""

import json
import subprocess
import sys
import os


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
        link = data.get("webViewLink") or data.get("link") or data.get("id", "")
        print(f"  ✓ Uploaded: {link}", file=sys.stderr)
        return data
    except json.JSONDecodeError:
        print(f"  Raw output: {result.stdout[:300]}", file=sys.stderr)
        return {"raw": result.stdout}


def main():
    links = {}

    # 1. Upload XLSX as Google Sheet
    xlsx_result = upload_file(
        "/tmp/nyc_assemblage/NYC_Assemblage_Scan_2026-03-24.xlsx",
        "NYC Assemblage Scan — 2026-03-24",
        convert=True  # Convert to Google Sheets
    )
    if xlsx_result:
        links["spreadsheet"] = xlsx_result.get("webViewLink") or xlsx_result.get("link") or str(xlsx_result)

    # 2. Upload Dashboard HTML
    dashboard_result = upload_file(
        "/tmp/nyc_assemblage/NYC_Assemblage_Dashboard_2026-03-24.html",
        "NYC Assemblage Dashboard — 2026-03-24.html",
    )
    if dashboard_result:
        links["dashboard"] = dashboard_result.get("webViewLink") or dashboard_result.get("link") or str(dashboard_result)

    # 3. Upload PDF Report
    pdf_result = upload_file(
        "/tmp/nyc_assemblage/report.pdf",
        "NYC Assemblage Report — 2026-03-24.pdf",
    )
    if pdf_result:
        links["report"] = pdf_result.get("webViewLink") or pdf_result.get("link") or str(pdf_result)

    print("\n=== Upload Summary ===", file=sys.stderr)
    for k, v in links.items():
        print(f"  {k}: {v}", file=sys.stderr)

    print(json.dumps(links))


if __name__ == "__main__":
    main()
