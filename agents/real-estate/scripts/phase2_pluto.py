#!/usr/bin/env python3
"""
Phase 2: PLUTO Geocoding + Zoning Filter
- Geocode each address to BBL
- Look up PLUTO lot data
- Filter to R7+ zoning only
"""

import json
import subprocess
import sys
import time
import urllib.request
import urllib.parse

R7_PLUS_ZONES = {
    "R7", "R7-1", "R7-2", "R7A", "R7B", "R7D", "R7X",
    "R8", "R8A", "R8B", "R8X",
    "R9", "R9A", "R9X",
    "R10", "R10A", "R10X",
    "C4-4", "C4-5",
    # MX zones with R7+ component handled below
}

def is_r7_plus(zone):
    """Check if a zone qualifies as R7+."""
    if not zone:
        return False
    z = zone.strip().upper()

    # Direct match
    if z in {x.upper() for x in R7_PLUS_ZONES}:
        return True

    # R7+ base zones
    for prefix in ["R7", "R8", "R9", "R10"]:
        if z.startswith(prefix):
            return True

    # MX zones (mixed use with residential component)
    if z.startswith("MX"):
        return True

    # C4-4+
    if z in ["C4-4", "C4-5", "C4-6", "C4-7"]:
        return True

    return False


def geocode_address(address):
    """Geocode an address to BBL using NYC GeoSearch."""
    try:
        encoded = urllib.parse.quote(address)
        url = f"https://geosearch.planninglabs.nyc/v2/search?text={encoded}&size=1"
        req = urllib.request.Request(url, headers={"User-Agent": "NYC-Assemblage/1.0"})
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read())

        features = data.get("features", [])
        if not features:
            return None, None, None

        feat = features[0]
        props = feat.get("properties", {})
        addendum = props.get("addendum", {})
        pad = addendum.get("pad", {})
        bbl = pad.get("bbl")

        if not bbl:
            return None, None, None

        # Parse BBL: borough(1) + block(5) + lot(4)
        bbl_str = str(bbl).zfill(10)
        borough_digit = bbl_str[0]
        block = bbl_str[1:6]
        lot = bbl_str[6:10]

        return bbl_str, block, lot

    except Exception as e:
        return None, None, None


def get_pluto_data(bbl):
    """Get PLUTO lot data for a BBL."""
    try:
        url = f"https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl={bbl}"
        req = urllib.request.Request(url, headers={"User-Agent": "NYC-Assemblage/1.0"})
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read())

        if not data:
            return None

        return data[0]
    except Exception as e:
        return None


def build_zola_url(bbl_str):
    """Build ZoLa URL from BBL."""
    try:
        bbl_str = str(bbl_str).zfill(10)
        borough = bbl_str[0]
        block = bbl_str[1:6].lstrip("0") or "0"
        lot = bbl_str[6:10].lstrip("0") or "0"
        # ZoLa format: /lot/{borough}/{block_padded}/{lot_padded}
        block_pad = bbl_str[1:6]
        lot_pad = bbl_str[6:10]
        return f"https://zola.planning.nyc.gov/lot/{borough}/{block_pad}/{lot_pad}"
    except:
        return "Unable to verify"


BOROUGH_DIGIT_MAP = {
    "Bronx": "2",
    "Brooklyn": "3",
    "Manhattan": "1",
    "Queens": "4",
    "Staten Island": "5",
}


def main():
    # Load Zillow results
    with open("/tmp/nyc_assemblage/zillow_results.json") as f:
        listings = json.load(f)

    print(f"Processing {len(listings)} listings for PLUTO geocoding...", file=sys.stderr)

    qualifying_r7 = []
    failed_geocode = 0
    failed_pluto = 0
    wrong_zone = 0

    for i, prop in enumerate(listings, 1):
        address = prop.get("address", "")
        zpid = prop.get("zpid", "")
        borough = prop.get("_borough", "")

        if i % 20 == 0 or i == 1:
            print(f"  [{i}/{len(listings)}] Processing {address[:50]}...", file=sys.stderr)

        # Geocode
        bbl, block, lot = geocode_address(address)

        if not bbl:
            # Try simplified address (remove unit/apt info)
            clean_addr = address.split("APT")[0].split("#")[0].strip().rstrip(",")
            bbl, block, lot = geocode_address(clean_addr)

        if not bbl:
            failed_geocode += 1
            prop["_bbl"] = None
            prop["_geocode_failed"] = True
            continue

        prop["_bbl"] = bbl
        prop["_block"] = block
        prop["_lot"] = lot
        prop["_borough_digit"] = bbl[0]

        # Get PLUTO data
        pluto = get_pluto_data(bbl)

        if not pluto:
            failed_pluto += 1
            prop["_pluto_failed"] = True
            # Still keep if we have zoning from another source
            prop["_zoning"] = "Unable to verify"
            continue

        # Extract PLUTO fields
        zone = pluto.get("zonedist1", "")
        prop["_zoning"] = zone
        prop["_lot_area"] = pluto.get("lotarea", "Not stated")
        prop["_bldg_area"] = pluto.get("bldgarea", "Not stated")
        prop["_year_built"] = pluto.get("yearbuilt", "Not stated")
        prop["_bldg_class"] = pluto.get("bldgclass", "Not stated")
        prop["_num_floors"] = pluto.get("numfloors", "Not stated")
        prop["_units_res"] = pluto.get("unitsres", "Not stated")
        prop["_units_total"] = pluto.get("unitstotal", "Not stated")
        prop["_lot_front"] = pluto.get("lotfront", "Not stated")
        prop["_lot_depth"] = pluto.get("lotdepth", "Not stated")
        prop["_zola_url"] = build_zola_url(bbl)

        # Zone filter
        if not is_r7_plus(zone):
            wrong_zone += 1
            continue

        # Filter out large multi-unit buildings (more than 5 units residential)
        try:
            units_res = int(float(pluto.get("unitsres", 0)))
            if units_res > 5:
                wrong_zone += 1  # Count as filtered
                continue
        except (ValueError, TypeError):
            pass

        qualifying_r7.append(prop)

        # Small delay every 10 geocodes to be nice to the API
        if i % 10 == 0:
            time.sleep(0.3)

    print(f"\n=== PLUTO Results ===", file=sys.stderr)
    print(f"  Geocode failures: {failed_geocode}", file=sys.stderr)
    print(f"  PLUTO lookup failures: {failed_pluto}", file=sys.stderr)
    print(f"  Wrong zone (< R7): {wrong_zone}", file=sys.stderr)
    print(f"  R7+ qualifying: {len(qualifying_r7)}", file=sys.stderr)

    # Borough breakdown
    from collections import Counter
    borough_counts = Counter(p.get("_borough", "Unknown") for p in qualifying_r7)
    zone_counts = Counter(p.get("_zoning", "Unknown") for p in qualifying_r7)
    for b, c in sorted(borough_counts.items()):
        print(f"  {b}: {c}", file=sys.stderr)
    print("\n  Top zones:", file=sys.stderr)
    for z, c in zone_counts.most_common(10):
        print(f"    {z}: {c}", file=sys.stderr)

    # Save
    with open("/tmp/nyc_assemblage/r7plus_properties.json", "w") as f:
        json.dump(qualifying_r7, f, indent=2)

    print(f"\n✓ Saved {len(qualifying_r7)} R7+ properties", file=sys.stderr)
    print(json.dumps({"count": len(qualifying_r7), "path": "/tmp/nyc_assemblage/r7plus_properties.json"}))


if __name__ == "__main__":
    main()
