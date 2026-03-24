"""Google Trends monitor.

Uses pytrends to check relative search interest for property addresses.
Free, no API key needed (but can be rate limited).
"""

from __future__ import annotations

import asyncio
from functools import partial

from src.monitors.base import BaseMonitor, MonitorResult


def _fetch_trends(address: str, borough: str) -> dict:
    """Synchronous pytrends call (run in executor)."""
    try:
        from pytrends.request import TrendReq

        pytrends = TrendReq(hl="en-US", tz=300, timeout=(10, 25))

        keywords = [f"{address} {borough}", f"{address} NYC"]
        # pytrends max 5 keywords
        pytrends.build_payload(keywords, cat=0, timeframe="today 3-m", geo="US-NY")
        interest = pytrends.interest_over_time()

        if interest.empty:
            return {"interest": {}, "max_value": 0, "keywords": keywords}

        # Get the max interest value across all keywords
        max_val = 0
        interest_data = {}
        for kw in keywords:
            if kw in interest.columns:
                series = interest[kw].tolist()
                kw_max = max(series) if series else 0
                interest_data[kw] = {
                    "max": kw_max,
                    "recent_values": series[-5:] if series else [],
                }
                max_val = max(max_val, kw_max)

        return {"interest": interest_data, "max_value": max_val, "keywords": keywords}

    except Exception as e:
        return {"error": str(e), "max_value": 0, "keywords": []}


class GoogleTrendsMonitor(BaseMonitor):
    name = "google_trends"

    async def check(self, address: str, borough: str) -> MonitorResult:
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(None, partial(_fetch_trends, address, borough))

        max_val = result.get("max_value", 0)
        # Google Trends values are 0-100, normalize to 0-1
        value = min(1.0, max_val / 100.0)

        has_error = "error" in result
        source = "error" if has_error and max_val == 0 else "live"

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source=source,
            raw_data=result,
        )
