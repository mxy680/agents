"""Search autocomplete monitor.

Checks if a property address appears in search engine autocomplete suggestions.
Uses DuckDuckGo (reliable, free) and Bing (free) autocomplete APIs.
If search engines suggest it, people are searching for it.
"""

from __future__ import annotations

import re

import httpx

from src.monitors.base import BaseMonitor, MonitorResult

DDG_AUTOCOMPLETE_URL = "https://duckduckgo.com/ac/"
BING_SUGGEST_URL = "https://www.bing.com/AS/Suggestions"


class GoogleAutocompleteMonitor(BaseMonitor):
    name = "google_autocomplete"

    async def check(self, address: str, borough: str) -> MonitorResult:
        queries = [
            address,
            f"{address} {borough}",
            f"{address} NYC",
            f"{address} property",
        ]

        match_count = 0
        total_suggestions = 0
        raw_responses: dict[str, dict] = {}
        has_any_data = False

        async with httpx.AsyncClient(timeout=15.0) as client:
            for query in queries:
                query_result: dict = {"ddg": [], "bing": []}

                # DuckDuckGo autocomplete (most reliable)
                try:
                    resp = await client.get(
                        DDG_AUTOCOMPLETE_URL,
                        params={"q": query, "type": "list"},
                        headers={"User-Agent": "Mozilla/5.0"},
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        # DDG returns [query, [suggestions]]
                        suggestions = data[1] if len(data) > 1 else []
                        query_result["ddg"] = suggestions
                        total_suggestions += len(suggestions)
                        has_any_data = True

                        addr_lower = address.lower()
                        for s in suggestions:
                            if addr_lower in s.lower():
                                match_count += 1
                except Exception as e:
                    query_result["ddg_error"] = str(e)

                # Bing autocomplete
                try:
                    resp = await client.get(
                        BING_SUGGEST_URL,
                        params={"qry": query, "cvid": "1"},
                        headers={
                            "User-Agent": (
                                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                                "AppleWebKit/537.36 (KHTML, like Gecko) "
                                "Chrome/131.0.0.0 Safari/537.36"
                            )
                        },
                    )
                    if resp.status_code == 200:
                        # Bing returns HTML with suggestions
                        html = resp.text
                        bing_suggestions = re.findall(r'query="([^"]+)"', html)
                        query_result["bing"] = bing_suggestions
                        total_suggestions += len(bing_suggestions)
                        has_any_data = True

                        addr_lower = address.lower()
                        for s in bing_suggestions:
                            if addr_lower in s.lower():
                                match_count += 1
                except Exception as e:
                    query_result["bing_error"] = str(e)

                raw_responses[query] = query_result

        # Normalize: each matching suggestion contributes
        max_possible = len(queries) * 4  # up to 4 matches per query (2 engines x 2)
        value = min(1.0, match_count / max_possible) if max_possible > 0 else 0.0

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if has_any_data else "error",
            raw_data={
                "match_count": match_count,
                "total_suggestions": total_suggestions,
                "total_queries": len(queries),
                "suggestions": raw_responses,
            },
        )
