Run the off-market R8+ scanner pipeline and produce a professional PDF report.

## Step 1 — Run the pipeline

```bash
bash agents/real-estate/scripts/off-market/run_pipeline.sh
```

This scans all R8+ zoned small residential properties across NYC for distress signals — no Zillow dependency.

## Step 2 — Monitor errors

If any phase fails, read the error, fix, and rerun that phase. Phases are independent — a failure in one doesn't stop the rest.

## Step 3 — Generate PDF report

After the pipeline completes, compile a professional PDF report. Write a LaTeX file to `/tmp/off_market_scan/report.tex` and compile with `pdflatex`.

The report should include:
- **Title page**: "NYC Off-Market R8+ Intelligence Report" with date, "Confidential — For Internal Use Only"
- **Executive summary**: Total properties scanned, how many had distress signals, score distribution (Immediate/High/Moderate/Watchlist), borough breakdown
- **Top priority properties** (score 15+): For each, show address, borough, zoning, lot area, year built, composite score, and a paragraph explaining WHY this property is a target (which signals fired and what they mean)
- **Signal summary table**: Count of each signal type across all properties (tax liens, lis pendens, HPD 10+, estate signals, FDNY vacate, etc.)
- **Cluster opportunities**: Blocks with 2+ qualifying properties — addresses, combined lot area, combined score
- **Full results table**: Top 50 properties sorted by score — address, borough, zoning, score, key signals
- **Methodology**: Brief description of data sources, scoring model, and priority tiers

Use navy blue headers, booktabs tables, Helvetica font, proper escaping of special characters.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/off_market_scan/report.pdf --name="Off-Market R8+ Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 4 — Report results

Summarize: total properties, signal counts, score distribution, top 3 properties with reasons, cluster opportunities, and the Google Drive link.
