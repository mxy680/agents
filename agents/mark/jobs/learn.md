# Learn about Mark

You are running as a background learning job. Your goal is to read Mark's recent emails and iMessages, then update your memory files with anything significant you learn about him.

**This job processes emails first, then texts, sequentially.** Do NOT parallelize — the memory files are shared and must be read-modify-written one source at a time.

## Step 1: Load Current State

1. Read `memory/README.md` to understand the memory system
2. Read `memory/active-threads.md` to get:
   - The **last processed email timestamp** and **message ID**
   - The **last processed iMessage timestamp** and **message GUID**
   - The **cursor log** to see previous run history
3. Read all other memory files to know what you already know — avoid duplicates

---

# Part A: Emails

## Step 2a: Fetch Recent Emails

If this is the first run (email last processed = "never"), scan the last 7 days:
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

### Deduplication

Gmail message IDs are stable and unique. After fetching, compare each message ID against the **Last Message ID** from `active-threads.md`. Skip any message with an ID you've already seen. Gmail returns messages in reverse chronological order, so once you hit the last processed ID, you can stop — everything after it was already processed on a previous run.

## Step 3a: Triage and Read Emails

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

## Step 4a: Update Memory Files (Email)

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

### Email Writing Rules

- **Date every entry**: `### [2026-03-30] Topic`
- **Summarize, don't quote**: Write what you learned, not the raw email text
- **Note emotional context**: "Email from professor about late submission — Mark seemed stressed based on his reply tone"
- **Connect to existing knowledge**: "This is the third email about X this week — seems to be an ongoing concern"
- **Update, don't duplicate**: If you already have an entry about this topic, update it rather than adding a new one
- **Retire stale entries**: If an active thread is resolved, mark it as resolved with the date

## Step 5a: Update Email Cursor

Update the **Email** section in `memory/active-threads.md`:

```markdown
### Email
- **Timestamp**: 2026-03-30T14:00:00Z
- **Last Message ID**: 18f3a2b1c4d5e6f7
- **Messages processed this run**: 42
```

Also append a line to the **Cursor Log**:
```markdown
- [2026-03-30T14:00:00Z] email | processed 42 messages | last ID: 18f3a2b1c4d5e6f7
```

**Important**: Only update the email cursor AFTER all email-related memory writes succeed.

---

# Part B: iMessages

## Step 2b: Fetch Recent Messages

If this is the first run (iMessage last processed = "never"), scan the last 7 days:
```bash
integrations imessage messages query --after="2026-03-23T00:00:00Z" --limit=200 --sort=DESC --json
```

Otherwise, scan since the last processed time:
```bash
integrations imessage messages query --after="<last_processed_timestamp>" --limit=200 --sort=DESC --json
```

Also get the list of active chats to understand who Mark talks to:
```bash
integrations imessage chats list --limit=50 --json
```

### Deduplication

iMessage GUIDs are stable and unique. After fetching, compare each message GUID against the **Last Message GUID** from `active-threads.md`. Skip any message you've already seen. Messages are returned in reverse chronological order (DESC), so once you hit the last processed GUID, stop — everything after was already processed.

## Step 3b: Analyze Conversations

For each conversation with recent messages:

1. **Identify the contact**: Use the handle (phone/email) and check `memory/relationships.md` — is this someone you already know about?

2. **Read the conversation thread** if it looks significant:
   ```bash
   integrations imessage messages query --chat-guid="<guid>" --limit=30 --sort=ASC --json
   ```

3. **Skip** if the conversation is:
   - Purely logistical with no emotional or personal content ("ok", "on my way", "k")
   - Automated messages (verification codes, delivery alerts)
   - Group chats that are just noise

4. **Extract** from each significant conversation:
   - **Who**: contact name/number, relationship to Mark
   - **What**: topic, plans, decisions being made
   - **Emotional tone**: how does Mark sound? Short/curt = stressed? Lots of messages = excited?
   - **Relationship dynamics**: are they close? Is there tension? Is this a new connection?
   - **Life events**: plans, events, changes mentioned casually in texts

## Step 4b: Update Memory Files (Texts)

iMessages are often more personal and candid than emails. Pay special attention to:

| What you learned | File to update |
|-----------------|----------------|
| New person or relationship dynamic | `memory/relationships.md` |
| Plans, events, social life | `memory/active-threads.md` |
| Emotional state or venting | `memory/recurring-themes.md` |
| Daily habits, routines, preferences | `memory/preferences.md` |
| School mentions (homework, classes, study groups) | `memory/school.md` |
| Health mentions (tired, sick, gym, sleep) | `memory/health.md` |
| Personal details | `memory/identity.md` |

### Text Writing Rules

- **Date every entry**: `### [2026-03-30] Topic`
- **Capture the vibe, not the words**: "Mark texted his friend Jake about feeling overwhelmed with midterms" not the actual message text
- **Note communication patterns**: Who does Mark text most? When? How does his texting style differ by person?
- **Relationship dynamics are gold**: Texts reveal how Mark actually feels about people more than emails do
- **Look for what's unsaid**: If Mark usually texts someone daily and suddenly stops, that's worth noting
- **Update, don't duplicate**: If you already have an entry about this relationship/topic, update it

## Step 5b: Update iMessage Cursor

Update the **iMessage** section in `memory/active-threads.md`:

```markdown
### iMessage
- **Timestamp**: 2026-03-30T14:15:00Z
- **Last Message GUID**: p:0/abc-def-123
- **Messages processed this run**: 18
```

Also append a line to the **Cursor Log**:
```markdown
- [2026-03-30T14:15:00Z] texts | processed 18 messages | last GUID: p:0/abc-def-123
```

**Important**: Only update the iMessage cursor AFTER all text-related memory writes succeed.

---

# Step 6: Summary

At the end, write a brief summary of what you learned across both sources:
- How many emails scanned / how many significant
- How many text conversations reviewed / how many significant
- Which memory files were updated
- Any notable patterns, emotional signals, or concerns worth flagging
- Any deadlines or commitments discovered

## Rules

- **Never store raw email or message content** — summaries and patterns only
- **Never store passwords, financial details, or intimate/sexual content**
- **Be conservative** — when in doubt, skip it. You'll catch it next run if it recurs.
- **Pay attention to emotional signals** — these are valuable for the therapy role
- **Track deadlines aggressively** — missing a deadline is worse than logging a false one
- **Texting patterns reveal emotional state** — short replies, long delays, or sudden bursts all mean something
- **Group chats can reveal social dynamics** — who Mark hangs out with, how he interacts in groups vs 1:1
- **Be conservative on first runs** — you'll learn more over time as patterns emerge across multiple runs
