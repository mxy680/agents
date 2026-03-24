#!/usr/bin/env python3
"""
Phase 3: Pre-Market Signal Checks
For each R7+ property, check:
- ACRIS: judgments, federal liens, tax lien sale certificates, LLC deed transfers on block
- DOB: demolition/new building permits on block (last 6 months)
- HPD: open violations count
- NYC Finance: tax lien list
"""

import json
import subprocess
import sys
import time
import urllib.request
import urllib.parse
from datetime import datetime, timedelta

TODAY = datetime.now()
DATE_90_DAYS_AGO = (TODAY - timedelta(days=90)).strftime("%Y-%m-%d")
DATE_180_DAYS_AGO = (TODAY - timedelta(days=180)).strftime("%Y-%m-%d")


def curl_socrata(url_base, params):
    """Make a Socrata API call using subprocess curl with proper encoding."""
    cmd = ["curl", "-s", "-G", url_base]
    for key, val in params.items():
        cmd.extend(["--data-urlencode", f"{key}={val}"])
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=20)
        if result.returncode != 0:
            return []
        parsed = json.loads(result.stdout)
        # Socrata may return a dict on error instead of a list
        return parsed if isinstance(parsed, list) else []
    except Exception:
        return []


def get_acris_docs(borough_digit, block, lot):
    """Get ACRIS document IDs for a lot."""
    # Zero-pad block to 5 digits, lot to 4 digits
    block_padded = str(block).zfill(5)
    lot_padded = str(lot).zfill(4)

    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/8h5j-fqxa.json",
        {
            "borough": str(borough_digit),
            "block": block_padded,
            "lot": lot_padded,
            "$limit": "500"
        }
    )
    doc_ids = [r.get("document_id") for r in records if r.get("document_id")]
    return doc_ids


def get_acris_masters(doc_ids):
    """Get document details from ACRIS Master table."""
    if not doc_ids:
        return []
    # Batch in groups of 50
    all_docs = []
    for i in range(0, len(doc_ids), 50):
        batch = doc_ids[i:i+50]
        ids_str = ",".join(f"'{d}'" for d in batch)
        records = curl_socrata(
            "https://data.cityofnewyork.us/resource/bnx9-e6tj.json",
            {"$where": f"document_id in({ids_str})"}
        )
        all_docs.extend(records)
    return all_docs


def get_acris_parties(doc_ids, party_type="2"):
    """Get parties for ACRIS documents (1=grantor/seller, 2=grantee/buyer)."""
    if not doc_ids:
        return []
    all_parties = []
    for i in range(0, min(len(doc_ids), 100), 50):
        batch = doc_ids[i:i+50]
        ids_str = ",".join(f"'{d}'" for d in batch)
        records = curl_socrata(
            "https://data.cityofnewyork.us/resource/636b-3b5g.json",
            {"$where": f"document_id in({ids_str}) AND party_type='{party_type}'"}
        )
        all_parties.extend(records)
    return all_parties


def get_acris_block_deeds(borough_digit, block):
    """Check for LLC deed transfers on the same block (last 12 months)."""
    block_padded = str(block).zfill(5)
    date_12mo = (TODAY - timedelta(days=365)).strftime("%Y-%m-%d")

    # Get all doc IDs for the block
    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/8h5j-fqxa.json",
        {
            "borough": str(borough_digit),
            "block": block_padded,
            "$limit": "500"
        }
    )
    doc_ids = [r.get("document_id") for r in records if r.get("document_id")]

    if not doc_ids:
        return False

    # Get masters, filter to recent DEEDs
    masters = get_acris_masters(doc_ids[:100])  # limit to 100 docs
    deed_ids = [
        m.get("document_id") for m in masters
        if m.get("doc_type") == "DEED"
        and m.get("document_date", "") >= date_12mo
    ]

    if not deed_ids:
        return False

    # Check if any buyer is an LLC
    parties = get_acris_parties(deed_ids[:50])
    for party in parties:
        name = party.get("name", "").upper()
        if "LLC" in name or "L.L.C" in name:
            return True

    return False


def check_acris_property(borough_digit, block, lot):
    """Check ACRIS for distress signals on a specific property."""
    doc_ids = get_acris_docs(borough_digit, block, lot)

    if not doc_ids:
        return {
            "acris_lis_pendens": False,
            "acris_lis_pendens_recent": False,
            "acris_federal_lien": False,
            "acris_tax_lien_cert": False,
            "acris_estate": False,
            "acris_doc_count": 0
        }

    masters = get_acris_masters(doc_ids)

    signals = {
        "acris_lis_pendens": False,
        "acris_lis_pendens_recent": False,
        "acris_federal_lien": False,
        "acris_tax_lien_cert": False,
        "acris_estate": False,
        "acris_doc_count": len(doc_ids)
    }

    for doc in masters:
        doc_type = doc.get("doc_type", "")
        doc_date = doc.get("document_date", "")

        if doc_type == "JUDG":
            signals["acris_lis_pendens"] = True
            if doc_date >= DATE_90_DAYS_AGO:
                signals["acris_lis_pendens_recent"] = True

        elif doc_type == "FL":
            signals["acris_federal_lien"] = True

        elif doc_type == "TLS":
            signals["acris_tax_lien_cert"] = True

    # Check for estate/probate signals in party names
    if doc_ids:
        estate_ids = doc_ids[:50]  # Check most recent 50 docs
        ids_str = ",".join(f"'{d}'" for d in estate_ids)
        estate_records = curl_socrata(
            "https://data.cityofnewyork.us/resource/636b-3b5g.json",
            {"$where": f"document_id in({ids_str}) AND (name like '%ESTATE OF%' OR name like '%EXECUTOR%' OR name like '%EXECUTRIX%')"}
        )
        if estate_records:
            signals["acris_estate"] = True

    return signals


def check_dob_block(borough_name, block):
    """Check for DM/NB permits on the same block in last 6 months."""
    block_padded = str(block).zfill(5)
    borough_upper = borough_name.upper()

    # DOB uses BRONX, BROOKLYN, MANHATTAN, QUEENS
    borough_map = {
        "BRONX": "BRONX",
        "BROOKLYN": "BROOKLYN",
        "MANHATTAN": "MANHATTAN",
        "QUEENS": "QUEENS",
    }
    dob_borough = borough_map.get(borough_upper, borough_upper)

    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/ipu4-2q9a.json",
        {
            "$where": f"borough='{dob_borough}' AND block='{block_padded}' AND job_type in('DM','NB') AND issuance_date > '{DATE_180_DAYS_AGO}'",
            "$limit": "50"
        }
    )

    has_demo = any(r.get("job_type") == "DM" for r in records)
    has_nb = any(r.get("job_type") == "NB" for r in records)

    return has_demo, has_nb


def check_hpd_violations(borough_digit, block, lot):
    """Count open HPD violations for a property."""
    block_padded = str(block).zfill(5).lstrip("0") or "0"
    lot_padded = str(lot).zfill(4).lstrip("0") or "0"

    # HPD uses boroid (not borough name), and unpadded block/lot
    results = curl_socrata(
        "https://data.cityofnewyork.us/resource/csn4-vhvf.json",
        {
            "$select": "count(*)",
            "$where": f"boroid='{borough_digit}' AND block='{block_padded}' AND lot='{lot_padded}'"
        }
    )

    if results and isinstance(results, list) and len(results) > 0:
        count_val = results[0].get("count", 0)
        try:
            return int(count_val)
        except (ValueError, TypeError):
            return 0
    return 0


def check_nyc_finance_lien(bbl):
    """Check NYC Finance tax lien list."""
    results = curl_socrata(
        "https://data.cityofnewyork.us/resource/9rz4-mjek.json",
        {"$where": f"bbl='{bbl}'"}
    )
    return len(results) > 0


def compute_score(prop):
    """Compute composite score for a property."""
    score = 0
    reasons = []

    # Zoning: R8+ gets bonus
    zone = prop.get("_zoning", "")
    if zone:
        z_upper = zone.upper()
        if any(z_upper.startswith(p) for p in ["R8", "R9", "R10"]):
            score += 3
            reasons.append("R8+ zoning (+3)")

    # Pre-war construction
    try:
        year_built = int(float(prop.get("_year_built", 0)))
        if 1800 < year_built < 1945:
            score += 2
            reasons.append(f"Pre-war ({year_built}) (+2)")
    except (ValueError, TypeError):
        pass

    # Small lot
    try:
        lot_area = int(float(prop.get("_lot_area", 0)))
        if 0 < lot_area < 2000:
            score += 2
            reasons.append(f"Small lot ({lot_area} SF) (+2)")
    except (ValueError, TypeError):
        pass

    # Active listing (we already filtered to active, so all get this)
    score += 3
    reasons.append("Active for-sale listing (+3)")

    # Days on market
    dom = prop.get("daysOnMarket", 0) or 0
    if dom > 180:
        score += 2
        reasons.append(f"DOM {dom} > 180 (+2)")

    # Tax lien
    if prop.get("_tax_lien"):
        score += 4
        reasons.append("Tax lien (+4)")

    # ACRIS signals
    if prop.get("_acris_lis_pendens_recent"):
        score += 5
        reasons.append("Lis pendens < 90 days (+5)")
    elif prop.get("_acris_lis_pendens"):
        score += 2
        reasons.append("Lis pendens (older) (+2)")

    if prop.get("_acris_federal_lien"):
        score += 3
        reasons.append("Federal/IRS lien (+3)")

    if prop.get("_acris_tax_lien_cert"):
        score += 3
        reasons.append("ACRIS tax lien certificate (+3)")

    if prop.get("_acris_estate"):
        score += 5
        reasons.append("Estate/probate signal (+5)")

    # Block signals
    if prop.get("_block_llc_deed"):
        score += 3
        reasons.append("LLC deed on block (+3)")

    if prop.get("_block_demo"):
        score += 3
        reasons.append("Demo permit on block (+3)")

    if prop.get("_block_new_building"):
        score += 2
        reasons.append("NB permit on block (+2)")

    # HPD violations
    hpd_count = prop.get("_hpd_violations", 0) or 0
    if hpd_count >= 10:
        score += 4
        reasons.append(f"HPD violations {hpd_count} (severe) (+4)")
    elif hpd_count >= 5:
        score += 2
        reasons.append(f"HPD violations {hpd_count} (+2)")

    # Price penalty — luxury properties are not realistic assemblage targets
    price = prop.get("price", 0) or 0
    if price > 5_000_000:
        score -= 5
        reasons.append(f"Price penalty >$5M (-5)")
    elif price > 3_000_000:
        score -= 3
        reasons.append(f"Price penalty >$3M (-3)")

    # Price bonus — affordable properties are better starter lots
    if 0 < price <= 800_000:
        score += 2
        reasons.append(f"Affordable price under $800K (+2)")

    prop["_score"] = max(score, 0)  # Floor at 0
    prop["_score_reasons"] = reasons

    if score >= 20:
        prop["_priority"] = "Immediate"
    elif score >= 15:
        prop["_priority"] = "High"
    elif score >= 10:
        prop["_priority"] = "Moderate"
    else:
        prop["_priority"] = "Watchlist"

    return score


def main():
    with open("/tmp/nyc_assemblage/r7plus_properties.json") as f:
        properties = json.load(f)

    print(f"Running signal checks on {len(properties)} R7+ properties...", file=sys.stderr)

    # Pre-load NYC Finance tax liens (single API call for efficiency)
    # Borough codes: 1=Manhattan, 2=Bronx, 3=Brooklyn, 4=Queens, 5=Staten Island
    print("\nLoading NYC Finance tax lien list...", file=sys.stderr)
    all_bbl_liens = set()
    for borough_code in ["2", "3", "1", "4"]:
        results = curl_socrata(
            "https://data.cityofnewyork.us/resource/9rz4-mjek.json",
            {"$where": f"borough='{borough_code}'", "$limit": "10000"}
        )
        if not isinstance(results, list):
            continue
        for r in results:
            if not isinstance(r, dict):
                continue
            # Dataset has no bbl field — construct from borough+block+lot
            b = str(r.get("borough", "")).zfill(1)
            blk = str(r.get("block", "")).zfill(5)
            lt = str(r.get("lot", "")).zfill(4)
            if b and blk and lt:
                constructed_bbl = b + blk + lt
                all_bbl_liens.add(constructed_bbl)
    print(f"  Loaded {len(all_bbl_liens)} properties on tax lien list", file=sys.stderr)

    # Cache for block-level signals (avoid redundant API calls)
    block_llc_cache = {}  # (borough_digit, block) -> bool
    block_dob_cache = {}  # (borough, block) -> (has_demo, has_nb)

    for i, prop in enumerate(properties, 1):
        bbl = prop.get("_bbl", "")
        block = prop.get("_block", "")
        lot = prop.get("_lot", "")
        borough = prop.get("_borough", "")
        borough_digit = prop.get("_borough_digit", "")

        address = prop.get("address", "")[:50]
        if i % 10 == 0 or i <= 3:
            print(f"  [{i}/{len(properties)}] {address}...", file=sys.stderr)

        # NYC Finance tax lien (from pre-loaded cache)
        prop["_tax_lien"] = bbl in all_bbl_liens

        # ACRIS signals
        if bbl and block and lot:
            try:
                acris_signals = check_acris_property(borough_digit, block, lot)
                prop.update(acris_signals)
            except Exception as e:
                prop["_acris_error"] = str(e)

            # Block-level LLC deed check (cached)
            cache_key = (borough_digit, block)
            if cache_key not in block_llc_cache:
                try:
                    block_llc_cache[cache_key] = get_acris_block_deeds(borough_digit, block)
                except Exception:
                    block_llc_cache[cache_key] = False
            prop["_block_llc_deed"] = block_llc_cache.get(cache_key, False)

        # DOB permits (cached per block)
        dob_key = (borough, block)
        if dob_key not in block_dob_cache:
            try:
                block_dob_cache[dob_key] = check_dob_block(borough, block)
            except Exception:
                block_dob_cache[dob_key] = (False, False)
        has_demo, has_nb = block_dob_cache.get(dob_key, (False, False))
        prop["_block_demo"] = has_demo
        prop["_block_new_building"] = has_nb

        # HPD violations
        if borough_digit and block and lot:
            try:
                hpd_count = check_hpd_violations(borough_digit, block, lot)
                prop["_hpd_violations"] = hpd_count
            except Exception:
                prop["_hpd_violations"] = 0
        else:
            prop["_hpd_violations"] = 0

        # Compute score
        compute_score(prop)

        # Rate limit
        if i % 5 == 0:
            time.sleep(0.5)

    # Sort by score descending
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    # Save
    with open("/tmp/nyc_assemblage/scored_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

    # Stats
    from collections import Counter
    priority_counts = Counter(p.get("_priority", "Watchlist") for p in properties)
    signal_counts = {
        "tax_lien": sum(1 for p in properties if p.get("_tax_lien")),
        "lis_pendens_recent": sum(1 for p in properties if p.get("_acris_lis_pendens_recent")),
        "lis_pendens_any": sum(1 for p in properties if p.get("_acris_lis_pendens")),
        "federal_lien": sum(1 for p in properties if p.get("_acris_federal_lien")),
        "estate": sum(1 for p in properties if p.get("_acris_estate")),
        "hpd_5plus": sum(1 for p in properties if (p.get("_hpd_violations") or 0) >= 5),
        "hpd_10plus": sum(1 for p in properties if (p.get("_hpd_violations") or 0) >= 10),
        "block_demo": sum(1 for p in properties if p.get("_block_demo")),
        "block_nb": sum(1 for p in properties if p.get("_block_new_building")),
        "block_llc": sum(1 for p in properties if p.get("_block_llc_deed")),
        "dom_180plus": sum(1 for p in properties if (p.get("daysOnMarket") or 0) > 180),
    }

    print(f"\n=== Scoring Results ===", file=sys.stderr)
    for tier, count in sorted(priority_counts.items()):
        print(f"  {tier}: {count}", file=sys.stderr)
    print(f"\n=== Signal Summary ===", file=sys.stderr)
    for signal, count in signal_counts.items():
        print(f"  {signal}: {count}", file=sys.stderr)

    print(f"\n✓ Saved {len(properties)} scored properties", file=sys.stderr)
    print(json.dumps({"count": len(properties), "signal_counts": signal_counts}))


if __name__ == "__main__":
    main()
