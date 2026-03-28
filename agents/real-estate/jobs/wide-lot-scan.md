Search PLUTO for wide-lot 1-2 family homes in the Bronx, deeply research the top candidates, and produce a professional PDF report.

## Step 1 — Query PLUTO

Search for 1-2 family homes with 50ft+ lot frontage in the Bronx. Start with R8+ zoning:

```bash
curl -s -G "https://data.cityofnewyork.us/resource/64uk-42ks.json" \
  --data-urlencode "\$where=borocode='2' AND lotfront >= 50 AND unitsres <= 2 AND (bldgclass like 'A%' OR bldgclass like 'B%') AND (zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%')" \
  --data-urlencode "\$select=bbl,address,zonedist1,lotarea,lotfront,lotdepth,bldgarea,yearbuilt,bldgclass,unitsres,numfloors,ownername" \
  --data-urlencode "\$limit=500" \
  --data-urlencode "\$order=lotfront DESC"
```

If R8+ returns fewer than 20 results, expand to include R7+:

```bash
curl -s -G "https://data.cityofnewyork.us/resource/64uk-42ks.json" \
  --data-urlencode "\$where=borocode='2' AND lotfront >= 50 AND unitsres <= 2 AND (bldgclass like 'A%' OR bldgclass like 'B%') AND (zonedist1 like 'R7%' OR zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%')" \
  --data-urlencode "\$select=bbl,address,zonedist1,lotarea,lotfront,lotdepth,bldgarea,yearbuilt,bldgclass,unitsres,numfloors,ownername" \
  --data-urlencode "\$limit=500" \
  --data-urlencode "\$order=lotfront DESC"
```

## Step 2 — Deep research the top 30 properties (by lot frontage)

For the 30 widest lots, do FULL research on each. Do NOT just list PLUTO data — actually investigate each property.

### A. ACRIS Research
For each BBL, check ACRIS for:
- **Last deed transfer** — when was it sold, for how much, who bought it?
- **Recent mortgages** — is there active financing? A satisfied mortgage (SAT) means the owner is unencumbered.
- **Judgments/liens** — lis pendens = foreclosure, tax lien sale certificates = distressed
- **Estate filings** — any "ESTATE OF" or "EXECUTOR" in the party names?

```bash
curl -s -G "https://data.cityofnewyork.us/resource/8h5j-fqxa.json" \
  --data-urlencode "borough=2" --data-urlencode "block=BLOCK" --data-urlencode "lot=LOT" \
  --data-urlencode "\$limit=30" --data-urlencode "\$order=recorded_datetime DESC"
```

### B. HPD Violations
```bash
curl -s -G "https://data.cityofnewyork.us/resource/csn4-vhvf.json" \
  --data-urlencode "\$select=count(*)" \
  --data-urlencode "\$where=boroid='2' AND block='BLOCK' AND lot='LOT'"
```

### C. NYC Finance — Tax Liens
```bash
curl -s "https://data.cityofnewyork.us/resource/9rz4-mjek.json?\$where=bbl='BBL'"
```

### D. DOB — Permits on Same Block
Check for demolition or new building permits that indicate development activity:
```bash
curl -s "https://data.cityofnewyork.us/resource/ipu4-2q9a.json?\$where=borough='BRONX' AND block='BLOCK' AND job_type in('DM','NB')&\$limit=10&\$order=issuance_date DESC"
```

### E. Cluster Detection
Check if multiple qualifying properties are on the same block:
- Group all 500 results by block number
- Any block with 2+ qualifying properties = assemblage opportunity
- Note these clusters prominently in the report

### F. Owner Analysis
From PLUTO `ownername`:
- Is the owner an individual or LLC?
- If LLC, does the name match the address? (e.g., "123 MAIN ST LLC" = likely investor holding)
- If individual, is it an elderly owner? (pre-war building, long ownership = estate risk)
- Search ACRIS for how long the current owner has held the property

## Step 3 — Generate PDF report

Write a LaTeX file and compile with `pdflatex -interaction=nonstopmode`.

### Title Page
"Bronx Wide Lot Development Opportunities" with date, "Confidential"

### Executive Summary
- Total qualifying properties found
- Zoning breakdown (R8+ vs R7)
- Average lot frontage, average lot area
- Properties with distress signals
- Key findings — the 3 most interesting properties and WHY
- Cluster opportunities identified

### Why Wide Lots Matter
Brief explanation of why 50ft+ frontage is valuable for development.

### Top Opportunities — Detailed Analysis (top 10-15)
For each top property, a UNIQUE detailed entry:

| Field | Content |
|-------|---------|
| Address | Full address |
| Current Owner | From PLUTO — **bold** |
| Zoning | District |
| Lot Frontage / Depth | Dimensions |
| Lot Area | SF |
| Year Built | Age |
| Building Class | Current use |
| Last Sale | Date and price from ACRIS |
| Current Financing | Active mortgages |
| HPD Violations | Count (flag 5+ as distress signal) |
| Tax Lien Status | From NYC Finance |
| Block Activity | Demo/NB permits on same block |
| Analysis | **UNIQUE assessment** — what makes this specific property interesting |
| Recommended Action | **SPECIFIC next step** |

**IMPORTANT**: The analysis for each property must be UNIQUE. Reference actual data: sale history, owner type, violation count, block activity, neighborhood context. Do NOT write "Pre-war (1930); likely obsolete" for every property.

Good analysis examples:
- "Last sold for $450K in 2003 to an individual owner. 23 years of ownership with no recent mortgage activity suggests an unencumbered, aging owner. The 80ft frontage on R7-1 zoning could support a 7-story, 30+ unit building. Two demolition permits were issued on the same block in 2025 — active development corridor."
- "Owned by 861 HORNADAY LLC since 2018 acquisition for $620K. The LLC structure and recent purchase suggest an investor hold. No HPD violations (well-maintained). The 66ft lot in R7-1 is valuable but the recent acquisition makes the owner less likely to sell at a discount."

### Cluster Opportunities
Table of blocks with 2+ qualifying properties:
- Block number, addresses, combined frontage, combined lot area
- Why the cluster matters (assemblage potential)

### Full Property Table
All remaining properties (not enriched) in a compact table: address, zone, frontage, lot area, year built, owner.

### Methodology
PLUTO query criteria, signal sources, scoring model.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/wide_lot_scan/report.pdf --name="Bronx Wide Lot Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 4 — Report results

Summarize: total properties, top 3 with specific analysis (not just dimensions), cluster opportunities, and the Google Drive link.
