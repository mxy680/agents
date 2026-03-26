#!/usr/bin/env python3
"""
Daily LLC Entity Monitor
Scans NY DOS for new LLC formations that look like real estate transactions.
"""
import json
import re
import sys
import os
import subprocess
from datetime import datetime, timedelta

OUT_DIR = "/tmp/llc_monitor"
TODAY = datetime.now()
YESTERDAY = (TODAY - timedelta(days=1)).strftime("%Y-%m-%d")

# Real estate keywords in entity names
RE_KEYWORDS = [
    "REALTY", "REALITY",  # common misspelling
    "HOLDINGS", "HOLDING",
    "PROPERTIES", "PROPERTY",
    "DEVELOPMENT", "DEVELOPERS",
    "EQUITIES", "EQUITY",
    "CAPITAL",
    "ASSOCIATES",
    "MANAGEMENT", "MGMT",
    "VENTURES",
    "ACQUISITIONS",
    "INVESTORS", "INVESTMENT",
    "HOUSING",
    "RESIDENTIAL",
    "REAL ESTATE",
]

# NYC street suffixes
STREET_SUFFIXES = [
    "AVE", "AVENUE", "ST", "STREET", "BLVD", "BOULEVARD",
    "PL", "PLACE", "DR", "DRIVE", "RD", "ROAD",
    "CT", "COURT", "WAY", "LN", "LANE", "PKWY", "PARKWAY",
]

# NYC borough indicators
BOROUGH_INDICATORS = [
    "BRONX", "BX", "BROOKLYN", "BK", "MANHATTAN", "MN",
    "QUEENS", "QN", "STATEN", "SI",
]

# Pattern: starts with a number (address-style LLC)
ADDRESS_PATTERN = re.compile(r"^\d+\s+\w+", re.IGNORECASE)

# Pattern: contains a number + street suffix
STREET_PATTERN = re.compile(
    r"\d+\s+[\w\s]+\b(" + "|".join(STREET_SUFFIXES) + r")\b",
    re.IGNORECASE
)


def looks_like_real_estate(name):
    """Check if an entity name looks like a real estate entity."""
    upper = name.upper()

    # Check for real estate keywords
    for kw in RE_KEYWORDS:
        if kw in upper:
            return True, f"keyword: {kw}"

    # Check for address pattern (number + street name)
    if ADDRESS_PATTERN.match(name) and any(s in upper for s in STREET_SUFFIXES):
        return True, "address pattern"

    # Check for street pattern anywhere in name
    if STREET_PATTERN.search(name):
        return True, "street pattern"

    # Check for borough indicators combined with LLC
    if "LLC" in upper:
        for bi in BOROUGH_INDICATORS:
            if bi in upper:
                return True, f"borough: {bi}"

    return False, ""


def get_recent_llcs():
    """Pull new LLC formations from NY DOS."""
    print(f"Pulling new LLC formations since {YESTERDAY}...", file=sys.stderr)
    cmd = [
        "integrations", "nydos", "entities", "recent",
        f"--since={YESTERDAY}", "--type=llc", "--limit=5000", "--json"
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        if result.returncode != 0:
            print(f"  [WARN] CLI error: {result.stderr[:200]}", file=sys.stderr)
            return []
        data = json.loads(result.stdout)
        return data if isinstance(data, list) else []
    except Exception as e:
        print(f"  [WARN] Failed: {e}", file=sys.stderr)
        return []


def main():
    os.makedirs(OUT_DIR, exist_ok=True)

    # Step 1: Get recent LLCs
    llcs = get_recent_llcs()
    print(f"  Total new LLCs: {len(llcs)}", file=sys.stderr)

    # Step 2: Pattern match
    matches = []
    for entity in llcs:
        name = entity.get("name", "")
        is_re, reason = looks_like_real_estate(name)
        if is_re:
            entity["match_reason"] = reason
            matches.append(entity)

    print(f"  Real estate pattern matches: {len(matches)}", file=sys.stderr)

    # Step 3: Sort by relevance (address patterns first, then keywords)
    def sort_key(e):
        reason = e.get("match_reason", "")
        if "address" in reason or "street" in reason:
            return 0
        if "borough" in reason:
            return 1
        return 2

    matches.sort(key=sort_key)

    # Save results
    output_path = f"{OUT_DIR}/matches_{TODAY.strftime('%Y-%m-%d')}.json"
    with open(output_path, "w") as f:
        json.dump(matches, f, indent=2)

    # Summary
    print(f"\n=== LLC Monitor Results ===", file=sys.stderr)
    print(f"  New LLCs scanned: {len(llcs)}", file=sys.stderr)
    print(f"  Real estate matches: {len(matches)}", file=sys.stderr)

    # Breakdown by match type
    from collections import Counter
    reason_counts = Counter(m.get("match_reason", "") for m in matches)
    if reason_counts:
        print(f"\n  By match type:", file=sys.stderr)
        for reason, count in reason_counts.most_common():
            print(f"    {reason}: {count}", file=sys.stderr)

    if matches:
        print(f"\n  TOP MATCHES:", file=sys.stderr)
        for m in matches[:20]:
            name = m.get("name", "")
            filer = m.get("filer_name", "")
            addr = m.get("process_address", "") or m.get("filer_address", "")
            reason = m.get("match_reason", "")
            print(f"    {name} | Filer: {filer} | {addr} | ({reason})", file=sys.stderr)

    print(f"\nSaved {len(matches)} matches to {output_path}", file=sys.stderr)
    print(json.dumps({"total_llcs": len(llcs), "matches": len(matches), "path": output_path}))


if __name__ == "__main__":
    main()
