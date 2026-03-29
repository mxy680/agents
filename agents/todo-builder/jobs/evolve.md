You are running an autonomous evolution cycle for the todo application.

## Step 0 — Read known issues

**BEFORE doing anything else**, read the known issues file:
```bash
cat agents/todo-builder/known-issues.md
```
This file contains issues from prior runs. Avoid repeating them. If you encounter a NEW issue during this run, append it to the file:
```bash
echo "- [$(date +%Y-%m-%d)] [description] → [solution]" >> agents/todo-builder/known-issues.md
```

## Hardcoded Facts

- **GitHub owner**: `engagentdev` (NOT `mxy680`)
- **Linear `--team` requires UUID** — list teams with `integrations linear teams list --json` to get the UUID. Do NOT use the key (e.g., "ENG").
- **Linear has no `--status` flag** — to close an issue, get the Done state ID from `integrations linear workflows list --team=<uuid> --json`, then use `--state=<done-state-id>`.
- **Use Next.js 15.5.14** — other versions may have build issues.
- **Vercel get/delete uses `--project`**, not `--name`.
- **GitHub contents create fails on empty repos** — create a README first.
- **Create a separate Linear team** called `Todo App` for this project — do NOT use the Engagent team.

## Step 1 — Set up Linear team

First, check if a `Todo App` Linear team exists:
```bash
integrations linear teams list --json
```
If there's no team with name "Todo App", create one. If the CLI doesn't support team creation, use the existing team but create issues with a `[todo-app]` prefix in the title.

Save the team UUID — you'll need it for all Linear operations in this run.

## Step 2 — Check current state

Check if the GitHub repo exists:

```bash
integrations github repos get --repo=todo-app --owner=engagentdev --json
```

### If the repo does NOT exist (first run):

Create everything from scratch:

1. **Create the GitHub repo** under the `engagentdev` org:
```bash
integrations github repos create --name=todo-app --description="Autonomous todo app built by Engagent" --private=false --json
```

2. **Create a Linear issue** for tracking (use team UUID, not key):
```bash
integrations linear issues create \
  --title="Initial scaffold: todo app" \
  --team=<todo-app-team-uuid> \
  --description="Create GitHub repo, scaffold Next.js 15 app, set up Supabase, deploy to Vercel." \
  --priority=2 \
  --json
```

3. **Create a NEW Supabase project** (NEVER use the existing `engagent` project):
```bash
integrations supabase orgs list --json
integrations supabase projects create --name=todo-app --org-id=<org_id> --db-pass=<generate_strong_password> --region=us-east-1 --json
```
Wait 2 minutes for provisioning, then get the project details:
```bash
integrations supabase projects list --json
integrations supabase projects api-keys --ref=<todo-app-ref> --json
```
Extract the project URL, anon key, and service role key.

4. **Create the todos table** on the NEW Supabase project:
```bash
curl -s -X POST "<NEW_PROJECT_URL>/rest/v1/rpc" \
  -H "apikey: <SERVICE_ROLE_KEY>" \
  -H "Authorization: Bearer <SERVICE_ROLE_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"query": "CREATE TABLE IF NOT EXISTS todos (id uuid DEFAULT gen_random_uuid() PRIMARY KEY, title text NOT NULL, completed boolean DEFAULT false, created_at timestamptz DEFAULT now()); ALTER TABLE todos ENABLE ROW LEVEL SECURITY; CREATE POLICY \"Public access\" ON todos FOR ALL USING (true);"}'
```

5. **Scaffold the Next.js app** — write files to temp, then push to GitHub:
   - Use **Next.js 15.5.14** (not 15.0.0 or latest — those have build issues)
   - Include: `package.json`, `tsconfig.json`, `next.config.mjs`, `tailwind.config.ts`, `postcss.config.mjs`, `.gitignore`, `app/globals.css`, `app/layout.tsx`, `app/page.tsx`, `lib/supabase.ts`
   - The Supabase URL and anon key should be hardcoded in the app's `.env.local` or referenced via `NEXT_PUBLIC_SUPABASE_URL` / `NEXT_PUBLIC_SUPABASE_ANON_KEY`

   For the FIRST commit to an empty repo, create a README first:
   ```bash
   cat > /tmp/readme.md << 'EOF'
   # Todo App
   Autonomous todo app built by Engagent.
   EOF
   CONTENT=$(base64 < /tmp/readme.md)
   integrations github repos contents create \
     --repo=todo-app --owner=engagentdev \
     --path=README.md --message="Initial commit" \
     --content="$CONTENT" --json
   ```
   Then create/update each subsequent file one at a time (each is a separate commit, that's fine).

6. **Create a Vercel project**:
```bash
integrations vercel projects create --name=todo-app --framework=nextjs --json
```

7. **Set env vars on Vercel** (use the NEW Supabase project's keys):
```bash
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_URL --value="<new_project_url>" --target=production --json
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_ANON_KEY --value="<new_anon_key>" --target=production --json
```

8. **Deploy to Vercel** — if auto-deploy from GitHub isn't working, deploy files directly:
   ```bash
   # For each file, compute SHA1 and upload
   SHA=$(shasum /tmp/myfile.tsx | cut -d' ' -f1)
   SIZE=$(wc -c < /tmp/myfile.tsx)
   curl -s -X POST "https://api.vercel.com/v2/files" \
     -H "Authorization: Bearer ${VERCEL_TOKEN}" \
     -H "Content-Type: application/octet-stream" \
     -H "x-vercel-digest: $SHA" \
     -H "Content-Length: $SIZE" \
     --data-binary @/tmp/myfile.tsx

   # Then create deployment with all file references
   curl -s -X POST "https://api.vercel.com/v13/deployments" \
     -H "Authorization: Bearer ${VERCEL_TOKEN}" \
     -H "Content-Type: application/json" \
     -d '{"name":"todo-app","files":[{"file":"package.json","sha":"...","size":N}, ...],"projectSettings":{"framework":"nextjs"},"target":"production"}'
   ```

9. **Verify deployment** — wait 60 seconds, then check:
```bash
sleep 60
curl -s -o /dev/null -w "%{http_code}" <deployment-url>
```

10. **Close the Linear issue** — first get the Done state ID:
```bash
integrations linear workflows list --team=<todo-app-team-uuid> --json
```
Find the state with `type: "completed"`, then:
```bash
integrations linear issues update --id=<issue-id> --state=<done-state-id> --json
```

### If the repo DOES exist (subsequent runs):

1. **Check what's already been built**:
```bash
integrations github repos contents list --repo=todo-app --owner=engagentdev --path=/ --json
integrations github repos commits list --repo=todo-app --owner=engagentdev --limit=10 --json
```

2. **Check Linear for completed work** (use team UUID):
```bash
integrations linear issues list --team=<todo-app-team-uuid> --json
```

3. **Decide what to build next**. Choose from this priority list (skip anything already done):
   - Authentication (Supabase Auth with email/password)
   - Categories/labels for todos (with color coding)
   - Due dates with overdue highlighting
   - Priority levels (high/medium/low)
   - Search and filter
   - Dark/light theme toggle
   - Drag-and-drop reordering
   - Keyboard shortcuts
   - Mobile responsive improvements
   - Analytics dashboard (completed vs pending over time)
   - Recurring todos
   - Sharing/collaboration
   - ...or anything else you think would improve the app

4. **Create a Linear issue** (use team UUID):
```bash
integrations linear issues create \
  --title="Add [feature]" \
  --team=<todo-app-team-uuid> \
  --description="Description of what will be built" \
  --priority=2 \
  --json
```

5. **Implement the feature** — get file SHA, update content:
```bash
# Get current file SHA
integrations github repos contents get --repo=todo-app --owner=engagentdev --path=<filepath> --json
# Update file
cat > /tmp/updated-file.tsx << 'EOF'
// content
EOF
CONTENT=$(base64 < /tmp/updated-file.tsx)
integrations github repos contents update \
  --repo=todo-app --owner=engagentdev \
  --path=<filepath> --message="feat: [description]" \
  --content="$CONTENT" --sha=<sha> --json
```

6. **Deploy** — wait for auto-deploy or deploy files directly (see step 8 above).

7. **Verify** — curl the production URL for HTTP 200.

8. **Close the Linear issue** using the Done state ID (see step 10 above).

## Step 2 — Generate report

Summarize:
- State before this run
- What was built/changed
- GitHub commits
- Deployment URL (with HTTP status)
- Linear issue created/closed
- Recommended next feature

## Important

- Use **Next.js 15.5.14** — other versions may have build issues.
- Write file content to temp files and encode with `base64 < /tmp/file`.
- Never delete existing functionality unless refactoring.
- If something fails, read the error and fix it. Don't retry the same thing.
- The GitHub org is **engagentdev** (not mxy680).
- Use the `Todo App` Linear team (find UUID via `integrations linear teams list --json`).
- NEVER use the existing `engagent` Supabase project.
