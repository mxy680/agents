#!/usr/bin/env python3
"""
Phase 4: Data Verification (Off-Market R8+ Properties)
Check for:
- Missing/invalid BBLs
- Duplicate addresses
- Inconsistent scoring

Input:  /tmp/off_market_scan/final_properties.json
Output: /tmp/off_market_scan/verified_properties.json
"""

import json
import sys
from collections import Counter


def main():
    with open("/tmp/off_market_scan/final_properties.json") as f:
        properties = json.load(f)

    print(f"Verifying {len(properties)} properties...", file=sys.stderr)
    issues = []
    fixed = 0

    # Check 1: Duplicate addresses
    address_counts = Counter(p.get("address", "").lower().strip() for p in properties)
    duplicates = {addr: count for addr, count in address_counts.items() if count > 1}
    if duplicates:
        print(f"  [WARN] {len(duplicates)} duplicate addresses found", file=sys.stderr)
        seen_addresses: dict = {}
        to_remove: set = set()
        for i, p in enumerate(properties):
            addr = p.get("address", "").lower().strip()
            bbl = p.get("_bbl", "")
            if addr in seen_addresses:
                # Keep the one we already have (it's already higher scored due to sort)
                to_remove.add(i)
                issues.append(f"Duplicate: {p.get('address')} (bbl {bbl})")
            else:
                seen_addresses[addr] = i
        properties = [p for i, p in enumerate(properties) if i not in to_remove]
        fixed += len(to_remove)
        print(f"  [FIXED] Removed {len(to_remove)} duplicates", file=sys.stderr)

    # Check 2: Duplicate BBLs
    bbl_counts = Counter(p.get("_bbl", "") for p in properties)
    dup_bbls = {b: c for b, c in bbl_counts.items() if c > 1 and b}
    if dup_bbls:
        print(f"  [WARN] {len(dup_bbls)} duplicate BBLs", file=sys.stderr)
        seen_bbls: set = set()
        to_remove_idx: set = set()
        for i, p in enumerate(properties):
            bbl = p.get("_bbl", "")
            if bbl in seen_bbls:
                to_remove_idx.add(i)
            else:
                seen_bbls.add(bbl)
        properties = [p for i, p in enumerate(properties) if i not in to_remove_idx]
        fixed += len(to_remove_idx)

    # Check 3: Missing BBLs
    missing_bbl = [p for p in properties if not p.get("_bbl")]
    if missing_bbl:
        print(f"  [INFO] {len(missing_bbl)} properties with missing BBL (kept but flagged)", file=sys.stderr)
        for p in missing_bbl:
            p["_data_quality_note"] = "BBL not geocoded"

    # Check 4: Inconsistent scoring
    # Any property with lis pendens + tax lien should score >= 12
    for p in properties:
        score = p.get("_score", 0)
        has_lis = p.get("_acris_lis_pendens", False)
        has_tax = p.get("_tax_lien", False)
        has_estate = p.get("_acris_estate", False)

        if has_lis and has_tax and score < 12:
            print(f"  [WARN] Potential underscore: {p.get('address')} score={score} but has lis pendens + tax lien", file=sys.stderr)
            issues.append(f"Potential underscore: {p.get('address')}")

        if has_estate and score < 10:
            print(f"  [WARN] Potential underscore: {p.get('address')} score={score} but has estate signal", file=sys.stderr)

    # Check 5: BBL format validation
    for p in properties:
        bbl = p.get("_bbl", "")
        if bbl:
            bbl_str = str(bbl).zfill(10)
            if len(bbl_str) != 10:
                issues.append(f"Invalid BBL length: {bbl} for {p.get('address')}")
            borough_digit = bbl_str[0]
            expected_digit = {
                "Manhattan": "1", "Bronx": "2", "Brooklyn": "3",
                "Queens": "4", "Staten Island": "5"
            }.get(p.get("_borough", ""), "")
            if expected_digit and borough_digit != expected_digit:
                issues.append(f"BBL borough mismatch: {bbl} borough={p.get('_borough')} for {p.get('address')}")

    # Final re-sort
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    # Summary
    priority_counts = Counter(p.get("_priority", "Watchlist") for p in properties)

    print(f"\n=== Verification Complete ===", file=sys.stderr)
    print(f"  Issues found: {len(issues)}", file=sys.stderr)
    print(f"  Records fixed/removed: {fixed}", file=sys.stderr)
    print(f"  Final property count: {len(properties)}", file=sys.stderr)
    print(f"\n  Priority distribution:", file=sys.stderr)
    for tier in ["Immediate", "High", "Moderate", "Watchlist"]:
        print(f"    {tier}: {priority_counts.get(tier, 0)}", file=sys.stderr)

    print(f"\n  Top 5 properties after verification:", file=sys.stderr)
    for p in properties[:5]:
        print(f"    Score {p.get('_score', 0)} | {p.get('address', 'N/A')[:60]} | {p.get('_zoning', 'N/A')}", file=sys.stderr)

    # Save verified data
    with open("/tmp/off_market_scan/verified_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

    if issues:
        with open("/tmp/off_market_scan/verification_issues.json", "w") as f:
            json.dump(issues, f, indent=2)

    print(f"\nVerification complete. {len(properties)} properties ready for output.", file=sys.stderr)
    print(json.dumps({"count": len(properties), "issues": len(issues), "fixed": fixed}))


if __name__ == "__main__":
    main()
