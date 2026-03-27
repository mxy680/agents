Run the daily LLC formation scan, analyze each candidate using your research tools, and produce a professional PDF report.

## Step 1 — Pull candidates

```bash
bash agents/real-estate/scripts/llc/run_scan.sh
```

This pulls all new LLC formations from NY DOS (last 3 days) and applies broad filtering to exclude obvious non-real-estate entities (restaurants, salons, tech companies, etc.). The result is a candidate list — NOT a final list. Your job is to analyze each candidate and determine which ones are actually real estate transactions.

Read the candidates from `/tmp/llc_monitor/candidates_$(date +%Y-%m-%d).json`.

## Step 2 — Analyze each candidate

For each candidate, use your reasoning and research tools to determine if it's a real estate acquisition:

### Strong signals (high confidence):
- **Entity name contains a NYC street address** (e.g., "540 WEST 29 LLC", "1776 SEMINOLE AVE LLC")
  - Geocode the address to verify it exists: `curl -s "https://geosearch.planninglabs.nyc/v2/search?text=ADDRESS+BOROUGH+NY"`
  - If the address resolves, look up PLUTO to check zoning: `curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=BBL"`
  - Check ACRIS for recent deed activity on that BBL
- **Filer is a known real estate law firm** — firms that file many address-style LLCs are deal shops
- **Entity name contains real estate keywords** (REALTY, HOLDINGS, DEVELOPMENT, PROPERTIES, etc.)

### Weak signals (needs more investigation):
- **Short numeric name** (e.g., "2964 LLC") — could be a block number, lot number, or unrelated
- **Borough name + LLC** — could be a real estate entity or just a business in that borough
- **Contains "CAPITAL" or "VENTURES"** — could be real estate investment or unrelated finance

### What to look for in your analysis:
1. **Does the entity name reference a real NYC address?** Verify via geocoding.
2. **Is the address in an R7+ zone?** Check PLUTO zoning.
3. **Are there any active listings or recent sales at that address?** Check ACRIS.
4. **Is the filer associated with other real estate LLCs?** Look for patterns in filer names.
5. **Does the entity name match common real estate LLC naming conventions?** (address + LLC, address + HOLDINGS, etc.)

### Batch your research efficiently:
- Process all "strong" matches first (address patterns, RE keywords)
- For "possible" matches, quickly check if the name could be an address — if not, skip
- Don't spend time on entities that are clearly not real estate after a quick look

## Step 3 — Generate PDF report

Write a LaTeX file to `/tmp/llc_monitor/report.tex` and compile with `pdflatex -interaction=nonstopmode`.

The report should include:

### Title Page
"NYC LLC Formation Intelligence Report" with date, "Confidential — Prepared for [broker]"

### Executive Summary
- Date range scanned
- Total new LLCs filed
- Candidates identified (strong + possible)
- Confirmed real estate entities after your analysis
- Key findings (most interesting discoveries)

### Confirmed Real Estate Entities (main section)
For each entity you've confirmed as real estate, include a detailed entry:

| Field | Content |
|-------|---------|
| Entity Name | e.g., "540 WEST 29TH STREET LLC" |
| Formation Date | Filing date |
| Property Address | The actual NYC address (if identified) |
| Borough / Neighborhood | Where the property is |
| Zoning | From PLUTO (R7-2, R8A, C4-4D, etc.) |
| Lot Area | From PLUTO |
| Current Use | Year built, building class, units |
| Recent ACRIS Activity | Any recent deeds, mortgages, judgments |
| Filer / Attorney | Name and address — who filed the LLC |
| Analysis | Your assessment: what's happening here (acquisition, development, holding company?) |
| Recommended Action | What the broker should do next |

### Watchlist (secondary section)
Entities that are ambiguous — might be real estate but you couldn't confirm. Brief table with entity name, reason flagged, and why it's uncertain.

### Pattern Analysis
- Which law firms/filers appeared most frequently
- Geographic clustering (which neighborhoods are active)
- Naming convention trends

### Methodology
Data source, date range, analysis approach

Use navy blue headers (`\definecolor{navy}{RGB}{0,0,128}`), booktabs tables, Helvetica font, fancyhdr. Escape LaTeX special characters (`$`, `#`, `%`, `&`, `_`).

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/llc_monitor/report.pdf --name="LLC Formation Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 4 — Report results

Summarize: total LLCs scanned, confirmed real estate entities found, top 3 most interesting finds with property details, and the Google Drive link.
