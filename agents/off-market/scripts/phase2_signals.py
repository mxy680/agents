#!/usr/bin/env python3
"""
Phase 2: Pre-Market Signal Checks (Off-Market R8+ Properties)
For each R8+ property (not listed for sale), check:
- ACRIS: judgments, federal liens, tax lien sale certificates, LLC deed transfers on block
- DOB: demolition/new building permits on block (last 6 months)
- HPD: open violations count
- NYC Finance: tax lien list
- 311 complaints
- ECB violations
- FDNY vacate orders
- DOB complaints
- Certificate of Occupancy on block
- Citi Bike density

Input:  /tmp/off_market_scan/r8plus_properties.json
Output: /tmp/off_market_scan/scored_properties.json
"""

import json
import subprocess
import sys
import time
from collections import Counter
from datetime import datetime, timedelta

TODAY = datetime.now()
DATE_90_DAYS_AGO = (TODAY - timedelta(days=90)).strftime("%Y-%m-%d")
DATE_180_DAYS_AGO = (TODAY - timedelta(days=180)).strftime("%Y-%m-%d")

# Simple rate limiter for Socrata (limit ~900/hr = ~15/min)
_socrata_call_times: list = []
_SOCRATA_MAX_PER_MIN = 14  # stay just under 15/min


def _socrata_rate_limit():
    """Block if we're approaching the Socrata rate limit."""
    now = time.monotonic()
    _socrata_call_times[:] = [t for t in _socrata_call_times if now - t < 60.0]
    if len(_socrata_call_times) >= _SOCRATA_MAX_PER_MIN:
        sleep_for = 60.0 - (now - _socrata_call_times[0]) + 0.1
        if sleep_for > 0:
            time.sleep(sleep_for)
    _socrata_call_times.append(time.monotonic())


def curl_socrata(url_base: str, params: dict) -> list:
    """Make a Socrata API call using subprocess curl with proper encoding."""
    _socrata_rate_limit()

    cmd = ["curl", "-s", "-G", url_base]
    for key, val in params.items():
        cmd.extend(["--data-urlencode", f"{key}={val}"])
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode != 0:
            return []
        if '"error"' in result.stdout and "429" in result.stdout:
            print("  [WARN] Socrata rate limit hit (429) — backing off 60s", file=sys.stderr)
            time.sleep(60)
            return []
        parsed = json.loads(result.stdout)
        return parsed if isinstance(parsed, list) else []
    except Exception:
        return []


def get_acris_docs(borough_digit: str, block: str, lot: str) -> list:
    """Get ACRIS document IDs for a lot."""
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
    return [r.get("document_id") for r in records if r.get("document_id")]


def get_acris_masters(doc_ids: list) -> list:
    """Get document details from ACRIS Master table."""
    if not doc_ids:
        return []
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


def get_acris_parties(doc_ids: list, party_type: str = "2") -> list:
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


def get_acris_block_deeds(borough_digit: str, block: str) -> bool:
    """Check for LLC deed transfers on the same block (last 12 months)."""
    block_padded = str(block).zfill(5)
    date_12mo = (TODAY - timedelta(days=365)).strftime("%Y-%m-%d")

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

    masters = get_acris_masters(doc_ids[:500])
    deed_ids = [
        m.get("document_id") for m in masters
        if m.get("doc_type") == "DEED"
        and m.get("document_date", "") >= date_12mo
    ]

    if not deed_ids:
        return False

    parties = get_acris_parties(deed_ids[:50])
    for party in parties:
        name = party.get("name", "").upper()
        if "LLC" in name or "L.L.C" in name:
            return True

    return False


def check_acris_property(borough_digit: str, block: str, lot: str) -> dict:
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

    if doc_ids:
        estate_ids = doc_ids[:50]
        ids_str = ",".join(f"'{d}'" for d in estate_ids)
        estate_records = curl_socrata(
            "https://data.cityofnewyork.us/resource/636b-3b5g.json",
            {"$where": f"document_id in({ids_str}) AND (name like '%ESTATE OF%' OR name like '%EXECUTOR%' OR name like '%EXECUTRIX%')"}
        )
        if estate_records:
            signals["acris_estate"] = True

    return signals


def check_dob_block(borough_name: str, block: str) -> tuple[bool, bool]:
    """Check for DM/NB permits on the same block in last 6 months."""
    block_padded = str(block).zfill(5)
    borough_upper = borough_name.upper()

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


def check_hpd_violations(borough_digit: str, block: str, lot: str) -> int:
    """Count open HPD violations for a property."""
    block_padded = str(block).zfill(5).lstrip("0") or "0"
    lot_padded = str(lot).zfill(4).lstrip("0") or "0"

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


def check_311_complaints(address: str, borough: str) -> int:
    """Count 311 complaints at an address in last 12 months."""
    date_12mo = (TODAY - timedelta(days=365)).strftime("%Y-%m-%d")
    borough_upper = borough.upper()
    addr_upper = address.upper().split(",")[0].strip()
    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/erm2-nwe9.json",
        {
            "$select": "count(*)",
            "$where": f"upper(incident_address)='{addr_upper}' AND upper(borough)='{borough_upper}' AND created_date > '{date_12mo}'"
        }
    )
    if records and isinstance(records, list) and len(records) > 0:
        try:
            return int(records[0].get("count", 0))
        except (ValueError, TypeError):
            return 0
    return 0


def check_ecb_violations(address: str) -> int:
    """Count defaulted ECB/OATH violations at an address."""
    parts = address.split(",")[0].strip().split(" ", 1)
    if len(parts) < 2:
        return 0
    house_num = parts[0]
    street = parts[1].strip().upper()
    for suffix in [" BRONX", " BROOKLYN", " MANHATTAN", " QUEENS", " NEW YORK", " NY"]:
        street = street.replace(suffix, "").strip()

    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/6bgk-3dad.json",
        {
            "$select": "count(*)",
            "$where": f"respondent_house_number='{house_num}' AND upper(respondent_street) like '%{street[:20]}%' AND violation_status='DEFAULT'"
        }
    )
    if records and isinstance(records, list) and len(records) > 0:
        try:
            return int(records[0].get("count", 0))
        except (ValueError, TypeError):
            return 0
    return 0


def check_fdny_vacate(borough_digit: str, block: str) -> bool:
    """Check for FDNY vacate orders on the block."""
    block_padded = str(block).zfill(5)
    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/frax-hfgs.json",
        {
            "$where": f"borough='{borough_digit}' AND block='{block_padded}'",
            "$limit": "5"
        }
    )
    return len(records) > 0


def check_dob_complaints(bbl: str) -> int:
    """Count open DOB complaints for a property."""
    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/eabe-havv.json",
        {
            "$select": "count(*)",
            "$where": f"bbl='{bbl}' AND status='OPEN'"
        }
    )
    if records and isinstance(records, list) and len(records) > 0:
        try:
            return int(records[0].get("count", 0))
        except (ValueError, TypeError):
            return 0
    return 0


def check_co_on_block(borough_digit: str, block: str) -> bool:
    """Check for new Certificates of Occupancy on the block in last 12 months."""
    block_padded = str(block).zfill(5)
    date_12mo = (TODAY - timedelta(days=365)).strftime("%Y-%m-%dT00:00:00.000")
    borough_names = {"1": "MANHATTAN", "2": "BRONX", "3": "BROOKLYN", "4": "QUEENS", "5": "STATEN ISLAND"}
    borough_name = borough_names.get(str(borough_digit), "")
    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/pkdm-hqz6.json",
        {
            "$where": f"borough='{borough_name}' AND block='{block_padded}' AND c_of_o_issuance_date > '{date_12mo}'",
            "$limit": "5"
        }
    )
    return len(records) > 0


def check_citibike_density(lat, lng) -> int:
    """Get Citi Bike station count within 1km."""
    if not lat or not lng:
        return 0
    cmd = [
        "integrations", "citibike", "stations", "density",
        f"--lat={lat}", f"--lng={lng}", "--radius=1000", "--json"
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode != 0:
            return 0
        data = json.loads(result.stdout)
        return data.get("count", 0)
    except Exception:
        return 0


def _has_distress_signal(prop: dict) -> bool:
    """Return True if any distress signal fired on this property."""
    if prop.get("_tax_lien"):
        return True
    if prop.get("_acris_lis_pendens"):
        return True
    if (prop.get("_hpd_violations") or 0) >= 5:
        return True
    if prop.get("_acris_estate"):
        return True
    if prop.get("_acris_federal_lien"):
        return True
    if prop.get("_fdny_vacate"):
        return True
    if (prop.get("_ecb_violations") or 0) > 0:
        return True
    return False


def compute_score(prop: dict) -> int:
    """Compute composite score for an off-market property."""
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
        year_built = int(float(str(prop.get("_year_built", 0) or 0)))
        if 1800 < year_built < 1945:
            score += 2
            reasons.append(f"Pre-war ({year_built}) (+2)")
    except (ValueError, TypeError):
        pass

    # Small lot
    try:
        lot_area = int(float(str(prop.get("_lot_area", 0) or 0).replace(",", "").replace("$", "")))
        if 0 < lot_area < 2000:
            score += 2
            reasons.append(f"Small lot ({lot_area} SF) (+2)")
    except (ValueError, TypeError):
        pass

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

    # 311 complaints (10+ in 12 months)
    complaints_311 = prop.get("_311_complaints", 0) or 0
    if complaints_311 >= 10:
        score += 3
        reasons.append(f"311 complaints {complaints_311} in 12mo (+3)")

    # ECB/OATH violations (defaulted)
    ecb = prop.get("_ecb_violations", 0) or 0
    if ecb > 0:
        score += 2
        reasons.append(f"Defaulted ECB violations {ecb} (+2)")

    # FDNY vacate order
    if prop.get("_fdny_vacate"):
        score += 5
        reasons.append("FDNY vacate order on block (+5)")

    # DOB complaints (open)
    dob_complaints = prop.get("_dob_complaints", 0) or 0
    if dob_complaints >= 3:
        score += 2
        reasons.append(f"Open DOB complaints {dob_complaints} (+2)")

    # Certificate of Occupancy on block
    if prop.get("_block_co"):
        score += 2
        reasons.append("New CO on block (active development) (+2)")

    # Citi Bike density
    cb_stations = prop.get("_citibike_stations", 0) or 0
    if cb_stations >= 5:
        score += 2
        reasons.append(f"Citi Bike {cb_stations} stations within 1km (+2)")

    # Off-market distress bonus: if ANY distress signal fires, add +3.
    # An off-market property WITH distress signals is the whole point of this tool.
    if _has_distress_signal(prop):
        score += 3
        reasons.append("Off-market with distress signal (+3)")

    prop["_score"] = max(score, 0)  # Floor at 0
    prop["_score_reasons"] = reasons

    # Priority tiers adjusted for lower scores (no listing bonus)
    if score >= 15:
        prop["_priority"] = "Immediate"
    elif score >= 10:
        prop["_priority"] = "High"
    elif score >= 6:
        prop["_priority"] = "Moderate"
    else:
        prop["_priority"] = "Watchlist"

    return score


def main():
    with open("/tmp/off_market_scan/r8plus_properties.json") as f:
        properties = json.load(f)

    print(f"Running signal checks on {len(properties)} R8+ off-market properties...", file=sys.stderr)

    # Pre-load NYC Finance tax liens using pagination to avoid truncation
    print("\nLoading NYC Finance tax lien list...", file=sys.stderr)
    all_bbl_liens: set = set()
    BATCH_SIZE = 50000
    for borough_code in ["2", "3", "1", "4"]:
        offset = 0
        while True:
            results = curl_socrata(
                "https://data.cityofnewyork.us/resource/9rz4-mjek.json",
                {
                    "$where": f"borough='{borough_code}'",
                    "$limit": str(BATCH_SIZE),
                    "$offset": str(offset),
                }
            )
            if not isinstance(results, list) or len(results) == 0:
                break
            for r in results:
                if not isinstance(r, dict):
                    continue
                b = str(r.get("borough", "")).zfill(1)
                blk = str(r.get("block", "")).zfill(5)
                lt = str(r.get("lot", "")).zfill(4)
                if b and blk and lt:
                    all_bbl_liens.add(b + blk + lt)
            if len(results) < BATCH_SIZE:
                break
            offset += BATCH_SIZE
    print(f"  Loaded {len(all_bbl_liens)} properties on tax lien list", file=sys.stderr)

    # Cache for block-level signals (avoid redundant API calls)
    block_llc_cache: dict = {}   # (borough_digit, block) -> bool
    block_dob_cache: dict = {}   # (borough, block) -> (has_demo, has_nb)
    fdny_cache: dict = {}        # (borough_digit, block) -> bool
    co_cache: dict = {}          # (borough_digit, block) -> bool

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
                prop["_hpd_violations"] = check_hpd_violations(borough_digit, block, lot)
            except Exception:
                prop["_hpd_violations"] = 0
        else:
            prop["_hpd_violations"] = 0

        # 311 complaints
        try:
            prop["_311_complaints"] = check_311_complaints(address, borough)
        except Exception:
            prop["_311_complaints"] = 0

        # ECB/OATH violations
        try:
            prop["_ecb_violations"] = check_ecb_violations(address)
        except Exception:
            prop["_ecb_violations"] = 0

        # FDNY vacate orders (cached per block)
        fdny_key = (borough_digit, block)
        if fdny_key not in fdny_cache:
            try:
                fdny_cache[fdny_key] = check_fdny_vacate(borough_digit, block)
            except Exception:
                fdny_cache[fdny_key] = False
        prop["_fdny_vacate"] = fdny_cache.get(fdny_key, False)

        # DOB complaints
        if bbl:
            try:
                prop["_dob_complaints"] = check_dob_complaints(bbl)
            except Exception:
                prop["_dob_complaints"] = 0

        # Certificate of Occupancy on block (cached)
        co_key = (borough_digit, block)
        if co_key not in co_cache:
            try:
                co_cache[co_key] = check_co_on_block(borough_digit, block)
            except Exception:
                co_cache[co_key] = False
        prop["_block_co"] = co_cache.get(co_key, False)

        # Citi Bike density
        lat = prop.get("latitude")
        lng = prop.get("longitude")
        if lat and lng:
            try:
                prop["_citibike_stations"] = check_citibike_density(lat, lng)
            except Exception:
                prop["_citibike_stations"] = 0

        # Compute score
        compute_score(prop)

        # Rate limit — 0.5s between every property, extra pause every 5
        time.sleep(0.5)
        if i % 5 == 0:
            time.sleep(0.5)

    # Sort by score descending
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    # Save
    with open("/tmp/off_market_scan/scored_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

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
        "311_10plus": sum(1 for p in properties if (p.get("_311_complaints") or 0) >= 10),
        "ecb_defaulted": sum(1 for p in properties if (p.get("_ecb_violations") or 0) > 0),
        "fdny_vacate": sum(1 for p in properties if p.get("_fdny_vacate")),
        "dob_complaints_3plus": sum(1 for p in properties if (p.get("_dob_complaints") or 0) >= 3),
        "block_co": sum(1 for p in properties if p.get("_block_co")),
        "citibike_5plus": sum(1 for p in properties if (p.get("_citibike_stations") or 0) >= 5),
        "has_any_distress": sum(1 for p in properties if _has_distress_signal(p)),
    }

    print(f"\n=== Scoring Results ===", file=sys.stderr)
    for tier, count in sorted(priority_counts.items()):
        print(f"  {tier}: {count}", file=sys.stderr)
    print(f"\n=== Signal Summary ===", file=sys.stderr)
    for signal, count in signal_counts.items():
        print(f"  {signal}: {count}", file=sys.stderr)

    print(f"\nSaved {len(properties)} scored properties", file=sys.stderr)
    print(json.dumps({"count": len(properties), "signal_counts": signal_counts}))


if __name__ == "__main__":
    main()
