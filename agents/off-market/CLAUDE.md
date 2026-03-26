# Off-Market R8+ Scanner

## Data Flow

1. **PLUTO query** — pull all R8+ small residential properties (paginated, ~2,000-5,000 lots)
2. **Signal checks** — for each property: ACRIS, DOB, HPD, NYC Finance, 311, ECB, FDNY, DOB complaints, CO, Citi Bike
3. **Scoring** — composite distress score (no Zillow signals)
4. **Cluster detection** — same-block groupings
5. **Verification** — deduplication, scoring consistency
6. **XLSX + PDF** — professional deliverables
7. **Upload** — Google Drive

---

## PLUTO Query

R8+ zoning filter: R8, R8A, R8B, R8X, R9, R9A, R9X, R10, R10A, R10X, C4-4, C4-5, C4-6, C4-7
Building class filter: A% (1-family), B% (2-family), C% (walk-up, 3-6 units), S% (mixed residential)
Units filter: unitsres <= 5

```bash
python3 agents/off-market/scripts/phase1_pluto_query.py
```

PLUTO has ~870K lots. R8+ small residential is approximately 2,000-5,000. Paginate with `$offset` in 5,000-row increments. BBL format: borough(1) + block(5, zero-padded) + lot(4, zero-padded).

---

## Signal Checks (Tools Available)

All NYC APIs are public Socrata endpoints — curl, no auth needed. Use `curl -s -G` with `--data-urlencode` for all `$where` clauses.

- **PLUTO** (`64uk-42ks`) — primary data source, zoning + lot data
- **ACRIS Legals** (`8h5j-fqxa`) — get document IDs by borough/block/lot
- **ACRIS Master** (`bnx9-e6tj`) — doc_type: JUDG (lis pendens), FL (federal lien), TLS (tax lien cert), CODP, CTOR
- **ACRIS Parties** (`636b-3b5g`) — detect "ESTATE OF", "EXECUTOR", LLC deed transfers
- **DOB Permits** (`ipu4-2q9a`) — job_type: DM (demo), NB (new building), SH (scaffolding)
- **HPD Violations** (`csn4-vhvf`) — open violations count; 10+ = owner burnout signal
- **NYC Finance Tax Liens** (`9rz4-mjek`) — any match = strong distress signal
- **311 Complaints** (`erm2-nwe9`) — HEAT/HOT WATER, Rodent; 10+ in 12 months = signal
- **ECB/OATH Violations** (`6bgk-3dad`) — defaulted fines
- **FDNY Vacate Orders** (`frax-hfgs`) — vacated buildings = motivated seller
- **DOB Complaints** (`eabe-havv`) — open unsafe/illegal complaints
- **Certificate of Occupancy** (`pkdm-hqz6`) — new CO on same block = active development
- **Citi Bike** — `integrations citibike stations density --lat=X --lng=Y --radius=1000 --json`
- **Google Drive** — `integrations drive files upload --path=... --name=... [--convert] --json`

No Zillow. No StreetEasy. No HMDA. No Trends. No Obituaries. No SLA. No Census. No NYSCEF.

---

## Scoring Model (Distress Signals Only)

| Signal | Points | Source |
|--------|--------|--------|
| R8+ zoning (vs R7) | +3 | PLUTO |
| Pre-war construction (before 1945) | +2 | PLUTO |
| Lot under 2,000 SF | +2 | PLUTO |
| Tax lien / delinquency | +4 | NYC Finance |
| Judgment filing (lis pendens) in last 90 days | +5 | ACRIS |
| Federal lien (IRS) | +3 | ACRIS |
| Deed transfer to LLC in last 12 months on same block | +3 | ACRIS |
| ACRIS party name contains "ESTATE OF" or "EXECUTOR" | +5 | ACRIS |
| Demolition permit on same block in last 6 months | +3 | DOB |
| New building permit on same block in last 6 months | +2 | DOB |
| Scaffolding permit 3+ years old | +2 | DOB |
| HPD open violations 5-9 | +2 | HPD |
| HPD open violations 10+ | +4 | HPD |
| 311 complaints 10+ in 12 months | +3 | 311 |
| Defaulted ECB/OATH violations | +2 | ECB |
| FDNY vacate order on building | +5 | FDNY |
| Open DOB complaints (unsafe/illegal) | +2 | DOB Complaints |
| New CO issued on same block (active development) | +2 | DOB CO |
| Citi Bike 5+ stations within 1km | +2 | Citi Bike |

**Priority tiers:**
- **20+** = Immediate outreach
- **15-19** = High priority
- **10-14** = Moderate priority
- **5-9** = Watchlist

---

## Important

- BBL: borough(1) + block(5, zero-padded) + lot(4, zero-padded)
- ACRIS block/lot fields must be zero-padded strings — never integers
- Process properties in batches — don't try to do all at once
- Escape special characters in LaTeX: `$` → `\$`, `#` → `\#`, `%` → `\%`, `&` → `\&`, `_` → `\_`
- XLSX and PDF are final deliverables — must be professional, accurate, and complete
