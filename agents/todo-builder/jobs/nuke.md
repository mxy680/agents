Delete EVERYTHING related to the todo app. This is a complete reset.

## WARNING: This is irreversible. Delete all resources across all platforms.

## Step 1 — Delete GitHub repo

```bash
integrations github repos delete --repo=todo-app --owner=engagentdev --confirm --json
```

If the repo doesn't exist, skip this step.

## Step 2 — Delete Vercel project

```bash
integrations vercel projects delete --project=todo-app --confirm --json
```

If the project doesn't exist, skip this step.

## Step 3 — Delete Supabase project

First find the todo-app project:
```bash
integrations supabase projects list --json
```

Find the project named `todo-app` (NOT `engagent`), then delete it:
```bash
integrations supabase projects delete --ref=<todo-app-ref> --confirm --json
```

**NEVER delete the `engagent` Supabase project.**

## Step 4 — Delete Linear issues and team

First find the Todo App team:
```bash
integrations linear teams list --json
```

If a `Todo App` team exists, delete all its issues:
```bash
integrations linear issues list --team=<todo-app-team-uuid> --json
```

Delete each issue:
```bash
integrations linear issues delete --id=<issue-id> --confirm --json
```

If the CLI supports team deletion, delete the team too.

Also check the Engagent team for any `[todo-app]` prefixed issues and delete those:
```bash
integrations linear issues list --team=<engagent-team-uuid> --json
```

## Step 5 — Clean up known issues file

Reset the known issues file:
```bash
cat > agents/todo-builder/known-issues.md << 'EOF'
# Known Issues

Write any issues you encounter here so you don't repeat them in future runs.
Format: `- [date] [issue description] → [solution]`

## Issues Log

EOF
```

## Step 6 — Report

Confirm what was deleted:
- GitHub repo: deleted / not found
- Vercel project: deleted / not found
- Supabase project: deleted / not found
- Linear issues: N deleted
- Known issues: reset

Everything is clean for a fresh start.
