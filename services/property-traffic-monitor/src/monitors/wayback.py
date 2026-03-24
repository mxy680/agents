"""Wayback Machine CDX API monitor.

Checks how many times the Wayback Machine has crawled property pages on target sites.
More frequent crawling = higher traffic/interest.
Free, no API key needed.
"""

from __future__ import annotations

import urllib.parse

import httpx

from src.monitors.base import BaseMonitor, MonitorResult

# URL patterns for target sites keyed by site name
SITE_URL_PATTERNS: dict[str, str] = {
    "propertyshark": "https://www.propertyshark.com/mason/Property/{address_slug}/",
    "zola": "https://zola.planning.nyc.gov/",
    "dob": "https://a810-bisweb.nyc.gov/bisweb/PropertyProfileOverviewServlet",
    "acris": "https://a836-acris.nyc.gov/DS/DocumentSearch/BBL",
}

CDX_API = "http://web.archive.org/cdx/search/cdx"


def _address_to_slug(address: str) -> str:
    """Convert address like '350 5th Ave' to URL-friendly slug."""
    return address.replace(" ", "-").replace(",", "").replace(".", "")


def _build_urls(address: str, borough: str) -> dict[str, str]:
    """Build probable URLs for the address on each target site."""
    slug = _address_to_slug(address)
    encoded = urllib.parse.quote(address)

    return {
        "propertyshark": f"https://www.propertyshark.com/mason/*{slug}*",
        "zola": f"https://zola.planning.nyc.gov/*",
        "dob_bis": f"https://a810-bisweb.nyc.gov/bisweb/*",
        "acris": f"https://a836-acris.nyc.gov/*",
        "nyc_dof": f"https://propertyinformationportal.nyc.gov/*",
    }


class WaybackMonitor(BaseMonitor):
    name = "wayback"

    async def check(self, address: str, borough: str) -> MonitorResult:
        urls = _build_urls(address, borough)
        total_captures = 0
        site_results: dict[str, int] = {}

        async with httpx.AsyncClient(timeout=30.0) as client:
            for site_name, url_pattern in urls.items():
                try:
                    resp = await client.get(
                        CDX_API,
                        params={
                            "url": url_pattern,
                            "output": "json",
                            "limit": 100,
                            "fl": "timestamp,original,statuscode",
                            "filter": "statuscode:200",
                            "from": "20240101",
                        },
                    )
                    if resp.status_code == 200:
                        data = resp.json() if resp.text.strip() else []
                        # First row is header
                        count = max(0, len(data) - 1) if data else 0
                        site_results[site_name] = count
                        total_captures += count
                except Exception as e:
                    site_results[site_name] = 0

        # Normalize: 0 captures = 0.0, 50+ captures = 1.0
        value = min(1.0, total_captures / 50.0)

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if total_captures > 0 else ("live" if site_results else "error"),
            raw_data={
                "total_captures": total_captures,
                "by_site": site_results,
                "query_urls": urls,
            },
        )
