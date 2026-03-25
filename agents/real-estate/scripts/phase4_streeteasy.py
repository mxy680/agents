#!/usr/bin/env python3
"""
Phase 4: StreetEasy enrichment for properties with preliminary score >= 3
Checks price history for drops and relisting cycles
"""

import json
import subprocess
import sys
import time
from datetime import datetime, timedelta

TODAY = datetime.now()
DATE_30_DAYS_AGO = (TODAY - timedelta(days=30)).strftime("%Y-%m-%d")


def get_streeteasy_history(address):
    """Get price history for an address from StreetEasy."""
    cmd = [
        "integrations", "streeteasy", "listings", "history",
        f"--address={address}",
        "--json"
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode != 0:
            return None
        data = json.loads(result.stdout)
        if isinstance(data, list):
            return data
        return None
    except Exception:
        return None


def analyze_price_history(history, current_price):
    """Analyze price history for drops and relisting cycles."""
    if not history:
        return {"se_price_drop_pct": 0, "se_cycles": 0, "se_recent_drop": False, "se_available": False}

    signals = {
        "se_price_drop_pct": 0.0,
        "se_cycles": 0,
        "se_recent_drop": False,
        "se_available": True,
        "se_original_price": None,
        "se_price_history": []
    }

    # Parse history entries
    prices = []
    listing_events = 0
    last_was_listed = False

    original_price = None
    min_price = None

    for entry in history:
        event = entry.get("event", "").lower()
        price = entry.get("price")
        date = entry.get("date", "")

        if price:
            try:
                price_val = int(str(price).replace(",", "").replace("$", ""))
                prices.append(price_val)

                if original_price is None:
                    original_price = price_val
                if min_price is None or price_val < min_price:
                    min_price = price_val

                # Check for recent price drop
                if date >= DATE_30_DAYS_AGO and "price" in event and "reduc" in event:
                    signals["se_recent_drop"] = True
                elif date >= DATE_30_DAYS_AGO and "price" in event:
                    signals["se_recent_drop"] = True

            except (ValueError, TypeError):
                pass

        # Count listing/delisting cycles
        if "listed" in event or "relisted" in event:
            if not last_was_listed:
                listing_events += 1
                last_was_listed = True
        elif "removed" in event or "delisted" in event or "taken off" in event:
            last_was_listed = False

    signals["se_cycles"] = listing_events
    signals["se_original_price"] = original_price

    # Calculate price drop percentage
    if original_price and min_price and original_price > 0:
        drop_pct = (original_price - min_price) / original_price * 100
        signals["se_price_drop_pct"] = round(drop_pct, 1)

    # Also check current price vs original
    if original_price and current_price and original_price > 0:
        current_drop_pct = (original_price - current_price) / original_price * 100
        if current_drop_pct > signals["se_price_drop_pct"]:
            signals["se_price_drop_pct"] = round(current_drop_pct, 1)

    return signals


def update_score_with_streeteasy(prop, se_signals):
    """Add StreetEasy signals to property score."""
    additional_score = 0
    additional_reasons = []

    drop_pct = se_signals.get("se_price_drop_pct", 0)
    cycles = se_signals.get("se_cycles", 0)
    recent_drop = se_signals.get("se_recent_drop", False)

    if drop_pct > 10:
        additional_score += 3
        additional_reasons.append(f"Price drop {drop_pct:.0f}% from original (+3)")

    if cycles >= 3:
        additional_score += 4
        additional_reasons.append(f"{cycles} listing/delisting cycles (+4)")

    if recent_drop:
        additional_score += 2
        additional_reasons.append("Price drop in last 30 days (+2)")

    prop.update(se_signals)
    prop["_score"] = prop.get("_score", 0) + additional_score
    prop["_score_reasons"] = prop.get("_score_reasons", []) + additional_reasons

    # Recompute priority
    score = prop["_score"]
    if score >= 20:
        prop["_priority"] = "Immediate"
    elif score >= 15:
        prop["_priority"] = "High"
    elif score >= 10:
        prop["_priority"] = "Moderate"
    else:
        prop["_priority"] = "Watchlist"

    return additional_score


def main():
    with open("/tmp/nyc_assemblage/scored_properties.json") as f:
        properties = json.load(f)

    # Only check properties with preliminary score >= 3
    eligible = [p for p in properties if p.get("_score", 0) >= 3]
    print(f"Running StreetEasy enrichment on {len(eligible)}/{len(properties)} eligible properties...", file=sys.stderr)

    enriched = 0
    errors = 0

    for i, prop in enumerate(eligible, 1):
        address = prop.get("address", "")
        zpid = prop.get("zpid", "")
        current_price = prop.get("price", 0)

        if i % 10 == 0 or i <= 3:
            print(f"  [{i}/{len(eligible)}] {address[:50]}...", file=sys.stderr)

        history = get_streeteasy_history(address)

        if history is not None:
            se_signals = analyze_price_history(history, current_price)
            added = update_score_with_streeteasy(prop, se_signals)
            if added > 0:
                enriched += 1
        else:
            errors += 1
            prop["se_available"] = False

        # Rate limit
        time.sleep(0.5)
        if i % 20 == 0:
            time.sleep(1)

    # Re-sort by updated score
    properties.sort(key=lambda p: p.get("_score", 0), reverse=True)

    print(f"\n=== StreetEasy Results ===", file=sys.stderr)
    print(f"  Enriched with new signals: {enriched}", file=sys.stderr)
    print(f"  Errors/not found: {errors}", file=sys.stderr)

    # Save
    with open("/tmp/nyc_assemblage/final_properties.json", "w") as f:
        json.dump(properties, f, indent=2)

    print(f"\n✓ Saved {len(properties)} final scored properties", file=sys.stderr)

    # Show top 10
    print("\n=== Top 10 Properties ===", file=sys.stderr)
    for p in properties[:10]:
        addr = p.get("address", "N/A")
        score = p.get("_score", 0)
        priority = p.get("_priority", "")
        zone = p.get("_zoning", "")
        reasons = ", ".join(p.get("_score_reasons", []))
        print(f"  Score {score} ({priority}) | {addr} | {zone} | {reasons[:80]}", file=sys.stderr)

    print(json.dumps({"count": len(properties), "top_score": properties[0].get("_score", 0) if properties else 0}))


if __name__ == "__main__":
    main()
