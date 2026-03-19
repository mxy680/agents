Generate a daily engagement report for my Instagram account. Use the `integrations` CLI:

1. Get profile overview: `integrations ig profile get --json`
2. Check recent activity: `integrations ig activity feed --limit=20 --json`
3. Get recent posts: `integrations ig media list --limit=5 --json`
4. For each recent post, get engagement: `integrations ig media likers --media-id=ID --limit=10 --json` and `integrations ig comments list --media-id=ID --limit=10 --json`

Organize the summary as:
- **Profile Stats**: Current follower/following counts
- **New Activity**: New followers, likes, and comments since yesterday
- **Top Performing Post**: Post with most engagement in the last 24 hours
- **Comment Highlights**: Notable comments that may need a response

Keep numbers prominent and include percentage changes if possible.
