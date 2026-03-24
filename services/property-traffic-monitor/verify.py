"""
Verification gate — runs after all monitors are built.
Exits 0 ONLY if real traffic data was retrieved from live sources.
Exits 1 with detailed failure report otherwise.
"""

import httpx
import sys
import json

API = "http://localhost:8000"


def verify():
    failures = []

    # 1. API is running
    try:
        r = httpx.get(f"{API}/addresses")
        addresses = r.json()
        assert len(addresses) >= 3, f"Need >= 3 monitored addresses, got {len(addresses)}"
    except Exception as e:
        print(f"FATAL: API not running or no addresses seeded: {e}")
        sys.exit(1)

    # 2. At least 3 monitors returned non-zero signals from LIVE data
    monitors_with_live_data = set()
    addresses_with_signals = 0

    for addr in addresses:
        signals = httpx.get(f"{API}/addresses/{addr['id']}/signals").json()
        score = httpx.get(f"{API}/addresses/{addr['id']}/score").json()

        has_signal = False
        for monitor_name, signal in signals.items():
            if signal.get("value", 0) > 0 and signal.get("source") == "live":
                monitors_with_live_data.add(monitor_name)
                has_signal = True
                print(
                    f"  \u2713 {addr['address']} \u2192 {monitor_name}: "
                    f"{signal['value']:.3f} (live)"
                )

        if has_signal:
            addresses_with_signals += 1

    # 3. Check minimum thresholds
    if len(monitors_with_live_data) < 3:
        failures.append(
            f"Need >= 3 monitors with live data, got {len(monitors_with_live_data)}: "
            f"{monitors_with_live_data or 'none'}"
        )

    if addresses_with_signals < 2:
        failures.append(
            f"Need >= 2 addresses with real signals, got {addresses_with_signals}"
        )

    # 4. Verify at least one composite score is non-zero
    scores = []
    for addr in addresses:
        score = httpx.get(f"{API}/addresses/{addr['id']}/score").json()
        scores.append(score.get("composite_score", 0))

    if max(scores) == 0:
        failures.append("All composite scores are 0 — no monitor produced real data")

    # 5. Verify signals contain actual data payloads (not just status codes)
    for addr in addresses[:1]:  # Spot-check first address
        signals = httpx.get(f"{API}/addresses/{addr['id']}/signals").json()
        for name, sig in signals.items():
            if sig.get("value", 0) > 0:
                if not sig.get("raw_data"):
                    failures.append(
                        f"{name} reports value > 0 but has no raw_data — "
                        f"might be fabricated. Must include actual API response."
                    )

    # VERDICT
    if failures:
        print("\n\u2717 VERIFICATION FAILED:")
        for f in failures:
            print(f"  - {f}")
        print("\nDo NOT output the completion promise. Fix the issues and re-verify.")
        sys.exit(1)
    else:
        print(f"\n\u2713 VERIFICATION PASSED")
        print(f"  Monitors with live data: {monitors_with_live_data}")
        print(f"  Addresses with signals: {addresses_with_signals}/{len(addresses)}")
        print(f"  Max composite score: {max(scores):.3f}")
        print(f"\nYou may now output the completion promise.")
        sys.exit(0)


if __name__ == "__main__":
    verify()
