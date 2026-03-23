# Bronx Assemblage Scout — Tool Documentation

## Authentication
All credentials are pre-configured via environment variables. Run commands directly — do not check for missing tokens.

## Tool 1: Zillow CLI

Search for properties:
```bash
integrations zillow properties search --location="Bronx, NY" --limit=40 --json
```

Get full details for a property:
```bash
integrations zillow properties get --zpid=ZPID --json
```

The search returns: zpid, address, price, beds, baths, sqft, homeType, status, zillowUrl, latitude, longitude, daysOnMarket.

The detail endpoint returns additional fields: lotSize, yearBuilt, description, photos, priceHistory, taxHistory, schools.

## Tool 2: NYC ZoLa (raw HTTP — no CLI needed)

### Step A: Geocode the address to get BBL

```bash
curl -s "https://geosearch.planninglabs.nyc/v2/search?text=1776+Seminole+Ave+Bronx+NY" | jq '.features[0].properties'
```

Response includes `addendum.pad.bbl` — the Borough-Block-Lot identifier. Example: `2037620044` where `2` = Bronx.

### Step B: Look up zoning for that BBL

```bash
curl -s "https://zola-api.planning.nyc.gov/api/v1/lot/2/03762/0044" | jq '.zoning_districts'
```

URL format: `/api/v1/lot/{borough}/{block}/{lot}` where borough=2 for Bronx, block and lot are zero-padded from the BBL.

The `zoning_districts` array contains objects with a `zonedist` field like "R7-1", "R8A", "C4-4", etc.

### Step C: Build the ZoLa URL

```
https://zola.planning.nyc.gov/lot/2/03762/0044
```

Same format as the API: `/lot/{borough}/{block}/{lot}`.

### Parsing the BBL

The BBL is a 10-digit number: `BBBBBBBLLLL` where:
- Digit 1: borough (2 = Bronx)
- Digits 2-6: block (5 digits, zero-padded)
- Digits 7-10: lot (4 digits, zero-padded)

Example: BBL `2037620044` → borough=2, block=03762, lot=0044

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

Returns `{ "spreadsheetId": "...", "spreadsheetUrl": "..." }`.

Write the header row:
```bash
integrations sheets values update --id=SPREADSHEET_ID --range="Sheet1!A1:M1" --values='[["Property Address","Asking Price","Units","Lot Size (SF)","Building SF","Year Built","Zoning","Starter Lot Potential","Block Context Note","Why This Could Be a Starting Point","Zillow Link","ZoLa Link","Notes"]]' --value-input=USER_ENTERED --json
```

Append data rows:
```bash
integrations sheets values append --id=SPREADSHEET_ID --range="Sheet1!A1" --values='[["123 Main St, Bronx, NY 10451","$500,000","2","2,500","1,800","1925","R7A","High","Narrow rowhouse block, similar buildings adjacent","Small 2-family in R7A zone, standard 25ft lot, older stock, attached context","https://zillow.com/...","https://zola.planning.nyc.gov/lot/2/...","Corner lot"]]' --value-input=USER_ENTERED --json
```

## Workflow

1. Search Zillow for Bronx 1-5 family homes for sale
2. For each result, get property details (lot size, year built)
3. Geocode the address via NYC GeoSearch → get BBL
4. Look up zoning via ZoLa API → get zoning district
5. Filter: only keep R7+ zoned properties
6. Score starter-lot potential (Low/Moderate/High) based on observable signals
7. Create Google Sheet and write results

## Important
- Process properties in batches — don't try to do all 40 at once
- If ZoLa API returns no results for an address, write "Unable to verify" in the zoning column
- If a property is a condo or co-op (check homeType), skip it
- The Google Sheet is the final deliverable — make sure every row is complete
