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

### Step 4 — Verify data

Before creating any output files, re-read all collected data and check for:
- Mismatched Zillow URLs (URL points to wrong property)
- Missing zoning data (should be "Unable to verify", not blank)
- Duplicate addresses
- Inconsistent scoring (e.g., R8 property scored Moderate when it should be High)
- Properties that should have been filtered out (condos, co-ops, <R7 zoning)

Fix any issues found.

### Step 5 — Create professional XLSX spreadsheet

Use Python with openpyxl to create a styled .xlsx file with:
- Dark blue header row with white bold text
- Color-coded "Starter Lot Potential" cells (green=High, yellow=Moderate, red=Low)
- Proper column widths for readability
- All 13 columns: Property Address | Asking Price | Units | Lot Size (SF) | Building SF | Year Built | Zoning | Starter Lot Potential | Block Context Note | Why This Could Be a Starting Point | Zillow Link | ZoLa Link | Notes

Save to /tmp/ and upload to Google Drive with `--mime-type=application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`. Google Drive will auto-convert it to a Google Sheet when opened.

After uploading, provide both links:
- The Google Drive file link (from the upload response)
- A direct Google Sheets link: `https://docs.google.com/spreadsheets/d/FILE_ID/edit` (use the file ID from the upload response)

### Step 6 — Create professional PDF report

Write a LaTeX document and compile with `pdflatex` to create a styled PDF report with:

1. **Title page** — "Bronx Assemblage Report" + date + "Prepared for Brokerage Team"
2. **Executive Summary** — zip codes searched, listings reviewed, R7+ filter results, score distribution
3. **Top Opportunities** — the 3-5 best starter lots with a paragraph each explaining why. Mention price, zoning, lot size, year built, block context.
4. **Cluster Opportunities** — cases where 2+ qualifying properties are on the same block. Include combined lot area estimate.
5. **Full Results Table** — all qualifying properties in a formatted table with alternating row colors
6. **Methodology** — data sources, process, limitations
7. **Link to companion spreadsheet** — the Google Drive URL of the XLSX file

Use professional LaTeX styling: booktabs tables, navy section headers, fancyhdr page headers, hyperlinked URLs. Escape all special characters (`$`, `#`, `%`, `&`, `_`).

Save the .tex file to /tmp/, compile with `pdflatex -interaction=nonstopmode` (run twice), and upload the PDF to Google Drive.

## Quality standard

- Accurate, clean, and useful for brokerage research and outreach planning.
- Write "Not stated" or "Unable to verify" for missing information.
- No speculation beyond visible listing and zoning context.
- The XLSX and PDF must be professional — these go to a brokerage team.
- Always run the verification step before creating output files.
