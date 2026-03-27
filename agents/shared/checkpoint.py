"""
Job checkpoint utility — tracks phase completion in Supabase.

Usage:
    from shared.checkpoint import is_phase_done, mark_phase_done, mark_phase_failed

    if is_phase_done("real-estate", "off-market-scan", "phase2_pluto"):
        print("Skipping phase 2 — already completed today")
    else:
        run_phase_2()
        mark_phase_done("real-estate", "off-market-scan", "phase2_pluto", {"count": 225})
"""

import json
import os
import urllib.request
from datetime import datetime

_SUPABASE_URL = None
_SERVICE_KEY = None


def _init():
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


def _today():
    return datetime.now().strftime("%Y-%m-%d")


def is_phase_done(agent_name, job_slug, phase):
    """Check if a phase was completed today."""
    if not _init():
        return False
    try:
        url = (
            f"{_SUPABASE_URL}/rest/v1/job_checkpoints"
            f"?agent_name=eq.{agent_name}"
            f"&job_slug=eq.{job_slug}"
            f"&run_date=eq.{_today()}"
            f"&phase=eq.{phase}"
            f"&status=eq.completed"
            f"&select=id&limit=1"
        )
        req = urllib.request.Request(url, headers=_headers())
        with urllib.request.urlopen(req, timeout=10) as resp:
            rows = json.loads(resp.read())
        return len(rows) > 0
    except Exception:
        return False


def mark_phase_done(agent_name, job_slug, phase, metadata=None):
    """Mark a phase as completed for today."""
    _upsert_checkpoint(agent_name, job_slug, phase, "completed", metadata)


def mark_phase_failed(agent_name, job_slug, phase, metadata=None):
    """Mark a phase as failed for today."""
    _upsert_checkpoint(agent_name, job_slug, phase, "failed", metadata)


def _upsert_checkpoint(agent_name, job_slug, phase, status, metadata=None):
    if not _init():
        return
    try:
        row = {
            "agent_name": agent_name,
            "job_slug": job_slug,
            "run_date": _today(),
            "phase": phase,
            "status": status,
            "metadata": metadata or {},
            "updated_at": datetime.now().isoformat(),
        }
        url = f"{_SUPABASE_URL}/rest/v1/job_checkpoints"
        body = json.dumps(row).encode()
        req = urllib.request.Request(url, data=body, headers=_headers(), method="POST")
        with urllib.request.urlopen(req, timeout=10) as resp:
            resp.read()
    except Exception:
        pass
