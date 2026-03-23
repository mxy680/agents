Your task is to identify potential "starter lot" opportunities for future residential assemblages in the Bronx.

You will use Zillow to locate active for-sale listings and NYC ZoLa to determine zoning.

## Objective

Create a structured Google Sheet of small residential properties currently for sale that could serve as the first acquisition in a larger assemblage strategy.

## Search criteria

Location: Bronx, New York only. Focus on high-density zoning areas by searching these zip codes individually:
- 10451 (Mott Haven / Port Morris)
- 10452 (Highbridge / Concourse)
- 10453 (Morris Heights / University Heights)
- 10454 (Mott Haven / Melrose)
- 10456 (Morrisania / Melrose)
- 10457 (East Tremont / Fordham)
- 10458 (Fordham / Belmont)
- 10459 (Longwood / Crotona Park East)
- 10460 (West Farms / Crotona Park)
- 10468 (Fordham / Bedford Park)

Run a separate Zillow search for each zip code. Do NOT just search "Bronx, NY" — that returns mostly low-density neighborhoods.

Listing status: Active for sale listings only.
Property types: 1-family, 2-family, 3-family, 4-family, 5-family homes.

Exclude: condos, co-ops, buildings with more than 5 units, pending/contingent/under-contract listings.

## Workflow

### Step 1 — Zillow search

Search for qualifying listings using the Zillow CLI.

For each listing, capture:
- Full property address
- Asking price
- Property type
- Number of units (if stated)
- Lot size (square feet)
- Building square footage (if stated)
- Year built (if stated)
- Days on market (if stated)
- Zillow URL

### Step 2 — ZoLa zoning verification

For each property address, look up the zoning on NYC ZoLa using the GeoSearch + ZoLa API.

Capture:
- Zoning designation (e.g., R6, R7-1, R7A, etc.)
- ZoLa URL

### Step 3 — Filter and score

Only include properties zoned R7 or higher.

Score each qualifying property as Low / Moderate / High starter-lot potential based on:
- Small or low-rise building in medium- or higher-density zoning
- Attached or rowhouse context
- Similar buildings on the block
- Corner lot positioning
- Older housing stock
- Standard narrow residential lots that could be assembled
- Lot size that appears modest relative to zoning

Do NOT calculate FAR or development yield.
Do NOT assume specific buildable square footage.

### Step 4 — Output to Google Sheet

Create a Google Sheet titled "Bronx Assemblage Scan — [today's date]".

Write a table with these columns:

| Property Address | Asking Price | Units | Lot Size (SF) | Building SF | Year Built | Zoning | Starter Lot Potential | Block Context Note | Why This Could Be a Starting Point | Zillow Link | ZoLa Link | Notes |

### Step 5 — Create Google Doc report

Create a formatted Google Doc titled "Bronx Assemblage Report — [today's date]".

Write it as an HTML file and upload via Google Drive. The report should include:

1. **Executive Summary** — how many zip codes searched, how many listings reviewed, how many passed the R7+ filter, how many scored High/Moderate/Low
2. **Top Opportunities** — the 3-5 best starter lots with a paragraph each explaining why they stand out. Mention price, zoning, lot size, year built, and block context.
3. **Cluster Opportunities** — any cases where 2+ qualifying properties are on the same block or adjacent. These are rare and especially valuable for assemblage.
4. **Full Results Table** — all qualifying properties in a formatted HTML table (same columns as the Google Sheet)
5. **Methodology** — brief note on data sources (Zillow search API, NYC PLUTO via Socrata, NYC GeoSearch) and any limitations encountered
6. **Link to the Google Sheet** — include the URL of the companion spreadsheet

Format the HTML cleanly with proper headings, paragraphs, and table styling. This document will be shared with a brokerage team.

## Quality standard

- Accurate, clean, and useful for brokerage research and outreach planning.
- Write "Not stated" or "Unable to verify" for missing information.
- No speculation beyond visible listing and zoning context.
- The Google Doc should be professional and readable — not a raw data dump.
