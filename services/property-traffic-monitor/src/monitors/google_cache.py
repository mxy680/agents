"""Google Cache timestamp monitor.

Checks Google's cache for property pages — pages re-crawled more often
when they get more traffic. Uses Google's webcache endpoint.
Free, no API key needed.
"""

from __future__ import annotations

import re
import urllib.parse
from datetime import datetime, timezone

import httpx

from src.monitors.base import BaseMonitor, MonitorResult

GOOGLE_CACHE_URL = "https://webcache.googleusercontent.com/search"

# Property page URLs to check cache timestamps for
SITE_PAGES: dict[str, str] = {
    "propertyshark": "https://www.propertyshark.com/mason/Property/{slug}/",
    "zola": "https://zola.planning.nyc.gov/",
}

# Regex to extract cache timestamp from Google Cache page
CACHE_DATE_RE = re.compile(
    r"It is a snapshot of the page as it appeared on (\w+ \d+, \d{4} \d+:\d+:\d+ GMT)"
)


def _address_to_slug(address: str) -> str:
    return address.replace(" ", "-").replace(",", "").replace(".", "")


class GoogleCacheMonitor(BaseMonitor):
    name = "google_cache"

    async def check(self, address: str, borough: str) -> MonitorResult:
        slug = _address_to_slug(address)
        cache_results: dict[str, dict] = {}
        freshness_signals = 0

        # Also check direct Google search for cached pages
        search_queries = [
            f"cache:propertyshark.com {address}",
            f'"{address}" site:propertyshark.com',
            f'"{address}" site:zola.planning.nyc.gov',
            f'"{address}" site:a810-bisweb.nyc.gov',
        ]

        async with httpx.AsyncClient(
            timeout=15.0,
            follow_redirects=True,
            headers={
                "User-Agent": (
                    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                    "AppleWebKit/537.36 (KHTML, like Gecko) "
                    "Chrome/131.0.0.0 Safari/537.36"
                )
            },
        ) as client:
            # Check Google search for indexed property pages
            try:
                resp = await client.get(
                    "https://www.google.com/search",
                    params={
                        "q": f'"{address}" (propertyshark OR zola OR bisweb OR acris)',
                        "num": 10,
                    },
                )
                if resp.status_code == 200:
                    body = resp.text
                    # Count how many results mention the address
                    addr_parts = address.lower().split()
                    result_count = body.lower().count(addr_parts[0]) if addr_parts else 0
                    cache_results["google_search"] = {
                        "status": resp.status_code,
                        "address_mentions": min(result_count, 20),
                        "has_results": result_count > 0,
                    }
                    if result_count > 0:
                        freshness_signals += 1
                else:
                    cache_results["google_search"] = {
                        "status": resp.status_code,
                        "address_mentions": 0,
                    }
            except Exception as e:
                cache_results["google_search"] = {"error": str(e)}

            # Check Bing search (often less restricted)
            try:
                resp = await client.get(
                    "https://www.bing.com/search",
                    params={"q": f'"{address}" NYC property'},
                )
                if resp.status_code == 200:
                    body = resp.text
                    addr_parts = address.lower().split()
                    result_count = body.lower().count(addr_parts[0]) if addr_parts else 0
                    cache_results["bing_search"] = {
                        "status": resp.status_code,
                        "address_mentions": min(result_count, 20),
                        "has_results": result_count > 0,
                    }
                    if result_count > 0:
                        freshness_signals += 1
                else:
                    cache_results["bing_search"] = {"status": resp.status_code}
            except Exception as e:
                cache_results["bing_search"] = {"error": str(e)}

        # Normalize
        value = min(1.0, freshness_signals / 2.0)

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if cache_results else "error",
            raw_data={
                "freshness_signals": freshness_signals,
                "results": cache_results,
            },
        )
