# Daily Probate Scan

Run the probate scan:

```bash
bash agents/probate-monitor/scripts/run_scan.sh
```

## Pipeline

1. Query ACRIS parties table for estate/probate filings from the last 7 days — party names containing "ESTATE OF", "EXECUTOR", or "EXECUTRIX"
2. Get BBLs from the ACRIS legals table for each matching document
3. Look up property details in PLUTO for each BBL
4. Filter: keep only properties with 15,000+ buildable SF (lot area * 6.0 FAR) or 5+ residential units
5. Deduplicate against previous alerts stored in `/tmp/probate_monitor/seen_bbls.json`
6. Output alert summary with property details

## Report

After the scan completes, report:
- Total estate documents found in ACRIS
- Total BBLs resolved
- Number of new (previously unseen) BBLs
- Number of qualifying alerts (15K+ buildable SF or 5+ units)
- For each new alert: address, borough, estate name, filing date, lot area, units, zoning, ZoLa URL
