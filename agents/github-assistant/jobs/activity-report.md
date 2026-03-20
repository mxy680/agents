Generate a daily activity report for my GitHub repositories. Use the `integrations` CLI:

1. List my repos: `integrations github repos list --sort=pushed --limit=10 --json`
2. For each recently active repo, check:
   - Open PRs: `integrations github pulls list --owner=OWNER --repo=REPO --state=open --json`
   - Recent issues: `integrations github issues list --owner=OWNER --repo=REPO --state=open --sort=updated --limit=5 --json`

Organize the summary as:
- **PRs Needing Review**: Open PRs awaiting review
- **CI Status**: Any failing workflow runs
- **New Issues**: Issues opened in the last 24 hours
- **Recently Merged**: PRs merged in the last 24 hours

Keep it concise — one line per item with a link.
