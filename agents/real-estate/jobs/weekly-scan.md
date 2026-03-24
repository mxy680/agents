Your task is to identify potential starter-lot opportunities for future residential assemblages across New York City, with pre-market distress signal detection.

## Objective

Produce a professional XLSX spreadsheet and PDF report of small residential properties in R7+ zones that show the highest combination of development potential and owner motivation to sell.

## Search Criteria

### Target zip codes by borough

**Bronx:**
10451, 10452, 10453, 10454, 10456, 10457, 10458, 10459, 10460, 10468

**Brooklyn (verified R7+ via PLUTO):**
11226 (Flatbush), 11215 (Park Slope/Windsor Terrace), 11238 (Prospect Heights), 11217 (Boerum Hill/Park Slope), 11225 (Crown Heights/Lefferts Gardens), 11205 (Fort Greene/Clinton Hill), 11220 (Sunset Park), 11230 (Midwood), 11213 (Crown Heights)

**Manhattan:**
10026 (Harlem), 10027 (Harlem), 10029 (East Harlem), 10030 (Harlem), 10031 (Hamilton Heights), 10032 (Washington Heights), 10033 (Washington Heights), 10034 (Inwood), 10039 (Harlem), 10040 (Washington Heights)

**Queens (verified R7+ via PLUTO):**
11373 (Elmhurst), 11374 (Rego Park), 11435 (Jamaica/Briarwood), 11372 (Jackson Heights), 11433 (Jamaica), 11432 (Jamaica Estates), 11355 (Flushing), 11104 (Sunnyside), 11375 (Forest Hills)

Run a separate Zillow search for each zip code. Do NOT search by borough name — that returns low-density neighborhoods.

### Listing filters
- Active for sale only
- Property types: 1-family, 2-family, 3-family, 4-family, 5-family homes
- Exclude: condos, co-ops, buildings with more than 5 units, pending/contingent

## Workflow

### Step 1 — Zillow search

Search all zip codes. For each qualifying listing capture: address, price, beds, baths, sqft, home type, status, days on market, Zillow URL.

### Step 2 — PLUTO verification

Geocode each address → BBL → PLUTO lookup. Get: zoning, lot area, building area, year built, building class. Filter to R7+ zoning only.

### Step 3 — Pre-market signal checks

For each qualifying R7+ property, run ALL of these checks:

**ACRIS (foreclosure + legal filings):**
- Get document IDs from Legals table using borough/block/lot
- Check Master table for recent JUDG (judgment/lis pendens), FL (federal lien), TLS (tax lien sale certificate) filings in the last 12 months
- Check if any recent DEED transfers on the same block went to an LLC (developer buying)

**DOB (permit activity on the block):**
- Search permits for the same block for DM (demolition) or NB (new building) permits in the last 6 months
- This detects competitor activity — someone is already assembling or developing on this block

**HPD (housing violations):**
- Count open violations for the property's BBL
- Flag properties with 5+ open violations (neglect) or 10+ (severe burnout)

**NYC Finance (tax distress):**
- Check if the BBL appears on the tax lien sale list
- Any hit = strong distress signal

**StreetEasy (price history + listing cycles):**
- Use `integrations streeteasy listings history --address="ADDRESS" --json`
- Check for price drops > 10% from original listing price
- Check for 3+ listing/delisting cycles (desperate seller signal)
- Check for recent price drops in the last 30 days (act-now signal)

### Step 4 — Composite scoring

Score each property using the model in CLAUDE.md. Assign a composite score and priority tier (Immediate / High / Moderate / Watchlist).

### Step 5 — Cluster detection

Identify blocks where 2+ qualifying properties are present. For each cluster:
- List all qualifying properties on the block
- Sum their lot areas
- Sum their asking prices
- Note if any share the same owner (ACRIS party data)

### Step 6 — Verify data

Before creating output files, check for:
- Mismatched Zillow URLs
- Missing or invalid BBLs
- Duplicate addresses
- Inconsistent scoring (e.g., property with lis pendens + tax lien scored below 15)
- Properties that should have been filtered out

Fix any issues.

### Step 7 — Create XLSX spreadsheet

Use Python openpyxl to create a styled spreadsheet with these columns:

| Borough | Property Address | Asking Price | Units | Lot Size (SF) | Building SF | Year Built | Zoning | Composite Score | Priority | Tax Lien? | Lis Pendens? | HPD Violations | Demo Permit on Block? | Price Drop? | Listing Cycles | Days on Market | Block Context | Zillow Link | ZoLa Link | Notes |

Color coding:
- Green rows = Immediate priority (score 20+)
- Light green = High priority (15-19)
- Yellow = Moderate (10-14)
- No color = Watchlist (5-9)

Sort by composite score descending.

Upload to Google Drive with `--convert` flag. Provide the Google Sheets link.

### Step 8 — Create PDF report

Write a LaTeX document and compile with pdflatex. Include:

1. **Executive Summary** — boroughs searched, listings reviewed, R7+ filter results, signal distribution, score distribution
2. **Immediate Priority Properties** (score 20+) — detailed paragraph for each explaining all signals present
3. **Top Cluster Opportunities** — same-block groupings with combined lot area and combined ask
4. **Pre-Market Signal Summary** — how many tax liens, lis pendens, HPD burnouts, and DOB permits were found
5. **Full Results Table** — all qualifying properties sorted by score with signal columns
6. **Methodology** — data sources, signal definitions, scoring model, limitations

Verify compilation is clean. Fix LaTeX errors and recompile if needed. Upload PDF to Google Drive.

## Quality standard

- Accurate, clean, and useful for brokerage outreach planning
- "Not stated" or "Unable to verify" for missing data — never guess
- No FAR calculations or development yield projections
- Always run verification step before output
- The XLSX and PDF must be professional — these go to a brokerage team
