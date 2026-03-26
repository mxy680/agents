Search PLUTO for wide-lot 1-2 family homes in the Bronx and produce a professional PDF report.

## Step 1 — Query PLUTO

Search for 1-2 family homes with 50ft+ lot frontage in the Bronx. Start with R8+ zoning:

```bash
curl -s -G "https://data.cityofnewyork.us/resource/64uk-42ks.json" \
  --data-urlencode "\$where=borocode='2' AND lotfront >= 50 AND unitsres <= 2 AND (bldgclass like 'A%' OR bldgclass like 'B%') AND (zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%')" \
  --data-urlencode "\$select=bbl,address,zonedist1,lotarea,lotfront,lotdepth,bldgarea,yearbuilt,bldgclass,unitsres,numfloors" \
  --data-urlencode "\$limit=500" \
  --data-urlencode "\$order=lotfront DESC"
```

If R8+ returns fewer than 20 results, expand to include R7+:

```bash
curl -s -G "https://data.cityofnewyork.us/resource/64uk-42ks.json" \
  --data-urlencode "\$where=borocode='2' AND lotfront >= 50 AND unitsres <= 2 AND (bldgclass like 'A%' OR bldgclass like 'B%') AND (zonedist1 like 'R7%' OR zonedist1 like 'R8%' OR zonedist1 like 'R9%' OR zonedist1 like 'R10%')" \
  --data-urlencode "\$select=bbl,address,zonedist1,lotarea,lotfront,lotdepth,bldgarea,yearbuilt,bldgclass,unitsres,numfloors" \
  --data-urlencode "\$limit=500" \
  --data-urlencode "\$order=lotfront DESC"
```

## Step 2 — Enrich with signals

For each property, check:
- ACRIS for recent deed transfers, lis pendens, estate signals
- HPD for open violations
- NYC Finance for tax liens
- DOB for permits on the block

Use the same curl patterns from the CLAUDE.md documentation.

## Step 3 — Generate PDF report

Write a LaTeX file and compile with `pdflatex`. The report should include:

- **Title page**: "Bronx Wide Lot Development Opportunities" with date, "Confidential"
- **Executive summary**: Total qualifying properties found, zoning breakdown (R8+ vs R7), average lot frontage, average lot area
- **Why wide lots matter**: Brief explanation — 50ft+ frontage allows larger as-of-right development, more efficient floor plates, better unit layouts, higher per-unit values
- **Property details table**: For each property sorted by lot frontage:
  - Address, zoning, lot frontage (ft), lot depth (ft), lot area (SF), year built, units, building class
  - Any distress signals (tax liens, HPD violations, estate)
  - ZoLa link
- **Top opportunities**: The 5-10 widest lots with analysis — why each is interesting for development (zoning potential, age, condition signals)
- **Methodology**: PLUTO query criteria, signal sources

Use navy blue headers, booktabs tables, Helvetica font.

Upload the PDF to Google Drive:
```bash
integrations drive files upload --path=/tmp/wide_lot_scan/report.pdf --name="Bronx Wide Lot Report — $(date +%Y-%m-%d).pdf" --json
```

## Step 4 — Report results

Summarize: total properties found (R8+ count and R7+ count), widest lots, any with distress signals, and the Google Drive link.
