Run the daily LLC formation scan and produce a professional PDF report.

## Step 1 — Run the scan

```bash
bash agents/real-estate/scripts/llc/run_scan.sh
```

This scans NY DOS for new LLC formations that match real estate naming patterns.

## Step 2 — Generate PDF report

After the scan completes, read the results from `/tmp/llc_monitor/matches_$(date +%Y-%m-%d).json` and compile a professional PDF report.

Write a LaTeX file to `/tmp/llc_monitor/report.tex` and compile with `pdflatex`.

The report should include:
- **Title page**: "NYC LLC Formation Intelligence Report" with date, "Confidential"
- **Executive summary**: Total new LLCs scanned today, how many matched real estate patterns, breakdown by match type (address pattern, keyword, borough indicator)
- **Top matches**: For each real estate match:
  - Entity name (e.g., "45 SEMINOLE LLC")
  - Formation date
  - Match reason (why it was flagged — address pattern, keyword like REALTY/HOLDINGS, borough reference)
  - Filer name (the attorney or formation service — reveals the buyer's law firm)
  - Filer/process address (mailing address on file)
  - Analysis: what this entity name likely references (property address, development project, holding company)
- **Pattern analysis**: Summary of naming patterns observed — how many use street addresses, how many use real estate keywords, which boroughs are most active
- **How to use this data**: Brief guide on next steps — search ACRIS for the address, look up the filer's other entities, identify the buyer before the deed records
- **Methodology**: Data source (NY DOS daily filings), pattern matching approach, date range

Use navy blue headers, booktabs tables, Helvetica font. Keep it actionable — each match should tell the reader what to do next.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/llc_monitor/report.pdf --name="LLC Formation Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 3 — Report results

Summarize: total LLCs scanned, real estate matches found, top 3 matches with entity names and filers, and the Google Drive link. If no matches, say so clearly.
