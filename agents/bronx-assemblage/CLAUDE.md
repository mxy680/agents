# Bronx Assemblage Scout — Tool Documentation

## Authentication
All credentials are pre-configured via environment variables. Run commands directly — do not check for missing tokens.

## Tool 1: Zillow CLI (search only)

Search for properties:
```bash
integrations zillow properties search --location="Bronx, NY 10451" --limit=40 --json
```

The search returns: zpid, address, price, beds, baths, sqft, homeType, status, zillowUrl, latitude, longitude, daysOnMarket.

**Do NOT use `integrations zillow properties get`** — the detail endpoint has CSRF issues. Use NYC PLUTO (Tool 2) for lot size, year built, and building area instead.

## Tool 2: NYC PLUTO + GeoSearch (zoning, lot data, year built)

PLUTO is the authoritative source for lot data. It returns zoning, lot area, building area, year built, and building class — all in one call.

### Step A: Geocode the address to get BBL

```bash
curl -s "https://geosearch.planninglabs.nyc/v2/search?text=1776+Seminole+Ave+Bronx+NY"
```

The BBL is in `.features[0].properties.addendum.pad.bbl`. Example: `2037620044` where `2` = Bronx.

### Step B: Look up lot data via PLUTO (Socrata)

```bash
curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=2037620044"
```

Returns: `zonedist1` (zoning), `lotarea` (lot SF), `bldgarea` (building SF), `yearbuilt`, `bldgclass`, `numbldgs`, `numfloors`, `unitsres`, `unitstotal`, `address`.

This is more reliable than Zillow for lot size and year built.

### Step C: Build the ZoLa URL

```
https://zola.planning.nyc.gov/lot/2/03762/0044
```

Format: `/lot/{borough}/{block}/{lot}` — parse from BBL where:
- Digit 1: borough (2 = Bronx)
- Digits 2-6: block (5 digits, zero-padded)
- Digits 7-10: lot (4 digits, zero-padded)

### Zoning classification guide

Only include properties zoned R7 or higher:
- **R7, R7-1, R7-2** — medium-density residential
- **R7A, R7B, R7D, R7X** — contextual medium-density
- **R8, R8A, R8B, R8X** — high-density residential
- **R9, R9A, R9X** — very high-density
- **R10, R10A, R10X** — highest density
- **C4-4, C4-5** and similar commercial zones with R7+ residential equivalent — include these too
- **M1-4/R7A** and similar MX zones — include if the R-component is R7+

Exclude: R1 through R6B (low/medium density — not worth assembling).

## Tool 3: Google Sheets CLI

Create a new spreadsheet:
```bash
integrations sheets spreadsheets create --title="Bronx Assemblage Scan — 2026-03-23" --json
```

Write the header row:
```bash
integrations sheets values update --id=SPREADSHEET_ID --range="Sheet1!A1:M1" --values='[["Property Address","Asking Price","Units","Lot Size (SF)","Building SF","Year Built","Zoning","Starter Lot Potential","Block Context Note","Why This Could Be a Starting Point","Zillow Link","ZoLa Link","Notes"]]' --value-input=USER_ENTERED --json
```

Append data rows:
```bash
integrations sheets values append --id=SPREADSHEET_ID --range="Sheet1!A1" --values='[["123 Main St, Bronx, NY 10451","$500,000","2","2,500","1,800","1925","R7A","High","Narrow rowhouse block, similar buildings adjacent","Small 2-family in R7A zone, standard 25ft lot, older stock, attached context","https://zillow.com/...","https://zola.planning.nyc.gov/lot/2/...","Corner lot"]]' --value-input=USER_ENTERED --json
```

## Tool 4: Google Docs CLI

Create a new document:
```bash
integrations docs documents create --title="Bronx Assemblage Report — 2026-03-23" --json
```

Returns `{ "id": "...", "title": "...", "url": "https://docs.google.com/document/d/.../edit" }`.

Append text to a document:
```bash
integrations docs documents append --document-id=DOC_ID --text="# Executive Summary\n\nThis report covers..." --json
```

The `--text` flag supports `\n` for newlines. Use `--text-file=PATH` to append from a file.

For rich formatting (headings, tables, bold), use batch-update with raw Docs API requests:
```bash
integrations docs documents batch-update --document-id=DOC_ID --requests-file=/tmp/requests.json --json
```

### Writing a formatted report

The easiest approach:
1. Create the doc: `docs documents create --title=...`
2. Write the full report content to a text file
3. Append it: `docs documents append --document-id=ID --text-file=/tmp/report.txt`

For the report, write plain text with clear section headers using markdown-style formatting. The agent should write content that reads well as plain text — Google Docs will display it cleanly.

## Workflow

1. Search Zillow for each target zip code (see job prompt for list)
2. Filter out condos, co-ops, pending/contingent listings
3. Geocode each address via NYC GeoSearch → get BBL
4. Look up zoning + lot data via PLUTO (Socrata) → get zoning, lot SF, building SF, year built
5. Filter: only keep R7+ zoned properties
6. Score starter-lot potential (Low/Moderate/High) based on observable signals
7. Create Google Sheet with raw data table
8. Create Google Doc report with analysis and recommendations

## Important
- **Use PLUTO for lot size, year built, building area** — not Zillow detail endpoint
- Process properties in batches — don't try to do all at once
- If PLUTO returns no results for a BBL, write "Unable to verify" in the zoning column
- If a property is a condo or co-op (check homeType from Zillow or bldgclass from PLUTO), skip it
- The Google Sheet is the final deliverable — make sure every row is complete
