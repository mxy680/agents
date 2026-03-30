# NYC Assemblage Intelligence — Tool Documentation

## CRITICAL RULES

1. **DO NOT prefix commands with `doppler run --`** — credentials are already in your environment
2. **DO NOT use the Zillow CLI** — it's blocked by PerimeterX. Zillow data comes from the Chrome extension scrape stored in Supabase
3. **`integrations trends` is rate-limited** — Google Trends returns 429 if used too frequently. Try once; if it fails, skip and move on.
4. **DO NOT use `integrations nysla`** — NY SLA dataset is locked behind login auth. No fix available. Skip it.
5. **`integrations obituaries` is untested** — Legacy.com API may be unreliable. Skip if it fails.
6. **DO NOT spawn sub-agents via the Agent tool** — run commands directly
7. **Use the pipeline scripts** when doing full scans: `bash scripts/run_pipeline.sh`
9. **For ad-hoc queries**, use curl for Socrata APIs and the `integrations` CLI for providers that work

## What Works

| Tool | Status | How to use |
|------|--------|-----------|
| PLUTO | Works | `curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=BBL"` |
| ACRIS | Works | `curl -s -G` with `--data-urlencode` (see below) |
| HPD | Works | `curl -s -G` with `--data-urlencode` |
| DOB | Works | `curl -s -G` with `--data-urlencode` |
| NYC Finance | Works | `curl -s -G` with `--data-urlencode` |
| 311 | Works | `curl -s -G` with `--data-urlencode` |
| ECB/OATH | Works | `curl -s -G` with `--data-urlencode` |
| Citi Bike CLI | Works | `integrations citibike stations density --lat=X --lng=Y --json` |
| HMDA CLI | Works | `integrations hmda loans summary --county=bronx --json` |
| Census CLI | Works | `integrations census tracts profile --tract=FIPS --json` |
| NY DOS CLI | Works | `integrations nydos entities match-address --address="..." --json` |
| NYC DOF CLI | Works | `integrations dof owners search --name="..." --json` |
| Google Drive CLI | Works | `integrations drive files upload --path=... --name=... --json` |
| Zillow CLI | BROKEN | Use Supabase scrape_data table instead |
| Google Trends CLI | Flaky | Rate limited (429) — may work after cooldown. Retry if needed. |
| NYSLA CLI | BROKEN | Dataset locked behind login — no fix |
| Obituaries CLI | Untested | Legacy.com API may be unreliable |

## Authentication
All credentials are pre-configured via environment variables. Run commands directly — no `doppler run` needed.
All NYC public APIs (PLUTO, ACRIS, DOB, HPD, Finance) require no authentication — just curl.

---

## Tool 1: Zillow Data (from Supabase, NOT CLI)

Zillow data is scraped via Chrome extension and stored in the `scrape_data` table in Supabase.
**DO NOT use the Zillow CLI** — it will get HTTP 403.

The pipeline script `scripts/phase1_zillow_search.py` reads from Supabase automatically.

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
curl -s -G "https://data.cityofnewyork.us/resource/8h5j-fqxa.json" \
  --data-urlencode "borough=2" \
  --data-urlencode "block=02964" \
  --data-urlencode "lot=0028" \
  --data-urlencode "\$limit=500"
```
Borough codes: 1=Manhattan, 2=Bronx, 3=Brooklyn, 4=Queens, 5=Staten Island.

**CRITICAL: block MUST be zero-padded to 5 digits, lot MUST be zero-padded to 4 digits.** Extract from BBL string, do NOT convert to integer. BBL `2029640028` → borough=`2`, block=`02964`, lot=`0028`.

**Step 2: Get document details from the Master table:**
```bash
curl -s -G "https://data.cityofnewyork.us/resource/bnx9-e6tj.json" \
  --data-urlencode "\$where=document_id in('DOC_ID_1','DOC_ID_2')"
```
Returns: `document_id`, `doc_type`, `document_date`, `document_amt`, `recorded_datetime`.

**Step 3 (optional): Get parties (buyer/seller names):**
```bash
curl -s -G "https://data.cityofnewyork.us/resource/636b-3b5g.json" \
  --data-urlencode "\$where=document_id in('DOC_ID_1') AND party_type='2'"
```
`party_type=1` = grantor/seller, `party_type=2` = grantee/buyer.

**ALWAYS use `curl -s -G` with `--data-urlencode` for ALL Socrata queries.** Never inline `$where` in the URL directly — it breaks on spaces and special characters.

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
  --data-urlencode "\$where=document_id in('DOC_ID') AND (name like '%ESTATE OF%' OR name like '%EXECUTOR%')"
```
If any party name matches, the property owner likely died and the estate is in probate. This is a strong sell signal — heirs often want to liquidate quickly.

### Detecting developer activity
Filter for `doc_type=DEED` where the buyer name contains "LLC":
```bash
curl -s -G "https://data.cityofnewyork.us/resource/636b-3b5g.json" \
  --data-urlencode "\$where=document_id='DOC_ID' AND party_type='2' AND name like '%LLC%'"
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

**IMPORTANT: Use dataset `csn4-vhvf` (Open HPD Violations), NOT `wvxf-dwi5`.**

The `csn4-vhvf` dataset contains only currently open violations — no status filter needed.

**Count open violations by BBL:**
```bash
curl -s -G "https://data.cityofnewyork.us/resource/csn4-vhvf.json" \
  --data-urlencode "\$select=count(*)" \
  --data-urlencode "\$where=boroid='2' AND block='02406' AND lot='0108'"
```

**Get violation details:**
```bash
curl -s -G "https://data.cityofnewyork.us/resource/csn4-vhvf.json" \
  --data-urlencode "\$where=boroid='2' AND block='02406' AND lot='0108'" \
  --data-urlencode "\$limit=100"
```

Returns: `violationid`, `inspectiondate`, `class` (A/B/C), `novdescription`, `currentstatus`.

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

## Tool 7: Additional Socrata Signals (no auth)

All use the same `curl -s -G` + `--data-urlencode` pattern as PLUTO/ACRIS/DOB/HPD.

### 311 Complaints — Neighborhood Neglect Trajectory
```bash
curl -s -G "https://data.cityofnewyork.us/resource/erm2-nwe9.json" \
  --data-urlencode "\$where=incident_address='1776 SEMINOLE AVE' AND borough='BRONX'" \
  --data-urlencode "\$select=complaint_type,count(*)" \
  --data-urlencode "\$group=complaint_type" \
  --data-urlencode "\$limit=50"
```
Key types: HEAT/HOT WATER (landlord neglect), Rodent (building decay), Noise - Residential (overcrowding). 10+ complaints in 12 months = distress signal.

### Local Law 97 — Energy Grades (buildings >25k sqft)
```bash
curl -s -G "https://data.cityofnewyork.us/resource/7x5e-2fxh.json" \
  --data-urlencode "\$where=borough_block_lot='2029640028'" \
  --data-urlencode "\$select=energy_star_score,source_eui_kbtu_ft,letter_grade"
```
Grade D or F = owner faces carbon fines starting 2024. Expensive retrofit or sell — strong distress signal for large buildings.

### DOF Rolling Sales — Actual Closed Sale Prices
```bash
curl -s -G "https://data.cityofnewyork.us/resource/usep-8jbt.json" \
  --data-urlencode "\$where=borough=2 AND block=2964 AND lot=28" \
  --data-urlencode "\$order=sale_date DESC" \
  --data-urlencode "\$limit=5"
```
Returns: `sale_price`, `sale_date`, `building_class_at_time_of_sale`. Compare last sale price to current Zillow listing — large gaps reveal motivation.

### ECB/OATH Violations — Environmental Fines
```bash
curl -s -G "https://data.cityofnewyork.us/resource/6bgk-3dad.json" \
  --data-urlencode "\$where=respondent_house_number='1776' AND respondent_street='SEMINOLE AVE' AND violation_status='DEFAULT'" \
  --data-urlencode "\$limit=50"
```
Defaulted environmental violations = unpaid fines accumulating. Combined with HPD violations, signals total owner disinvestment.

### Eviction Filings (Housing Court)
```bash
curl -s -G "https://data.cityofnewyork.us/resource/6z8x-wfk4.json" \
  --data-urlencode "\$where=borough='BRONX' AND street_address like '%SEMINOLE%'" \
  --data-urlencode "\$order=executed_date DESC" \
  --data-urlencode "\$limit=20"
```
Multiple eviction filings = problem tenants or landlord trying to clear building. Either way, signals a building the owner wants to exit.

### Scaffolding / Sidewalk Sheds — Stalled Projects
```bash
curl -s -G "https://data.cityofnewyork.us/resource/ipu4-2q9a.json" \
  --data-urlencode "\$where=bbl='2029640028' AND job_type='SH'" \
  --data-urlencode "\$order=issuance_date DESC"
```
Scaffolding up 3+ years = stalled repair, owner can't afford to fix or remove. Uses same DOB permits endpoint (job_type `SH`).

### ULURP / City Planning — Rezoning Pipeline
```bash
curl -s -G "https://data.cityofnewyork.us/resource/n5mv-nfpy.json" \
  --data-urlencode "\$where=borough='BX' AND ulurp_status='Active'" \
  --data-urlencode "\$limit=50"
```
Active rezoning applications near target zone = properties may be upzoned soon. Current R6 lots in a pending R7+ upzone are undervalued.

### Certificate of Occupancy — Use Changes + New Buildings
```bash
curl -s -G "https://data.cityofnewyork.us/resource/pkdm-hqz6.json" \
  --data-urlencode "\$where=borough='BRONX' AND block='02964'" \
  --data-urlencode "\$order=c_of_o_issuance_date DESC" \
  --data-urlencode "\$limit=20"
```
Returns: `c_of_o_filing_type`, `c_of_o_issuance_date`, `number_of_dwelling_units`, `house_no`, `street_name`. A new CO on the same block = active development. CO type change (e.g. residential to commercial) = owner repositioning.

### FDNY Vacate Orders — Fire-Damaged / Unsafe Buildings
```bash
curl -s -G "https://data.cityofnewyork.us/resource/frax-hfgs.json" \
  --data-urlencode "\$where=borough='BRONX' AND block='02964'" \
  --data-urlencode "\$order=vacate_date DESC" \
  --data-urlencode "\$limit=20"
```
Vacated buildings = uninhabitable. Owner faces costly repairs or demolition. Very strong motivated seller signal, especially on R7+ lots.

### DOB Complaints — Active Building Issues
```bash
curl -s -G "https://data.cityofnewyork.us/resource/eabe-havv.json" \
  --data-urlencode "\$where=community_board='202' AND status='OPEN'" \
  --data-urlencode "\$select=complaint_category,count(*)" \
  --data-urlencode "\$group=complaint_category" \
  --data-urlencode "\$limit=50"
```
Open DOB complaints (illegal conversion, unsafe structure, construction without permit) add to the distress picture alongside HPD violations.

---

## Tool 8: Citi Bike CLI — Transit Density Signal

Station density near a property = transit accessibility = value signal.

**Search nearby stations:**
```bash
integrations citibike stations search --lat=40.8176 --lng=-73.9209 --radius=500 --json
```

**Get density metrics for scoring:**
```bash
integrations citibike stations density --lat=40.8176 --lng=-73.9209 --radius=1000 --json
```
Returns: `count`, `avg_capacity`, `total_capacity`, `radius_m`.

### Signal interpretation
- 5+ stations within 1km = excellent transit access (+2 points)
- 0 stations within 1km = poor transit, skip unless other signals are very strong

---

## Tool 9: HMDA CLI — Mortgage Origination Intelligence

CFPB mortgage data reveals where institutional investors are buying.

**County-level summary:**
```bash
integrations hmda loans summary --county=bronx --year=2023 --json
```

**Census tract detail:**
```bash
integrations hmda loans tract --tract=36005000100 --year=2023 --json
```
Returns: total originations, dollar volume, avg loan size, top tracts. A spike in non-owner-occupied loans = institutional money moving in.

---

## Tool 10: Google Trends CLI — Neighborhood Momentum

Rising search interest for a neighborhood predicts price appreciation 12-18 months out.

**Get momentum score (key metric):**
```bash
integrations trends interest momentum --keyword="mott haven apartments" --json
```
Returns: `recent_avg`, `earlier_avg`, `momentum_pct`, `trend` (rising/stable/declining). Compares last 3 months vs first 3 months of a 12-month window.

**Compare neighborhoods:**
```bash
integrations trends interest compare --keywords="mott haven,east new york,bed stuy" --json
```

**Raw interest over time:**
```bash
integrations trends interest search --keyword="mott haven" --time="today 12-m" --json
```

### Signal interpretation
- Momentum > +15% = "rising" → neighborhood is gaining attention before prices move (+3 points)
- Momentum < -15% = "declining" → skip or discount other signals

---

## Tool 11: Obituaries CLI — Estate Property Detection

Cross-reference deceased names with ACRIS property ownership to find estate properties before they hit the market.

**Extract names for ACRIS cross-ref (key command):**
```bash
integrations obituaries names --city=Bronx --date-range=Last30Days --json
```
Returns: `[{first, last, full, publish_date}]`. Pipe `last` names into ACRIS party search to find properties.

**Full obituary search:**
```bash
integrations obituaries search --city=Bronx --state="New York" --date-range=Last30Days --limit=50 --json
```

### Workflow
1. Run `obituaries names --city=<borough>` for each target borough
2. For each last name, search ACRIS parties: `curl ... name like '%LASTNAME%'`
3. If ACRIS match found with "ESTATE OF" or "EXECUTOR" → property is in probate → strong acquisition signal

---

## Tool 12: NY SLA CLI — Liquor License Gentrification Signal

New bar/restaurant licenses precede residential price appreciation by 2-3 years.

**Count new licenses (key command):**
```bash
integrations nysla licenses count --borough=bronx --since=2025-09-01 --json
```
Returns: `{new_licenses, breakdown: [{type, count}]}`. A spike vs prior period = gentrification signal.

**Search licenses:**
```bash
integrations nysla licenses search --borough=bronx --zip=10451 --since=2025-01-01 --json
```

**License density by ZIP:**
```bash
integrations nysla licenses density --borough=bronx --zip=10451 --json
```

### Signal interpretation
- 5+ new restaurant/bar licenses in a ZIP in 6 months = strong gentrification signal (+3 points)
- High existing density = established commercial area (neutral — already priced in)

---

## Tool 13: Census ACS CLI — Demographic Trends

ACS 5-year estimates by census tract. Rising rent burden + population growth = development demand.

**Tract profile (key command):**
```bash
integrations census tracts profile --tract=36005000100 --json
```
Returns: `population`, `median_income`, `median_rent`, `median_home_value`, `vacancy_rate`, `owner_occupied_pct`, `renter_occupied_pct`.

**Compare tracts in a borough:**
```bash
integrations census tracts compare --borough=bronx --sort=vacancy --limit=20 --json
```

**Borough-wide summary:**
```bash
integrations census tracts summary --borough=bronx --json
```

### Signal interpretation
- Vacancy > 10% = weak demand or transitional area (context-dependent)
- Rent burden > 30% (median_rent / monthly_income) = tenants stretched, development demand for new supply (+2 points)
- Population growth (compare years) = rising demand

---

## Tool 14: NYSCEF CLI — Court Records (direct lookup)

Look up a specific court case by docket ID (no CAPTCHA required):
```bash
integrations nyscef cases get --docket-id=ENCODED_ID --json
```

Note: NYSCEF search requires hCaptcha and cannot be automated. Use ACRIS party name data to detect estate/probate signals instead (see "Detecting estate/probate signals" above). If you find an estate signal in ACRIS, you can construct the NYSCEF case URL manually for the report.

---

## Tool 15: Professional XLSX Spreadsheet (via openpyxl)

Create styled .xlsx, upload to Google Drive with `--convert` flag for native Google Sheet:
```bash
integrations drive files upload --path=/tmp/scan.xlsx --name="NYC Assemblage Scan — 2026-03-24" --convert --json
```

Use openpyxl with: dark blue headers, color-coded potential scores (green=High, yellow=Moderate, red=Low), proper column widths.

---

## Tool 17: Professional PDF Report (via LaTeX)

Write a .tex file, compile with `pdflatex -interaction=nonstopmode`, upload to Drive:
```bash
integrations drive files upload --path=/tmp/report.pdf --name="NYC Assemblage Report — 2026-03-24.pdf" --json
```

Use booktabs tables, navy section headers, fancyhdr, hyperlinked URLs. Escape `$`, `#`, `%`, `&`, `_` characters. Verify compilation is clean — fix and recompile if errors.

---

## Tool 18: Google Drive CLI

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
| HPD open violations 5-9 | +2 | HPD |
| HPD open violations 10+ | +4 (not cumulative with above) | HPD |
| Adjacent lot also for sale | +4 | Zillow + PLUTO |
| ACRIS party name contains "ESTATE OF" or "EXECUTOR" | +5 | ACRIS |
| 311 complaints 10+ in 12 months | +3 | 311 |
| Local Law 97 grade D or F | +3 | LL97 Energy |
| Defaulted ECB/OATH violations | +2 | ECB |
| Eviction filings in last 12 months | +2 | Housing Court |
| Scaffolding permit 3+ years old | +2 | DOB |
| Active ULURP rezoning nearby (upzone) | +3 | City Planning |
| Citi Bike 5+ stations within 1km | +2 | Citi Bike |
| HMDA investor loan spike in census tract | +2 | HMDA |
| Google Trends momentum > +15% (rising) | +3 | Trends |
| Obituary name matches ACRIS property owner | +5 | Obituaries + ACRIS |
| 5+ new liquor licenses in ZIP in 6 months | +3 | NY SLA |
| FDNY vacate order on building | +5 | FDNY |
| New CO issued on same block (active development) | +2 | DOB CO |
| Open DOB complaints (unsafe/illegal) | +2 | DOB Complaints |
| Census tract rent burden > 30% | +2 | Census ACS |

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
   - 311: complaint volume at address in last 12 months
   - LL97 energy grade (if building >25k sqft)
   - ECB/OATH: defaulted environmental violations
   - Housing Court: recent eviction filings
   - DOB sidewalk sheds: scaffolding permits 3+ years old
   - Citi Bike: station density within 1km (transit signal)
   - HMDA: investor loan activity in census tract
6. **For each qualifying block, check for cluster signals:**
   - Multiple Zillow listings on same block?
   - Recent ACRIS deed transfers to LLCs on same block?
   - DOB permits on same block?
   - Active ULURP rezoning applications in the area?
   - Google Trends: neighborhood momentum (rising/stable/declining)?
   - NY SLA: new liquor license count in ZIP (gentrification signal)
   - Obituaries: cross-ref recent deaths with ACRIS property ownership
   - FDNY: vacate orders on building
   - DOB CO: new certificates of occupancy on same block
   - DOB Complaints: open complaints (unsafe, illegal conversion)
   - Census ACS: tract demographics (rent burden, vacancy, income)
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
