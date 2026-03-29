You are running an autonomous evolution cycle for the todo application.

## Step 1 — Check current state

First, determine if the project exists:

```bash
integrations github repos get --repo=todo-app --owner=mxy680 --json
```

### If the repo does NOT exist (first run):

Create everything from scratch:

1. **Create the GitHub repo**:
```bash
integrations github repos create --name=todo-app --description="Autonomous todo app built by Engagent" --private=false --json
```

2. **Scaffold the app**: Create a complete Next.js 15 todo application with:
   - App Router
   - Tailwind CSS + shadcn/ui styling
   - Supabase for data storage (todos table)
   - CRUD operations: add, complete, delete todos
   - Clean, modern dark UI
   - A `package.json`, `tsconfig.json`, `next.config.mjs`, and all necessary files

   Upload each file to GitHub using:
   ```bash
   # Write content to a temp file first, then base64-encode it
   cat > /tmp/myfile.tsx << 'EOF'
   // file content here
   EOF
   CONTENT=$(base64 < /tmp/myfile.tsx)
   integrations github repos contents create \
     --repo=todo-app \
     --owner=mxy680 \
     --path=<filepath> \
     --message="Initial scaffold" \
     --content="$CONTENT" \
     --json
   ```

   Note: Always write content to a temp file and pipe through `base64` — do NOT use `echo -n "..." | base64` for multi-line content.

3. **Create a Vercel project** and link to the GitHub repo:
```bash
integrations vercel projects create --name=todo-app --framework=nextjs --json
```

4. **Set environment variables on Vercel**:
```bash
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_URL --value="$NEXT_PUBLIC_SUPABASE_URL" --target=production --json
integrations vercel env set --project=todo-app --key=NEXT_PUBLIC_SUPABASE_ANON_KEY --value="$NEXT_PUBLIC_SUPABASE_ANON_KEY" --target=production --json
```

5. **Create the Supabase table** via curl:
```bash
curl -s -X POST "${NEXT_PUBLIC_SUPABASE_URL}/rest/v1/rpc/exec_sql" \
  -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"query": "CREATE TABLE IF NOT EXISTS todos (id uuid DEFAULT gen_random_uuid() PRIMARY KEY, title text NOT NULL, completed boolean DEFAULT false, created_at timestamptz DEFAULT now())"}'
```

6. **Create a Linear issue** documenting the initial scaffold:
```bash
integrations linear issues create \
  --title="Initial scaffold: todo app" \
  --team=ENG \
  --description="Created GitHub repo, scaffolded Next.js 15 app with Tailwind CSS and Supabase, deployed to Vercel." \
  --priority=2 \
  --json
```

### If the repo DOES exist (subsequent runs):

1. **Check what's already been built** — list the repo contents and recent commits:
```bash
integrations github repos contents list --repo=todo-app --owner=mxy680 --path=/ --json
integrations github repos commits list --repo=todo-app --owner=mxy680 --limit=10 --json
```

2. **Check Linear for completed work**:
```bash
integrations linear issues list --team=ENG --status=Done --json
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
   - Email reminders for due todos
   - Import/export (CSV)
   - ...or anything else you think would improve the app

4. **Create a Linear issue** for the feature you're about to build:
```bash
integrations linear issues create \
  --title="Add [feature]" \
  --team=ENG \
  --description="Description of what will be built" \
  --priority=2 \
  --json
```

5. **Implement the feature** — update/create files on GitHub.

   To get the current SHA of a file before updating it:
   ```bash
   integrations github repos contents get --repo=todo-app --owner=mxy680 --path=<filepath> --json
   ```

   To update an existing file (requires SHA from the get command above):
   ```bash
   cat > /tmp/updated-file.tsx << 'EOF'
   // updated content here
   EOF
   CONTENT=$(base64 < /tmp/updated-file.tsx)
   integrations github repos contents update \
     --repo=todo-app \
     --owner=mxy680 \
     --path=<filepath> \
     --message="feat: add [feature]" \
     --content="$CONTENT" \
     --sha=<sha-from-get> \
     --json
   ```

   To create a new file:
   ```bash
   cat > /tmp/new-file.tsx << 'EOF'
   // new content here
   EOF
   CONTENT=$(base64 < /tmp/new-file.tsx)
   integrations github repos contents create \
     --repo=todo-app \
     --owner=mxy680 \
     --path=<filepath> \
     --message="feat: add [feature]" \
     --content="$CONTENT" \
     --json
   ```

6. **Wait for Vercel deployment** — Vercel auto-deploys from GitHub pushes. Wait 30 seconds, then check:
```bash
sleep 30
integrations vercel deployments list --project=todo-app --limit=1 --json
```

7. **Verify the deployment** — curl the production URL and check for HTTP 200:
```bash
curl -s -o /dev/null -w "%{http_code}" https://todo-app.vercel.app
```

8. **Close the Linear issue** with a summary:
```bash
integrations linear issues update --id=<issue-id> --status=Done --json
```

## Step 2 — Generate report

After completing your work, summarize:
- What was the state of the app before this run
- What feature/improvement you added
- The GitHub commit(s) made
- The deployment URL
- The Linear issue created/closed
- What you'd recommend building next

## Important

- Write CLEAN, MODERN code. Use TypeScript, Tailwind CSS, and modern React patterns.
- Always write file content to a temp file and encode with `base64 < /tmp/file` — never inline multi-line content.
- Never delete existing functionality unless explicitly refactoring.
- If something fails, try to fix it. Don't give up on the first error.
- The app should look professional — good spacing, typography, colors.
- Always test that the deployed site loads before closing the issue.
