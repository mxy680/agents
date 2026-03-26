# Probate Monitor

## Data Flow

1. Query ACRIS parties table for recent documents with party names containing "ESTATE OF", "EXECUTOR", or "EXECUTRIX"
2. Get the BBLs for those documents from ACRIS legals table
3. Look up each BBL in PLUTO for building details (units, lot area, building area, zoning)
4. Filter: keep only properties with `unitsres >= 5` OR `(lotarea * FAR_for_zone) >= 15000`
5. Check if we've already alerted on this property (dedup against previous runs via `/tmp/probate_monitor/seen_bbls.json`)
6. Output: alert list with property address, estate name, filing date, lot details

---

## ACRIS Probate Query

```bash
# Step 1: Find recent documents with estate party names
curl -s -G "https://data.cityofnewyork.us/resource/636b-3b5g.json" \
  --data-urlencode "$where=name like '%ESTATE OF%' OR name like '%EXECUTOR%' OR name like '%EXECUTRIX%'" \
  --data-urlencode "$order=good_through_date DESC" \
  --data-urlencode "$limit=500"

# Step 2: Get BBLs from legals table using document_ids
curl -s -G "https://data.cityofnewyork.us/resource/8h5j-fqxa.json" \
  --data-urlencode "$where=document_id in('DOC_ID_1','DOC_ID_2')"

# Step 3: Look up PLUTO
curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=BBL_HERE"
```

Always use `curl -s -G` with `--data-urlencode` for Socrata queries. Never inline `$where` directly in the URL.

---

## Alert Format

Each alert includes:
- `address`, `borough`, `bbl`
- `estate_name` (from ACRIS parties)
- `filing_date`
- `lot_area`, `bldg_area`, `zoning`
- `units_res`, `units_total`, `year_built`
- `zola_url` — `https://zola.planning.nyc.gov/lot/{boro}/{block}/{lot}`

---

## Tools

- **ACRIS parties** (`636b-3b5g`) — probate detection
- **ACRIS legals** (`8h5j-fqxa`) — BBL lookup by document_id
- **PLUTO** (`64uk-42ks`) — property details (units, lot area, zoning)
- **Google Drive** — `integrations drive files upload` for uploading alert reports

## Important

- BBL format: borough(1) + block(5, zero-padded) + lot(4, zero-padded)
- Seen BBLs stored at `/tmp/probate_monitor/seen_bbls.json` — persist across runs
- FAR estimate: use 6.0 as conservative default for buildable SF calculation
- Borough codes: 1=Manhattan, 2=Bronx, 3=Brooklyn, 4=Queens, 5=Staten Island
