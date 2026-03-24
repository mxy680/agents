# Ralph Loop Prompt — Property Traffic Spike Monitor

## Usage
```bash
/ralph-loop:ralph-loop "$(cat docs/property-traffic-monitor-ralph-prompt.md | sed -n '/^## PROMPT$/,$ p' | tail -n +2)" --completion-promise="SYSTEM DEPLOYED AND VERIFIED"
```

## PROMPT

Build a working system that detects when people are actively researching a specific NYC property address on PropertyShark, ZoLa, DOB (Department of Buildings), Actovia, NYC Department of Finance (ACRIS), and CoStar — BEFORE any public filing or listing appears.

The goal is NOT to monitor filings (permits, deeds, etc.) — those come AFTER the deal is in motion. The goal is to detect the RESEARCH ACTIVITY itself: the moment when multiple people start looking up the same property on these sites because word is spreading through real estate networks.

### CONTEXT — What We Already Know

Prior research (see `docs/property-traffic-spike-detection.md`) found that:
- None of these sites publicly expose view counts or page traffic data
- CoStar Marketing Center shows traffic but only to the listing broker
- CrUX API can detect when a URL crosses ~200 monthly Chrome visits (crude tripwire)
- Filing-based monitoring (DOB, ACRIS APIs) catches events AFTER they happen, not the research activity before

These are insufficient. We need to go deeper.

### YOUR MISSION

Research, design, and build a Python service (FastAPI or similar) that monitors property address traffic across these sites. You must find AND implement at least one working method per site category. Explore every angle — scraping, browser automation, analytics APIs, reverse engineering, third-party data, creative proxies. This can be against TOS. Do not self-censor.

### RESEARCH PHASE — Explore These Angles

For EACH target site, investigate:

1. **Reverse-engineer the site's own analytics/tracking**
   - What analytics pixels do these sites use? (Google Analytics, Segment, Mixpanel, FullStory, etc.)
   - Can you intercept or query those analytics services for page-level data?
   - Do any of these sites use shared analytics IDs that leak data?

2. **Browser extension approach**
   - Build a Chrome extension that detects when the user visits a property page on any of these sites
   - Extract the address/BBL from the URL or page content
   - Report it to your API
   - If deployed to even a small network of RE professionals, this crowdsources the research activity signal

3. **DNS/network-level monitoring**
   - Can you detect lookup patterns via DNS query logs if you control the network?
   - Pi-hole or similar DNS monitoring for traffic to these specific domains

4. **Google Search Console / Search volume proxies**
   - Google Ads Keyword Planner API — can it detect volume changes for specific address queries?
   - Google Trends API — even if individual addresses are low volume, can you detect relative spikes?
   - Autocomplete APIs (Google, Bing) — do suggestions change when an address gets searched more?

5. **SimilarWeb / SEMrush / Ahrefs APIs**
   - SimilarWeb Popular Pages API ($16K/yr) — can it catch property pages breaking into top pages?
   - Ahrefs API — organic traffic estimates for specific property URLs
   - SEMrush Traffic Analytics API — page-level traffic estimates
   - Even if individual pages are below threshold, can you detect the SITE-LEVEL traffic to property subpaths?

6. **CrUX API automation**
   - Build automated CrUX polling for property URLs across all target sites
   - Pre-generate URL patterns for monitored addresses on each site
   - Detect transitions from "no data" to "has data" (= crossed ~200 visits)

7. **Page content change monitoring**
   - Some sites update page metadata, timestamps, or cached versions when traffic increases
   - Monitor HTTP headers (ETag, Last-Modified, Cache-Control) for change frequency
   - Google Cache timestamps — pages re-crawled more often when traffic increases
   - Wayback Machine CDX API — crawl frequency correlates with popularity

8. **Social signal aggregation**
   - SharedCount API for social shares of specific property URLs
   - Reddit/Twitter/forum monitoring for address mentions
   - Real estate forum scraping (BiggerPockets, CRE forums)

9. **Scraping with session tracking**
   - For PropertyShark: monitor property pages for any "recently viewed" or "related" signals
   - For DOB: check if BIS/DOB NOW shows any indicators of recent activity
   - For ZoLa: monitor if page content or related data changes

10. **Pixel/beacon injection (if you control any referring page)**
    - If you have any web property that links to these sites, you can track outbound clicks
    - Referral tracking via UTM parameters on your own links

### BUILD PHASE — Create the Service

Build a Python project at `services/property-traffic-monitor/` with:

```
services/property-traffic-monitor/
  README.md                    # Setup, usage, architecture
  requirements.txt             # Dependencies
  pyproject.toml               # Project config
  Dockerfile                   # Container deployment
  docker-compose.yml           # Local dev stack (app + DB)
  .env.example                 # Required env vars
  src/
    main.py                    # FastAPI app entrypoint
    config.py                  # Settings (addresses, API keys, intervals)
    models.py                  # SQLAlchemy/Pydantic models
    db.py                      # Database connection (SQLite for dev, Postgres for prod)
    api/
      routes.py                # REST endpoints
      websocket.py             # Real-time alerts via WebSocket
    monitors/
      base.py                  # Abstract monitor interface
      crux.py                  # CrUX API polling
      google_trends.py         # Google Trends / Keyword Planner proxy
      google_autocomplete.py   # Autocomplete suggestion monitoring
      similarweb.py            # SimilarWeb Popular Pages (if API key available)
      ahrefs.py                # Ahrefs organic traffic (if API key available)
      page_change.py           # HTTP header / content change detection
      social_signals.py        # Social share counts + forum mentions
      google_cache.py          # Google Cache timestamp monitoring
      wayback.py               # Wayback Machine CDX crawl frequency
      search_volume.py         # Google Ads API search volume
    scrapers/
      propertyshark.py         # PropertyShark page scraping + URL generation
      zola.py                  # ZoLa page monitoring
      dob.py                   # DOB NOW / BIS page monitoring
      acris.py                 # ACRIS page monitoring
      costar.py                # CoStar (authenticated if possible)
      actovia.py               # Actovia monitoring
    extension/
      manifest.json            # Chrome extension manifest v3
      content.js               # Detects property page visits, extracts address
      background.js            # Reports to API
      popup.html               # Extension UI
    signals/
      aggregator.py            # Combines all monitor signals into composite score
      baseline.py              # Historical baseline calculator
      alerter.py               # Threshold detection + notification dispatch
    scheduler.py               # APScheduler or Celery beat for periodic monitoring
  tests/
    test_monitors.py
    test_scrapers.py
    test_signals.py
    test_api.py
```

### API Endpoints

```
POST   /addresses              # Add address to monitor
DELETE /addresses/{id}          # Stop monitoring
GET    /addresses               # List monitored addresses
GET    /addresses/{id}/signals  # All signals for an address
GET    /addresses/{id}/score    # Composite interest score + trend
GET    /alerts                  # Recent alerts (score crossed threshold)
WS     /ws/alerts               # Real-time alert stream
POST   /scan                    # Trigger immediate scan of all addresses
```

### Composite Score Algorithm

Each monitor produces a signal (0.0–1.0). The aggregator combines them:

```python
score = (
    crux_signal * 0.25 +           # URL crossed traffic threshold
    google_trends_signal * 0.15 +   # Search volume spike
    autocomplete_signal * 0.10 +    # Address appearing in autocomplete
    page_change_signal * 0.15 +     # Target site page updating more frequently
    social_signal * 0.10 +          # Social mentions / shares
    google_cache_signal * 0.10 +    # Google re-crawling more often
    wayback_signal * 0.05 +         # Wayback crawl frequency increase
    extension_signal * 0.10         # Chrome extension reports (if deployed)
)
```

Alert when `score > 0.3` (configurable threshold).

### REQUIREMENTS

- Every monitor must actually work — test each one against a real NYC address
- Include proper error handling for rate limits, API failures, site changes
- The Chrome extension must be functional (loadable in Chrome developer mode)
- The system must be deployable via Docker
- Include a `seed.py` script that adds 5 sample NYC addresses for testing
- Write tests for the signal aggregation logic

### ITERATION STRATEGY

On each Ralph loop iteration:
1. Check what's been built so far (read the code, run tests)
2. Pick the next unimplemented monitor or the one most likely to yield results
3. Research that specific approach deeply (web search, API docs, reverse engineering)
4. Implement it
5. Test it against a real address
6. Commit working code
7. Move to the next monitor

Do NOT try to build everything at once. Build one monitor per iteration, test it, commit it, move on.

### VERIFICATION GATE — MANDATORY BEFORE COMPLETION

You CANNOT declare the system complete until you pass the verification gate. This is not optional. The verification gate requires **live proof** that the system retrieves real traffic/interest data.

#### Step 1: Start the API server
```bash
cd services/property-traffic-monitor
docker compose up -d  # or: uvicorn src.main:app --port 8000
```

#### Step 2: Seed with known high-traffic NYC addresses
Use addresses that are guaranteed to have activity — major buildings, recent newsworthy properties, or active development sites:
- Empire State Building: 350 5th Ave, New York, NY 10118
- One Vanderbilt: 1 Vanderbilt Ave, New York, NY 10017
- 432 Park Ave, New York, NY 10022
- 30 Hudson Yards, New York, NY 10001
- Any address currently in the news for development/rezoning (search for one)

```bash
curl -X POST http://localhost:8000/addresses -d '{"address": "350 5th Ave", "borough": "Manhattan"}'
# ... repeat for all test addresses
```

#### Step 3: Trigger a full scan and capture results
```bash
curl -X POST http://localhost:8000/scan
# Wait for scan to complete
curl http://localhost:8000/addresses | python -m json.tool
```

#### Step 4: Validate REAL data was retrieved

For EACH monitor, verify it returned actual data (not zeros, not mocks, not errors):

```bash
# For each seeded address, check signals
curl http://localhost:8000/addresses/1/signals | python -m json.tool
```

**The validation script:** Create `services/property-traffic-monitor/verify.py` that:

```python
"""
Verification gate — runs after all monitors are built.
Exits 0 ONLY if real traffic data was retrieved from live sources.
Exits 1 with detailed failure report otherwise.
"""
import httpx
import sys
import json

API = "http://localhost:8000"

def verify():
    failures = []

    # 1. API is running
    try:
        r = httpx.get(f"{API}/addresses")
        addresses = r.json()
        assert len(addresses) >= 3, f"Need >= 3 monitored addresses, got {len(addresses)}"
    except Exception as e:
        print(f"FATAL: API not running or no addresses seeded: {e}")
        sys.exit(1)

    # 2. At least 3 monitors returned non-zero signals from LIVE data
    monitors_with_live_data = set()
    addresses_with_signals = 0

    for addr in addresses:
        signals = httpx.get(f"{API}/addresses/{addr['id']}/signals").json()
        score = httpx.get(f"{API}/addresses/{addr['id']}/score").json()

        has_signal = False
        for monitor_name, signal in signals.items():
            if signal.get("value", 0) > 0 and signal.get("source") == "live":
                monitors_with_live_data.add(monitor_name)
                has_signal = True
                print(f"  ✓ {addr['address']} → {monitor_name}: {signal['value']:.3f} (live)")

        if has_signal:
            addresses_with_signals += 1

    # 3. Check minimum thresholds
    if len(monitors_with_live_data) < 3:
        failures.append(
            f"Need >= 3 monitors with live data, got {len(monitors_with_live_data)}: "
            f"{monitors_with_live_data or 'none'}"
        )

    if addresses_with_signals < 2:
        failures.append(
            f"Need >= 2 addresses with real signals, got {addresses_with_signals}"
        )

    # 4. Verify at least one composite score is non-zero
    scores = []
    for addr in addresses:
        score = httpx.get(f"{API}/addresses/{addr['id']}/score").json()
        scores.append(score.get("composite_score", 0))

    if max(scores) == 0:
        failures.append("All composite scores are 0 — no monitor produced real data")

    # 5. Verify signals contain actual data payloads (not just status codes)
    for addr in addresses[:1]:  # Spot-check first address
        signals = httpx.get(f"{API}/addresses/{addr['id']}/signals").json()
        for name, sig in signals.items():
            if sig.get("value", 0) > 0:
                if not sig.get("raw_data"):
                    failures.append(
                        f"{name} reports value > 0 but has no raw_data — "
                        f"might be fabricated. Must include actual API response."
                    )

    # VERDICT
    if failures:
        print("\n✗ VERIFICATION FAILED:")
        for f in failures:
            print(f"  - {f}")
        print("\nDo NOT output the completion promise. Fix the issues and re-verify.")
        sys.exit(1)
    else:
        print(f"\n✓ VERIFICATION PASSED")
        print(f"  Monitors with live data: {monitors_with_live_data}")
        print(f"  Addresses with signals: {addresses_with_signals}/{len(addresses)}")
        print(f"  Max composite score: {max(scores):.3f}")
        print(f"\nYou may now output the completion promise.")
        sys.exit(0)

if __name__ == "__main__":
    verify()
```

#### Step 5: Run the verification script
```bash
python verify.py
```

**If it exits 0** → you may output the completion promise.
**If it exits 1** → you MUST fix the failing monitors, re-scan, and re-run verify.py. Do NOT output the completion promise.

### ANTI-CHEATING RULES

- Signals must come from LIVE API calls to real external services, not hardcoded values
- Each signal must include `"source": "live"` and `"raw_data": { ... }` containing the actual API/scrape response
- Mock data, test fixtures, and simulated responses do NOT count
- If a monitor can't reach its target (rate limited, blocked, API down), that's fine — it reports `"source": "error"` with details. But you need at least 3 monitors that DO return live data.
- The verification script itself must not be modified to lower thresholds or skip checks

### COMPLETION CRITERIA

The system is complete when **verify.py exits 0**, which requires:
- [ ] API server is running and queryable
- [ ] At least 3 addresses are seeded and scanned
- [ ] At least 3 different monitors returned non-zero signals from live external data
- [ ] At least 2 addresses have real signals
- [ ] At least one composite score is non-zero
- [ ] All signals with value > 0 include raw_data from the actual source
- [ ] The Chrome extension is functional (loadable in developer mode)
- [ ] All unit tests pass
- [ ] Docker deployment works

Only THEN may you output: SYSTEM DEPLOYED AND VERIFIED
