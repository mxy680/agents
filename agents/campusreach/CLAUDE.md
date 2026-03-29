## CRITICAL RULES

1. DO NOT prefix commands with `doppler run --` — credentials are already in your environment
2. Always read `known-issues.md` FIRST before starting work
3. Always WRITE to `known-issues.md` when you encounter errors
4. Create Linear issues in the CampusReach team before starting work
5. The GitHub repo is `engagentdev/campusreach`
6. NEVER use or modify the Engagent Supabase project — CampusReach has its own
7. Vercel auto-deploys from GitHub — just commit and wait
8. For file operations use `integrations github repos contents create/update`

## Available Integrations

### GitHub
- `integrations github repos contents list/get/create/update --repo=campusreach --owner=engagentdev --json`
- `integrations github repos commits list --repo=campusreach --owner=engagentdev --json`

### Vercel
- `integrations vercel projects get --project=campusreach --json`
- `integrations vercel deployments list --project=campusreach --json`
- `integrations vercel env list/set --project=campusreach --json`

### Linear
- `integrations linear teams list --json` (find CampusReach team UUID)
- `integrations linear issues list/create/update --team=<uuid> --json`
- `integrations linear workflows list --team=<uuid> --json` (get Done state ID)

### Supabase
- `integrations supabase projects get --ref=gzbkcivzxcctzvrhgqos --json`
- `integrations supabase auth get-config --ref=gzbkcivzxcctzvrhgqos --json`

---

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development
pnpm dev                    # Start Next.js dev server (localhost:3000)

# Build
pnpm build                  # Generate Prisma client and build Next.js

# Linting
pnpm lint                   # Run ESLint

# Database
pnpm db:generate            # Generate Prisma client
pnpm db:push                # Push schema to database (no migration)
pnpm db:migrate             # Create and apply migration
pnpm db:studio              # Open Prisma Studio
pnpm db:pull                # Pull schema from database
```

## Architecture

CampusReach is a volunteer management platform connecting student volunteers with organizations. It uses Next.js 16 App Router with Supabase for authentication and PostgreSQL via Prisma for data.

### Two User Types
- **Volunteers** (`/vol/*`): Students who sign up for events, track hours, rate experiences
- **Organizations** (`/org/*`): Nonprofits that create events and manage volunteers

User type is determined by checking `OrganizationMember` (org) or `Volunteer` (volunteer) records linked to the Supabase auth user ID. See `lib/user-type.ts` for `getUserType()` and `getUserData()` functions.

### Key Directories
- `app/api/vol/` - Volunteer API routes (profile, opportunities, messaging, ratings)
- `app/api/org/` - Organization API routes (opportunities, team, volunteers, messaging)
- `app/vol/` - Volunteer pages (dashboard, explore, profile, settings, messaging)
- `app/org/` - Organization pages (dashboard, opportunities, profile, settings, volunteers, messaging)
- `app/auth/` - Authentication flows (signin, signup/volunteer, signup/organization, callback)
- `components/` - React components including charts and sidebars
- `components/ui/` - Radix-based UI primitives (shadcn/ui pattern)
- `lib/supabase/` - Supabase client helpers (server.ts for Server Components, client.ts for Client Components)

### Database Schema (prisma/schema.prisma)
Core models:
- `Volunteer` - Student profile linked to Supabase user via `userId`
- `Organization` - Org profile with members
- `OrganizationMember` - Links Supabase users to organizations
- `Event` - Volunteer opportunities with signups, time entries, ratings
- `EventSignup` - Volunteer registration for events
- `TimeEntry` - Logged volunteer hours
- `EventRating` - Post-event volunteer feedback
- `GroupChat`/`ChatMessage` - Event-based messaging

### Auth Flow
OAuth via Supabase (Google). The callback at `app/auth/callback/route.ts` handles:
1. Completing OAuth exchange
2. Creating Volunteer or Organization records based on signup intent
3. Redirecting to appropriate dashboard (`/vol` or `/org`)

### Path Aliases
`@/*` maps to the project root (configured in tsconfig.json).
