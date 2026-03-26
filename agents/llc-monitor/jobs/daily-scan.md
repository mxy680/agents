Run the daily LLC formation scan to identify potential real estate buyers before deals close.

## Steps

1. Run the scan script:
   ```bash
   bash agents/llc-monitor/scripts/run_scan.sh
   ```

2. The script will:
   - Pull new LLC formations from NY DOS (last 24 hours)
   - Pattern match against real estate naming conventions
   - For matches, include filer info and mailing address for related entity lookup
   - Save results to `/tmp/llc_monitor/matches_YYYY-MM-DD.json`

3. Report the following in your response:
   - How many new LLCs were filed today
   - How many matched real estate patterns
   - Breakdown by match type (keyword, address pattern, borough indicator)
   - Top 20 matches with: entity name, filer, mailing address, match reason
