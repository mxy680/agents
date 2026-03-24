"""Tests for the signal aggregator."""

from src.signals.aggregator import WEIGHTS, compute_composite_score


def test_empty_signals():
    assert compute_composite_score({}) == 0.0


def test_all_max_signals():
    signals = {name: 1.0 for name in WEIGHTS}
    score = compute_composite_score(signals)
    expected = sum(WEIGHTS.values())
    assert abs(score - expected) < 0.001


def test_single_signal():
    score = compute_composite_score({"wayback": 0.5})
    expected = 0.5 * WEIGHTS["wayback"]
    assert abs(score - expected) < 0.001


def test_clamps_above_one():
    score = compute_composite_score({"wayback": 2.0})
    expected = 1.0 * WEIGHTS["wayback"]
    assert abs(score - expected) < 0.001


def test_clamps_below_zero():
    score = compute_composite_score({"wayback": -1.0})
    assert score == 0.0


def test_unknown_monitor_ignored():
    score = compute_composite_score({"unknown_monitor": 1.0})
    assert score == 0.0


def test_partial_signals():
    signals = {"wayback": 1.0, "google_trends": 0.5, "crux": 0.8}
    score = compute_composite_score(signals)
    expected = (
        1.0 * WEIGHTS["wayback"]
        + 0.5 * WEIGHTS["google_trends"]
        + 0.8 * WEIGHTS["crux"]
    )
    assert abs(score - round(expected, 4)) < 0.001


def test_weights_sum_to_one():
    total = sum(WEIGHTS.values())
    assert abs(total - 1.0) < 0.001


def test_all_monitors_have_weights():
    expected_monitors = [
        "crux",
        "google_trends",
        "google_autocomplete",
        "page_change",
        "social_signals",
        "google_cache",
        "wayback",
        "extension",
    ]
    for m in expected_monitors:
        assert m in WEIGHTS, f"Missing weight for monitor: {m}"
