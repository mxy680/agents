#!/usr/bin/env python3
"""
Bronx Wide Lot Scanner
Finds 1-2 family homes with 50ft+ lot frontage in R8+ (or R7+) zones.
"""
import json
import sys
import os
import urllib.request
import urllib.parse

OUT_DIR = "/tmp/wide_lot_scan"

BOROUGH_CODE = "2"  # Bronx

R8_PLUS = "zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%'"
R7_PLUS = "zonedist1 like 'R7%' OR zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%'"


def query_pluto(zoning_filter):
    """Query PLUTO for wide-lot 1-2 family homes."""
    where = (
        f"borocode='{BOROUGH_CODE}' AND lotfront >= 50 AND unitsres <= 2 "
        f"AND (bldgclass like 'A%' OR bldgclass like 'B%') "
        f"AND ({zoning_filter})"
    )
    params = urllib.parse.urlencode({
        "$where": where,
        "$select": "bbl,address,zonedist1,lotarea,lotfront,lotdepth,bldgarea,yearbuilt,bldgclass,unitsres,numfloors",
        "$limit": "500",
        "$order": "lotfront DESC",
    })
    url = f"https://data.cityofnewyork.us/resource/64uk-42ks.json?{params}"
    req = urllib.request.Request(url, headers={"User-Agent": "Emdash-Agents/1.0"})
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            data = json.loads(resp.read())
        return data if isinstance(data, list) else []
    except Exception as e:
        print(f"  [WARN] PLUTO query failed: {e}", file=sys.stderr)
        return []


def main():
    os.makedirs(OUT_DIR, exist_ok=True)

    # Start with R8+
    print("Querying PLUTO for R8+ wide lots (50ft+) in Bronx...", file=sys.stderr)
    r8_results = query_pluto(R8_PLUS)
    print(f"  R8+ results: {len(r8_results)}", file=sys.stderr)

    # If fewer than 20, expand to R7+
    if len(r8_results) < 20:
        print(f"  Few R8+ hits ({len(r8_results)}), expanding to R7+...", file=sys.stderr)
        r7_results = query_pluto(R7_PLUS)
        print(f"  R7+ results: {len(r7_results)}", file=sys.stderr)
        results = r7_results
        zoning_level = "R7+"
    else:
        results = r8_results
        zoning_level = "R8+"

    # Deduplicate by BBL
    seen = {}
    for r in results:
        bbl = r.get("bbl", "")
        if bbl and bbl not in seen:
            seen[bbl] = r
    properties = list(seen.values())

    # Build ZoLa URLs
    for p in properties:
        bbl = str(p.get("bbl", "")).zfill(10)
        boro = bbl[0]
        block = bbl[1:6]
        lot = bbl[6:10]
        p["_zola_url"] = f"https://zola.planning.nyc.gov/lot/{boro}/{block}/{lot}"

    # Zone breakdown
    from collections import Counter
    zone_counts = Counter(p.get("zonedist1", "?") for p in properties)

    print(f"\n=== Wide Lot Scan Results ===", file=sys.stderr)
    print(f"  Zoning level: {zoning_level}", file=sys.stderr)
    print(f"  Total qualifying properties: {len(properties)}", file=sys.stderr)
    print(f"\n  By zone:", file=sys.stderr)
    for z, c in zone_counts.most_common():
        print(f"    {z}: {c}", file=sys.stderr)

    if properties:
        print(f"\n  Top 5 widest lots:", file=sys.stderr)
        for p in properties[:5]:
            addr = p.get("address", "?")
            zone = p.get("zonedist1", "?")
            front = p.get("lotfront", "?")
            area = p.get("lotarea", "?")
            yr = p.get("yearbuilt", "?")
            print(f"    {addr} | {zone} | {front}ft front | {area} SF | built {yr}", file=sys.stderr)

    # Save
    output_path = f"{OUT_DIR}/wide_lots.json"
    with open(output_path, "w") as f:
        json.dump(properties, f, indent=2)

    print(f"\n✓ Saved {len(properties)} properties to {output_path}", file=sys.stderr)
    print(json.dumps({
        "count": len(properties),
        "zoning_level": zoning_level,
        "r8_count": len(r8_results),
        "path": output_path,
    }))


if __name__ == "__main__":
    main()
