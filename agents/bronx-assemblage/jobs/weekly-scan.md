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

## Quality standard

- Accurate, clean, and useful for brokerage research and outreach planning.
- Write "Not stated" or "Unable to verify" for missing information.
- No speculation beyond visible listing and zoning context.
