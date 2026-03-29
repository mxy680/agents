# Nuke Everything — CampusReach Complete Reset

WARNING: This job deletes ALL CampusReach resources. It is irreversible. Only run this manually when you need a complete reset.

IMPORTANT: NEVER touch the Engagent Supabase project. Only delete the CampusReach project (ref: gzbkcivzxcctzvrhgqos).

## Step 1: Delete GitHub Repository Contents

List all files at the root of the repo and delete them one by one, or delete the repo itself if that's simpler:

```bash
# List all files
integrations github repos contents list --repo=campusreach --owner=engagentdev --path=. --json
```

For each file/directory, get the SHA and delete it:
```bash
integrations github repos contents get --repo=campusreach --owner=engagentdev --path=<path> --json
# Then delete using the SHA
integrations github repos contents delete \
  --repo=campusreach \
  --owner=engagentdev \
  --path=<path> \
  --message="chore: nuke — complete reset" \
  --sha=<sha> \
  --json
```

Alternatively, delete the entire repository:
```bash
integrations github repos delete --repo=campusreach --owner=engagentdev --json
```

## Step 2: Delete Vercel Project

```bash
integrations vercel projects delete --project=campusreach --json
```

## Step 3: Delete Supabase Project

CRITICAL: Only delete the CampusReach project (ref: gzbkcivzxcctzvrhgqos). NEVER delete the Engagent project.

Verify the project name before deleting:
```bash
integrations supabase projects get --ref=gzbkcivzxcctzvrhgqos --json
```

Confirm the name contains "campusreach" or "campus" before proceeding. Then delete:
```bash
integrations supabase projects delete --ref=gzbkcivzxcctzvrhgqos --json
```

## Step 4: Delete Linear Issues

Find the CampusReach team and delete all issues:
```bash
integrations linear teams list --json
integrations linear issues list --team=<uuid> --json
```

For each issue, delete or archive it:
```bash
integrations linear issues delete --issue=<id> --json
```

## Step 5: Reset Known Issues File

Reset `known-issues.md` to its blank template. Get the current SHA first, then update with blank content:

```bash
integrations github repos contents get \
  --repo=campusreach \
  --owner=engagentdev \
  --path=agents/campusreach/known-issues.md \
  --json
```

Then update with the blank template content (base64 encode the following):
```
# Known Issues

Write any issues you encounter here so you don't repeat them in future runs.
Format: `- [date] [issue description] → [solution]`

## Issues Log
```

```bash
echo "# Known Issues

Write any issues you encounter here so you don't repeat them in future runs.
Format: \`- [date] [issue description] → [solution]\`

## Issues Log
" | base64

integrations github repos contents update \
  --repo=campusreach \
  --owner=engagentdev \
  --path=agents/campusreach/known-issues.md \
  --message="chore: reset known issues after nuke" \
  --content=<base64> \
  --sha=<current-sha> \
  --json
```

## Step 6: Confirm

Report what was deleted:
- GitHub: repo deleted / contents cleared
- Vercel: project deleted
- Supabase: project deleted (ref: gzbkcivzxcctzvrhgqos)
- Linear: issues cleared
- known-issues.md: reset

The CampusReach slate is now clean.
