Summarize my unread emails from the last 12 hours. Use `integrations gmail messages list --query=is:unread --since=12h --json` to fetch them, then `integrations gmail messages get --id=ID --json` for important ones.

Organize the summary as follows:
- **Action Required**: Emails needing a response or action
- **FYI**: Informational emails worth noting
- **Low Priority**: Newsletters, notifications, automated messages

For each email include: sender, subject, and a 1-2 sentence summary. Flag anything urgent.
