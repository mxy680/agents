Run the daily probate scan and produce a professional PDF report.

## Step 1 — Run the scan

```bash
bash agents/real-estate/scripts/probate/run_scan.sh
```

This monitors ACRIS for estate/probate filings on properties with 15K+ buildable SF or 5+ units.

## Step 2 — Generate PDF report

After the scan completes, read the results from `/tmp/probate_monitor/alerts_$(date +%Y-%m-%d).json` and compile a professional PDF report.

Write a LaTeX file to `/tmp/probate_monitor/report.tex` and compile with `pdflatex`.

The report should include:
- **Title page**: "NYC Probate Property Alert Report" with date, "Confidential"
- **Executive summary**: How many estate documents were found in ACRIS, how many BBLs resolved, how many passed the qualification filter (15K+ SF or 5+ units), how many are new (not seen before)
- **Alert details**: For each qualifying property:
  - Address, borough, BBL
  - Estate name (from ACRIS party records)
  - Filing date
  - Property details: zoning, lot area, building area, units residential, year built
  - Why it matters: explain what probate means for this property (heirs want to sell, estate liquidation timeline, opportunity window)
  - ZoLa link for zoning context
- **Market context**: Brief note on why probate properties are valuable (>50% sell, heirs are motivated, below-market pricing typical)
- **Methodology**: Data sources (ACRIS parties table, PLUTO), qualification criteria, deduplication approach

Use navy blue headers, booktabs tables, Helvetica font. Keep it concise — this is a daily alert, not a weekly deep-dive.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/probate_monitor/report.pdf --name="Probate Alert Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 3 — Report results

Summarize: number of new alerts, top 3 properties with estate names and unit counts, and the Google Drive link. If no new alerts, say so clearly.
