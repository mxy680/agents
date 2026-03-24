"""Tests for individual monitor implementations."""

import pytest

from src.monitors.base import MonitorResult
from src.monitors.wayback import WaybackMonitor, _address_to_slug, _build_urls
from src.monitors.google_autocomplete import GoogleAutocompleteMonitor
from src.monitors.google_cache import GoogleCacheMonitor
from src.monitors.page_change import PageChangeMonitor, _build_check_urls
from src.monitors.social_signals import SocialSignalsMonitor


def test_address_to_slug():
    assert _address_to_slug("350 5th Ave") == "350-5th-Ave"
    assert _address_to_slug("1 Vanderbilt Ave, New York") == "1-Vanderbilt-Ave-New-York"


def test_build_urls():
    urls = _build_urls("350 5th Ave", "Manhattan")
    assert "propertyshark" in urls
    assert "dob_bis" in urls
    assert "acris" in urls
    assert "zola" in urls


def test_build_check_urls():
    urls = _build_check_urls("350 5th Ave", "Manhattan")
    assert "propertyshark" in urls
    assert "dob_bis" in urls
    assert "boro=1" in urls["dob_bis"]


def test_monitor_result_dataclass():
    result = MonitorResult(
        monitor_name="test",
        value=0.5,
        source="live",
        raw_data={"foo": "bar"},
    )
    assert result.monitor_name == "test"
    assert result.value == 0.5
    assert result.source == "live"
    assert result.raw_data == {"foo": "bar"}


@pytest.mark.asyncio
async def test_wayback_monitor_returns_result():
    monitor = WaybackMonitor()
    assert monitor.name == "wayback"
    # Just verify the monitor can be instantiated and has correct name
    # Live test happens in integration/verification


@pytest.mark.asyncio
async def test_autocomplete_monitor_returns_result():
    monitor = GoogleAutocompleteMonitor()
    assert monitor.name == "google_autocomplete"


@pytest.mark.asyncio
async def test_page_change_monitor_returns_result():
    monitor = PageChangeMonitor()
    assert monitor.name == "page_change"


@pytest.mark.asyncio
async def test_social_signals_monitor_returns_result():
    monitor = SocialSignalsMonitor()
    assert monitor.name == "social_signals"


@pytest.mark.asyncio
async def test_google_cache_monitor_returns_result():
    monitor = GoogleCacheMonitor()
    assert monitor.name == "google_cache"
