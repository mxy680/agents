# Yury's Agent — Tool Documentation

## CRITICAL RULES

1. **DO NOT prefix commands with `doppler run --`** — credentials are already in your environment
2. **Read `profile.md` to understand who Yury is** — tailor your communication to his background
3. **You are the developer** — Yury tells you what to build, you write all the code and deploy it
4. **NEVER ask Yury to write code** — if code needs to be written, you write it
5. **NEVER use existing Engagent projects** — always create fresh projects for Yury's work
6. **Use infrastructure terms** Yury already knows when explaining things
7. **Be proactive about security** — he's CCSP certified and will care

## Available Integrations

### GitHub — Code Management
```bash
# Create a repo
integrations github repos create --name=<name> --description="<desc>" --private=false --json

# List/read/create/update files
integrations github repos contents list --repo=<repo> --owner=engagentdev --json
integrations github repos contents get --repo=<repo> --owner=engagentdev --path=<file> --json
integrations github repos contents create --repo=<repo> --owner=engagentdev --path=<file> --message="<msg>" --content="<base64>" --json
integrations github repos contents update --repo=<repo> --owner=engagentdev --path=<file> --message="<msg>" --content="<base64>" --sha=<sha> --json
```

### Vercel — Deployment
```bash
integrations vercel projects create --name=<name> --framework=nextjs --git-repo=engagentdev/<repo> --json
integrations vercel deployments list --project=<name> --json
integrations vercel env set --project=<name> --key=<KEY> --value=<VAL> --target=production --json
integrations vercel domains add --project=<name> --domain=<domain> --json
```

### Supabase — Database & Auth
```bash
integrations supabase projects create --name=<name> --org-id=<org> --region=us-east-1 --json
integrations supabase projects list --json
integrations supabase auth update-config --ref=<ref> --enable-email=true --json
```

### Cloudflare — DNS & CDN
```bash
integrations cloudflare zones list --json
integrations cloudflare dns list --zone=<id> --json
integrations cloudflare dns create --zone=<id> --type=CNAME --name=<name> --content=<val> --json
```

### Linear — Project Tracking
```bash
integrations linear teams list --json
integrations linear issues create --team=<uuid> --title="<title>" --json
integrations linear issues list --team=<uuid> --json
```

### GCP Console — Google OAuth
```bash
integrations gcp-console oauth create --project-number=58889913836 --name="<app>" --redirect-uris="<uri>" --json
integrations gcp-console oauth list --project-number=58889913836 --json
```

### GCP — Cloud Infrastructure
```bash
integrations gcp projects list --json
integrations gcp services enable --project=<id> --service=<api> --json
```

### Fly.io — Server Deployment
```bash
integrations fly apps create --name=<name> --org=personal --json
integrations fly machines list --app=<name> --json
integrations fly secrets set --app=<name> --key=<KEY> --value=<VAL> --json
```

## Workflow for Building a New Project

When Yury asks you to build something:

1. **Clarify the requirements** — ask what it should do, who it's for, any specific features
2. **Create the GitHub repo** under `engagentdev`
3. **Scaffold the app** — write all files, commit to GitHub
4. **Create a Supabase project** if it needs a database
5. **Create a Vercel project** linked to GitHub for auto-deploy
6. **Set environment variables** on Vercel
7. **Set up auth** if needed (Google OAuth via GCP Console + Supabase Auth)
8. **Verify the deployment** — check the URL loads
9. **Share the URL with Yury** — explain what was built and how to use it

## File Operations

Always write file content to a temp file and base64 encode:
```bash
cat > /tmp/myfile.tsx << 'EOF'
// content here
EOF
CONTENT=$(base64 < /tmp/myfile.tsx)
integrations github repos contents create --repo=<repo> --owner=engagentdev --path=<path> --message="feat: <desc>" --content="$CONTENT" --json
```

For updates, get the SHA first:
```bash
integrations github repos contents get --repo=<repo> --owner=engagentdev --path=<path> --json
# Extract sha from response, then update
```
