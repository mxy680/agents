# Agent Marketplace - Integration CLI

## Overview
Go CLI binary (`integrations`) that AI agents call inside Docker containers to interact with external services. Currently supports Gmail and Google Sheets.

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

## Architecture
- `cmd/integrations/main.go` — entrypoint, registers providers
- `internal/cli/` — Cobra root command, output helpers (JSON/text)
- `internal/auth/` — Google OAuth token management with auto-refresh
- `internal/providers/gmail/` — Gmail provider with injectable ServiceFactory
- `internal/providers/sheets/` — Sheets provider with dual ServiceFactory (Sheets + Drive)
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
  values.go              # values get/update/append/clear/batch-get/batch-update commands
  tabs.go                # tabs list/create/delete/rename commands
  helpers_test.go        # Unit tests for helpers (parseValuesJSON, formatCellsTable, etc.)
  spreadsheets_test.go   # Tests for spreadsheets commands
  values_test.go         # Tests for values commands
  tabs_test.go           # Tests for tabs commands
  sheets_test.go         # Provider-level tests (TestProviderNew, TestProviderRegisterCommands)
  mock_server_test.go    # httptest mock server helpers for Sheets + Drive APIs
```

## Testing
- Gmail commands use `ServiceFactory` for dependency injection
- Sheets commands use dual `SheetsServiceFactory` + `DriveServiceFactory` for DI
- Tests use `httptest.NewServer` to mock APIs via `newFullMockServer(t)`
- Coverage target: 80%+ (gmail: 93.2%, sheets: 84.8%)

## Environment Variables
```
GOOGLE_DESKTOP_CLIENT_ID, GOOGLE_DESKTOP_CLIENT_SECRET
GOOGLE_ACCESS_TOKEN (fallback: GMAIL_ACCESS_TOKEN)
GOOGLE_REFRESH_TOKEN (fallback: GMAIL_REFRESH_TOKEN)
```

# currentDate
Today's date is 2026-03-16.
