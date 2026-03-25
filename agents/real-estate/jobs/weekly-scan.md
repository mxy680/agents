Your task is to run the NYC assemblage intelligence pipeline and deliver the results.

## What to do

The pipeline scripts are pre-built in `agents/real-estate/scripts/`. Do NOT rewrite them. Just run them and handle any errors.

### Step 1 — Run the pipeline

```bash
bash agents/real-estate/scripts/run_pipeline.sh
```

This runs 9 phases in sequence:
1. Zillow search across 137 NYC zip codes (handles 40-result cap with price splits)
2. PLUTO geocoding + R7+ zoning filter
3. Signal checks — ACRIS, DOB, HPD, NYC Finance, 311, ECB/OATH, FDNY vacate, DOB complaints, Certificates of Occupancy, Citi Bike density, NY SLA liquor licenses
4. StreetEasy price history enrichment (if cookies available)
5. Cluster detection (same-block groupings)
6. Data verification (duplicates, missing fields, scoring)
7. XLSX spreadsheet generation
8. LaTeX PDF report
9. Upload all deliverables to Google Drive

### Step 2 — Monitor and fix errors

Watch the pipeline output. If any phase fails:
- Read the error message
- Fix the issue (usually a data format problem or API error)
- Rerun just that phase: `python3 agents/real-estate/scripts/phaseN_name.py`
- Then continue with the remaining phases

### Step 3 — Report results

After the pipeline completes, summarize:
- How many zip codes searched
- How many Zillow listings found
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
- If Zillow cookies are expired (403 errors), tell the user to refresh cookies in the portal and stop.
