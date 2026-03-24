"""Chrome User Experience Report (CrUX) API monitor.

Checks if property page URLs have enough traffic to appear in CrUX data.
CrUX only includes URLs with ~200+ monthly Chrome visits, so presence = signal.
Requires a free Google Cloud API key.
"""

from __future__ import annotations

import httpx

from src.config import settings
from src.monitors.base import BaseMonitor, MonitorResult

CRUX_API_URL = "https://chromeuxreport.googleapis.com/v1/records:queryRecord"

# Target site origins to check
SITE_ORIGINS: list[str] = [
    "https://www.propertyshark.com",
    "https://zola.planning.nyc.gov",
    "https://a810-bisweb.nyc.gov",
    "https://a836-acris.nyc.gov",
    "https://propertyinformationportal.nyc.gov",
    "https://www.actovia.com",
    "https://www.costar.com",
]


class CruxMonitor(BaseMonitor):
    name = "crux"

    async def check(self, address: str, borough: str) -> MonitorResult:
        if not settings.crux_api_key:
            return MonitorResult(
                monitor_name=self.name,
                value=0.0,
                source="error",
                raw_data={"error": "CRUX_API_KEY not configured"},
            )

        results: dict[str, dict] = {}
        origins_with_data = 0

        async with httpx.AsyncClient(timeout=15.0) as client:
            for origin in SITE_ORIGINS:
                try:
                    resp = await client.post(
                        f"{CRUX_API_URL}?key={settings.crux_api_key}",
                        json={"origin": origin},
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        has_record = "record" in data
                        results[origin] = {
                            "has_data": has_record,
                            "metrics": list(data.get("record", {}).get("metrics", {}).keys())
                            if has_record
                            else [],
                        }
                        if has_record:
                            origins_with_data += 1
                    elif resp.status_code == 404:
                        results[origin] = {"has_data": False, "reason": "not_enough_traffic"}
                    else:
                        results[origin] = {
                            "has_data": False,
                            "error": f"HTTP {resp.status_code}",
                        }
                except Exception as e:
                    results[origin] = {"has_data": False, "error": str(e)}

        # Signal: fraction of target sites that have CrUX data
        value = origins_with_data / len(SITE_ORIGINS) if SITE_ORIGINS else 0.0

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if results else "error",
            raw_data={
                "origins_checked": len(SITE_ORIGINS),
                "origins_with_data": origins_with_data,
                "results": results,
            },
        )
