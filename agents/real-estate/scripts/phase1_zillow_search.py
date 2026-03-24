#!/usr/bin/env python3
"""
Phase 1: Zillow Search across all NYC zip codes
Searches every zip code in Bronx, Brooklyn, Manhattan, Queens
Deduplicates by ZPID, filters to 1-5 family homes only
"""

import json
import subprocess
import sys
import time
import os

# All target zip codes
BRONX_ZIPS = [
    "10451","10452","10453","10454","10455","10456","10457","10458","10459","10460",
    "10461","10462","10463","10464","10465","10466","10467","10468","10469","10470",
    "10471","10472","10473","10474","10475"
]

BROOKLYN_ZIPS = [
    "11201","11203","11204","11205","11206","11207","11208","11209","11210","11211",
    "11212","11213","11214","11215","11216","11217","11218","11219","11220","11221",
    "11222","11223","11224","11225","11226","11228","11229","11230","11231","11232",
    "11233","11234","11235","11236","11237","11238","11239"
]

MANHATTAN_ZIPS = [
    "10001","10002","10003","10009","10010","10011","10012","10013","10014","10016",
    "10019","10021","10022","10023","10024","10025","10026","10027","10028","10029",
    "10030","10031","10032","10033","10034","10035","10037","10039","10040","10128"
]

QUEENS_ZIPS = [
    "11101","11102","11103","11104","11105","11106","11109","11354","11355","11356",
    "11357","11358","11359","11360","11361","11362","11363","11364","11365","11366",
    "11367","11368","11369","11370","11372","11373","11374","11375","11377","11378",
    "11379","11385","11411","11412","11413","11414","11415","11416","11417","11418",
    "11419","11420","11421","11422","11423"
]

BOROUGH_ZIPS = {
    "Bronx": ("Bronx", BRONX_ZIPS),
    "Brooklyn": ("Brooklyn", BROOKLYN_ZIPS),
    "Manhattan": ("Manhattan", MANHATTAN_ZIPS),
    "Queens": ("Queens", QUEENS_ZIPS),
}

# Acceptable homeTypes for 1-5 family residential
ACCEPTABLE_HOME_TYPES = {
    "SINGLE_FAMILY", "MULTI_FAMILY", "TOWNHOUSE"
}

# Status strings that indicate NOT active
EXCLUDE_STATUS = {
    "pending", "contingent", "under contract", "accepting backup offers"
}


def search_zip(borough_label, zip_code, min_price=None, max_price=None):
    """Run Zillow search for a single zip code."""
    cmd = [
        "integrations", "zillow", "properties", "search",
        f"--location={borough_label}, NY {zip_code}",
        "--limit=40",
        "--json"
    ]
    if min_price:
        cmd.append(f"--min-price={min_price}")
    if max_price:
        cmd.append(f"--max-price={max_price}")

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        if result.returncode != 0:
            print(f"  [WARN] CLI error for {zip_code}: {result.stderr[:200]}", file=sys.stderr)
            return []
        data = json.loads(result.stdout)
        if isinstance(data, list):
            return data
        return []
    except subprocess.TimeoutExpired:
        print(f"  [WARN] Timeout for {zip_code}", file=sys.stderr)
        return []
    except json.JSONDecodeError:
        print(f"  [WARN] JSON parse error for {zip_code}: {result.stdout[:200]}", file=sys.stderr)
        return []


def is_qualifying(prop):
    """Filter to 1-5 family homes that are active for sale."""
    home_type = prop.get("homeType", "").upper()
    status = prop.get("status", "").lower()
    address = prop.get("address", "").upper()

    # Must be acceptable home type
    if home_type not in ACCEPTABLE_HOME_TYPES:
        return False

    # Exclude pending/contingent
    for excl in EXCLUDE_STATUS:
        if excl in status:
            return False

    # Exclude condo/co-op indicators in address (APT, #, UNIT)
    # These are often mis-classified single units
    # We'll be more permissive here and let PLUTO filter later
    # But exclude obvious condo/coop addresses
    if " APT " in address or "APT#" in address:
        # Could be a multi-family unit - still let through for PLUTO check
        # PLUTO will reveal the true building class
        pass

    return True


def search_borough(borough_key):
    """Search all zips in a borough, handling 40-result truncation."""
    borough_label, zips = BOROUGH_ZIPS[borough_key]
    all_results = {}  # zpid -> prop
    total_zips = len(zips)

    for i, zip_code in enumerate(zips, 1):
        print(f"  [{i}/{total_zips}] {borough_key} {zip_code}...", file=sys.stderr, end=" ")
        results = search_zip(borough_label, zip_code)
        count = len(results)

        # Add to collection
        for prop in results:
            zpid = prop.get("zpid")
            if zpid:
                prop["_zip"] = zip_code
                prop["_borough"] = borough_key
                all_results[zpid] = prop

        print(f"{count} results", file=sys.stderr, end="")

        # If we hit the 40-result cap, split by price
        if count >= 40:
            print(f" [HIT CAP — splitting by price]", file=sys.stderr)
            price_ranges = [
                (None, 500000),
                (500001, 800000),
                (800001, 1200000),
                (1200001, None),
            ]
            for min_p, max_p in price_ranges:
                sub_results = search_zip(borough_label, zip_code, min_p, max_p)
                sub_count = len(sub_results)

                # If still at cap, split by home type too
                if sub_count >= 40:
                    for ht in ["house", "multi_family"]:
                        cmd = [
                            "integrations", "zillow", "properties", "search",
                            f"--location={borough_label}, NY {zip_code}",
                            f"--home-type={ht}",
                            "--limit=40",
                            "--json"
                        ]
                        if min_p:
                            cmd.append(f"--min-price={min_p}")
                        if max_p:
                            cmd.append(f"--max-price={max_p}")
                        try:
                            r = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
                            ht_results = json.loads(r.stdout) if r.returncode == 0 else []
                            for prop in ht_results:
                                zpid = prop.get("zpid")
                                if zpid:
                                    prop["_zip"] = zip_code
                                    prop["_borough"] = borough_key
                                    all_results[zpid] = prop
                        except Exception:
                            pass
                else:
                    for prop in sub_results:
                        zpid = prop.get("zpid")
                        if zpid:
                            prop["_zip"] = zip_code
                            prop["_borough"] = borough_key
                            all_results[zpid] = prop

                time.sleep(0.5)
        else:
            print("", file=sys.stderr)

        # Rate limiting: small delay every 5 zips
        if i % 5 == 0:
            time.sleep(1)
        else:
            time.sleep(0.3)

    return list(all_results.values())


def main():
    os.makedirs("/tmp/nyc_assemblage", exist_ok=True)

    all_properties = {}  # zpid -> prop (deduped across boroughs)

    for borough_key in ["Bronx", "Brooklyn", "Manhattan", "Queens"]:
        print(f"\n=== Searching {borough_key} ===", file=sys.stderr)
        results = search_borough(borough_key)

        # Deduplicate
        new_count = 0
        for prop in results:
            zpid = prop.get("zpid")
            if zpid and zpid not in all_properties:
                all_properties[zpid] = prop
                new_count += 1

        print(f"  {borough_key}: {len(results)} raw results, {new_count} new unique", file=sys.stderr)

    print(f"\n=== Total raw results: {len(all_properties)} ===", file=sys.stderr)

    # Filter to qualifying home types
    qualifying = []
    excluded = {"condo": 0, "coop": 0, "pending": 0, "other": 0}

    for zpid, prop in all_properties.items():
        home_type = prop.get("homeType", "").upper()
        status = prop.get("status", "").lower()

        # Exclude condos/coops
        if "CONDO" in home_type or "APARTMENT" in home_type or "COOP" in home_type:
            excluded["condo"] += 1
            continue
        if "condo" in status or "co-op" in status or "coop" in status:
            excluded["coop"] += 1
            continue

        # Exclude pending/contingent
        is_pending = any(s in status for s in ["pending", "contingent", "under contract", "backup"])
        if is_pending:
            excluded["pending"] += 1
            continue

        if home_type in ACCEPTABLE_HOME_TYPES:
            qualifying.append(prop)
        else:
            excluded["other"] += 1

    print(f"\n=== After filtering ===", file=sys.stderr)
    print(f"  Qualifying: {len(qualifying)}", file=sys.stderr)
    print(f"  Excluded condos/coops: {excluded['condo'] + excluded['coop']}", file=sys.stderr)
    print(f"  Excluded pending: {excluded['pending']}", file=sys.stderr)
    print(f"  Excluded other types: {excluded['other']}", file=sys.stderr)

    # Borough breakdown
    from collections import Counter
    borough_counts = Counter(p.get("_borough", "Unknown") for p in qualifying)
    for b, c in sorted(borough_counts.items()):
        print(f"  {b}: {c}", file=sys.stderr)

    # Save results
    output_path = "/tmp/nyc_assemblage/zillow_results.json"
    with open(output_path, "w") as f:
        json.dump(qualifying, f, indent=2)

    print(f"\n✓ Saved {len(qualifying)} qualifying listings to {output_path}", file=sys.stderr)
    print(json.dumps({"count": len(qualifying), "path": output_path}))


if __name__ == "__main__":
    main()
