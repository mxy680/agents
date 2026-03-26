# LLC Entity Monitor

## Data Flow

1. Pull all new LLC formations from NY DOS daily filings (last 24 hours)
2. Pattern match entity names against real estate conventions:
   - Contains real estate keywords: REALTY, HOLDINGS, PROPERTIES, DEVELOPMENT, EQUITIES, CAPITAL, ASSOCIATES
   - Looks like an address-based LLC (e.g., "1776 SEMINOLE AVE LLC")
   - Contains NYC address patterns (numbers + street names + borough indicators)
3. For matches, look up filer info and related entities at the same mailing address
4. Output daily digest of likely real estate entity formations

---

## Manual Search Mode

When given a property address, search for entities formed in the last 90 days that match:

```bash
integrations nydos entities match-address --address="1776 Seminole Ave" --since=2025-12-01 --json
```

---

## NY DOS API

```bash
# Recent LLC formations
integrations nydos entities recent --since=2026-03-25 --type=llc --json

# Search by name
integrations nydos entities search --name="SEMINOLE" --json

# Match property address
integrations nydos entities match-address --address="1776 Seminole Ave" --json
```

---

## NYC DOF for Pattern Learning

```bash
# Find properties owned by LLCs (learn naming patterns)
integrations dof owners by-entity --pattern="LLC" --borough=2 --limit=100 --json
```

---

## Tools

- **NY DOS entities CLI** (`nydos`) — entity formations, search, address matching
- **NYC DOF owners CLI** (`dof`) — property owner lookup, entity pattern learning
- **Google Drive** — `integrations drive files upload` for uploading alert reports

## Output

Matches saved to `/tmp/llc_monitor/matches_YYYY-MM-DD.json`. Each entry includes:
- `name` — entity name
- `filer_name` — person/firm who filed
- `process_address` or `filer_address` — mailing address on file
- `match_reason` — why it was flagged (keyword, address pattern, borough)
