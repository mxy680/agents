# Emdash Agents — Internal Integration Platform

## Overview
Internal admin tool for managing AI agent integrations. Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Supports Gmail, Google Sheets, Google Calendar, Google Drive, Google Places, GitHub, Instagram, LinkedIn, Framer, Supabase, X (Twitter), iMessage (via BlueBubbles), Canvas LMS, and Zillow. Includes an admin-only Next.js portal for centralized credential management, and a Go orchestrator that deploys Claude Agent SDK containers to Kubernetes.

Mark owns all integrations centrally. Clients get specialized agents configured via the admin dashboard. Session-bound integrations (Instagram, LinkedIn, X, Canvas) use Playwright browser automation for cookie capture — no manual cookie pasting or Chrome extensions.

## Quick Start
```bash
make build          # → bin/integrations
make test           # run tests with coverage
make lint           # go vet

# Orchestrator
make orchestrator       # → bin/orchestrator
make orchestrator-dev   # run with doppler (localhost:8080)
make kind-setup         # create local k8s cluster
make sync-templates     # sync agents/*/template.yaml → Supabase

# Docker images
make docker-agent-base    # build agent base image
make docker-export-creds  # build export-creds image

# Portal
make portal-install             # npm install
make portal-dev                 # npm run dev (localhost:3000)
make portal-build               # npm run build
make portal-lint                # npm run lint
make portal-playwright-install  # install Playwright chromium for session capture
```

## Commands — Gmail
```
integrations gmail messages list [--query=QUERY] [--limit=N] [--since=Xh] [--page-token=TOKEN] [--json]
integrations gmail messages get --id=MESSAGE_ID [--json]
integrations gmail messages send --to=EMAIL --subject=SUBJECT [--body=TEXT | --body-file=PATH] [--cc=EMAIL] [--reply-to=MSG_ID] [--dry-run] [--json]
```

`messages` has alias `msg`. The old `list-unread` and `search` commands are unified into `messages list`:
- Unread: `messages list --query=is:unread --since=24h`
- Search: `messages list --query=from:boss`

## Commands — Google Sheets
```
# Spreadsheet management (list/delete use Drive API)
integrations sheets spreadsheets list [--limit=N] [--page-token=TOKEN] [--json]
integrations sheets spreadsheets get --id=SPREADSHEET_ID [--json]
integrations sheets spreadsheets create --title=TITLE [--json]
integrations sheets spreadsheets delete --id=SPREADSHEET_ID [--confirm] [--json]

# Cell values (core read/write)
integrations sheets values get --id=SPREADSHEET_ID --range=RANGE [--major-dimension=ROWS|COLUMNS] [--json]
integrations sheets values update --id=SPREADSHEET_ID --range=RANGE [--values=JSON | --values-file=PATH] [--value-input=RAW|USER_ENTERED] [--json]
integrations sheets values append --id=SPREADSHEET_ID --range=RANGE [--values=JSON | --values-file=PATH] [--value-input=RAW|USER_ENTERED] [--json]
integrations sheets values clear --id=SPREADSHEET_ID --range=RANGE [--confirm] [--json]
integrations sheets values batch-get --id=SPREADSHEET_ID --ranges=R1,R2 [--major-dimension=ROWS|COLUMNS] [--json]
integrations sheets values batch-update --id=SPREADSHEET_ID [--data=JSON | --data-file=PATH] [--value-input=RAW|USER_ENTERED] [--json]

# Tab/sheet management
integrations sheets tabs list --id=SPREADSHEET_ID [--json]
integrations sheets tabs create --id=SPREADSHEET_ID --title=TITLE [--json]
integrations sheets tabs delete --id=SPREADSHEET_ID --sheet-id=SHEET_ID [--confirm] [--json]
integrations sheets tabs rename --id=SPREADSHEET_ID --sheet-id=SHEET_ID --title=NEW_TITLE [--json]
```

`spreadsheets` has alias `ss`. `values` has alias `val`. `tabs` has alias `tab`.

## Commands — Calendar
```
integrations calendar events list [--calendar-id=ID] [--query=Q] [--time-min=RFC3339] [--time-max=RFC3339] [--limit=N] [--single-events] [--order-by=startTime|updated] [--json]
integrations calendar events get --event-id=ID [--calendar-id=ID] [--json]
integrations calendar events create --summary=TEXT --start=RFC3339 --end=RFC3339 [--description=TEXT] [--location=TEXT] [--attendees=EMAIL,...] [--timezone=TZ] [--all-day] [--dry-run] [--json]
integrations calendar events quick-add --text=TEXT [--calendar-id=ID] [--dry-run] [--json]
integrations calendar events update --event-id=ID [--summary=TEXT] [--start=RFC3339] [--end=RFC3339] [--description=TEXT] [--location=TEXT] [--attendees=EMAIL,...] [--dry-run] [--json]
integrations calendar events delete --event-id=ID [--confirm] [--dry-run] [--json]
integrations calendar events move --event-id=ID --destination=CALENDAR_ID [--json]
integrations calendar events instances --event-id=ID [--time-min=RFC3339] [--time-max=RFC3339] [--limit=N] [--json]

integrations calendar calendars list [--limit=N] [--show-hidden] [--json]
integrations calendar calendars get [--calendar-id=ID] [--json]
integrations calendar calendars create --summary=TEXT [--description=TEXT] [--timezone=TZ] [--dry-run] [--json]
integrations calendar calendars update [--calendar-id=ID] [--summary=TEXT] [--description=TEXT] [--timezone=TZ] [--dry-run] [--json]
integrations calendar calendars delete --calendar-id=ID [--confirm] [--dry-run] [--json]

integrations calendar freebusy query --time-min=RFC3339 --time-max=RFC3339 [--calendar-ids=ID,...] [--json]
```

`events` has aliases `event`, `ev`. `calendars` has alias `cal`. `freebusy` has alias `fb`.
`--calendar-id` defaults to `"primary"` on all events commands and calendars get/update/delete.

## Commands — Drive
```
integrations drive files list [--query=Q] [--limit=N] [--page-token=TOKEN] [--order-by=ORDER] [--corpora=CORPORA] [--drive-id=ID] [--include-trashed] [--json]
integrations drive files get --file-id=ID [--json]
integrations drive files download --file-id=ID [--output=PATH] [--export-mime=MIME]
integrations drive files upload --path=PATH [--name=NAME] [--parent=FOLDER_ID] [--mime-type=MIME] [--description=TEXT] [--dry-run] [--json]
integrations drive files copy --file-id=ID [--name=NAME] [--parent=FOLDER_ID] [--dry-run] [--json]
integrations drive files move --file-id=ID --parent=FOLDER_ID [--dry-run] [--json]
integrations drive files trash --file-id=ID [--dry-run] [--json]
integrations drive files untrash --file-id=ID [--dry-run] [--json]
integrations drive files delete --file-id=ID [--confirm] [--dry-run] [--json]

integrations drive permissions list --file-id=ID [--json]
integrations drive permissions get --file-id=ID --permission-id=ID [--json]
integrations drive permissions create --file-id=ID --role=ROLE --type=TYPE [--email=EMAIL] [--domain=DOMAIN] [--dry-run] [--json]
integrations drive permissions delete --file-id=ID --permission-id=ID [--confirm] [--dry-run] [--json]
```

`files` has aliases `file`, `f`. `permissions` has aliases `permission`, `perm`.
`--include-trashed` defaults to false (auto-adds `trashed = false` to query).
`--export-mime` is for Google Workspace files (Docs→PDF, Sheets→CSV, etc.).

## Commands — GitHub
```
# Repos
integrations github repos list [--owner=OWNER] [--type=all|owner|public|private|member] [--sort=created|updated|pushed|full_name] [--limit=N] [--page-token=N] [--json]
integrations github repos get --owner=OWNER --repo=REPO [--json]
integrations github repos create --name=NAME [--description=TEXT] [--private] [--dry-run] [--json]
integrations github repos fork --owner=OWNER --repo=REPO [--org=ORG] [--dry-run] [--json]
integrations github repos delete --owner=OWNER --repo=REPO [--confirm] [--dry-run] [--json]

# Issues
integrations github issues list --owner=OWNER --repo=REPO [--state=open|closed|all] [--labels=L1,L2] [--assignee=USER] [--sort=created|updated|comments] [--limit=N] [--json]
integrations github issues get --owner=OWNER --repo=REPO --number=N [--json]
integrations github issues create --owner=OWNER --repo=REPO --title=TEXT [--body=TEXT] [--labels=L1,L2] [--assignees=U1,U2] [--dry-run] [--json]
integrations github issues update --owner=OWNER --repo=REPO --number=N [--title=TEXT] [--body=TEXT] [--state=open|closed] [--labels=L1,L2] [--assignees=U1,U2] [--dry-run] [--json]
integrations github issues close --owner=OWNER --repo=REPO --number=N [--dry-run] [--json]
integrations github issues comment --owner=OWNER --repo=REPO --number=N --body=TEXT [--dry-run] [--json]

# Pull Requests
integrations github pulls list --owner=OWNER --repo=REPO [--state=open|closed|all] [--head=BRANCH] [--base=BRANCH] [--sort=created|updated|popularity|long-running] [--limit=N] [--json]
integrations github pulls get --owner=OWNER --repo=REPO --number=N [--json]
integrations github pulls create --owner=OWNER --repo=REPO --title=TEXT --head=BRANCH --base=BRANCH [--body=TEXT] [--draft] [--dry-run] [--json]
integrations github pulls update --owner=OWNER --repo=REPO --number=N [--title=TEXT] [--body=TEXT] [--state=open|closed] [--base=BRANCH] [--dry-run] [--json]
integrations github pulls merge --owner=OWNER --repo=REPO --number=N [--method=merge|squash|rebase] [--commit-title=TEXT] [--commit-message=TEXT] [--dry-run] [--json]
integrations github pulls review --owner=OWNER --repo=REPO --number=N --event=APPROVE|REQUEST_CHANGES|COMMENT [--body=TEXT] [--dry-run] [--json]

# Workflow Runs
integrations github runs list --owner=OWNER --repo=REPO [--workflow-id=ID] [--branch=BRANCH] [--status=completed|in_progress|queued] [--limit=N] [--json]
integrations github runs get --owner=OWNER --repo=REPO --run-id=ID [--json]
integrations github runs re-run --owner=OWNER --repo=REPO --run-id=ID [--dry-run] [--json]
integrations github runs workflows --owner=OWNER --repo=REPO [--json]

# Releases
integrations github releases list --owner=OWNER --repo=REPO [--limit=N] [--json]
integrations github releases get --owner=OWNER --repo=REPO [--tag=TAG | --release-id=ID | --latest] [--json]
integrations github releases create --owner=OWNER --repo=REPO --tag=TAG [--name=TEXT] [--body=TEXT] [--target=COMMITISH] [--draft] [--prerelease] [--dry-run] [--json]
integrations github releases delete --owner=OWNER --repo=REPO --release-id=ID [--confirm] [--dry-run] [--json]

# Gists
integrations github gists list [--limit=N] [--page-token=N] [--json]
integrations github gists get --gist-id=ID [--json]
integrations github gists create [--description=TEXT] [--files=JSON | --files-file=PATH] [--public] [--dry-run] [--json]
integrations github gists update --gist-id=ID [--description=TEXT] [--files=JSON | --files-file=PATH] [--dry-run] [--json]
integrations github gists delete --gist-id=ID [--confirm] [--dry-run] [--json]

# Search
integrations github search repos --query=Q [--sort=stars|forks|updated] [--order=asc|desc] [--limit=N] [--json]
integrations github search code --query=Q [--sort=indexed] [--order=asc|desc] [--limit=N] [--json]
integrations github search issues --query=Q [--sort=created|updated|comments] [--order=asc|desc] [--limit=N] [--json]
integrations github search commits --query=Q [--sort=author-date|committer-date] [--order=asc|desc] [--limit=N] [--json]
integrations github search users --query=Q [--sort=followers|repositories|joined] [--order=asc|desc] [--limit=N] [--json]

# Git (low-level)
integrations github git refs list --owner=OWNER --repo=REPO [--namespace=heads|tags] [--json]
integrations github git refs get --owner=OWNER --repo=REPO --ref=REF [--json]
integrations github git refs create --owner=OWNER --repo=REPO --ref=REF --sha=SHA [--dry-run] [--json]
integrations github git refs update --owner=OWNER --repo=REPO --ref=REF --sha=SHA [--force] [--dry-run] [--json]
integrations github git refs delete --owner=OWNER --repo=REPO --ref=REF [--confirm] [--dry-run] [--json]
integrations github git commits get --owner=OWNER --repo=REPO --sha=SHA [--json]
integrations github git commits create --owner=OWNER --repo=REPO --message=TEXT --tree=SHA --parents=SHA,... [--dry-run] [--json]
integrations github git trees get --owner=OWNER --repo=REPO --sha=SHA [--recursive] [--json]
integrations github git trees create --owner=OWNER --repo=REPO [--tree=JSON | --tree-file=PATH] [--base-tree=SHA] [--dry-run] [--json]
integrations github git blobs get --owner=OWNER --repo=REPO --sha=SHA [--json]
integrations github git blobs create --owner=OWNER --repo=REPO --content=TEXT [--encoding=utf-8|base64] [--dry-run] [--json]
integrations github git tags get --owner=OWNER --repo=REPO --sha=SHA [--json]
integrations github git tags create --owner=OWNER --repo=REPO --tag=NAME --message=TEXT --object=SHA --type=commit|tree|blob [--dry-run] [--json]

# Organizations
integrations github orgs list [--limit=N] [--json]
integrations github orgs get --org=ORG [--json]
integrations github orgs members --org=ORG [--role=all|admin|member] [--limit=N] [--json]
integrations github orgs repos --org=ORG [--type=all|public|private|forks|sources|member] [--limit=N] [--json]

# Teams
integrations github teams list --org=ORG [--limit=N] [--json]
integrations github teams get --org=ORG --team-slug=SLUG [--json]
integrations github teams members --org=ORG --team-slug=SLUG [--role=all|member|maintainer] [--limit=N] [--json]
integrations github teams repos --org=ORG --team-slug=SLUG [--limit=N] [--json]
integrations github teams add-repo --org=ORG --team-slug=SLUG --owner=OWNER --repo=REPO [--permission=pull|push|admin] [--dry-run] [--json]
integrations github teams remove-repo --org=ORG --team-slug=SLUG --owner=OWNER --repo=REPO [--confirm] [--dry-run] [--json]

# Labels
integrations github labels list --owner=OWNER --repo=REPO [--limit=N] [--json]
integrations github labels get --owner=OWNER --repo=REPO --name=NAME [--json]
integrations github labels create --owner=OWNER --repo=REPO --name=NAME [--color=HEX] [--description=TEXT] [--dry-run] [--json]
integrations github labels update --owner=OWNER --repo=REPO --name=NAME [--new-name=NAME] [--color=HEX] [--description=TEXT] [--dry-run] [--json]
integrations github labels delete --owner=OWNER --repo=REPO --name=NAME [--confirm] [--dry-run] [--json]

# Branches
integrations github branches list --owner=OWNER --repo=REPO [--protected] [--limit=N] [--json]
integrations github branches get --owner=OWNER --repo=REPO --branch=NAME [--json]
integrations github branches protection get --owner=OWNER --repo=REPO --branch=NAME [--json]
integrations github branches protection update --owner=OWNER --repo=REPO --branch=NAME [--settings=JSON | --settings-file=PATH] [--dry-run] [--json]
integrations github branches protection delete --owner=OWNER --repo=REPO --branch=NAME [--confirm] [--dry-run] [--json]
```

`repos` has alias `repo`. `issues` has alias `issue`. `pulls` has aliases `pull`, `pr`. `runs` has alias `run`. `releases` has alias `release`. `gists` has alias `gist`. `orgs` has alias `org`. `teams` has alias `team`. `labels` has alias `label`. `branches` has alias `branch`.

## Architecture
- `cmd/integrations/main.go` — entrypoint, registers providers
- `internal/cli/` — Cobra root command, output helpers (JSON/text)
- `internal/auth/` — Google OAuth + GitHub OAuth token management with auto-refresh
- `internal/providers/gmail/` — Gmail provider with injectable ServiceFactory
- `internal/providers/sheets/` — Sheets provider with dual ServiceFactory (Sheets + Drive)
- `internal/providers/calendar/` — Calendar provider with injectable ServiceFactory
- `internal/providers/drive/` — Drive provider with injectable ServiceFactory
- `internal/providers/github/` — GitHub provider with injectable ClientFactory
- `internal/providers/provider.go` — Provider interface

## Architecture — Gmail Package Layout
```
internal/providers/gmail/
  gmail.go              # Provider struct, RegisterCommands (nested: gmail → messages → list/get/send)
  helpers.go            # Shared types (EmailSummary, EmailDetail, SendResult) and shared functions
  messages.go           # messages list, messages get, messages send commands
  helpers_test.go       # Unit tests for helpers (parseSinceDuration, truncate, extractHeaders, etc.)
  messages_test.go      # Integration tests for all messages sub-commands
  mock_server_test.go   # httptest mock server helpers, captureStdout, newTestRootCmd
  gmail_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
  integration_test.go   # Real Gmail API tests (build tag: integration)
```

## Architecture — Sheets Package Layout
```
internal/providers/sheets/
  sheets.go              # Provider struct, RegisterCommands, dual ServiceFactory (Sheets + Drive)
  helpers.go             # Shared types (SpreadsheetSummary, CellData, etc.) and helper functions
  spreadsheets.go        # spreadsheets list/get/create/delete (list/delete via Drive API)
  values_read.go         # values get, batch-get commands
  values_write.go        # values update, append, clear, batch-update commands
  tabs.go                # tabs list/create/delete/rename commands
  helpers_test.go        # Unit tests for helpers (parseValuesJSON, formatCellsTable, etc.)
  spreadsheets_test.go   # Tests for spreadsheets commands
  values_test.go         # Tests for values commands
  tabs_test.go           # Tests for tabs commands
  sheets_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
  mock_server_test.go    # httptest mock server helpers for Sheets + Drive APIs
```

## Architecture — Calendar Package Layout
```
internal/providers/calendar/
  calendar.go           # Provider struct, RegisterCommands (nested: calendar → events/calendars/freebusy)
  helpers.go            # Shared types (EventSummary, EventDetail, CalendarSummary, FreeBusyResult) and helpers
  events.go             # 8 event commands (list, get, create, quick-add, update, delete, move, instances)
  calendars.go          # 5 calendar commands (list, get, create, update, delete)
  freebusy.go           # 1 freebusy command (query)
  helpers_test.go       # Unit tests for helpers (formatEventTime, toEventSummary, parseAttendees, etc.)
  events_test.go        # Integration tests for all events sub-commands
  calendars_test.go     # Integration tests for all calendars sub-commands
  freebusy_test.go      # Integration tests for freebusy query
  mock_server_test.go   # httptest mock server helpers, captureStdout, newTestRootCmd
  calendar_test.go      # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
```

## Architecture — Drive Package Layout
```
internal/providers/drive/
  drive.go              # Provider struct, RegisterCommands (nested: drive → files/permissions)
  helpers.go            # Shared types (FileSummary, FileDetail, PermissionInfo) and helpers
  files.go              # 9 file commands (list, get, download, upload, copy, move, trash, untrash, delete)
  permissions.go        # 4 permission commands (list, get, create, delete)
  helpers_test.go       # Unit tests for helpers (toFileSummary, formatSize, truncate, etc.)
  files_test.go         # Integration tests for all files sub-commands
  permissions_test.go   # Integration tests for all permissions sub-commands
  mock_server_test.go   # httptest mock server helpers, captureStdout, newTestRootCmd
  drive_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
```

## Commands — Google Places (Scraper)
```
# Search — find businesses and places by query (no API key needed)
integrations places search --query="coffee shops in Cleveland" [--geo=LAT,LNG] [--zoom=N] [--depth=N] [--email] [--concurrency=N] [--lang=CODE] [--limit=N] [--json]

# Lookup — get full details for a specific Google Maps URL
integrations places lookup --url=MAPS_URL [--email] [--json]
```

`places` has alias `place`. `search` has alias `find`.

Powered by [gosom/google-maps-scraper](https://github.com/gosom/google-maps-scraper) — scrapes Google Maps directly, no API key or billing required. Returns 34+ data fields per place including address, phone, hours, reviews, ratings, emails, images, and more.

Requires the `google-maps-scraper` binary in PATH or set `GOOGLE_MAPS_SCRAPER_BIN` env var. The agent Docker image includes the binary + Chromium.

## Architecture — Places Package Layout
```
internal/providers/places/
  places.go              # Provider struct with ScraperFunc, RegisterCommands (search, lookup)
  scraper.go             # ScraperFunc type, ScraperOptions, exec-based implementation
  helpers.go             # Entry struct (34+ fields from scraper), PlaceSummary, formatters
  search.go              # places search command
  lookup.go              # places lookup command
  helpers_test.go        # Unit tests for helpers (truncate, parseLatLng, toPlaceSummary, etc.)
  search_test.go         # Tests for search command with mock ScraperFunc
  lookup_test.go         # Tests for lookup command with mock ScraperFunc
  scraper_test.go        # Tests for scraper binary resolution, output parsing
  mock_scraper_test.go   # Mock ScraperFunc factory, captureStdout, newTestRootCmd
  places_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
```

## Commands — Instagram
```
# Profile
integrations instagram profile get [--username=USERNAME | --user-id=ID] [--json]
integrations instagram profile edit-form [--json]

# Media/Posts
integrations instagram media list [--user-id=ID] [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram media get --media-id=ID [--json]
integrations instagram media delete --media-id=ID [--confirm] [--dry-run] [--json]
integrations instagram media archive --media-id=ID [--dry-run] [--json]
integrations instagram media unarchive --media-id=ID [--dry-run] [--json]
integrations instagram media likers --media-id=ID [--limit=N] [--json]
integrations instagram media save --media-id=ID [--collection-id=ID] [--dry-run] [--json]
integrations instagram media unsave --media-id=ID [--dry-run] [--json]

# Stories
integrations instagram stories list [--user-id=ID] [--json]
integrations instagram stories get --story-id=ID [--json]
integrations instagram stories viewers --story-id=ID [--limit=N] [--json]
integrations instagram stories feed [--limit=N] [--json]
integrations instagram stories delete --story-id=ID [--confirm] [--dry-run] [--json]

# Reels
integrations instagram reels list [--user-id=ID] [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram reels get --reel-id=ID [--json]
integrations instagram reels feed [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram reels delete --reel-id=ID [--confirm] [--dry-run] [--json]

# Comments
integrations instagram comments list --media-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram comments replies --media-id=ID --comment-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram comments create --media-id=ID --text=TEXT [--reply-to=COMMENT_ID] [--dry-run] [--json]
integrations instagram comments delete --media-id=ID --comment-id=ID [--confirm] [--dry-run] [--json]
integrations instagram comments like --comment-id=ID [--dry-run] [--json]
integrations instagram comments unlike --comment-id=ID [--dry-run] [--json]
integrations instagram comments disable --media-id=ID [--dry-run] [--json]
integrations instagram comments enable --media-id=ID [--dry-run] [--json]

# Likes
integrations instagram likes like --media-id=ID [--dry-run] [--json]
integrations instagram likes unlike --media-id=ID [--dry-run] [--json]
integrations instagram likes list --media-id=ID [--limit=N] [--json]
integrations instagram likes liked [--limit=N] [--cursor=TOKEN] [--json]

# Relationships
integrations instagram relationships followers [--user-id=ID] [--limit=N] [--cursor=TOKEN] [--query=Q] [--json]
integrations instagram relationships following [--user-id=ID] [--limit=N] [--cursor=TOKEN] [--query=Q] [--json]
integrations instagram relationships follow --user-id=ID [--dry-run] [--json]
integrations instagram relationships unfollow --user-id=ID [--dry-run] [--json]
integrations instagram relationships remove-follower --user-id=ID [--dry-run] [--json]
integrations instagram relationships block --user-id=ID [--dry-run] [--json]
integrations instagram relationships unblock --user-id=ID [--dry-run] [--json]
integrations instagram relationships blocked [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram relationships mute --user-id=ID [--stories] [--posts] [--dry-run] [--json]
integrations instagram relationships unmute --user-id=ID [--stories] [--posts] [--dry-run] [--json]
integrations instagram relationships restrict --user-id=ID [--dry-run] [--json]
integrations instagram relationships unrestrict --user-id=ID [--dry-run] [--json]
integrations instagram relationships status --user-id=ID [--json]

# Search
integrations instagram search users --query=Q [--limit=N] [--json]
integrations instagram search tags --query=Q [--limit=N] [--json]
integrations instagram search locations --query=Q [--lat=LAT] [--lng=LNG] [--limit=N] [--json]
integrations instagram search top --query=Q [--limit=N] [--json]
integrations instagram search clear [--dry-run] [--json]
integrations instagram search explore [--limit=N] [--cursor=TOKEN] [--json]

# Collections
integrations instagram collections list [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram collections get --collection-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram collections create --name=NAME [--dry-run] [--json]
integrations instagram collections edit --collection-id=ID --name=NAME [--dry-run] [--json]
integrations instagram collections delete --collection-id=ID [--confirm] [--dry-run] [--json]
integrations instagram collections saved [--limit=N] [--cursor=TOKEN] [--json]

# Tags/Hashtags
integrations instagram tags get --name=TAG [--json]
integrations instagram tags feed --name=TAG [--tab=top|recent] [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram tags follow --name=TAG [--dry-run] [--json]
integrations instagram tags unfollow --name=TAG [--dry-run] [--json]
integrations instagram tags following [--json]
integrations instagram tags related --name=TAG [--json]

# Locations
integrations instagram locations get --location-id=ID [--json]
integrations instagram locations feed --location-id=ID [--tab=ranked|recent] [--limit=N] [--cursor=TOKEN] [--json]
integrations instagram locations search --query=Q [--lat=LAT] [--lng=LNG] [--limit=N] [--json]
integrations instagram locations stories --location-id=ID [--json]

# Activity
integrations instagram activity feed [--limit=N] [--json]
integrations instagram activity mark-checked [--json]

# Live
integrations instagram live list [--json]
integrations instagram live get --broadcast-id=ID [--json]
integrations instagram live comments --broadcast-id=ID [--json]
integrations instagram live heartbeat --broadcast-id=ID [--json]
integrations instagram live like --broadcast-id=ID [--dry-run] [--json]
integrations instagram live post-comment --broadcast-id=ID --text=TEXT [--dry-run] [--json]

# Highlights
integrations instagram highlights list [--user-id=ID] [--json]
integrations instagram highlights get --highlight-id=ID [--json]
integrations instagram highlights create --title=TITLE --story-ids=ID,ID [--dry-run] [--json]
integrations instagram highlights edit --highlight-id=ID [--title=TITLE] [--add-stories=ID,ID] [--remove-stories=ID,ID] [--dry-run] [--json]
integrations instagram highlights delete --highlight-id=ID [--confirm] [--dry-run] [--json]

# Close Friends
integrations instagram closefriends list [--json]
integrations instagram closefriends add --user-id=ID [--dry-run] [--json]
integrations instagram closefriends remove --user-id=ID [--dry-run] [--json]

# Settings
integrations instagram settings get [--json]
integrations instagram settings set-private [--dry-run] [--json]
integrations instagram settings set-public [--dry-run] [--json]
integrations instagram settings login-activity [--json]
```

`instagram` has alias `ig`. `media` has aliases `post`, `posts`. `stories` has aliases `story`, `st`. `reels` has alias `reel`. `comments` has alias `comment`. `likes` has alias `like`. `relationships` has aliases `rel`, `friendship`. `search` has alias `find`. `collections` has aliases `collection`, `saved`. `tags` has aliases `tag`, `hashtag`. `locations` has aliases `location`, `loc`. `activity` has aliases `notifications`, `notif`. `live` has alias `broadcast`. `highlights` has aliases `highlight`, `hl`. `closefriends` has aliases `cf`, `besties`. `settings` has aliases `setting`, `account`.

## Architecture — Instagram Package Layout
```
internal/providers/instagram/
  instagram.go          # Provider struct, RegisterCommands (17 resource subcommand groups)
  client.go             # HTTP client: web (www) + mobile (i.instagram.com), CSRF rotation, rate limit detection
  helpers.go            # Shared types (UserSummary, UserDetail, MediaSummary, etc.) and helpers
  profile.go            # profile get, edit-form (web API)
  media.go              # media list, get, delete, archive, unarchive, likers, save, unsave
  stories.go            # stories list, get, viewers, feed, delete
  reels.go              # reels list, get, feed, delete
  comments.go           # comments list, replies, create, delete, like, unlike, disable, enable
  likes.go              # likes like, unlike, list, liked
  relationships.go      # 13 relationship commands (follow, block, mute, restrict, etc.)
  search.go             # search users, tags, locations, top, clear, explore
  collections.go        # collections list, get, create, edit, delete, saved
  tags.go               # tags get, feed, follow, unfollow, following, related
  locations.go          # locations get, feed, search, stories
  activity.go           # activity feed, mark-checked
  live.go               # live list, get, comments, heartbeat, like, post-comment
  highlights.go         # highlights list, get, create, edit, delete
  closefriends.go       # closefriends list, add, remove
  settings.go           # settings get, set-private, set-public, login-activity
  *_test.go             # Tests for each command file + helpers + provider
  mock_server_test.go   # httptest mock server helpers for all endpoints
```

## Testing
- All providers use `ServiceFactory`/`ClientFactory` for dependency injection
- Tests use `httptest.NewServer` to mock APIs via `newFullMockServer(t)`
- Coverage target: 80%+ (gmail: 93.2%, sheets: 85.5%, calendar: 92.9%, drive: 88.9%, places: 94.6%, instagram: 85.0%)

## Architecture — GitHub Package Layout
```
internal/providers/github/
  github.go              # Provider struct, RegisterCommands (nested: github → repos/issues/pulls/runs/releases/gists/search/git/orgs/teams/labels/branches)
  helpers.go             # Shared types, doGitHub API wrapper, text formatters, JSON extraction helpers
  repos.go               # 5 repos commands (list, get, create, fork, delete)
  issues.go              # 6 issues commands (list, get, create, update, close, comment)
  pulls.go               # 6 pulls commands (list, get, create, update, merge, review)
  runs.go                # 4 runs commands (list, get, re-run, workflows)
  releases.go            # 4 releases commands (list, get, create, delete)
  gists.go               # 5 gists commands (list, get, create, update, delete)
  search.go              # 5 search commands (repos, code, issues, commits, users)
  git.go                 # 13 git commands (refs list/get/create/update/delete, commits get/create, trees get/create, blobs get/create, tags get/create)
  orgs.go                # 4 orgs commands (list, get, members, repos)
  teams.go               # 6 teams commands (list, get, members, repos, add-repo, remove-repo)
  labels.go              # 5 labels commands (list, get, create, update, delete)
  branches.go            # 5 branches commands (list, get, protection get/update/delete)
  helpers_test.go        # Unit tests for helpers
  repos_test.go          # Tests for repos commands
  issues_test.go         # Tests for issues commands
  pulls_test.go          # Tests for pulls commands
  runs_test.go           # Tests for runs commands
  releases_test.go       # Tests for releases commands
  gists_test.go          # Tests for gists commands
  search_test.go         # Tests for search commands
  git_test.go            # Tests for git commands
  orgs_test.go           # Tests for orgs commands
  teams_test.go          # Tests for teams commands
  labels_test.go         # Tests for labels commands
  branches_test.go       # Tests for branches commands
  mock_server_test.go    # httptest mock server helpers, captureStdout, newTestRootCmd
  github_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
```

## Testing
- All providers use `ServiceFactory` (or `ClientFactory` for GitHub) for dependency injection
- Tests use `httptest.NewServer` to mock APIs via `newFullMockServer(t)`
- Orchestrator uses `sqlmock` + `fake.NewSimpleClientset()` for DB and K8s tests
- Coverage target: 80%+ (gmail: 93.2%, sheets: 85.5%, calendar: 92.9%, drive: 88.9%, instagram: 85.0%, github: 85.8%, linkedin: 86.5%, framer: 80.5%, supabase: 82.5%, x: 84.2%, imessage: 83.9%, canvas: 80.0%, zillow: 86.9%)

## Commands — Framer
```
# Project
integrations framer project info [--json]
integrations framer project user [--json]
integrations framer project changed-paths [--json]
integrations framer project contributors [--from-version=V] [--to-version=V] [--json]

# Publishing & Deployment
integrations framer publish create [--dry-run] [--json]
integrations framer publish deploy --deployment-id=ID [--domains=D1,D2] [--dry-run] [--json]
integrations framer publish list [--json]
integrations framer publish info [--json]

# CMS Collections
integrations framer collections list [--json]
integrations framer collections get --id=ID [--json]
integrations framer collections create --name=NAME [--dry-run] [--json]
integrations framer collections fields --id=ID [--json]
integrations framer collections add-fields --id=ID --fields=JSON [--dry-run] [--json]
integrations framer collections remove-fields --id=ID --field-ids=ID,ID [--confirm] [--json]
integrations framer collections set-field-order --id=ID --field-ids=ID,ID [--json]
integrations framer collections items --id=ID [--json]
integrations framer collections add-items --id=ID [--items=JSON | --items-file=PATH] [--dry-run] [--json]
integrations framer collections remove-items --id=ID --item-ids=ID,ID [--confirm] [--json]
integrations framer collections set-item-order --id=ID --item-ids=ID,ID [--json]

# Managed Collections
integrations framer managed-collections list [--json]
integrations framer managed-collections create --name=NAME [--dry-run] [--json]
integrations framer managed-collections fields --id=ID [--json]
integrations framer managed-collections set-fields --id=ID [--fields=JSON | --fields-file=PATH] [--dry-run] [--json]
integrations framer managed-collections items --id=ID [--json]
integrations framer managed-collections add-items --id=ID [--items=JSON | --items-file=PATH] [--dry-run] [--json]
integrations framer managed-collections remove-items --id=ID --item-ids=ID,ID [--confirm] [--json]
integrations framer managed-collections set-item-order --id=ID --item-ids=ID,ID [--json]

# Canvas Nodes
integrations framer nodes get --node-id=ID [--json]
integrations framer nodes children --node-id=ID [--json]
integrations framer nodes parent --node-id=ID [--json]
integrations framer nodes list-by-type --type=FrameNode|TextNode|SVGNode|ComponentInstanceNode|WebPageNode|DesignPageNode|ComponentNode [--json]
integrations framer nodes create-frame --attributes=JSON [--parent-id=ID] [--dry-run] [--json]
integrations framer nodes create-text --attributes=JSON [--parent-id=ID] [--dry-run] [--json]
integrations framer nodes create-component --name=NAME [--dry-run] [--json]
integrations framer nodes create-web-page --path=PATH [--dry-run] [--json]
integrations framer nodes create-design-page --name=NAME [--dry-run] [--json]
integrations framer nodes clone --node-id=ID [--dry-run] [--json]
integrations framer nodes remove --node-ids=ID,ID [--confirm] [--dry-run] [--json]
integrations framer nodes set-attributes --node-id=ID --attributes=JSON [--dry-run] [--json]
integrations framer nodes set-parent --node-id=ID --parent-id=ID [--index=N] [--dry-run] [--json]
integrations framer nodes rect --node-id=ID [--json]

# AI Agent
integrations framer agent system-prompt [--json]
integrations framer agent context [--json]
integrations framer agent read --queries=JSON [--json]
integrations framer agent apply [--dsl=TEXT | --dsl-file=PATH] [--dry-run] [--json]

# Color Styles
integrations framer styles colors list [--json]
integrations framer styles colors get --id=ID [--json]
integrations framer styles colors create --attributes=JSON [--dry-run] [--json]

# Text Styles
integrations framer styles text list [--json]
integrations framer styles text get --id=ID [--json]
integrations framer styles text create --attributes=JSON [--dry-run] [--json]

# Fonts
integrations framer fonts list [--json]
integrations framer fonts get --family=NAME [--json]

# Localization
integrations framer locales list [--json]
integrations framer locales default [--json]
integrations framer locales create --language=CODE [--region=CODE] [--dry-run] [--json]
integrations framer locales languages [--json]
integrations framer locales regions --language=CODE [--json]
integrations framer locales groups [--json]
integrations framer locales set-data [--data=JSON | --data-file=PATH] [--dry-run] [--json]

# Redirects
integrations framer redirects list [--json]
integrations framer redirects add [--redirects=JSON | --redirects-file=PATH] [--dry-run] [--json]
integrations framer redirects remove --ids=ID,ID [--confirm] [--json]
integrations framer redirects set-order --ids=ID,ID [--json]

# Code Files
integrations framer code list [--json]
integrations framer code get --id=ID [--json]
integrations framer code create --name=NAME [--code=TEXT | --code-file=PATH] [--dry-run] [--json]
integrations framer code typecheck --name=NAME [--content=TEXT | --content-file=PATH] [--json]
integrations framer code custom-get [--json]
integrations framer code custom-set --html=TEXT --location=headStart|headEnd|bodyStart|bodyEnd [--dry-run] [--json]

# Images
integrations framer images upload --path=PATH [--json]
integrations framer images upload-batch --paths=P1,P2 [--json]

# Files
integrations framer files upload --path=PATH [--json]
integrations framer files upload-batch --paths=P1,P2 [--json]

# SVG
integrations framer svg add [--svg=TEXT | --svg-file=PATH] [--dry-run] [--json]
integrations framer svg vector-sets [--json]

# Screenshots
integrations framer screenshot take --node-id=ID [--format=png|jpeg] [--scale=N] [--output=PATH] [--json]
integrations framer screenshot export-svg --node-id=ID [--output=PATH] [--json]

# Plugin Data
integrations framer plugin-data get --key=KEY [--json]
integrations framer plugin-data set --key=KEY --value=TEXT [--dry-run] [--json]
integrations framer plugin-data keys [--json]
```

`framer` has alias `fr`. `collections` has alias `col`. `managed-collections` has alias `mcol`. `nodes` has alias `node`. `styles` has alias `style`. `fonts` has alias `font`. `locales` has alias `locale`. `redirects` has alias `redirect`. `screenshot` has alias `ss`. `plugin-data` has alias `pd`. `publish` has alias `pub`. `project` has alias `proj`. `images` has alias `img`. `files` has alias `file`.

## Architecture — Framer Package Layout
```
internal/providers/framer/
  framer.go              # Provider struct, RegisterCommands (15 resource subcommand groups)
  bridge_client.go       # BridgeClient: spawns Node.js subprocess, JSON-RPC over stdin/stdout
  helpers.go             # Shared types (ProjectInfo, CollectionSummary, NodeSummary, etc.) and helpers
  bridge/
    bridge.js            # Node.js sidecar: connects to Framer WebSocket API, handles JSON-RPC commands
    package.json         # framer-api ^0.1.3 dependency
  project.go             # project info, user, changed-paths, contributors
  publish.go             # publish create, deploy, list, info
  collections.go         # 11 collection commands (list, get, create, fields CRUD, items CRUD)
  managed_collections.go # 8 managed collection commands
  nodes.go               # 14 node commands (get, create, clone, remove, set-attributes, etc.)
  agent.go               # agent system-prompt, context, read, apply
  styles.go              # 6 style commands (colors list/get/create, text list/get/create)
  fonts.go               # fonts list, get
  locales.go             # 7 locale commands (list, default, create, languages, regions, groups, set-data)
  redirects.go           # redirects list, add, remove, set-order
  code.go                # 6 code commands (list, get, create, typecheck, custom-get, custom-set)
  images.go              # images upload, upload-batch
  files.go               # files upload, upload-batch
  svg.go                 # svg add, vector-sets
  screenshot.go          # screenshot take, export-svg
  plugin_data.go         # plugin-data get, set, keys
  *_test.go              # Tests for each command file + helpers + provider
  mock_bridge_test.go    # In-process mock bridge for testing without Node.js
```

## Commands — LinkedIn
```
# Profile
integrations linkedin profile get --public-id=SLUG [--json]
integrations linkedin profile me [--json]

# Connections
integrations linkedin connections list [--limit=N] [--cursor=TOKEN] [--sort=RECENTLY_ADDED|LAST_NAME|FIRST_NAME] [--json]
integrations linkedin connections get --urn=URN [--json]
integrations linkedin connections remove --urn=URN [--confirm] [--dry-run] [--json]

# Invitations
integrations linkedin invitations list [--direction=received|sent] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin invitations send --urn=URN [--message=TEXT] [--dry-run] [--json]
integrations linkedin invitations accept --invitation-id=ID [--dry-run] [--json]
integrations linkedin invitations reject --invitation-id=ID [--dry-run] [--json]
integrations linkedin invitations withdraw --invitation-id=ID [--dry-run] [--json]

# Posts
integrations linkedin posts list [--username=USERNAME] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin posts get --post-urn=URN [--json]
integrations linkedin posts create --text=TEXT [--visibility=public|connections] [--dry-run] [--json]
integrations linkedin posts delete --post-urn=URN [--confirm] [--dry-run] [--json]
integrations linkedin posts reactions --post-urn=URN [--limit=N] [--json]
integrations linkedin posts react --post-urn=URN --type=LIKE|CELEBRATE|SUPPORT|LOVE|INSIGHTFUL|FUNNY [--dry-run] [--json]

# Comments
integrations linkedin comments list --post-urn=URN [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin comments create --post-urn=URN --text=TEXT [--reply-to=COMMENT_URN] [--dry-run] [--json]
integrations linkedin comments delete --comment-urn=URN [--confirm] [--dry-run] [--json]
integrations linkedin comments like --comment-urn=URN [--dry-run] [--json]
integrations linkedin comments unlike --comment-urn=URN [--dry-run] [--json]

# Messages
integrations linkedin messages conversations [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin messages list --conversation-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin messages send --conversation-id=ID --text=TEXT [--dry-run] [--json]
integrations linkedin messages new --recipients=URN,URN --text=TEXT [--subject=TEXT] [--dry-run] [--json]
integrations linkedin messages delete --conversation-id=ID [--confirm] [--dry-run] [--json]
integrations linkedin messages mark-read --conversation-id=ID [--dry-run] [--json]

# Feed
integrations linkedin feed list [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin feed hashtag --tag=TAG [--limit=N] [--cursor=TOKEN] [--json]

# Companies
integrations linkedin companies get --company-id=ID [--json]
integrations linkedin companies search --query=Q [--limit=N] [--json]
integrations linkedin companies employees --company-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin companies follow --company-id=ID [--dry-run] [--json]
integrations linkedin companies unfollow --company-id=ID [--dry-run] [--json]
integrations linkedin companies jobs --company-id=ID [--limit=N] [--cursor=TOKEN] [--json]

# Jobs
integrations linkedin jobs search --query=Q [--location=TEXT] [--experience=ENTRY|ASSOCIATE|MID_SENIOR|DIRECTOR|EXECUTIVE] [--type=FULL_TIME|PART_TIME|CONTRACT|TEMPORARY|INTERNSHIP] [--remote=ON_SITE|REMOTE|HYBRID] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin jobs get --job-id=ID [--json]
integrations linkedin jobs save --job-id=ID [--dry-run] [--json]
integrations linkedin jobs unsave --job-id=ID [--dry-run] [--json]
integrations linkedin jobs saved [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin jobs recommended [--limit=N] [--cursor=TOKEN] [--json]

# Search
integrations linkedin search people --query=Q [--network=F|S|O] [--company=ID] [--location=TEXT] [--title=TEXT] [--industry=TEXT] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin search companies --query=Q [--industry=TEXT] [--size=RANGE] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin search jobs --query=Q [--location=TEXT] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin search posts --query=Q [--author=URN] [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin search groups --query=Q [--limit=N] [--cursor=TOKEN] [--json]

# Groups
integrations linkedin groups list [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin groups get --group-id=ID [--json]
integrations linkedin groups members --group-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin groups posts --group-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin groups join --group-id=ID [--dry-run] [--json]
integrations linkedin groups leave --group-id=ID [--confirm] [--dry-run] [--json]

# Notifications
integrations linkedin notifications list [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin notifications mark-read [--dry-run] [--json]

# Network
integrations linkedin network followers [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin network following [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin network follow --urn=URN [--dry-run] [--json]
integrations linkedin network unfollow --urn=URN [--dry-run] [--json]
integrations linkedin network suggestions [--limit=N] [--json]

# Skills
integrations linkedin skills list [--username=USERNAME] [--json]
integrations linkedin skills endorse --urn=URN --skill-id=ID [--dry-run] [--json]
integrations linkedin skills endorsements --skill-id=ID [--limit=N] [--json]

# Analytics
integrations linkedin analytics profile-views [--json]
integrations linkedin analytics search-appearances [--json]
integrations linkedin analytics post-impressions --post-urn=URN [--json]

# Events
integrations linkedin events list [--limit=N] [--cursor=TOKEN] [--json]
integrations linkedin events get --event-id=ID [--json]
integrations linkedin events attend --event-id=ID [--dry-run] [--json]
integrations linkedin events unattend --event-id=ID [--dry-run] [--json]

# Settings
integrations linkedin settings get [--json]
integrations linkedin settings privacy [--json]
integrations linkedin settings visibility --field=FIELD --value=VALUE [--dry-run] [--json]
```

`linkedin` has alias `li`. `connections` has alias `conn`. `invitations` has alias `invite`. `posts` has alias `post`. `comments` has alias `comment`. `messages` has alias `msg`. `companies` has aliases `company`, `org`. `jobs` has alias `job`. `search` has alias `find`. `groups` has alias `group`. `notifications` has alias `notif`. `events` has alias `event`. `skills` has alias `skill`. `settings` has alias `setting`. `profile` has alias `prof`.

## Architecture — LinkedIn Package Layout
```
internal/providers/linkedin/
  linkedin.go           # Provider struct, RegisterCommands (17 resource subcommand groups)
  client.go             # HTTP client: Voyager API (www.linkedin.com), CSRF rotation, rate limit detection
  helpers.go            # Shared types (ProfileSummary, PostSummary, JobSummary, etc.) and helpers
  profile.go            # profile get, me (Voyager API)
  connections.go        # connections list, get, remove
  invitations.go        # invitations list, send, accept, reject, withdraw
  posts.go              # posts list, get, create, delete, reactions, react
  comments.go           # comments list, create, delete, like, unlike
  messages.go           # messages conversations, list, send, new, delete, mark-read
  feed.go               # feed list, hashtag
  companies.go          # companies get, search, employees, follow, unfollow, jobs
  jobs.go               # jobs search, get, save, unsave, saved, recommended
  search.go             # search people, companies, jobs, posts, groups
  groups.go             # groups list, get, members, posts, join, leave
  notifications.go      # notifications list, mark-read
  network.go            # network followers, following, follow, unfollow, suggestions
  skills.go             # skills list, endorse, endorsements
  analytics.go          # analytics profile-views, search-appearances, post-impressions
  events.go             # events list, get, attend, unattend
  settings.go           # settings get, privacy, visibility
  *_test.go             # Tests for each command file + helpers + provider
  mock_server_test.go   # httptest mock server helpers for all endpoints
```

## Commands — Supabase
```
# Projects [alias: proj]
integrations supabase projects list [--json]
integrations supabase projects get --ref=REF [--json]
integrations supabase projects create --name=NAME --region=REGION --org-id=ID [--db-pass=PASS] [--plan=free|pro] [--dry-run] [--json]
integrations supabase projects update --ref=REF [--name=NAME] [--dry-run] [--json]
integrations supabase projects delete --ref=REF [--confirm] [--dry-run] [--json]
integrations supabase projects pause --ref=REF [--dry-run] [--json]
integrations supabase projects restore --ref=REF [--dry-run] [--json]
integrations supabase projects health --ref=REF [--json]
integrations supabase projects regions [--json]

# Organizations [alias: org]
integrations supabase orgs list [--json]
integrations supabase orgs create --name=NAME [--dry-run] [--json]

# Branches (Preview Environments) [alias: branch]
integrations supabase branches list --ref=REF [--json]
integrations supabase branches get --branch-id=ID [--json]
integrations supabase branches create --ref=REF --git-branch=NAME [--region=REGION] [--dry-run] [--json]
integrations supabase branches update --branch-id=ID [--git-branch=NAME] [--reset-on-push] [--dry-run] [--json]
integrations supabase branches delete --branch-id=ID [--confirm] [--dry-run] [--json]
integrations supabase branches push --branch-id=ID [--dry-run] [--json]
integrations supabase branches merge --branch-id=ID [--dry-run] [--json]
integrations supabase branches reset --branch-id=ID [--dry-run] [--json]
integrations supabase branches diff --branch-id=ID [--json]
integrations supabase branches disable --ref=REF [--confirm] [--dry-run] [--json]

# API Keys [alias: key]
integrations supabase keys list --ref=REF [--json]
integrations supabase keys get --ref=REF --key-id=ID [--json]
integrations supabase keys create --ref=REF --name=NAME [--type=anon|service_role] [--dry-run] [--json]
integrations supabase keys update --ref=REF --key-id=ID [--name=NAME] [--dry-run] [--json]
integrations supabase keys delete --ref=REF --key-id=ID [--confirm] [--dry-run] [--json]

# Secrets (Edge Function Secrets) [alias: secret]
integrations supabase secrets list --ref=REF [--json]
integrations supabase secrets create --ref=REF --name=NAME --value=VALUE [--dry-run] [--json]
integrations supabase secrets delete --ref=REF --name=NAME [--confirm] [--dry-run] [--json]

# Auth Config
integrations supabase auth get --ref=REF [--json]
integrations supabase auth update --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]
integrations supabase auth signing-keys list --ref=REF [--json]
integrations supabase auth signing-keys get --ref=REF --key-id=ID [--json]
integrations supabase auth signing-keys create --ref=REF [--dry-run] [--json]
integrations supabase auth signing-keys update --ref=REF --key-id=ID [--dry-run] [--json]
integrations supabase auth signing-keys delete --ref=REF --key-id=ID [--confirm] [--dry-run] [--json]
integrations supabase auth third-party list --ref=REF [--json]
integrations supabase auth third-party get --ref=REF --tpa-id=ID [--json]
integrations supabase auth third-party create --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]
integrations supabase auth third-party delete --ref=REF --tpa-id=ID [--confirm] [--dry-run] [--json]

# Database [alias: db]
integrations supabase db migrations --ref=REF [--json]
integrations supabase db types --ref=REF [--lang=typescript] [--json]
integrations supabase db ssl-enforcement get --ref=REF [--json]
integrations supabase db ssl-enforcement update --ref=REF --enabled [--dry-run] [--json]
integrations supabase db jit-access get --ref=REF [--json]
integrations supabase db jit-access update --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]

# Network [alias: net]
integrations supabase network restrictions get --ref=REF [--json]
integrations supabase network restrictions update --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]
integrations supabase network restrictions apply --ref=REF [--dry-run] [--json]
integrations supabase network bans list --ref=REF [--json]
integrations supabase network bans remove --ref=REF [--ips=IP,...] [--confirm] [--dry-run] [--json]

# Domains [alias: domain]
integrations supabase domains custom get --ref=REF [--json]
integrations supabase domains custom delete --ref=REF [--confirm] [--dry-run] [--json]
integrations supabase domains custom initialize --ref=REF --hostname=HOST [--dry-run] [--json]
integrations supabase domains custom verify --ref=REF [--dry-run] [--json]
integrations supabase domains custom activate --ref=REF [--dry-run] [--json]
integrations supabase domains vanity get --ref=REF [--json]
integrations supabase domains vanity delete --ref=REF [--confirm] [--dry-run] [--json]
integrations supabase domains vanity check --ref=REF --subdomain=NAME [--json]
integrations supabase domains vanity activate --ref=REF --subdomain=NAME [--dry-run] [--json]

# PostgREST [alias: rest]
integrations supabase rest get --ref=REF [--json]
integrations supabase rest update --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]

# Analytics [alias: logs]
integrations supabase analytics logs --ref=REF [--json]
integrations supabase analytics api-counts --ref=REF [--json]
integrations supabase analytics api-requests --ref=REF [--json]
integrations supabase analytics functions --ref=REF [--json]

# Advisors [alias: advisor]
integrations supabase advisors performance --ref=REF [--json]
integrations supabase advisors security --ref=REF [--json]

# Billing [alias: bill]
integrations supabase billing addons list --ref=REF [--json]
integrations supabase billing addons apply --ref=REF --addon=VARIANT [--dry-run] [--json]
integrations supabase billing addons remove --ref=REF --addon=VARIANT [--confirm] [--dry-run] [--json]

# Snippets [alias: snippet]
integrations supabase snippets list [--json]
integrations supabase snippets get --snippet-id=ID [--json]

# Actions (CI/CD) [alias: action]
integrations supabase actions list --ref=REF [--json]
integrations supabase actions get --ref=REF --run-id=ID [--json]
integrations supabase actions logs --ref=REF --run-id=ID [--json]
integrations supabase actions update-status --ref=REF --run-id=ID --status=STATUS [--dry-run] [--json]

# Encryption (pgsodium) [alias: encrypt]
integrations supabase encryption get --ref=REF [--json]
integrations supabase encryption update --ref=REF [--config=JSON | --config-file=PATH] [--dry-run] [--json]
```

`supabase` has alias `sb`. `projects` has alias `proj`. `branches` has alias `branch`. `keys` has alias `key`. `secrets` has alias `secret`. `signing-keys` has alias `sk`. `third-party` has alias `tpa`. `domains` has alias `domain`. `network` has alias `net`. `analytics` has alias `logs`. `advisors` has alias `advisor`. `billing` has alias `bill`. `snippets` has alias `snippet`. `actions` has alias `action`. `encryption` has alias `encrypt`.

## Architecture — Supabase Package Layout
```
internal/providers/supabase/
  supabase.go           # Provider struct, RegisterCommands (16 resource subcommand groups)
  helpers.go            # Shared types (ProjectSummary, BranchSummary, etc.), doSupabase HTTP helper
  projects.go           # 9 project commands (list, get, create, update, delete, pause, restore, health, regions)
  orgs.go               # 2 org commands (list, create)
  branches.go           # 10 branch commands (list, get, create, update, delete, push, merge, reset, diff, disable)
  keys.go               # 5 API key commands (list, get, create, update, delete)
  secrets.go            # 3 secret commands (list, create, delete)
  auth.go               # 11 auth commands (get, update, signing-keys CRUD, third-party CRUD)
  database.go           # 6 database commands (migrations, types, ssl-enforcement, jit-access)
  network.go            # 5 network commands (restrictions get/update/apply, bans list/remove)
  domains.go            # 9 domain commands (custom get/delete/init/verify/activate, vanity get/delete/check/activate)
  rest.go               # 2 PostgREST config commands (get, update)
  analytics.go          # 4 analytics commands (logs, api-counts, api-requests, functions)
  advisors.go           # 2 advisor commands (performance, security)
  billing.go            # 3 billing commands (addons list/apply/remove)
  snippets.go           # 2 snippet commands (list, get)
  actions.go            # 4 CI/CD action commands (list, get, logs, update-status)
  encryption.go         # 2 encryption commands (get, update)
  *_test.go             # Tests for each command file + helpers + provider
  mock_server_test.go   # httptest mock server helpers for all endpoints
```

## Commands — X (Twitter)
```
# Posts [alias: post, tweet]
integrations x posts get --tweet-id=ID [--json]
integrations x posts lookup --ids=ID,ID [--json]
integrations x posts similar --tweet-id=ID [--limit=N] [--json]
integrations x posts create --text=TEXT [--reply-to=TWEET_ID] [--quote-url=URL] [--media-ids=ID,ID] [--sensitive] [--dry-run] [--json]
integrations x posts delete --tweet-id=ID [--confirm] [--dry-run] [--json]
integrations x posts search --query=Q [--type=top|latest|users|photos|videos] [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts timeline [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts latest-timeline [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts user-tweets --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts user-replies --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts retweeters --tweet-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x posts favoriters --tweet-id=ID [--limit=N] [--cursor=TOKEN] [--json]

# Scheduled Tweets [alias: sched]
integrations x scheduled list [--json]
integrations x scheduled create --text=TEXT --date=RFC3339 [--media-ids=ID,ID] [--dry-run] [--json]
integrations x scheduled delete --tweet-id=ID [--confirm] [--dry-run] [--json]

# Likes [alias: like]
integrations x likes list --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x likes like --tweet-id=ID [--dry-run] [--json]
integrations x likes unlike --tweet-id=ID [--dry-run] [--json]

# Retweets [alias: retweet, rt]
integrations x retweets retweet --tweet-id=ID [--dry-run] [--json]
integrations x retweets undo --tweet-id=ID [--dry-run] [--json]

# Bookmarks [alias: bookmark, bm]
integrations x bookmarks list [--limit=N] [--cursor=TOKEN] [--json]
integrations x bookmarks add --tweet-id=ID [--folder-id=ID] [--dry-run] [--json]
integrations x bookmarks remove --tweet-id=ID [--dry-run] [--json]
integrations x bookmarks clear [--confirm] [--dry-run] [--json]
integrations x bookmarks folders [--json]
integrations x bookmarks folder-tweets --folder-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x bookmarks create-folder --name=NAME [--dry-run] [--json]
integrations x bookmarks edit-folder --folder-id=ID --name=NAME [--dry-run] [--json]
integrations x bookmarks delete-folder --folder-id=ID [--confirm] [--dry-run] [--json]

# Users [alias: user]
integrations x users get --username=USER [--json]
integrations x users get-by-id --user-id=ID [--json]
integrations x users search --query=Q [--limit=N] [--cursor=TOKEN] [--json]
integrations x users highlights --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x users media --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x users subscriptions --user-id=ID [--json]

# Follows [alias: follow]
integrations x follows followers --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x follows following --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x follows verified-followers --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x follows followers-you-know --user-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x follows follow --user-id=ID [--dry-run] [--json]
integrations x follows unfollow --user-id=ID [--dry-run] [--json]

# Blocks [alias: block]
integrations x blocks block --user-id=ID [--dry-run] [--json]
integrations x blocks unblock --user-id=ID [--dry-run] [--json]

# Mutes [alias: mute]
integrations x mutes mute --user-id=ID [--dry-run] [--json]
integrations x mutes unmute --user-id=ID [--dry-run] [--json]

# Direct Messages [alias: dm]
integrations x dm inbox [--json]
integrations x dm conversation --conversation-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x dm send --user-id=ID --text=TEXT [--media-id=ID] [--dry-run] [--json]
integrations x dm send-group --conversation-id=ID --text=TEXT [--media-id=ID] [--dry-run] [--json]
integrations x dm delete --message-id=ID [--confirm] [--dry-run] [--json]
integrations x dm react --message-id=ID --emoji=EMOJI [--dry-run] [--json]
integrations x dm unreact --message-id=ID --emoji=EMOJI [--dry-run] [--json]
integrations x dm add-members --conversation-id=ID --user-ids=ID,ID [--dry-run] [--json]
integrations x dm rename-group --conversation-id=ID --name=NAME [--dry-run] [--json]

# Lists [alias: list]
integrations x lists get --list-id=ID [--json]
integrations x lists owned [--limit=N] [--cursor=TOKEN] [--json]
integrations x lists search --query=Q [--limit=N] [--cursor=TOKEN] [--json]
integrations x lists tweets --list-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x lists members --list-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x lists subscribers --list-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x lists create --name=NAME [--description=TEXT] [--private] [--dry-run] [--json]
integrations x lists update --list-id=ID [--name=NAME] [--description=TEXT] [--private] [--dry-run] [--json]
integrations x lists delete --list-id=ID [--confirm] [--dry-run] [--json]
integrations x lists add-member --list-id=ID --user-id=ID [--dry-run] [--json]
integrations x lists remove-member --list-id=ID --user-id=ID [--dry-run] [--json]
integrations x lists set-banner --list-id=ID --path=PATH [--dry-run] [--json]
integrations x lists remove-banner --list-id=ID [--dry-run] [--json]

# Communities [alias: community]
integrations x communities get --community-id=ID [--json]
integrations x communities search --query=Q [--limit=N] [--json]
integrations x communities tweets --community-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x communities media --community-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x communities members --community-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x communities moderators --community-id=ID [--limit=N] [--cursor=TOKEN] [--json]
integrations x communities timeline [--limit=N] [--cursor=TOKEN] [--json]
integrations x communities join --community-id=ID [--dry-run] [--json]
integrations x communities leave --community-id=ID [--dry-run] [--json]
integrations x communities request-join --community-id=ID [--dry-run] [--json]
integrations x communities search-tweets --community-id=ID --query=Q [--limit=N] [--json]

# Notifications [alias: notif]
integrations x notifications all [--limit=N] [--cursor=TOKEN] [--json]
integrations x notifications mentions [--limit=N] [--cursor=TOKEN] [--json]
integrations x notifications verified [--limit=N] [--cursor=TOKEN] [--json]

# Media [alias: upload]
integrations x media upload --path=PATH [--alt-text=TEXT] [--json]
integrations x media status --media-id=ID [--json]
integrations x media set-alt-text --media-id=ID --alt-text=TEXT [--dry-run] [--json]

# Trends [alias: trend]
integrations x trends list [--json]
integrations x trends locations [--json]
integrations x trends by-place --woeid=ID [--json]

# Polls [alias: poll]
integrations x polls create --options=O1,O2,O3 --duration=MINUTES [--json]
integrations x polls vote --tweet-id=ID --choice=N [--dry-run] [--json]

# Geo [alias: location]
integrations x geo reverse --lat=LAT --lng=LNG [--json]
integrations x geo search --query=Q [--lat=LAT] [--lng=LNG] [--json]
integrations x geo get --place-id=ID [--json]
```

`x` has alias `twitter`. `posts` has aliases `post`, `tweet`. `scheduled` has alias `sched`. `likes` has alias `like`. `retweets` has aliases `retweet`, `rt`. `bookmarks` has aliases `bookmark`, `bm`. `users` has alias `user`. `follows` has alias `follow`. `blocks` has alias `block`. `mutes` has alias `mute`. `lists` has alias `list`. `communities` has alias `community`. `notifications` has alias `notif`. `media` has alias `upload`. `trends` has alias `trend`. `polls` has alias `poll`. `geo` has alias `location`.

Powered by X's internal GraphQL and v1.1 APIs — no API key or billing required. Uses cookie-based session auth (auth_token + ct0).

## Architecture — X Package Layout
```
internal/providers/x/
  x.go                    # Provider struct, RegisterCommands (19 resource subcommand groups)
  client.go               # HTTP client: GraphQL + v1.1 + upload + caps helpers, CSRF rotation, static bearer token
  helpers.go              # Shared types (TweetSummary, UserSummary, ListSummary, etc.) and helpers
  posts.go                # 12 post commands (GraphQL)
  scheduled.go            # 3 scheduled tweet commands (GraphQL)
  likes.go                # 3 like commands (GraphQL)
  retweets.go             # 2 retweet commands (GraphQL)
  bookmarks.go            # 9 bookmark commands (GraphQL)
  users.go                # 6 user commands (GraphQL)
  follows.go              # 6 follow commands (GraphQL + v1.1)
  blocks.go               # 2 block commands (v1.1)
  mutes.go                # 2 mute commands (v1.1)
  dm.go                   # 9 DM commands (v1.1 + GraphQL)
  lists.go                # 13 list commands (GraphQL)
  communities.go          # 11 community commands (GraphQL)
  notifications.go        # 3 notification commands (v2 internal)
  media.go                # 3 media commands (upload.x.com + v1.1)
  trends.go               # 3 trend commands (v1.1 + v2)
  polls.go                # 2 poll commands (caps.x.com)
  geo.go                  # 3 geo commands (v1.1)
  helpers_test.go          # Unit tests for helpers
  client_test.go           # Unit tests for client
  *_test.go                # Tests for each command file + provider
  mock_server_test.go      # httptest mock server helpers for all endpoints
```

## Commands — iMessage (via BlueBubbles)
```
# Chats [alias: chat]
integrations imessage chats list [--limit=N] [--offset=N] [--sort=lastmessage] [--with-participants] [--with-last-message] [--query=Q] [--json]
integrations imessage chats get --guid=GUID [--json]
integrations imessage chats create --participants=PHONE,PHONE [--message=TEXT] [--dry-run] [--json]
integrations imessage chats update --guid=GUID [--name=TEXT] [--dry-run] [--json]
integrations imessage chats delete --guid=GUID [--confirm] [--dry-run] [--json]
integrations imessage chats messages --guid=GUID [--limit=N] [--offset=N] [--after=DATETIME] [--before=DATETIME] [--json]
integrations imessage chats read --guid=GUID [--dry-run] [--json]
integrations imessage chats unread --guid=GUID [--dry-run] [--json]
integrations imessage chats leave --guid=GUID [--dry-run] [--json]
integrations imessage chats typing --guid=GUID [--stop] [--dry-run] [--json]
integrations imessage chats count [--json]
integrations imessage chats icon get --guid=GUID [--output=PATH] [--json]
integrations imessage chats icon set --guid=GUID --path=PATH [--dry-run] [--json]
integrations imessage chats icon remove --guid=GUID [--confirm] [--dry-run] [--json]

# Participants [alias: participant]
integrations imessage participants add --guid=GUID --address=PHONE_OR_EMAIL [--dry-run] [--json]
integrations imessage participants remove --guid=GUID --address=PHONE_OR_EMAIL [--confirm] [--dry-run] [--json]

# Messages [alias: msg]
integrations imessage messages send --to=PHONE_OR_EMAIL --text=TEXT [--subject=TEXT] [--effect=TEXT] [--dry-run] [--json]
integrations imessage messages send-group --guid=GUID --text=TEXT [--subject=TEXT] [--effect=TEXT] [--dry-run] [--json]
integrations imessage messages send-attachment --to=PHONE_OR_EMAIL --path=PATH [--text=TEXT] [--dry-run] [--json]
integrations imessage messages send-multipart --to=PHONE_OR_EMAIL [--parts=JSON | --parts-file=PATH] [--dry-run] [--json]
integrations imessage messages get --guid=GUID [--json]
integrations imessage messages query [--chat-guid=GUID] [--limit=N] [--offset=N] [--after=DATETIME] [--before=DATETIME] [--sort=ASC|DESC] [--with=chat,attachment,handle] [--json]
integrations imessage messages edit --guid=GUID --text=TEXT [--part=N] [--dry-run] [--json]
integrations imessage messages unsend --guid=GUID [--part=N] [--dry-run] [--json]
integrations imessage messages react --chat-guid=GUID --message-guid=GUID --type=love|like|dislike|laugh|emphasis|question [--dry-run] [--json]
integrations imessage messages delete --chat-guid=GUID --message-guid=GUID [--confirm] [--dry-run] [--json]
integrations imessage messages count [--after=DATETIME] [--before=DATETIME] [--chat-guid=GUID] [--json]
integrations imessage messages count-updated [--after=DATETIME] [--before=DATETIME] [--chat-guid=GUID] [--json]
integrations imessage messages count-sent [--json]
integrations imessage messages embedded-media --guid=GUID [--json]
integrations imessage messages notify --guid=GUID [--dry-run] [--json]

# Scheduled Messages [alias: sched]
integrations imessage scheduled list [--json]
integrations imessage scheduled get --id=ID [--json]
integrations imessage scheduled create --chat-guid=GUID --text=TEXT --send-at=RFC3339 [--dry-run] [--json]
integrations imessage scheduled update --id=ID [--text=TEXT] [--send-at=RFC3339] [--dry-run] [--json]
integrations imessage scheduled delete --id=ID [--confirm] [--dry-run] [--json]

# Attachments [alias: attach]
integrations imessage attachments get --guid=GUID [--json]
integrations imessage attachments download --guid=GUID --output=PATH
integrations imessage attachments download-force --guid=GUID --output=PATH
integrations imessage attachments upload --path=PATH [--dry-run] [--json]
integrations imessage attachments live --guid=GUID --output=PATH
integrations imessage attachments blurhash --guid=GUID [--json]
integrations imessage attachments count [--json]

# Handles [alias: handle]
integrations imessage handles list [--limit=N] [--offset=N] [--query=Q] [--json]
integrations imessage handles get --guid=GUID [--json]
integrations imessage handles count [--json]
integrations imessage handles focus --guid=GUID [--json]
integrations imessage handles availability --address=PHONE_OR_EMAIL [--service=imessage|facetime] [--json]

# Contacts [alias: contact]
integrations imessage contacts list [--json]
integrations imessage contacts get [--query=Q] [--json]
integrations imessage contacts create [--data=JSON | --data-file=PATH] [--dry-run] [--json]

# FaceTime [alias: ft]
integrations imessage facetime call --addresses=PHONE,PHONE [--dry-run] [--json]
integrations imessage facetime answer --call-uuid=UUID [--dry-run] [--json]
integrations imessage facetime leave --call-uuid=UUID [--dry-run] [--json]

# FindMy [alias: fm]
integrations imessage findmy devices [--json]
integrations imessage findmy devices-refresh [--json]
integrations imessage findmy friends [--json]
integrations imessage findmy friends-refresh [--json]

# iCloud
integrations imessage icloud account [--json]
integrations imessage icloud change-alias --alias=EMAIL [--dry-run] [--json]
integrations imessage icloud contact-card [--json]

# Server
integrations imessage server info [--json]
integrations imessage server logs [--json]
integrations imessage server restart [--soft] [--dry-run] [--json]
integrations imessage server update-check [--json]
integrations imessage server update-install [--dry-run] [--json]
integrations imessage server alerts [--json]
integrations imessage server alerts-read [--dry-run] [--json]
integrations imessage server stats [--type=totals|media|media-by-chat] [--json]

# Webhooks [alias: webhook, wh]
integrations imessage webhooks list [--json]
integrations imessage webhooks create --url=URL [--events=E1,E2] [--dry-run] [--json]
integrations imessage webhooks delete --id=ID [--confirm] [--dry-run] [--json]

# Mac
integrations imessage mac lock [--dry-run] [--json]
integrations imessage mac restart-messages [--dry-run] [--json]
```

`imessage` has alias `imsg`. `chats` has alias `chat`. `messages` has alias `msg`. `attachments` has alias `attach`. `handles` has alias `handle`. `contacts` has alias `contact`. `scheduled` has alias `sched`. `participants` has alias `participant`. `webhooks` has aliases `webhook`, `wh`. `facetime` has alias `ft`. `findmy` has alias `fm`.

Powered by [BlueBubbles](https://bluebubbles.app) — self-hosted iMessage REST API running on a Mac. Requires a Mac with Messages.app signed in and BlueBubbles Server installed.

## Architecture — iMessage Package Layout
```
internal/providers/imessage/
  imessage.go           # Provider struct, RegisterCommands (13 resource subcommand groups)
  client.go             # HTTP client: BlueBubbles REST API, password auth via query param
  helpers.go            # Shared types (ChatSummary, MessageSummary, AttachmentSummary, etc.) and helpers
  chats.go              # 14 chat commands (list, get, create, update, delete, messages, read, unread, leave, typing, count, icon get/set/remove)
  participants.go       # 2 participant commands (add, remove)
  messages.go           # 15 message commands (send, send-group, send-attachment, send-multipart, get, query, edit, unsend, react, delete, count, count-updated, count-sent, embedded-media, notify)
  scheduled.go          # 5 scheduled message commands (list, get, create, update, delete)
  attachments.go        # 7 attachment commands (get, download, download-force, upload, live, blurhash, count)
  handles.go            # 5 handle commands (list, get, count, focus, availability)
  contacts.go           # 3 contact commands (list, get, create)
  facetime.go           # 3 FaceTime commands (call, answer, leave)
  findmy.go             # 4 FindMy commands (devices, devices-refresh, friends, friends-refresh)
  icloud.go             # 3 iCloud commands (account, change-alias, contact-card)
  server.go             # 8 server commands (info, logs, restart, update-check, update-install, alerts, alerts-read, stats)
  webhooks.go           # 3 webhook commands (list, create, delete)
  mac.go                # 2 macOS commands (lock, restart-messages)
  *_test.go             # Tests for each command file + helpers + provider + client
  mock_server_test.go   # httptest mock server helpers for all endpoints
```

## Commands — Canvas LMS
```
# Courses [alias: course]
integrations canvas courses list [--enrollment-type=teacher|student|ta|observer|designer] [--limit=N] [--json]
integrations canvas courses get --course-id=ID [--json]

# Assignments [alias: assignment, assign]
integrations canvas assignments list --course-id=ID [--search=Q] [--limit=N] [--json]
integrations canvas assignments get --course-id=ID --assignment-id=ID [--json]
integrations canvas assignments delete --course-id=ID --assignment-id=ID [--confirm] [--dry-run] [--json]

# Assignment Groups [alias: assign-group, ag]
integrations canvas assignment-groups list --course-id=ID [--json]
integrations canvas assignment-groups get --course-id=ID --group-id=ID [--json]
integrations canvas assignment-groups create --course-id=ID --name=NAME [--position=N] [--weight=N] [--dry-run] [--json]
integrations canvas assignment-groups update --course-id=ID --group-id=ID [--name=NAME] [--position=N] [--weight=N] [--dry-run] [--json]
integrations canvas assignment-groups delete --course-id=ID --group-id=ID [--confirm] [--dry-run] [--json]

# Submissions [alias: submission, sub]
integrations canvas submissions list --course-id=ID --assignment-id=ID [--limit=N] [--json]
integrations canvas submissions get --course-id=ID --assignment-id=ID [--user-id=ID] [--json]

# Quizzes [alias: quiz]
integrations canvas quizzes list --course-id=ID [--search=Q] [--limit=N] [--json]
integrations canvas quizzes get --course-id=ID --quiz-id=ID [--json]
integrations canvas quizzes create --course-id=ID --title=TEXT [--quiz-type=practice_quiz|assignment|graded_survey|survey] [--time-limit=N] [--points=N] [--published] [--dry-run] [--json]
integrations canvas quizzes update --course-id=ID --quiz-id=ID [--title=TEXT] [--time-limit=N] [--published] [--dry-run] [--json]
integrations canvas quizzes delete --course-id=ID --quiz-id=ID [--confirm] [--dry-run] [--json]
integrations canvas quizzes questions --course-id=ID --quiz-id=ID [--limit=N] [--json]
integrations canvas quizzes submissions --course-id=ID --quiz-id=ID [--limit=N] [--json]

# Discussions [alias: discuss, disc]
integrations canvas discussions list --course-id=ID [--scope=locked|unlocked|pinned|unpinned] [--search=Q] [--limit=N] [--json]
integrations canvas discussions get --course-id=ID --topic-id=ID [--json]
integrations canvas discussions create --course-id=ID --title=TEXT [--message=TEXT] [--type=side_comment|threaded] [--published] [--pinned] [--dry-run] [--json]
integrations canvas discussions update --course-id=ID --topic-id=ID [--title=TEXT] [--message=TEXT] [--dry-run] [--json]
integrations canvas discussions delete --course-id=ID --topic-id=ID [--confirm] [--dry-run] [--json]
integrations canvas discussions entries --course-id=ID --topic-id=ID [--limit=N] [--json]
integrations canvas discussions reply --course-id=ID --topic-id=ID --message=TEXT [--entry-id=ID] [--dry-run] [--json]
integrations canvas discussions mark-read --course-id=ID --topic-id=ID [--dry-run] [--json]

# Announcements [alias: announce, ann]
integrations canvas announcements list --course-ids=ID,ID [--start-date=RFC3339] [--end-date=RFC3339] [--active-only] [--limit=N] [--json]
integrations canvas announcements get --course-id=ID --announcement-id=ID [--json]
integrations canvas announcements create --course-id=ID --title=TEXT --message=TEXT [--published] [--dry-run] [--json]
integrations canvas announcements update --course-id=ID --announcement-id=ID [--title=TEXT] [--message=TEXT] [--dry-run] [--json]
integrations canvas announcements delete --course-id=ID --announcement-id=ID [--confirm] [--dry-run] [--json]

# Modules [alias: mod]
integrations canvas modules list --course-id=ID [--search=Q] [--limit=N] [--json]
integrations canvas modules get --course-id=ID --module-id=ID [--json]
integrations canvas modules create --course-id=ID --name=NAME [--position=N] [--unlock-at=RFC3339] [--require-sequential-progress] [--dry-run] [--json]
integrations canvas modules update --course-id=ID --module-id=ID [--name=NAME] [--published] [--dry-run] [--json]
integrations canvas modules delete --course-id=ID --module-id=ID [--confirm] [--dry-run] [--json]
integrations canvas modules items --course-id=ID --module-id=ID [--limit=N] [--json]
integrations canvas modules add-item --course-id=ID --module-id=ID --type=File|Page|Discussion|Assignment|Quiz|SubHeader|ExternalUrl|ExternalTool --content-id=ID [--title=TEXT] [--dry-run] [--json]
integrations canvas modules remove-item --course-id=ID --module-id=ID --item-id=ID [--confirm] [--dry-run] [--json]

# Pages [alias: page]
integrations canvas pages list --course-id=ID [--sort=title|created_at|updated_at] [--search=Q] [--published] [--limit=N] [--json]
integrations canvas pages get --course-id=ID --url=URL_SLUG [--json]
integrations canvas pages create --course-id=ID --title=TEXT [--body=TEXT] [--published] [--front-page] [--editing-roles=teachers|students|members|public] [--dry-run] [--json]
integrations canvas pages update --course-id=ID --url=URL_SLUG [--title=TEXT] [--body=TEXT] [--published] [--dry-run] [--json]
integrations canvas pages delete --course-id=ID --url=URL_SLUG [--confirm] [--dry-run] [--json]
integrations canvas pages revisions --course-id=ID --url=URL_SLUG [--limit=N] [--json]

# Files [alias: file, f]
integrations canvas files list --course-id=ID [--search=Q] [--sort=name|size|created_at|updated_at] [--limit=N] [--json]
integrations canvas files get --file-id=ID [--json]
integrations canvas files download --file-id=ID --output=PATH
integrations canvas files update --file-id=ID [--name=NAME] [--parent-folder-id=ID] [--locked] [--hidden] [--dry-run] [--json]
integrations canvas files delete --file-id=ID [--confirm] [--dry-run] [--json]
integrations canvas files folders --course-id=ID [--json]
integrations canvas files folder-contents --folder-id=ID [--limit=N] [--json]
integrations canvas files create-folder --course-id=ID --name=NAME [--parent-folder=PATH] [--locked] [--hidden] [--dry-run] [--json]

# Enrollments [alias: enroll]
integrations canvas enrollments list --course-id=ID [--type=StudentEnrollment|TeacherEnrollment|TaEnrollment|ObserverEnrollment|DesignerEnrollment] [--state=active|invited|completed|inactive|rejected] [--limit=N] [--json]
integrations canvas enrollments get --enrollment-id=ID [--json]
integrations canvas enrollments create --course-id=ID --user-id=ID --type=StudentEnrollment|TeacherEnrollment [--enrollment-state=active|invited] [--dry-run] [--json]
integrations canvas enrollments deactivate --course-id=ID --enrollment-id=ID [--dry-run] [--json]
integrations canvas enrollments reactivate --course-id=ID --enrollment-id=ID [--dry-run] [--json]
integrations canvas enrollments conclude --course-id=ID --enrollment-id=ID [--dry-run] [--json]
integrations canvas enrollments delete --course-id=ID --enrollment-id=ID [--confirm] [--dry-run] [--json]

# Sections [alias: section, sec]
integrations canvas sections list --course-id=ID [--limit=N] [--json]
integrations canvas sections get --section-id=ID [--json]
integrations canvas sections create --course-id=ID --name=NAME [--start-at=RFC3339] [--end-at=RFC3339] [--dry-run] [--json]
integrations canvas sections update --section-id=ID [--name=NAME] [--start-at=RFC3339] [--end-at=RFC3339] [--dry-run] [--json]
integrations canvas sections delete --section-id=ID [--confirm] [--dry-run] [--json]
integrations canvas sections crosslist --section-id=ID --new-course-id=ID [--dry-run] [--json]
integrations canvas sections uncrosslist --section-id=ID [--dry-run] [--json]

# Calendar [alias: cal]
integrations canvas calendar list [--type=event|assignment] [--start-date=RFC3339] [--end-date=RFC3339] [--context-codes=course_ID,...] [--limit=N] [--json]
integrations canvas calendar get --event-id=ID [--json]
integrations canvas calendar create --context-code=course_ID --title=TEXT --start-at=RFC3339 --end-at=RFC3339 [--description=TEXT] [--location-name=TEXT] [--all-day] [--dry-run] [--json]
integrations canvas calendar update --event-id=ID [--title=TEXT] [--start-at=RFC3339] [--end-at=RFC3339] [--dry-run] [--json]
integrations canvas calendar delete --event-id=ID [--confirm] [--dry-run] [--json]

# Conversations [alias: conv, msg]
integrations canvas conversations list [--scope=unread|starred|archived|sent] [--filter=course_ID] [--limit=N] [--json]
integrations canvas conversations get --conversation-id=ID [--json]
integrations canvas conversations create --recipients=ID,ID --subject=TEXT --body=TEXT [--group-conversation] [--context-code=course_ID] [--dry-run] [--json]
integrations canvas conversations reply --conversation-id=ID --body=TEXT [--dry-run] [--json]
integrations canvas conversations update --conversation-id=ID [--starred] [--workflow-state=read|unread|archived] [--dry-run] [--json]
integrations canvas conversations delete --conversation-id=ID [--confirm] [--dry-run] [--json]
integrations canvas conversations mark-all-read [--dry-run] [--json]
integrations canvas conversations unread-count [--json]

# Users [alias: user]
integrations canvas users me [--json]
integrations canvas users todo [--json]
integrations canvas users upcoming [--json]
integrations canvas users missing [--json]

# Groups [alias: group, grp]
integrations canvas groups list [--context-type=Account|Course] [--context-id=ID] [--limit=N] [--json]
integrations canvas groups get --group-id=ID [--json]
integrations canvas groups create --name=NAME --group-category-id=ID [--description=TEXT] [--join-level=parent_context_auto_join|parent_context_request|invitation_only] [--dry-run] [--json]
integrations canvas groups update --group-id=ID [--name=NAME] [--description=TEXT] [--dry-run] [--json]
integrations canvas groups delete --group-id=ID [--confirm] [--dry-run] [--json]
integrations canvas groups members --group-id=ID [--limit=N] [--json]
integrations canvas groups categories --course-id=ID [--json]

# Rubrics [alias: rubric]
integrations canvas rubrics list --course-id=ID [--limit=N] [--json]
integrations canvas rubrics get --course-id=ID --rubric-id=ID [--json]
integrations canvas rubrics create --course-id=ID --title=TEXT [--criteria=JSON] [--points=N] [--dry-run] [--json]
integrations canvas rubrics update --course-id=ID --rubric-id=ID [--title=TEXT] [--criteria=JSON] [--dry-run] [--json]
integrations canvas rubrics delete --course-id=ID --rubric-id=ID [--confirm] [--dry-run] [--json]

# Grades [alias: grade]
integrations canvas grades list --course-id=ID [--json]
integrations canvas grades history --course-id=ID [--assignment-id=ID] [--student-id=ID] [--json]

# Outcomes [alias: outcome]
integrations canvas outcomes list [--context-type=Account|Course] [--context-id=ID] [--json]
integrations canvas outcomes get --outcome-id=ID [--json]
integrations canvas outcomes create --title=TEXT [--description=TEXT] [--mastery-points=N] [--context-type=Course] [--context-id=ID] [--dry-run] [--json]
integrations canvas outcomes update --outcome-id=ID [--title=TEXT] [--description=TEXT] [--dry-run] [--json]
integrations canvas outcomes delete --outcome-id=ID [--confirm] [--dry-run] [--json]
integrations canvas outcomes groups --context-type=Account|Course --context-id=ID [--json]
integrations canvas outcomes results --course-id=ID [--user-ids=ID,...] [--outcome-ids=ID,...] [--json]

# Planner [alias: plan]
integrations canvas planner list [--start-date=RFC3339] [--end-date=RFC3339] [--context-codes=course_ID,...] [--limit=N] [--json]
integrations canvas planner notes [--start-date=RFC3339] [--end-date=RFC3339] [--json]
integrations canvas planner create-note --title=TEXT [--details=TEXT] [--course-id=ID] [--todo-date=RFC3339] [--dry-run] [--json]
integrations canvas planner update-note --note-id=ID [--title=TEXT] [--details=TEXT] [--todo-date=RFC3339] [--dry-run] [--json]
integrations canvas planner delete-note --note-id=ID [--confirm] [--dry-run] [--json]
integrations canvas planner overrides [--json]
integrations canvas planner override --plannable-type=TYPE --plannable-id=ID --marked-complete [--dry-run] [--json]

# Bookmarks [alias: bookmark, bm]
integrations canvas bookmarks list [--json]
integrations canvas bookmarks get --bookmark-id=ID [--json]
integrations canvas bookmarks create --name=NAME --url=URL [--position=N] [--dry-run] [--json]
integrations canvas bookmarks update --bookmark-id=ID [--name=NAME] [--url=URL] [--position=N] [--dry-run] [--json]
integrations canvas bookmarks delete --bookmark-id=ID [--confirm] [--dry-run] [--json]

# External Tools [alias: lti, tool]
integrations canvas external-tools list --course-id=ID [--search=Q] [--limit=N] [--json]
integrations canvas external-tools get --course-id=ID --tool-id=ID [--json]
integrations canvas external-tools create --course-id=ID --name=NAME --url=URL --consumer-key=KEY --shared-secret=SECRET [--privacy-level=anonymous|name_only|public] [--dry-run] [--json]
integrations canvas external-tools update --course-id=ID --tool-id=ID [--name=NAME] [--url=URL] [--dry-run] [--json]
integrations canvas external-tools delete --course-id=ID --tool-id=ID [--confirm] [--dry-run] [--json]
integrations canvas external-tools sessionless-launch --course-id=ID --tool-id=ID [--json]

# Content Migrations [alias: migration, migrate]
integrations canvas content-migrations list --course-id=ID [--limit=N] [--json]
integrations canvas content-migrations get --course-id=ID --migration-id=ID [--json]
integrations canvas content-migrations create --course-id=ID --type=course_copy_importer|canvas_cartridge_importer|zip_file_importer [--source-course-id=ID] [--dry-run] [--json]
integrations canvas content-migrations progress --course-id=ID --migration-id=ID [--json]
integrations canvas content-migrations content-list --course-id=ID --migration-id=ID [--json]

# Content Exports [alias: export]
integrations canvas content-exports list --course-id=ID [--limit=N] [--json]
integrations canvas content-exports get --course-id=ID --export-id=ID [--json]
integrations canvas content-exports create --course-id=ID --type=common_cartridge|qti|zip [--dry-run] [--json]

# Analytics [alias: stats]
integrations canvas analytics course --course-id=ID [--json]
integrations canvas analytics assignments --course-id=ID [--json]
integrations canvas analytics student --course-id=ID --student-id=ID [--json]
integrations canvas analytics student-assignments --course-id=ID --student-id=ID [--json]

# Notifications [alias: notif]
integrations canvas notifications list [--json]
integrations canvas notifications preferences [--json]
integrations canvas notifications update-preference --category=TEXT --frequency=immediately|daily|weekly|never [--dry-run] [--json]

# Peer Reviews [alias: review]
integrations canvas peer-reviews list --course-id=ID --assignment-id=ID [--json]
integrations canvas peer-reviews create --course-id=ID --assignment-id=ID --user-id=ID --reviewer-id=ID [--dry-run] [--json]
integrations canvas peer-reviews delete --course-id=ID --assignment-id=ID --user-id=ID --reviewer-id=ID [--confirm] [--dry-run] [--json]

# Search [alias: find]
integrations canvas search recipients --search=Q [--context=course_ID] [--type=user|context] [--limit=N] [--json]
integrations canvas search courses --search=Q [--limit=N] [--json]
integrations canvas search all --search=Q [--context=course_ID] [--limit=N] [--json]

# Favorites [alias: fav]
integrations canvas favorites courses [--json]
integrations canvas favorites groups [--json]
integrations canvas favorites add-course --course-id=ID [--dry-run] [--json]
integrations canvas favorites remove-course --course-id=ID [--dry-run] [--json]
integrations canvas favorites add-group --group-id=ID [--dry-run] [--json]
integrations canvas favorites remove-group --group-id=ID [--dry-run] [--json]
```

`canvas` has alias `cvs`. `courses` has alias `course`. `assignments` has aliases `assignment`, `assign`. `assignment-groups` has aliases `assign-group`, `ag`. `submissions` has aliases `submission`, `sub`. `quizzes` has alias `quiz`. `discussions` has aliases `discuss`, `disc`. `announcements` has aliases `announce`, `ann`. `modules` has alias `mod`. `pages` has alias `page`. `files` has aliases `file`, `f`. `enrollments` has alias `enroll`. `sections` has aliases `section`, `sec`. `calendar` has alias `cal`. `conversations` has aliases `conv`, `msg`. `users` has alias `user`. `groups` has aliases `group`, `grp`. `rubrics` has alias `rubric`. `grades` has alias `grade`. `outcomes` has alias `outcome`. `planner` has alias `plan`. `bookmarks` has aliases `bookmark`, `bm`. `external-tools` has aliases `lti`, `tool`. `content-migrations` has aliases `migration`, `migrate`. `content-exports` has alias `export`. `analytics` has alias `stats`. `notifications` has alias `notif`. `peer-reviews` has alias `review`. `search` has alias `find`. `favorites` has alias `fav`.

Powered by Canvas LMS's `/api/v1/` REST API — uses session cookie auth (`_normandy_session` + `_csrf_token`), no API key or access token required.

## Architecture — Canvas Package Layout
```
internal/providers/canvas/
  canvas.go               # Provider struct, RegisterCommands (29 resource subcommand groups)
  client.go               # HTTP client: session cookie auth, CSRF rotation, rate limit detection, Link header pagination
  helpers.go              # Shared types (CourseSummary, AssignmentSummary, SubmissionSummary, etc.) and helpers
  courses.go              # 2 course commands (list, get)
  assignments.go          # 3 assignment commands (list, get, delete)
  assignment_groups.go    # 5 assignment group commands (list, get, create, update, delete)
  submissions.go          # 2 submission commands (list, get)
  quizzes.go              # 7 quiz commands (list, get, create, update, delete, questions, submissions)
  discussions.go          # 8 discussion commands (list, get, create, update, delete, entries, reply, mark-read)
  announcements.go        # 5 announcement commands (list, get, create, update, delete)
  modules.go              # 8 module commands (list, get, create, update, delete, items, add-item, remove-item)
  pages.go                # 6 page commands (list, get, create, update, delete, revisions)
  files.go                # 8 file commands (list, get, download, update, delete, folders, folder-contents, create-folder)
  enrollments.go          # 7 enrollment commands (list, get, create, deactivate, reactivate, conclude, delete)
  sections.go             # 7 section commands (list, get, create, update, delete, crosslist, uncrosslist)
  calendar.go             # 5 calendar commands (list, get, create, update, delete)
  conversations.go        # 8 conversation commands (list, get, create, reply, update, delete, mark-all-read, unread-count)
  users.go                # 4 user commands (me, todo, upcoming, missing)
  groups.go               # 7 group commands (list, get, create, update, delete, members, categories)
  rubrics.go              # 5 rubric commands (list, get, create, update, delete)
  grades.go               # 2 grade commands (list, history)
  outcomes.go             # 7 outcome commands (list, get, create, update, delete, groups, results)
  planner.go              # 7 planner commands (list, notes, create-note, update-note, delete-note, overrides, override)
  bookmarks.go            # 5 bookmark commands (list, get, create, update, delete)
  external_tools.go       # 6 external tool commands (list, get, create, update, delete, sessionless-launch)
  content_migrations.go   # 5 content migration commands (list, get, create, progress, content-list)
  content_exports.go      # 3 content export commands (list, get, create)
  analytics.go            # 4 analytics commands (course, assignments, student, student-assignments)
  notifications.go        # 3 notification commands (list, preferences, update-preference)
  peer_reviews.go         # 3 peer review commands (list, create, delete)
  search.go               # 3 search commands (recipients, courses, all)
  favorites.go            # 6 favorite commands (courses, groups, add-course, remove-course, add-group, remove-group)
  *_test.go               # Tests for each command file + helpers + provider + client
  mock_server_test.go     # httptest mock server helpers for all endpoints
```

## Commands — Zillow
```
# Properties [alias: property, prop]
integrations zillow properties search --location=TEXT [--status=for_sale|for_rent|sold] [--min-price=N] [--max-price=N] [--min-beds=N] [--max-beds=N] [--min-baths=N] [--max-baths=N] [--min-sqft=N] [--max-sqft=N] [--home-type=house|condo|townhouse|multi_family|land|manufactured|apartment] [--sort=newest|price_low|price_high|beds|baths|sqft|lot_size] [--days-on-zillow=N] [--limit=N] [--page=N] [--json]
integrations zillow properties search-map --ne-lat=LAT --ne-lng=LNG --sw-lat=LAT --sw-lng=LNG [--status=for_sale|for_rent|sold] [--zoom=N] [--limit=N] [--page=N] [--json]
integrations zillow properties get --zpid=ID [--json]
integrations zillow properties get-by-url --url=ZILLOW_URL [--json]
integrations zillow properties photos --zpid=ID [--json]
integrations zillow properties price-history --zpid=ID [--json]
integrations zillow properties tax-history --zpid=ID [--json]
integrations zillow properties similar --zpid=ID [--limit=N] [--json]
integrations zillow properties nearby --zpid=ID [--limit=N] [--json]

# Zestimates [alias: zestimate, zest]
integrations zillow zestimates get --zpid=ID [--json]
integrations zillow zestimates rent --zpid=ID [--json]
integrations zillow zestimates chart --zpid=ID [--duration=1y|5y|10y] [--json]

# Agents [alias: agent]
integrations zillow agents search --location=TEXT [--name=TEXT] [--specialty=buying|selling] [--rating=N] [--limit=N] [--json]
integrations zillow agents get --agent-id=ID [--json]
integrations zillow agents reviews --agent-id=ID [--limit=N] [--json]
integrations zillow agents listings --agent-id=ID [--status=for_sale|for_rent|sold] [--limit=N] [--json]

# Mortgage [alias: mort]
integrations zillow mortgage rates [--state=ST] [--program=Fixed30Year|Fixed15Year|Fixed20Year|ARM5|ARM7] [--loan-type=Conventional|FHA|VA|USDA|Jumbo] [--credit-score=Low|High|VeryHigh] [--json]
integrations zillow mortgage rates-history [--state=ST] [--program=Fixed30Year|Fixed15Year] [--duration-days=N] [--aggregation=Daily|Weekly|Monthly] [--json]
integrations zillow mortgage calculate --price=N [--down-payment=N] [--rate=N] [--term=N] [--json]
integrations zillow mortgage lender-reviews --nmls-id=ID [--company=TEXT] [--limit=N] [--json]

# Search [alias: find]
integrations zillow search autocomplete --query=TEXT [--json]
integrations zillow search by-address --address=TEXT [--json]

# Walk Score [alias: ws]
integrations zillow walkscore get --zpid=ID [--json]

# Schools [alias: school]
integrations zillow schools nearby --zpid=ID [--limit=N] [--json]

# Neighborhoods [alias: neighborhood, hood]
integrations zillow neighborhoods get --region-id=ID [--json]
integrations zillow neighborhoods search --location=TEXT [--json]
integrations zillow neighborhoods market-stats --region-id=ID [--json]

# Builders [alias: builder]
integrations zillow builders search --location=TEXT [--limit=N] [--json]
integrations zillow builders get --builder-id=ID [--json]
integrations zillow builders communities --builder-id=ID [--json]
integrations zillow builders reviews --builder-id=ID [--limit=N] [--json]

# Rentals [alias: rental, rent]
integrations zillow rentals search --location=TEXT [--min-price=N] [--max-price=N] [--min-beds=N] [--max-beds=N] [--home-type=apartment|house|condo|townhouse] [--limit=N] [--page=N] [--json]
integrations zillow rentals get --zpid=ID [--json]
integrations zillow rentals estimate --zpid=ID [--json]
```

`zillow` has alias `zw`. `properties` has aliases `property`, `prop`. `zestimates` has aliases `zestimate`, `zest`. `agents` has alias `agent`. `mortgage` has alias `mort`. `search` has alias `find`. `walkscore` has alias `ws`. `schools` has alias `school`. `neighborhoods` has aliases `neighborhood`, `hood`. `builders` has alias `builder`. `rentals` has aliases `rental`, `rent`.

Powered by Zillow's internal APIs — no API key or billing required. Uses Zillow's search API, GraphQL property details, autocomplete, and the public mortgage API. Optional proxy support via `ZILLOW_PROXY_URL` for production use.

## Architecture — Zillow Package Layout
```
internal/providers/zillow/
  zillow.go              # Provider struct, RegisterCommands (10 resource subcommand groups)
  client.go              # HTTP client: search + GraphQL + mortgage APIs, proxy support, rate limit/block detection
  helpers.go             # Shared types (PropertySummary, PropertyDetail, AgentSummary, etc.) and helpers
  properties.go          # 9 property commands (search, search-map, get, get-by-url, photos, price-history, tax-history, similar, nearby)
  zestimates.go          # 3 zestimate commands (get, rent, chart)
  agents.go              # 4 agent commands (search, get, reviews, listings)
  mortgage.go            # 4 mortgage commands (rates, rates-history, calculate, lender-reviews)
  search.go              # 2 search commands (autocomplete, by-address)
  walkscore.go           # 1 walkscore command (get)
  schools.go             # 1 schools command (nearby)
  neighborhoods.go       # 3 neighborhood commands (get, search, market-stats)
  builders.go            # 4 builder commands (search, get, communities, reviews)
  rentals.go             # 3 rental commands (search, get, estimate)
  helpers_test.go        # Unit tests for helpers
  client_test.go         # Unit tests for client
  properties_test.go     # Tests for properties commands
  zestimates_test.go     # Tests for zestimates commands
  agents_test.go         # Tests for agents commands
  mortgage_test.go       # Tests for mortgage commands
  search_test.go         # Tests for search commands
  walkscore_test.go      # Tests for walkscore commands
  schools_test.go        # Tests for schools commands
  neighborhoods_test.go  # Tests for neighborhoods commands
  builders_test.go       # Tests for builders commands
  rentals_test.go        # Tests for rentals commands
  mock_server_test.go    # httptest mock server helpers for all endpoints
  zillow_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
```

## Web Portal (Next.js 15 + Supabase) — Admin Only

### Architecture
- `portal/` — Next.js 15 App Router, TypeScript, Tailwind CSS
- Auth: Supabase Auth with Google OAuth (narrow scopes for login), admin-only access enforced in proxy.ts
- Database: Supabase PostgreSQL with RLS policies
- Integration tokens: AES-256-GCM encrypted, stored in `user_integrations` table (single admin owner)
- Session-bound auth: Playwright browser automation captures cookies for Instagram, LinkedIn, X, Canvas
- OAuth providers: Google, GitHub, Supabase use standard OAuth connect flows
- Manual config: Framer (API key), iMessage/BlueBubbles (server URL + password)

### Portal Directory Layout
```
portal/
  supabase/migrations/
    00001_create_tables.sql        # integrations table + RLS
    00008_single_tenant_pivot.sql  # clients table, drop user_agents/marketplace columns
  app/
    page.tsx                       # Redirect to /login
    login/page.tsx                 # Admin sign-in (Emdash Admin branding)
    integrations/page.tsx          # Admin integration dashboard with provider cards
    admin/
      page.tsx                     # Admin overview dashboard (stats)
      clients/page.tsx             # Client management CRUD
    chat/[agentName]/              # Agent chat interface
    jobs/                          # Job management pages
    auth/callback|sign-out/        # Supabase Auth flows
    api/
      integrations/
        playwright/connect|status/ # Playwright session capture API
        google/connect|callback/   # Google OAuth
        github/connect|callback/   # GitHub OAuth
        supabase/connect|callback/ # Supabase OAuth
        bluebubbles/save|disconnect/
        framer/save|disconnect/
        canvas/disconnect/
      admin/clients/               # Client CRUD API
      chat/[agentName]/            # Agent chat API (admin credentials)
      jobs/cron|trigger/           # Job scheduling (admin credentials)
  lib/
    playwright/                    # Playwright session automation
      session-capture.ts           # Core: launch browser, poll cookies, capture
      providers/                   # Per-provider cookie configs + mappers
        instagram.ts, linkedin.ts, x.ts, canvas.ts
    supabase/server.ts|client.ts|admin.ts
    crypto.ts                      # AES-256-GCM (Go-compatible wire format)
    credentials.ts                 # resolveAdminCredentials() — single-tenant
    job-runner.ts                  # Job execution with admin credentials
  components/
    app-sidebar.tsx                # Admin nav: Dashboard, Agents, Jobs, Clients
    playwright-connect-dialog.tsx  # Browser launch + status polling UI
    connect-dialog.tsx             # OAuth connect flow
    framer-connect-dialog.tsx      # Framer API key form
    bluebubbles-connect-dialog.tsx # BlueBubbles config form
  proxy.ts                         # Admin-only middleware (ADMIN_EMAILS check)
```

### Token Bridge (Go)
```
internal/tokenbridge/
  crypto.go       # AES-256-GCM decrypt (Go side)
  bridge.go       # ExportEnvForUser() → reads integrations table, decrypts to env map
  bridge_test.go  # 94% coverage via sqlmock
```

### Cross-Language Encryption Wire Format
`base64(nonce [12 bytes] || ciphertext || auth_tag [16 bytes])`
Shared key: `ENCRYPTION_MASTER_KEY` (64 hex chars = 32 bytes)

## Architecture — Orchestrator

```
Portal (Next.js)  ──HTTP──►  Orchestrator (Go :8080)  ──K8s API──►  Agent Pods
                                    │                                    │
                                    ├── Supabase (templates, instances)  │
                                    └── tokenbridge (decrypt creds)──────┘
                                                                   (init container)
```

### Orchestrator REST API
All endpoints under `/api/v1/`, auth via Supabase JWT.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/templates` | List active templates |
| GET | `/templates/{id}` | Template details |
| POST | `/agents/deploy` | Deploy agent (`{template_id, config_overrides?}`) |
| GET | `/agents` | List user's instances |
| GET | `/agents/{id}` | Instance status |
| GET | `/agents/{id}/logs` | Stream logs (SSE) |
| POST | `/agents/{id}/stop` | Stop running agent |
| DELETE | `/agents/{id}` | Delete stopped instance |

### Orchestrator Package Layout
```
cmd/orchestrator/main.go           # HTTP server entrypoint
cmd/sync-templates/main.go         # Template sync CLI

internal/orchestrator/
  config.go                        # Config struct
  models.go                        # AgentTemplate, AgentInstance structs
  store.go + store_test.go         # DB CRUD (sqlmock tests)
  k8s.go + k8s_test.go            # K8s client wrapper (fake clientset tests)
  pod_spec.go + pod_spec_test.go   # Pod spec builder (pure function)
  credentials.go + credentials_test.go  # Credential resolution (reuses tokenbridge)
  server.go                        # chi router + JWT middleware
  handlers.go                      # REST handlers
  reconciler.go                    # Background pod status sync
```

### Agent Templates (git)
```
agents/email-assistant/
  template.yaml      # name, description, required_integrations, docker_image
  role.md            # Agent persona
  CLAUDE.md          # Claude instructions
  entrypoint.py      # SDK entry point
  requirements.txt   # Python deps
```

### Docker Images
- `docker/agent-base/Dockerfile` — Python 3.12 + Anthropic Agent SDK
- `docker/export-creds/Dockerfile` — Debian slim + export-creds binary

## Environment Variables
```
# Google (Gmail, Sheets, Calendar, Drive)
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET
GOOGLE_ACCESS_TOKEN (fallback: GMAIL_ACCESS_TOKEN)
GOOGLE_REFRESH_TOKEN (fallback: GMAIL_REFRESH_TOKEN)

# Google Places (scraper-based, no API key needed)
GOOGLE_MAPS_SCRAPER_BIN   # path to google-maps-scraper binary (optional, falls back to PATH)

# Instagram (cookie-based session auth)
INSTAGRAM_SESSION_ID       # sessionid cookie (required)
INSTAGRAM_CSRF_TOKEN       # csrftoken cookie (required)
INSTAGRAM_DS_USER_ID       # ds_user_id cookie (required)
INSTAGRAM_MID              # mid cookie (optional, reduces challenges)
INSTAGRAM_IG_DID           # ig_did cookie (optional, reduces challenges)
INSTAGRAM_USER_AGENT       # User-Agent override (optional)

# LinkedIn (cookie-based session auth via Voyager API)
LINKEDIN_LI_AT            # li_at cookie (required)
LINKEDIN_JSESSIONID       # JSESSIONID cookie (required, also used as CSRF token)
LINKEDIN_USER_AGENT       # User-Agent override (optional)

# Framer (API key auth, project-scoped)
FRAMER_API_KEY            # Server API key (required)
FRAMER_PROJECT_URL        # Project URL like https://framer.com/projects/... (required)

# GitHub
GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET
GITHUB_ACCESS_TOKEN, GITHUB_REFRESH_TOKEN
GITHUB_API_BASE_URL (optional, defaults to https://api.github.com)

# iMessage (BlueBubbles self-hosted server)
BLUEBUBBLES_URL           # BlueBubbles server URL, e.g. https://my-mac.ngrok.io (required)
BLUEBUBBLES_PASSWORD      # Server password (required)

# Supabase (OAuth 2.1 with PKCE)
SUPABASE_INTEGRATION_CLIENT_ID, SUPABASE_INTEGRATION_CLIENT_SECRET
SUPABASE_ACCESS_TOKEN, SUPABASE_REFRESH_TOKEN
SUPABASE_API_BASE_URL (optional, defaults to https://api.supabase.com)

# X (Twitter) (cookie-based session auth, no API key needed)
X_AUTH_TOKEN              # auth_token cookie (required)
X_CSRF_TOKEN              # ct0 cookie (required)
X_USER_AGENT              # User-Agent override (optional)

# Canvas LMS (cookie-based session auth, no API key needed)
CANVAS_BASE_URL           # Canvas instance URL, e.g. https://canvas.university.edu (required)
CANVAS_SESSION_COOKIE     # _normandy_session cookie (required)
CANVAS_CSRF_TOKEN         # _csrf_token cookie (required)
CANVAS_LOG_SESSION_ID     # log_session_id cookie (optional)
CANVAS_USER_AGENT         # User-Agent override (optional)

# Zillow (session cookie auth via Playwright, no API key needed)
ZILLOW_COOKIES            # All cookies from Playwright session capture (required for search/property APIs)
ZILLOW_PROXY_URL          # HTTP/SOCKS5 proxy URL (optional, recommended for production)
ZILLOW_USER_AGENT         # User-Agent override (optional)

# Orchestrator
SUPABASE_DB_URL, ENCRYPTION_MASTER_KEY, SUPABASE_JWT_SECRET
PORT (default: 8080)
KUBE_NAMESPACE (default: agents)
AGENT_BASE_IMAGE, EXPORT_CREDS_IMAGE
ANTHROPIC_API_KEY_SECRET (K8s secret name, default: anthropic-api-key)

# Portal (Next.js) — additional env vars
ORCHESTRATOR_URL          # Orchestrator HTTP base URL, e.g. http://localhost:8080 or https://agents.markshteyn.com:8080 (required for chat API)
```

# currentDate
Today's date is 2026-03-16.
