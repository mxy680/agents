"""Social signals monitor.

Checks Reddit, Hacker News, and web search APIs for mentions of property addresses.
Free APIs, no key needed.
"""

from __future__ import annotations

import httpx

from src.monitors.base import BaseMonitor, MonitorResult

REDDIT_SEARCH_URL = "https://www.reddit.com/search.json"
HN_SEARCH_URL = "https://hn.algolia.com/api/v1/search"


class SocialSignalsMonitor(BaseMonitor):
    name = "social_signals"

    async def check(self, address: str, borough: str) -> MonitorResult:
        results: dict[str, dict] = {}
        total_mentions = 0

        async with httpx.AsyncClient(
            timeout=15.0,
            headers={
                "User-Agent": (
                    "PropertyTrafficMonitor/1.0 (research tool; "
                    "contact: admin@example.com)"
                )
            },
        ) as client:
            # Reddit search
            try:
                resp = await client.get(
                    REDDIT_SEARCH_URL,
                    params={
                        "q": f'"{address}" NYC',
                        "sort": "new",
                        "limit": 25,
                        "t": "month",
                    },
                )
                if resp.status_code == 200:
                    data = resp.json()
                    posts = data.get("data", {}).get("children", [])
                    reddit_count = len(posts)
                    total_mentions += reddit_count
                    results["reddit"] = {
                        "count": reddit_count,
                        "posts": [
                            {
                                "title": p["data"]["title"],
                                "subreddit": p["data"]["subreddit"],
                                "created_utc": p["data"]["created_utc"],
                                "url": f"https://reddit.com{p['data']['permalink']}",
                            }
                            for p in posts[:5]
                        ],
                    }
                else:
                    results["reddit"] = {"error": f"HTTP {resp.status_code}", "count": 0}
            except Exception as e:
                results["reddit"] = {"error": str(e), "count": 0}

            # Hacker News (Algolia API)
            try:
                resp = await client.get(
                    HN_SEARCH_URL,
                    params={
                        "query": f'"{address}"',
                        "tags": "story",
                        "hitsPerPage": 10,
                    },
                )
                if resp.status_code == 200:
                    data = resp.json()
                    hits = data.get("hits", [])
                    hn_count = len(hits)
                    total_mentions += hn_count
                    results["hackernews"] = {
                        "count": hn_count,
                        "posts": [
                            {
                                "title": h.get("title", ""),
                                "url": h.get("url", ""),
                                "points": h.get("points", 0),
                                "created_at": h.get("created_at", ""),
                            }
                            for h in hits[:5]
                        ],
                    }
                else:
                    results["hackernews"] = {"error": f"HTTP {resp.status_code}", "count": 0}
            except Exception as e:
                results["hackernews"] = {"error": str(e), "count": 0}

            # BiggerPockets forum search (public)
            try:
                resp = await client.get(
                    "https://www.biggerpockets.com/search",
                    params={"q": address, "type": "forums"},
                    follow_redirects=True,
                )
                # Just check if the page is reachable and has results
                bp_found = resp.status_code == 200 and address.split()[0] in resp.text
                results["biggerpockets"] = {
                    "reachable": resp.status_code == 200,
                    "address_found": bp_found,
                    "count": 1 if bp_found else 0,
                }
                if bp_found:
                    total_mentions += 1
            except Exception as e:
                results["biggerpockets"] = {"error": str(e), "count": 0}

        # Normalize: 0 mentions = 0, 10+ = 1.0
        value = min(1.0, total_mentions / 10.0)

        return MonitorResult(
            monitor_name=self.name,
            value=value,
            source="live" if results else "error",
            raw_data={
                "total_mentions": total_mentions,
                "sources": results,
            },
        )
