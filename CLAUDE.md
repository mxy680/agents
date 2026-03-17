# Agent Marketplace - Integration CLI

## Overview
Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Supports Gmail, Google Calendar, and Google Drive.

## Quick Start
```bash
make build          # → bin/integrations
make test           # run tests with coverage
make lint           # go vet
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

## Architecture
- `cmd/integrations/main.go` — entrypoint, registers providers
- `internal/cli/` — Cobra root command, output helpers (JSON/text)
- `internal/auth/` — Google OAuth token management with auto-refresh (Gmail + Calendar + Drive)
- `internal/providers/gmail/` — Gmail provider with injectable ServiceFactory
- `internal/providers/calendar/` — Calendar provider with injectable ServiceFactory
- `internal/providers/drive/` — Drive provider with injectable ServiceFactory
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

## Testing
- All providers use `ServiceFactory` for dependency injection
- Tests use `httptest.NewServer` to mock APIs via `newFullMockServer(t)`
- Coverage target: 80%+ (currently 93.2% gmail, 92.9% calendar, 84.7% drive)

## Environment Variables
```
GOOGLE_DESKTOP_CLIENT_ID, GOOGLE_DESKTOP_CLIENT_SECRET, GMAIL_ACCESS_TOKEN, GMAIL_REFRESH_TOKEN
```

# currentDate
Today's date is 2026-03-16.
