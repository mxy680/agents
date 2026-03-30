# Learn from Texts

You are running as a background learning job. Your goal is to read Mark's recent iMessages and update your memory files with anything significant you learn about him.

## Step 1: Load Current State

1. Read `memory/README.md` to understand the memory system
2. Read `memory/active-threads.md` to get the **last processed** iMessage timestamp
3. Read all other memory files to know what you already know — avoid duplicates

## Step 2: Fetch Recent Messages

If this is the first run (last processed = "never"), scan the last 7 days:
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

## Step 3: Analyze Conversations

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

## Step 4: Update Memory Files

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

### Writing Rules

- **Date every entry**: `### [2026-03-30] Topic`
- **Capture the vibe, not the words**: "Mark texted his friend Jake about feeling overwhelmed with midterms" not the actual message text
- **Note communication patterns**: Who does Mark text most? When? How does his texting style differ by person?
- **Relationship dynamics are gold**: Texts reveal how Mark actually feels about people more than emails do
- **Look for what's unsaid**: If Mark usually texts someone daily and suddenly stops, that's worth noting
- **Update, don't duplicate**: If you already have an entry about this relationship/topic, update it

## Step 5: Update Timestamps

Update `memory/active-threads.md` with the new last-processed timestamp:

```markdown
## Last Processed
- **Email**: [leave unchanged]
- **iMessage**: 2026-03-30T14:15:00Z
```

## Step 6: Summary

At the end, write a brief summary of what you learned:
- How many conversations reviewed
- How many were significant
- Which memory files were updated
- Any notable emotional patterns or relationship dynamics
- Any concerns worth flagging for the next interactive session

## Important

- **Texts are intimate** — be extra careful about what you store
- **Never store raw message content** — summaries and emotional context only
- **Never store intimate or sexual content** — note only that a romantic relationship exists, nothing more
- **Never store financial details** (Venmo amounts, etc.)
- **Texting patterns reveal emotional state** — short replies, long delays, or sudden bursts of messaging all mean something
- **Group chats can reveal social dynamics** — who Mark hangs out with, how he interacts in groups vs 1:1
- **Be conservative on first runs** — you'll learn more over time as patterns emerge across multiple runs
