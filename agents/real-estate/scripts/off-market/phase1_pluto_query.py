#!/usr/bin/env python3
"""
Phase 1: Query PLUTO for all R8+ zoned small residential properties.
No Zillow dependency — starts directly from city property records.
"""
import json
import sys
import os
import time
import urllib.request
import urllib.parse

# Add shared module to path
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".."))
try:
    from shared.cache import get_cached, put_cached, put_cached_batch
except ImportError:
    def get_cached(p, e): return None
    def put_cached(p, e, d, **kw): return False
    def put_cached_batch(p, items): return False

OUT_DIR = "/tmp/off_market_scan"

R8_PLUS_ZONES = [
    "R8", "R8A", "R8B", "R8X",
    "R9", "R9A", "R9X",
    "R10", "R10A", "R10X",
]

# Also include C4-4+ and MX zones
EXTRA_ZONES = ["C4-4", "C4-5", "C4-6", "C4-7"]

# Building classes for 1-5 family residential
# A = One Family Dwellings, B = Two Family Dwellings, C = Walk-up Apartments (3-6 units), S = Residence (Multiple Use)
BLDG_CLASS_PREFIXES = ["A", "B", "C", "S"]

BOROUGH_CODES = {"1": "Manhattan", "2": "Bronx", "3": "Brooklyn", "4": "Queens", "5": "Staten Island"}


def query_pluto(where_clause, offset=0, limit=5000):
    """Query PLUTO via Socrata API."""
    params = urllib.parse.urlencode({
        "$where": where_clause,
        "$select": "bbl,address,zonedist1,lotarea,bldgarea,yearbuilt,bldgclass,numfloors,unitsres,unitstotal,borocode,block,lot,lotfront,lotdepth",
        "$limit": str(limit),
        "$offset": str(offset),
        "$order": "bbl",
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


def build_where_clause():
    """Build the SoQL WHERE clause for R8+ small residential."""
    # Zone filter
    zone_list = R8_PLUS_ZONES + EXTRA_ZONES
    zone_sql = ",".join(f"'{z}'" for z in zone_list)

    # Building class filter (starts with A, B, C, or S)
    bldg_filters = " OR ".join(f"bldgclass like '{p}%'" for p in BLDG_CLASS_PREFIXES)

    # Combine: R8+ zone AND small residential AND 5 or fewer units
    return f"zonedist1 in({zone_sql}) AND ({bldg_filters}) AND unitsres <= '5'"


def main():
    os.makedirs(OUT_DIR, exist_ok=True)

    where = build_where_clause()
    print(f"Querying PLUTO for R8+ small residential properties...", file=sys.stderr)
    print(f"  WHERE: {where[:100]}...", file=sys.stderr)

    all_properties = []
    offset = 0
    page_size = 5000

    while True:
        print(f"  Fetching offset {offset}...", file=sys.stderr, end=" ")
        rows = query_pluto(where, offset, page_size)
        print(f"{len(rows)} rows", file=sys.stderr)

        if not rows:
            break

        for row in rows:
            bbl = row.get("bbl", "")
            boro_code = row.get("borocode", "")
            borough = BOROUGH_CODES.get(str(boro_code), "Unknown")
            block = row.get("block", "")
            lot = row.get("lot", "")

            prop = {
                "_bbl": bbl,
                "_borough": borough,
                "_borough_digit": str(boro_code),
                "_block": str(block).zfill(5),
                "_lot": str(lot).zfill(4),
                "address": row.get("address", ""),
                "_zoning": row.get("zonedist1", ""),
                "_lot_area": row.get("lotarea", ""),
                "_bldg_area": row.get("bldgarea", ""),
                "_year_built": row.get("yearbuilt", ""),
                "_bldg_class": row.get("bldgclass", ""),
                "_num_floors": row.get("numfloors", ""),
                "_units_res": row.get("unitsres", ""),
                "_units_total": row.get("unitstotal", ""),
                "_lot_front": row.get("lotfront", ""),
                "_lot_depth": row.get("lotdepth", ""),
                "_zola_url": f"https://zola.planning.nyc.gov/lot/{boro_code}/{str(block).zfill(5)}/{str(lot).zfill(4)}",
            }
            all_properties.append(prop)

        # Cache PLUTO data for each row (annual data, cache forever)
        cache_items = []
        for row in rows:
            bbl = row.get("bbl", "")
            if bbl:
                cache_items.append((bbl, row, bbl))
        if cache_items:
            put_cached_batch("pluto", cache_items)

        if len(rows) < page_size:
            break
        offset += page_size
        time.sleep(0.5)

    # Deduplicate by BBL
    seen = {}
    for p in all_properties:
        bbl = p.get("_bbl", "")
        if bbl and bbl not in seen:
            seen[bbl] = p
    properties = list(seen.values())

    # Borough breakdown
    from collections import Counter
    borough_counts = Counter(p.get("_borough", "Unknown") for p in properties)
    zone_counts = Counter(p.get("_zoning", "Unknown") for p in properties)

    print(f"\n=== PLUTO Results ===", file=sys.stderr)
    print(f"  Total R8+ small residential: {len(properties)}", file=sys.stderr)
    print(f"\n  By borough:", file=sys.stderr)
    for b, c in sorted(borough_counts.items()):
        print(f"    {b}: {c}", file=sys.stderr)
    print(f"\n  Top zones:", file=sys.stderr)
    for z, c in zone_counts.most_common(10):
        print(f"    {z}: {c}", file=sys.stderr)

    # Save
    output_path = f"{OUT_DIR}/r8plus_properties.json"
    with open(output_path, "w") as f:
        json.dump(properties, f, indent=2)

    print(f"\n✓ Saved {len(properties)} properties to {output_path}", file=sys.stderr)
    print(json.dumps({"count": len(properties), "path": output_path}))


if __name__ == "__main__":
    main()
