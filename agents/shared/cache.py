"""
Supabase scrape_data cache for agent scripts.

Usage:
    from shared.cache import get_cached, put_cached, get_cached_batch

    # Check cache first, fetch if missing
    pluto = get_cached("pluto", bbl)
    if pluto is None:
        pluto = fetch_from_socrata(bbl)
        put_cached("pluto", bbl, pluto)

    # Batch get (for checking many at once)
    cached = get_cached_batch("pluto", [bbl1, bbl2, bbl3])
    # Returns {bbl1: data, bbl2: data} — missing keys not in dict
"""

import json
import os
import urllib.request
import urllib.parse

_SUPABASE_URL = None
_SERVICE_KEY = None


def _init():
    """Lazy-init Supabase credentials from env."""
    global _SUPABASE_URL, _SERVICE_KEY
    if _SUPABASE_URL is None:
        _SUPABASE_URL = os.environ.get("NEXT_PUBLIC_SUPABASE_URL", "")
        _SERVICE_KEY = os.environ.get("SUPABASE_SERVICE_ROLE_KEY", "")
    return bool(_SUPABASE_URL and _SERVICE_KEY)


def _headers():
    return {
        "apikey": _SERVICE_KEY,
        "Authorization": f"Bearer {_SERVICE_KEY}",
        "Content-Type": "application/json",
        "Accept": "application/json",
        "Prefer": "resolution=merge-duplicates",
    }


def get_cached(provider: str, external_id: str):
    """Get a single cached item. Returns the data dict or None."""
    if not _init():
        return None
    try:
        eid = urllib.parse.quote(external_id, safe="")
        url = f"{_SUPABASE_URL}/rest/v1/scrape_data?provider=eq.{provider}&external_id=eq.{eid}&select=data&limit=1"
        req = urllib.request.Request(url, headers=_headers())
        with urllib.request.urlopen(req, timeout=10) as resp:
            rows = json.loads(resp.read())
        if rows and rows[0].get("data"):
            return rows[0]["data"]
        return None
    except Exception:
        return None


def get_cached_batch(provider: str, external_ids: list):
    """Get multiple cached items. Returns {external_id: data} for found items."""
    if not _init() or not external_ids:
        return {}
    try:
        result = {}
        # Supabase REST API supports IN via csv: external_id=in.(a,b,c)
        # Batch in groups of 100 to avoid URL length limits
        for i in range(0, len(external_ids), 100):
            batch = external_ids[i:i + 100]
            ids_csv = ",".join(urllib.parse.quote(str(eid), safe="") for eid in batch)
            url = (
                f"{_SUPABASE_URL}/rest/v1/scrape_data"
                f"?provider=eq.{provider}&external_id=in.({ids_csv})"
                f"&select=external_id,data&limit={len(batch)}"
            )
            req = urllib.request.Request(url, headers=_headers())
            with urllib.request.urlopen(req, timeout=30) as resp:
                rows = json.loads(resp.read())
            for row in rows:
                if row.get("data"):
                    result[row["external_id"]] = row["data"]
        return result
    except Exception:
        return {}


def put_cached(provider: str, external_id: str, data: dict, batch_id: str = None, bbl: str = None):
    """Upsert a single item into the cache."""
    if not _init():
        return False
    try:
        import datetime
        row = {
            "provider": provider,
            "external_id": str(external_id),
            "data": data,
            "scraped_at": datetime.datetime.now().isoformat(),
        }
        if batch_id:
            row["batch_id"] = batch_id
        if bbl:
            row["bbl"] = bbl

        url = f"{_SUPABASE_URL}/rest/v1/scrape_data"
        body = json.dumps(row).encode()
        req = urllib.request.Request(url, data=body, headers=_headers(), method="POST")
        with urllib.request.urlopen(req, timeout=10) as resp:
            resp.read()
        return True
    except Exception:
        return False


def put_cached_batch(provider: str, items: list):
    """Upsert multiple items. Each item is (external_id, data, bbl=None).

    Args:
        provider: The provider name (e.g., 'pluto', 'citibike')
        items: List of tuples (external_id, data_dict) or (external_id, data_dict, bbl)
    """
    if not _init() or not items:
        return False
    try:
        import datetime
        now = datetime.datetime.now().isoformat()
        rows = []
        for item in items:
            if len(item) == 3:
                eid, data, bbl = item
            else:
                eid, data = item
                bbl = None
            row = {
                "provider": provider,
                "external_id": str(eid),
                "data": data,
                "scraped_at": now,
            }
            if bbl:
                row["bbl"] = bbl
            rows.append(row)

        # Upsert in chunks of 500
        for i in range(0, len(rows), 500):
            chunk = rows[i:i + 500]
            url = f"{_SUPABASE_URL}/rest/v1/scrape_data"
            body = json.dumps(chunk).encode()
            req = urllib.request.Request(url, data=body, headers=_headers(), method="POST")
            with urllib.request.urlopen(req, timeout=30) as resp:
                resp.read()
        return True
    except Exception:
        return False
