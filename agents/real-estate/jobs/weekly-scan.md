Your task is to run the NYC assemblage intelligence pipeline and deliver the results.

## Prerequisites

Before triggering this job, the user must have already run the Zillow scrape from the Chrome extension. The extension scrapes all 137 NYC ZIP codes and saves results to `/tmp/nyc_assemblage/zillow_results.json`. If that file doesn't exist or is empty, tell the user to run the extension scrape first and stop.

## What to do

### Step 1 — Verify Zillow data exists

```bash
ls -la /tmp/nyc_assemblage/zillow_results.json
python3 -c "import json; d=json.load(open('/tmp/nyc_assemblage/zillow_results.json')); print(f'{len(d)} listings')"
```

If the file doesn't exist or has 0 listings, stop and tell the user:
> "Run the Zillow scrape from the Chrome extension first. Click the extension icon and press Scrape Zillow."

### Step 2 — Run the pipeline

```bash
bash agents/real-estate/scripts/run_pipeline.sh
```

This runs 9 phases in sequence:
1. **Zillow search** — reads pre-scraped results from the extension (skips CLI)
2. **PLUTO geocoding** — geocodes addresses to BBL, filters to R7+ zoning
3. **Signal checks** — ACRIS, DOB, HPD, NYC Finance, 311, ECB/OATH, FDNY vacate, DOB complaints, Certificates of Occupancy, Citi Bike density (all free public APIs, no auth needed)
4. **StreetEasy enrichment** — price history for properties with score >= 3
5. **Cluster detection** — same-block groupings
6. **Data verification** — duplicates, missing fields, scoring consistency
7. **XLSX spreadsheet** — color-coded, multi-sheet workbook
8. **LaTeX PDF report** — professional report with detailed property analysis
9. **Upload** — XLSX + PDF to Google Drive

### Step 3 — Monitor and fix errors

Watch the pipeline output. If any phase fails:
- Read the error message
- Fix the issue (usually a data format problem or API timeout)
- Rerun just that phase: `python3 agents/real-estate/scripts/phaseN_name.py`
- Then continue with the remaining phases

Common issues:
- Phase 3 signal checks may timeout on some APIs — they are wrapped in try/except and will skip gracefully
- Phase 4 StreetEasy may fail if cookies are expired — non-critical, pipeline continues without StreetEasy data
- Phase 8 LaTeX compilation may fail — check for unescaped special characters, fix and recompile

### Step 4 — Report results

After the pipeline completes, summarize:
- How many Zillow listings the extension scraped
- How many passed R7+ filter
- How many had distress signals
- Score distribution (Immediate / High / Moderate / Watchlist)
- Top 3 properties and why
- Cluster opportunities
- Links to all deliverables (Google Sheet, PDF)

## Execution rules
- Do NOT rewrite the pipeline scripts. They are permanent and tested.
- Do NOT spawn sub-agents.
- If a phase fails, diagnose and fix the specific issue, then rerun that phase.
- Phase 1 should always use the pre-scraped extension data. If it falls back to CLI and gets 403 errors, stop and tell the user to run the extension scrape.
