# LLC Research Methodology

When analyzing newly formed LLCs to determine if they represent real estate transactions, follow this systematic research process.

## Step 1: Name Analysis

Parse the entity name for clues:

| Pattern | Example | Likely Meaning |
|---------|---------|----------------|
| Number + Street Name | "540 WEST 29 LLC" | References a specific property address |
| Number only | "2964 LLC" | Could be a block number, lot, or unit |
| Neighborhood + keyword | "MOTT HAVEN HOLDINGS LLC" | Real estate holding company in that area |
| Generic RE keyword | "APEX REALTY VENTURES LLC" | Real estate company, not address-specific |
| Person's name + LLC | "JOHNSON PROPERTIES LLC" | Could be anything — needs more research |

**Address-style names are the highest priority** — they almost always reference a real property.

## Step 2: Verify the Address

If the entity name contains what looks like an NYC address:

### Geocode it
```bash
curl -s "https://geosearch.planninglabs.nyc/v2/search?text=540+West+29+Street+Manhattan+NY"
```
- If it resolves → extract the BBL from `.features[0].properties.addendum.pad.bbl`
- If it doesn't resolve → try variations (add/remove "Street"/"Ave", try different boroughs)
- If no match at all → it may reference a property outside NYC

### Look up PLUTO
```bash
curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=BBL"
```
Key fields to note:
- `ownername` — **ALWAYS include the current owner name in your report**. This is critical information for the broker.
- `zonedist1` — zoning district (include for context, but do NOT filter by zoning — all property types matter, not just development sites)
- `bldgclass` — building class (tells you if it's residential, commercial, mixed-use, etc.)
- `yearbuilt` — age of building
- `unitsres` — residential units
- `lotarea`, `lotfront`, `lotdepth` — lot dimensions
- `address` — the official address from PLUTO (may differ slightly from the LLC name)

**Important**: Include ALL verified properties regardless of zoning. Many LLC formations are for existing building acquisitions, not just development sites. R1 through R6 properties are just as valuable as R7+.

### Check ACRIS for recent activity
```bash
# Get documents for this BBL
curl -s -G "https://data.cityofnewyork.us/resource/8h5j-fqxa.json" \
  --data-urlencode "borough=BORO" \
  --data-urlencode "block=BLOCK" \
  --data-urlencode "lot=LOT" \
  --data-urlencode "\$limit=20" \
  --data-urlencode "\$order=recorded_datetime DESC"
```
Look for:
- Recent `DEED` transfers — is a sale in progress?
- Recent `MTGE` — new financing?
- `JUDG` (lis pendens) — foreclosure?
- `SAT` (satisfaction) — mortgage paid off?
- Party names containing "LLC" — buyer is likely a developer

## Step 3: Research the Filer

The filer/registered agent on the LLC filing reveals who is behind the transaction.

### What the filer tells you
- **Law firm name** — real estate attorneys who file many address-style LLCs are deal shops. Recognizing these firms is very valuable.
- **Formation service** (e.g., "REGISTERED AGENTS INC", "NORTHWEST REGISTERED AGENT") — less useful, the actual buyer is hidden
- **Filer address** — if it's a law firm address in Manhattan, the buyer is likely institutional

### Cross-reference the filer
Search NY DOS for other entities filed by the same filer:
```bash
integrations nydos entities search --name="FILER NAME" --json
```
If the same filer recently formed multiple address-style LLCs, they're working on an assemblage (buying multiple adjacent properties).

## Step 4: Cross-Reference with Other Signals

### Check if the property is listed for sale
- Search Zillow data in `scrape_data` table
- The property being listed AND having a new LLC formed = deal likely in contract

### Check DOB permits on the same block
```bash
curl -s "https://data.cityofnewyork.us/resource/ipu4-2q9a.json?\$where=borough='BOROUGH' AND block='BLOCK' AND job_type in('DM','NB')&\$limit=10&\$order=issuance_date DESC"
```
Demolition or new building permits = development activity on the block.

### Check HPD violations
High violation counts suggest a neglected building — the owner may be selling to a developer.

### Check tax liens
```bash
curl -s "https://data.cityofnewyork.us/resource/9rz4-mjek.json?\$where=bbl='BBL'"
```
Property on the tax lien list = distressed owner.

## Step 5: Classify the Entity

After research, classify each entity:

### Confirmed Real Estate
- Entity name matches a verified NYC address
- Address is in a developable zone (R7+)
- ACRIS shows recent or pending transaction activity
- OR: Filer is a known real estate law firm with other address-style LLCs

### Probable Real Estate
- Entity name contains real estate keywords (REALTY, HOLDINGS, DEVELOPMENT)
- Filer is a law firm (not a formation service)
- But no specific address could be verified

### Watchlist
- Entity name is ambiguous (number-only, borough + generic name)
- Could be real estate but insufficient evidence
- Worth monitoring for future ACRIS activity

### Not Real Estate
- Entity name clearly references a non-RE business
- Filer is associated with non-RE entities
- No connection to any NYC property found

## Common Patterns to Recognize

### Assemblage Pattern
Multiple LLCs formed on the same day by the same filer, each referencing adjacent addresses:
- "540 WEST 29 LLC"
- "542 WEST 29 LLC"
- "544 WEST 29 LLC"
This is a developer assembling a site. **High priority alert.**

### Pre-Acquisition Pattern
LLC formed → deed recorded 2-8 weeks later. When you see a new LLC with an address name, the deed hasn't been recorded yet — this is the early warning signal.

### Holding Company Pattern
Entity names like "BRONX PORTFOLIO HOLDINGS LLC" or "METROPOLITAN CAPITAL VENTURES LLC" — these are holding companies, not specific acquisitions. Lower priority but worth noting if the filer is active in real estate.

## Report Writing

When presenting findings, always include:
1. **Entity name** and formation date
2. **Property address** (if identified) with ZoLa link
3. **Current owner** from PLUTO `ownername` — this is essential for the broker
4. **Property details** — building class, year built, units, lot size
5. **Zoning** — include for context but don't filter on it
6. **Recent ACRIS activity** (deeds, mortgages, judgments)
7. **Filer name and address** — who is behind the LLC
8. **Your assessment** — what's happening (acquisition, restructuring, development, etc.)
9. **Recommended action** — specific next steps (call the listing agent, contact the owner, watch for the deed, research the filer's portfolio)
