from __future__ import annotations

# Weight map for composite score calculation
WEIGHTS: dict[str, float] = {
    "crux": 0.25,
    "google_trends": 0.15,
    "google_autocomplete": 0.10,
    "page_change": 0.15,
    "social_signals": 0.10,
    "google_cache": 0.10,
    "wayback": 0.05,
    "extension": 0.10,
}


def compute_composite_score(signals: dict[str, float]) -> float:
    """Compute weighted composite score from monitor signals.

    Each signal value should be 0.0-1.0.
    Unrecognized monitor names are ignored.
    """
    total = 0.0
    for name, value in signals.items():
        weight = WEIGHTS.get(name, 0.0)
        clamped = max(0.0, min(1.0, value))
        total += weight * clamped
    return round(total, 4)
