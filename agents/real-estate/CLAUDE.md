# NYC Assemblage Intelligence — Tool Documentation

## Authentication
Zillow and Google credentials are pre-configured via environment variables. Run commands directly. All NYC public APIs (PLUTO, ACRIS, DOB, HPD, Finance) require no authentication — just curl.

---

## Tool 1: Zillow CLI (search only)

```bash
integrations zillow properties search --location="Bronx, NY 10451" --limit=40 --json
```

Returns: zpid, address, price, beds, baths, sqft, homeType, status, zillowUrl, latitude, longitude, daysOnMarket.

**Do NOT use `integrations zillow properties get`** — use NYC PLUTO for lot data instead.

---

## Tool 2: NYC PLUTO — Zoning, Lot Data, Year Built

The authoritative source for lot data. Returns zoning, lot area, building area, year built, building class — all in one call.

**Geocode address to BBL:**
```bash
curl -s "https://geosearch.planninglabs.nyc/v2/search?text=1776+Seminole+Ave+Bronx+NY"
```
BBL is in `.features[0].properties.addendum.pad.bbl`. Example: `2037620044` (2=Bronx).

**Look up lot data:**
```bash
curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=2037620044"
```
Returns: `zonedist1`, `lotarea`, `bldgarea`, `yearbuilt`, `bldgclass`, `numfloors`, `unitsres`, `unitstotal`, `lotfront`, `lotdepth`.

**Build ZoLa URL:** `https://zola.planning.nyc.gov/lot/2/03762/0044`

BBL parsing: digit 1 = borough (1=Manhattan, 2=Bronx, 3=Brooklyn, 4=Queens, 5=SI), digits 2-6 = block, digits 7-10 = lot.

### Zoning filter
Only include R7+: R7, R7-1, R7-2, R7A, R7B, R7D, R7X, R8, R8A, R8B, R8X, R9, R9A, R9X, R10, R10A, R10X. Also include C4-4, C4-5 and MX zones with R7+ residential component. Exclude R1 through R6B.

---

## Tool 3: ACRIS — Deed Transfers, Foreclosures, Liens

NYC's official property document filing system. All data is on Socrata — no auth needed.

**Step 1: Get document IDs for a BBL from the Legals table:**
```bash
curl -s "https://data.cityofnewyork.us/resource/8h5j-fqxa.json?borough=2&block=02964&lot=0028&\$limit=500"
```
Borough codes: 1=Manhattan, 2=Bronx, 3=Brooklyn, 4=Queens, 5=Staten Island.

**Step 2: Get document details from the Master table:**
```bash
curl -s "https://data.cityofnewyork.us/resource/bnx9-e6tj.json?\$where=document_id in('DOC_ID_1','DOC_ID_2')"
```
Returns: `document_id`, `doc_type`, `document_date`, `document_amt`, `recorded_datetime`.

**Step 3 (optional): Get parties (buyer/seller names):**
```bash
curl -s "https://data.cityofnewyork.us/resource/636b-3b5g.json?\$where=document_id in('DOC_ID_1')&party_type=2"
```
`party_type=1` = grantor/seller, `party_type=2` = grantee/buyer.

### Key document types to monitor

| Code | Meaning | Signal |
|------|---------|--------|
| `DEED` | Deed transfer | Ownership change — check if buyer is LLC (developer) |
| `MTGE` | Mortgage | New financing activity |
| `SAT` | Satisfaction of mortgage | Owner paid off mortgage (unencumbered, easier to sell) |
| `ASST` | Assignment of mortgage | Loan sold to new lender |
| `JUDG` | Judgment | **Includes lis pendens (foreclosure filings)** |
| `FL` | Federal lien (IRS) | Tax trouble — distressed owner |
| `TLS` | Tax lien sale certificate | **Property was in a tax lien sale** |
| `CTOR` | Court order | Legal action against property |
| `CODP` | Condemnation proceedings | Government taking property |

### Detecting foreclosures (lis pendens)
Filter for `doc_type=JUDG` filed in the last 90 days:
```bash
curl -s "https://data.cityofnewyork.us/resource/bnx9-e6tj.json?\$where=document_id in('ID1','ID2')AND doc_type='JUDG' AND document_date > '2025-12-01'"
```

### Detecting estate/probate signals
Filter ACRIS parties for names containing "ESTATE OF" or "AS EXECUTOR":
```bash
curl -s -G "https://data.cityofnewyork.us/resource/636b-3b5g.json" \
  --data-urlencode "\$where=document_id in('DOC_ID') AND (name like '%25ESTATE OF%25' OR name like '%25EXECUTOR%25')"
```
If any party name matches, the property owner likely died and the estate is in probate. This is a strong sell signal — heirs often want to liquidate quickly.

### Detecting developer activity
Filter for `doc_type=DEED` where the buyer name contains "LLC":
```bash
curl -s "https://data.cityofnewyork.us/resource/636b-3b5g.json?\$where=document_id='DOC_ID' AND party_type='2' AND name like '%25LLC%25'"
```

---

## Tool 4: DOB — Permits, Demolitions, New Buildings

NYC Department of Buildings permit data on Socrata.

**Search permits by BBL:**
```bash
curl -s "https://data.cityofnewyork.us/resource/ipu4-2q9a.json?bbl=2029640028&\$limit=50&\$order=issuance_date DESC"
```

**Search permits by block (to find activity on the same block):**
```bash
curl -s "https://data.cityofnewyork.us/resource/ipu4-2q9a.json?\$where=borough='BRONX' AND block='02964'&\$limit=50&\$order=issuance_date DESC"
```

### Key permit types

| Job Type | Meaning | Signal |
|----------|---------|--------|
| `DM` | Demolition | **Someone is tearing down a building on this block** |
| `NB` | New Building | **A new development is coming to this block** |
| `A1` | Major alteration | Significant renovation — owner investing or converting |
| `A2` | Minor alteration | Routine work |

**Detecting competitor activity:** Search for DM or NB permits on the same block filed in the last 6 months:
```bash
curl -s "https://data.cityofnewyork.us/resource/ipu4-2q9a.json?\$where=borough='BRONX' AND block='02964' AND job_type in('DM','NB') AND issuance_date > '2025-09-01'"
```

---

## Tool 5: HPD — Housing Violations

Open violations dataset on Socrata. No auth needed.

**Search open violations by BBL:**
```bash
curl -s "https://data.cityofnewyork.us/resource/wvxf-dwi5.json?\$where=boroid='2' AND block='02964' AND lot='0028' AND currentstatus='VIOLATION OPEN'&\$limit=100"
```

Returns: `violationid`, `inspectiondate`, `violationstatus`, `class` (A/B/C), `novdescription`.

**Count open violations:**
```bash
curl -s "https://data.cityofnewyork.us/resource/wvxf-dwi5.json?\$select=count(*)&\$where=boroid='2' AND block='02964' AND lot='0028' AND currentstatus='VIOLATION OPEN'"
```

### Signal interpretation
- 0-2 open violations = normal
- 3-5 open violations = some neglect
- 5-10 open violations = significant neglect
- **10+ open violations = owner burnout — strong sell signal**

Class C (immediately hazardous) violations are especially significant — they indicate the building has serious safety issues the owner isn't fixing.

---

## Tool 6: NYC Finance — Tax Liens

Tax lien sale lists on Socrata.

**Search tax lien list by BBL:**
```bash
curl -s "https://data.cityofnewyork.us/resource/9rz4-mjek.json?\$where=bbl='2029640028'"
```

**Search all tax liens in a borough:**
```bash
curl -s "https://data.cityofnewyork.us/resource/9rz4-mjek.json?\$where=borough='BRONX'&\$limit=5000"
```

Returns: `bbl`, `borough`, `block`, `lot`, `building_class`, `community_district`, `council_district`, `cycle`, `ecb_penalty`, `lien_sale_indicator`, `tax_class`.

### Signal interpretation
Any property appearing on the tax lien list = **strong distress signal**. The owner owes multiple years of back taxes. NYC will auction the lien. The owner is very likely to accept a below-market off-market offer.

---

## Tool 7: NYSCEF CLI — Court Records (direct lookup)

Look up a specific court case by docket ID (no CAPTCHA required):
```bash
integrations nyscef cases get --docket-id=ENCODED_ID --json
```

Note: NYSCEF search requires hCaptcha and cannot be automated. Use ACRIS party name data to detect estate/probate signals instead (see "Detecting estate/probate signals" above). If you find an estate signal in ACRIS, you can construct the NYSCEF case URL manually for the report.

---

## Tool 8: StreetEasy CLI — Price History + Listing Cycles

Requires STREETEASY_COOKIES env var (captured via Playwright in the portal).

**Search listings:**
```bash
integrations streeteasy listings search --location="Bronx, NY 10452" --status=for_sale --limit=20 --json
```

**Get price history for a property:**
```bash
integrations streeteasy listings history --address="1226 Shakespeare Ave Bronx NY" --json
```

Returns: array of `{date, event, price}` entries showing every list, delist, relist, and price change.

### Signal interpretation
- Property listed → delisted → relisted at lower price = **motivated seller**
- 3+ listing cycles with declining prices = **desperate seller**
- Price drop > 10% from original listing = **significant negotiation leverage**
- Fresh price drop (last 7 days) = **act now — make an offer this week**

### Composite scoring additions

| Signal | Points | Source |
|--------|--------|--------|
| Price drop > 10% from original | +3 | StreetEasy |
| 3+ listing/delisting cycles | +4 | StreetEasy |
| Price drop in last 30 days | +2 | StreetEasy |

---

## Tool 8: Professional XLSX Spreadsheet (via openpyxl)

Create styled .xlsx, upload to Google Drive with `--convert` flag for native Google Sheet:
```bash
integrations drive files upload --path=/tmp/scan.xlsx --name="NYC Assemblage Scan — 2026-03-24" --convert --json
```

Use openpyxl with: dark blue headers, color-coded potential scores (green=High, yellow=Moderate, red=Low), proper column widths.

---

## Tool 9: Professional PDF Report (via LaTeX)

Write a .tex file, compile with `pdflatex -interaction=nonstopmode`, upload to Drive:
```bash
integrations drive files upload --path=/tmp/report.pdf --name="NYC Assemblage Report — 2026-03-24.pdf" --json
```

Use booktabs tables, navy section headers, fancyhdr, hyperlinked URLs. Escape `$`, `#`, `%`, `&`, `_` characters. Verify compilation is clean — fix and recompile if errors.

---

## Tool 10: Google Drive CLI

```bash
integrations drive files upload --path=/tmp/file --name="Name" [--convert] --json
```

---

## Composite Scoring Model

Each qualifying R7+ lot gets a composite score:

| Signal | Points | Source |
|--------|--------|--------|
| R8+ zoning (vs R7) | +3 | PLUTO |
| Pre-war construction (before 1945) | +2 | PLUTO |
| Lot under 2,000 SF | +2 | PLUTO |
| Active for-sale listing on Zillow | +3 | Zillow |
| Days on market > 180 | +2 | Zillow |
| Tax lien / delinquency | +4 | NYC Finance |
| Judgment filing (lis pendens) in last 90 days | +5 | ACRIS |
| Federal lien (IRS) | +3 | ACRIS |
| Deed transfer to LLC in last 12 months on same block | +3 | ACRIS |
| Demolition permit on same block in last 6 months | +3 | DOB |
| New building permit on same block in last 6 months | +2 | DOB |
| HPD open violations > 5 | +2 | HPD |
| HPD open violations > 10 | +4 | HPD |
| Adjacent lot also for sale | +4 | Zillow + PLUTO |
| Price drop > 10% from original | +3 | StreetEasy |
| 3+ listing/delisting cycles | +4 | StreetEasy |
| Price drop in last 30 days | +2 | StreetEasy |
| ACRIS party name contains "ESTATE OF" or "EXECUTOR" | +5 | ACRIS |

**Priority tiers:**
- **20+** = Immediate outreach (multiple strong signals converging)
- **15-19** = High priority
- **10-14** = Moderate priority
- **5-9** = Watchlist

---

## Workflow

1. Search Zillow for each target zip code across all boroughs
2. Filter out condos, co-ops, pending/contingent listings
3. Geocode each address via NYC GeoSearch → get BBL
4. Look up zoning + lot data via PLUTO → filter to R7+ only
5. **For each qualifying property, run all signal checks in parallel:**
   - ACRIS: check for recent judgments, federal liens, tax lien sale certificates
   - DOB: check for demolition/new building permits on the same block
   - HPD: count open violations
   - NYC Finance: check tax lien list
   - StreetEasy: check price history for drops and relisting cycles
6. **For each qualifying block, check for cluster signals:**
   - Multiple Zillow listings on same block?
   - Recent ACRIS deed transfers to LLCs on same block?
   - DOB permits on same block?
7. Calculate composite score for each property
8. **Verify data:** Check for duplicates, mismatched URLs, inconsistent scoring. Fix issues.
9. Create professional XLSX with all properties, signals, and scores. Upload to Drive with --convert.
10. Create professional LaTeX PDF report. Verify compilation. Upload to Drive.

## Important
- **Use PLUTO for lot size, year built, building area** — not Zillow
- **All NYC APIs are public Socrata endpoints** — just curl, no auth needed
- Process properties in batches — don't try to do all at once
- BBL format: borough(1) + block(5) + lot(4), zero-padded
- Escape special characters in LaTeX: `$` → `\$`, `#` → `\#`, `%` → `\%`, `&` → `\&`, `_` → `\_`
- The XLSX and PDF are the final deliverables — they must be professional, accurate, and complete
