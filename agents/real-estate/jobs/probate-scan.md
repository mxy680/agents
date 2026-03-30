Run the daily probate scan, deeply research each qualifying property, and produce a professional PDF report.

## Step 1 — Run the scan

```bash
bash agents/real-estate/scripts/probate/run_scan.sh
```

This monitors ACRIS for estate/probate filings on properties with 15K+ buildable SF or 5+ units.

Read the results from `/tmp/probate_monitor/alerts_$(date +%Y-%m-%d).json`.

## Step 2 — Deep research each property

For EACH qualifying property, do the following research. Do NOT use generic boilerplate — every property should have a unique, specific analysis.

### A. ACRIS Deep Dive
```bash
curl -s -G "https://data.cityofnewyork.us/resource/8h5j-fqxa.json" \
  --data-urlencode "borough=BORO" --data-urlencode "block=BLOCK" --data-urlencode "lot=LOT" \
  --data-urlencode "\$limit=50" --data-urlencode "\$order=recorded_datetime DESC"
```
Then look up the master table for document details. Look for:
- **Last sale date and price** — when was the property last sold and for how much?
- **Recent mortgages** — is there financing in place? How much?
- **Judgments/liens** — any lis pendens, tax liens, federal liens?
- **Other estate filings** — are there multiple estate documents (suggests complex probate)?

### B. HPD Violations
Check open violations — a heavily violated building is harder to sell and more likely to trade at a discount:
```bash
curl -s -G "https://data.cityofnewyork.us/resource/csn4-vhvf.json" \
  --data-urlencode "\$select=count(*)" \
  --data-urlencode "\$where=boroid='BORO' AND block='BLOCK' AND lot='LOT'"
```

### C. Current Owner Analysis
From PLUTO `ownername`:
- Is the owner an individual (personal estate) or a corporation/LLC (entity restructuring)?
- Does the owner name match the estate name? (If yes, the deceased was the direct owner — simpler probate)
- Is the owner a management company? (If yes, the deceased may have been a passive investor — heirs may want to liquidate quickly)

### D. Property Classification
Classify each property into one of these categories:
- **Trophy Asset** — 100+ units in Manhattan/prime Brooklyn, R8+, institutional-quality. Buyers: REITs, private equity.
- **Mid-Market Multifamily** — 20-99 units, any borough. Buyers: local operators, syndicators.
- **Small Portfolio** — 5-19 units. Buyers: individual investors, family offices.
- **Development Site** — large lot with underbuilt structure in high-density zone. Buyers: developers.

### E. Actionable Assessment
For each property, write a SPECIFIC analysis (not generic). Include:
- What makes this property valuable RIGHT NOW
- Who the likely buyers are
- What the timeline looks like (early probate = 6-18 months, late probate = imminent)
- Specific recommended action (contact executor, monitor ACRIS for deed, approach management company)
- Any risks (complex ownership, rent stabilization, environmental issues)

## Step 3 — Generate PDF report

Write a LaTeX file to `/tmp/probate_monitor/report.tex` and compile with `pdflatex -interaction=nonstopmode`.

The report should include:

### Title Page
"NYC Probate Property Alert Report" with date, "Confidential"

### Executive Summary
- Scan metrics (documents found, BBLs resolved, qualifying properties)
- Borough distribution
- Key findings — the 2-3 most interesting properties and WHY
- Total estimated value of the portfolio (if sale prices are available from ACRIS)

### Alert Details (main section)
For each property, a detailed entry with:

| Field | Content |
|-------|---------|
| Address | Full address with borough |
| Priority | HIGH (100+ units or R8+) / MODERATE (20-99 units) / STANDARD (5-19 units) |
| Estate Name | From ACRIS |
| Filing Date | When the estate document was recorded |
| Current Owner | From PLUTO — bold this |
| Property Type | Classification (Trophy/Mid-Market/Small Portfolio/Dev Site) |
| Zoning | From PLUTO |
| Lot Area / Building Area | Square footage |
| Units | Residential unit count |
| Year Built / Floors | Building age and size |
| Last Sale | Date and price from ACRIS (if available) |
| Current Financing | Recent mortgages from ACRIS |
| HPD Violations | Count of open violations |
| ACRIS Activity | Summary of recent deed/mortgage/judgment activity |
| Analysis | YOUR SPECIFIC ASSESSMENT — not generic boilerplate |
| Recommended Action | SPECIFIC next step |

**IMPORTANT**: The analysis and recommended action for each property MUST be unique and specific. Do not copy-paste the same paragraph for every property. Reference the actual data: sale price, unit count, violation count, owner type, neighborhood context.

### Market Context
- Why probate properties represent acquisition opportunities
- Typical timeline from filing to sale
- Pricing dynamics (estate sales typically close 10-20% below market)

### Methodology
Data sources, qualification criteria, date range

Use navy blue headers, booktabs tables, Helvetica font. Color-code priorities: HIGH in red, MODERATE in orange, STANDARD in blue.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/probate_monitor/report.pdf --name="Probate Alert Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 4 — Report results

Summarize: total alerts, top 3 properties with specific details (not just unit counts — include last sale price, owner type, why it's interesting), and the Google Drive link.
