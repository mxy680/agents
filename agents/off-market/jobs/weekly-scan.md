# Weekly Off-Market R8+ Scan

Run the off-market pipeline:

```bash
bash agents/off-market/scripts/run_pipeline.sh
```

## Pipeline Phases

1. **PLUTO query** — pull all R8+ small residential properties (paginated across all 5 boroughs)
2. **Signal checks** — for each property: ACRIS (lis pendens, federal liens, tax lien certs, estate signals), DOB (permits, scaffolding), HPD (violations), NYC Finance (tax liens), 311 complaints, ECB violations, FDNY vacate orders, DOB complaints, Certificate of Occupancy, Citi Bike density
3. **Scoring** — composite distress score (no Zillow or listing signals — distress only)
4. **Cluster detection** — group properties by block; flag blocks with multiple distress signals
5. **Verification** — deduplicate by BBL, check scoring consistency, fix any data issues
6. **XLSX spreadsheet** — all properties with signals and scores, color-coded by priority tier
7. **PDF report** — professional summary with top opportunities and methodology
8. **Upload to Google Drive** — XLSX as native Google Sheet, PDF as-is

## Report Results

After the pipeline completes, summarize:
- Total R8+ small residential properties found in PLUTO
- How many had at least one distress signal
- Score distribution (20+, 15-19, 10-14, 5-9, <5)
- Top 5 properties by composite score with key signals
- Google Drive links to the XLSX and PDF
