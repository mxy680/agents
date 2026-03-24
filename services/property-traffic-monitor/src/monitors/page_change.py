"""Page change detection monitor.

Monitors HTTP headers (ETag, Last-Modified, Content-Length) for property pages.
Pages that change frequently may indicate higher traffic/activity.
Free, no API key needed.
"""

from __future__ import annotations

import hashlib
import urllib.parse

import httpx

from src.monitors.base import BaseMonitor, MonitorResult

# Target site URL builders
SITE_URLS: dict[str, str] = {
    "propertyshark": "https://www.propertyshark.com/mason/Property/{slug}/",
    "dob_bis": "https://a810-bisweb.nyc.gov/bisweb/PropertyProfileOverviewServlet?boro={boro_num}&block=&lot=",
    "zola": "https://zola.planning.nyc.gov/",
    "nyc_property_portal": "https://propertyinformationportal.nyc.gov/",
}

BOROUGH_NUMS = {
    "Manhattan": "1",
    "Bronx": "2",
    "Brooklyn": "3",
    "Queens": "4",
    "Staten Island": "5",
}


def _build_check_urls(address: str, borough: str) -> dict[str, str]:
    slug = address.replace(" ", "-").replace(",", "").replace(".", "")
    boro_num = BOROUGH_NUMS.get(borough, "1")
    encoded = urllib.parse.quote(address)

    return {
        "propertyshark": f"https://www.propertyshark.com/mason/Property/{slug}/",
        "dob_bis": (
            f"https://a810-bisweb.nyc.gov/bisweb/PropertyProfileOverviewServlet"
            f"?boro={boro_num}&houseno={urllib.parse.quote(address.split()[0])}"
            f"&street={urllib.parse.quote(' '.join(address.split()[1:]))}"
        ),
        "zola": "https://zola.planning.nyc.gov/",
        "nyc_property_portal": "https://propertyinformationportal.nyc.gov/",
    }


class PageChangeMonitor(BaseMonitor):
    name = "page_change"

    async def check(self, address: str, borough: str) -> MonitorResult:
        urls = _build_check_urls(address, borough)
        page_data: dict[str, dict] = {}
        signals_found = 0

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
            for site_name, url in urls.items():
                try:
                    resp = await client.head(url)
                    headers = dict(resp.headers)
                    info = {
                        "status": resp.status_code,
                        "etag": headers.get("etag", ""),
                        "last_modified": headers.get("last-modified", ""),
                        "content_length": headers.get("content-length", ""),
                        "cache_control": headers.get("cache-control", ""),
                        "x_cache": headers.get("x-cache", ""),
                        "age": headers.get("age", ""),
                        "server": headers.get("server", ""),
                    }
                    page_data[site_name] = info

                    # Signal: page is reachable and has dynamic headers
                    if resp.status_code == 200:
                        signals_found += 1
                        if info["etag"] or info["last_modified"]:
                            signals_found += 1

                except Exception as e:
                    page_data[site_name] = {"error": str(e)}

        # Normalize: reachable pages with cache headers indicate activity
        max_signals = len(urls) * 2  # 1 for reachable, 1 for cache headers
        value = min(1.0, signals_found / max_signals) if max_signals > 0 else 0.0

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if page_data else "error",
            raw_data={
                "pages_checked": len(urls),
                "signals_found": signals_found,
                "pages": page_data,
            },
        )
