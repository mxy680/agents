#!/usr/bin/env python3
"""
Daily Probate Scanner
Monitors ACRIS for estate/probate filings on high-value properties.
"""
import json
import sys
import os
import time
import urllib.request
import urllib.parse
from datetime import datetime

# Add shared module to path
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".."))
try:
    from shared.cache import get_cached, put_cached
except ImportError:
    def get_cached(p, e): return None
    def put_cached(p, e, d, **kw): return False

OUT_DIR = "/tmp/probate_monitor"
TODAY = datetime.now()
LOOKBACK_DAYS = 7


def curl_socrata(url_base, params):
    """Make a Socrata API call."""
    query_params = urllib.parse.urlencode(params)
    url = f"{url_base}?{query_params}"
    req = urllib.request.Request(url, headers={"User-Agent": "Emdash-Agents/1.0"})
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            data = json.loads(resp.read())
        return data if isinstance(data, list) else []
    except Exception as e:
        print(f"  [WARN] Socrata query failed: {e}", file=sys.stderr)
        return []


def load_seen() -> set:
    """Load previously seen BBLs from Supabase cache."""
    cached = get_cached("probate_seen", "all_bbls")
    if cached and isinstance(cached, dict):
        return set(cached.get("bbls", []))
    return set()


def save_seen(seen: set) -> None:
    """Save seen BBLs to Supabase cache."""
    put_cached("probate_seen", "all_bbls", {"bbls": sorted(seen)})


def get_probate_documents():
    """Find recent ACRIS documents with estate/probate party names."""
    print("Querying ACRIS for estate/probate filings...", file=sys.stderr)

    records = curl_socrata(
        "https://data.cityofnewyork.us/resource/636b-3b5g.json",
        {
            "$where": "(name like '%ESTATE OF%' OR name like '%EXECUTOR%' OR name like '%EXECUTRIX%')",
            "$order": "good_through_date DESC",
            "$limit": "1000",
            "$select": "document_id,name,good_through_date",
        }
    )

    print(f"  Found {len(records)} estate party records", file=sys.stderr)

    # Deduplicate by document_id and extract estate names
    doc_info = {}
    for r in records:
        doc_id = r.get("document_id", "")
        if doc_id and doc_id not in doc_info:
            doc_info[doc_id] = {
                "document_id": doc_id,
                "estate_name": r.get("name", ""),
                "date": r.get("good_through_date", ""),
            }

    print(f"  Unique documents: {len(doc_info)}", file=sys.stderr)
    return list(doc_info.values())


def get_bbls_for_documents(doc_ids):
    """Get BBLs from ACRIS legals table for the given document IDs."""
    if not doc_ids:
        return {}

    doc_to_bbl = {}
    # Batch in groups of 50
    for i in range(0, len(doc_ids), 50):
        batch = doc_ids[i:i+50]
        ids_str = ",".join(f"'{d}'" for d in batch)
        records = curl_socrata(
            "https://data.cityofnewyork.us/resource/8h5j-fqxa.json",
            {"$where": f"document_id in({ids_str})"}
        )
        for r in records:
            doc_id = r.get("document_id", "")
            borough = r.get("borough", "")
            block = str(r.get("block", "")).zfill(5)
            lot = str(r.get("lot", "")).zfill(4)
            if borough and block and lot:
                bbl = f"{borough}{block}{lot}"
                doc_to_bbl[doc_id] = bbl
        time.sleep(0.5)

    return doc_to_bbl


def get_pluto_data(bbl):
    """Get PLUTO lot data for a BBL. Checks Supabase cache first."""
    cached = get_cached("pluto", bbl)
    if cached:
        return cached
    try:
        url = f"https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl={bbl}"
        req = urllib.request.Request(url, headers={"User-Agent": "Emdash-Agents/1.0"})
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read())
        if not data:
            return None
        result = data[0]
        put_cached("pluto", bbl, result, bbl=bbl)
        return result
    except Exception:
        return None


def qualifies(pluto):
    """Check if property meets criteria: 15K+ buildable SF or 5+ units."""
    if not pluto:
        return False

    # Check units
    try:
        units = int(float(pluto.get("unitsres", 0) or 0))
        if units >= 5:
            return True
    except (ValueError, TypeError):
        pass

    # Check buildable SF (rough estimate: lot_area * FAR)
    # For R8+, FAR is typically 6.02-10.0. Use conservative 6.0.
    try:
        lot_area = float(pluto.get("lotarea", 0) or 0)
        if lot_area * 6.0 >= 15000:
            return True
    except (ValueError, TypeError):
        pass

    return False


BOROUGH_NAMES = {"1": "Manhattan", "2": "Bronx", "3": "Brooklyn", "4": "Queens", "5": "Staten Island"}


def main():
    os.makedirs(OUT_DIR, exist_ok=True)
    seen = load_seen()

    # Step 1: Get probate documents
    docs = get_probate_documents()

    # Step 2: Get BBLs
    doc_ids = [d["document_id"] for d in docs]
    doc_to_bbl = get_bbls_for_documents(doc_ids)
    print(f"  Resolved {len(doc_to_bbl)} BBLs", file=sys.stderr)

    # Build BBL -> estate info mapping
    bbl_info = {}
    for doc in docs:
        doc_id = doc["document_id"]
        bbl = doc_to_bbl.get(doc_id)
        if bbl and bbl not in bbl_info:
            bbl_info[bbl] = doc

    # Step 3: Filter by already-seen
    new_bbls = {bbl: info for bbl, info in bbl_info.items() if bbl not in seen}
    print(f"  New (unseen) BBLs: {len(new_bbls)}", file=sys.stderr)

    # Step 4: Look up PLUTO and filter
    alerts = []
    for i, (bbl, info) in enumerate(new_bbls.items(), 1):
        if i % 20 == 0:
            print(f"  [{i}/{len(new_bbls)}] Checking PLUTO...", file=sys.stderr)

        pluto = get_pluto_data(bbl)
        if not pluto:
            continue

        if not qualifies(pluto):
            continue

        boro_code = bbl[0]
        block = bbl[1:6]
        lot = bbl[6:10]

        alert = {
            "bbl": bbl,
            "address": pluto.get("address", "Unknown"),
            "borough": BOROUGH_NAMES.get(boro_code, "Unknown"),
            "estate_name": info.get("estate_name", ""),
            "filing_date": info.get("date", "")[:10],
            "zoning": pluto.get("zonedist1", ""),
            "lot_area": pluto.get("lotarea", ""),
            "bldg_area": pluto.get("bldgarea", ""),
            "units_res": pluto.get("unitsres", ""),
            "units_total": pluto.get("unitstotal", ""),
            "year_built": pluto.get("yearbuilt", ""),
            "zola_url": f"https://zola.planning.nyc.gov/lot/{boro_code}/{block}/{lot}",
        }
        alerts.append(alert)

        # Mark as seen
        seen.add(bbl)

        time.sleep(0.3)

    # Save seen list
    save_seen(seen)

    # Save alerts
    alert_path = f"{OUT_DIR}/alerts_{TODAY.strftime('%Y-%m-%d')}.json"
    with open(alert_path, "w") as f:
        json.dump(alerts, f, indent=2)

    # Summary
    print(f"\n=== Probate Alert Results ===", file=sys.stderr)
    print(f"  Estate documents found: {len(docs)}", file=sys.stderr)
    print(f"  BBLs resolved: {len(doc_to_bbl)}", file=sys.stderr)
    print(f"  New BBLs (not seen before): {len(new_bbls)}", file=sys.stderr)
    print(f"  Qualifying alerts (15K+ SF or 5+ units): {len(alerts)}", file=sys.stderr)

    if alerts:
        print(f"\n  NEW ALERTS:", file=sys.stderr)
        for a in alerts:
            print(f"    {a['address']}, {a['borough']} | {a['estate_name']} | Units: {a['units_res']} | Lot: {a['lot_area']} SF | {a['zoning']}", file=sys.stderr)
    else:
        print(f"\n  No new qualifying probate alerts today.", file=sys.stderr)

    print(f"\nSaved {len(alerts)} alerts to {alert_path}", file=sys.stderr)
    print(json.dumps({"alerts": len(alerts), "path": alert_path}))


if __name__ == "__main__":
    main()
