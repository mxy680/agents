# Learn from Emails

You are running as a background learning job. Your goal is to read Mark's recent emails and update your memory files with anything significant you learn about him.

## Step 1: Load Current State

1. Read `memory/README.md` to understand the memory system
2. Read `memory/active-threads.md` to get the **last processed** email timestamp
3. Read all other memory files to know what you already know — avoid duplicates

## Step 2: Fetch Recent Emails

If this is the first run (last processed = "never"), scan the last 7 days:
```bash
integrations gmail messages list --since=7d --limit=200 --json
```

Otherwise, scan since the last processed time. Use Gmail's `after:` operator:
```bash
integrations gmail messages list --query="after:YYYY/MM/DD" --limit=200 --json
```

Also check for unread emails specifically:
```bash
integrations gmail messages list --query="is:unread" --json
```

## Step 3: Triage and Read

For each email in the list:

1. **Skip immediately** if it's clearly:
   - A newsletter or marketing email (check sender domain, subject patterns)
   - An automated notification (GitHub, Vercel, Linear, etc.)
   - Spam or promotional content
   - A package tracking or shipping notification

2. **Read the full email** if it looks significant:
   ```bash
   integrations gmail messages get --id=<message_id> --json
   ```

3. **Extract** from each significant email:
   - **Who**: sender/recipient name, email, relationship to Mark
   - **What**: topic, context, any action items or deadlines
   - **Emotional tone**: is this stressful, exciting, routine, concerning?
   - **Patterns**: does this connect to something you already know?

## Step 4: Update Memory Files

For each piece of significant information, update the appropriate memory file:

| What you learned | File to update |
|-----------------|----------------|
| New person or relationship context | `memory/relationships.md` and `memory/email-contacts.md` |
| School-related (courses, deadlines, professors) | `memory/school.md` |
| Work/career-related | `memory/work.md` |
| Personal details (location, background) | `memory/identity.md` |
| Preferences or habits | `memory/preferences.md` |
| Emotional patterns or stressors | `memory/recurring-themes.md` |
| Ongoing situation or commitment | `memory/active-threads.md` |
| Health-related mentions | `memory/health.md` |

### Writing Rules

- **Date every entry**: `### [2026-03-30] Topic`
- **Summarize, don't quote**: Write what you learned, not the raw email text
- **Note emotional context**: "Email from professor about late submission — Mark seemed stressed based on his reply tone"
- **Connect to existing knowledge**: "This is the third email about X this week — seems to be an ongoing concern"
- **Update, don't duplicate**: If you already have an entry about this topic, update it rather than adding a new one
- **Retire stale entries**: If an active thread is resolved, mark it as resolved with the date

## Step 5: Update Timestamps

Update `memory/active-threads.md` with the new last-processed timestamp:

```markdown
## Last Processed
- **Email**: 2026-03-30T14:00:00Z
- **iMessage**: [leave unchanged]
```

## Step 6: Summary

At the end, write a brief summary of what you learned:
- How many emails scanned
- How many were significant
- Which memory files were updated
- Any notable patterns or concerns worth flagging

## Important

- **Never store raw email content** — summaries and patterns only
- **Never store passwords, financial details, or intimate content**
- **Be conservative** — when in doubt about whether something is significant, skip it. You'll catch it next run if it comes up again.
- **Pay attention to emotional signals** — these are valuable for the therapy role
- **Track deadlines aggressively** — missing a deadline is worse than logging a false one
