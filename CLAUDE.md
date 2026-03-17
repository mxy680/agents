# Agent Marketplace - Integration CLI

## Overview
Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Supports Gmail, Google Sheets, Google Calendar, Google Drive, GitHub, and Instagram. Includes a Next.js web portal for self-service OAuth and token management.

## Quick Start
```bash
make build          # → bin/integrations
make test           # run tests with coverage
make lint           # go vet

# Portal
make portal-install # npm install
make portal-dev     # npm run dev (localhost:3000)
make portal-build   # npm run build
make portal-lint    # npm run lint
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
- Coverage target: 80%+ (gmail: 93.2%, sheets: 85.5%, calendar: 92.9%, drive: 88.9%, instagram: 85.0%)

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
- Coverage target: 80%+ (gmail: 93.2%, sheets: 85.5%, calendar: 92.9%, drive: 88.9%, instagram: 85.0%, github: 85.8%)

## Web Portal (Next.js 15 + Supabase)

### Architecture
- `portal/` — Next.js 15 App Router, TypeScript, Tailwind CSS
- Auth: Supabase Auth with Google OAuth (narrow scopes for login)
- Database: Supabase PostgreSQL with RLS policies
- Integration tokens: AES-256-GCM encrypted, stored in `integrations` table
- Two Google OAuth flows: (1) login via Supabase Auth, (2) full-scope connect via custom API route

### Portal Directory Layout
```
portal/
  supabase/migrations/00001_create_tables.sql  # integrations table + RLS
  src/
    app/
      page.tsx                     # Landing page
      login/page.tsx               # Google sign-in via Supabase Auth
      integrations/page.tsx        # Dashboard with provider cards
      auth/callback/route.ts       # Supabase Auth code exchange
      api/integrations/
        route.ts                   # GET: list user integrations
        google/connect|callback|disconnect/route.ts
        github/connect|callback|disconnect/route.ts
        instagram/save|disconnect/route.ts
    lib/
      supabase/server.ts|client.ts|middleware.ts
      crypto.ts                    # AES-256-GCM (Go-compatible wire format)
      providers.ts                 # Provider metadata and scopes
    components/
      navbar.tsx, provider-card.tsx, instagram-form.tsx, sign-out-button.tsx
    middleware.ts                   # Protects /integrations route
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

## Environment Variables
```
# Google (Gmail, Sheets, Calendar, Drive)
GOOGLE_DESKTOP_CLIENT_ID, GOOGLE_DESKTOP_CLIENT_SECRET
GOOGLE_ACCESS_TOKEN (fallback: GMAIL_ACCESS_TOKEN)
GOOGLE_REFRESH_TOKEN (fallback: GMAIL_REFRESH_TOKEN)

# Instagram (cookie-based session auth)
INSTAGRAM_SESSION_ID       # sessionid cookie (required)
INSTAGRAM_CSRF_TOKEN       # csrftoken cookie (required)
INSTAGRAM_DS_USER_ID       # ds_user_id cookie (required)
INSTAGRAM_MID              # mid cookie (optional, reduces challenges)
INSTAGRAM_IG_DID           # ig_did cookie (optional, reduces challenges)
INSTAGRAM_USER_AGENT       # User-Agent override (optional)

# GitHub
GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET
GITHUB_ACCESS_TOKEN, GITHUB_REFRESH_TOKEN
GITHUB_API_BASE_URL (optional, defaults to https://api.github.com)
```

# currentDate
Today's date is 2026-03-16.
