# Property Traffic Monitor

Detects research activity on NYC property addresses across PropertyShark, ZoLa, DOB, ACRIS, and more — before any public filing or listing appears.

## Setup

```bash
pip install -r requirements.txt
```

## Run

```bash
# Start the API server
uvicorn src.main:app --port 8000

# Or with Docker
docker compose up -d
```

## Seed & Scan

```bash
python seed.py              # Add sample NYC addresses
curl -X POST localhost:8000/scan   # Trigger a scan
```

## Verify

```bash
python verify.py   # Run verification gate
```

## Chrome Extension

1. Open Chrome → `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked" → select `src/extension/`

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/addresses` | Add address to monitor |
| GET | `/addresses` | List monitored addresses |
| DELETE | `/addresses/{id}` | Stop monitoring |
| GET | `/addresses/{id}/signals` | All signals for an address |
| GET | `/addresses/{id}/score` | Composite interest score |
| GET | `/alerts` | Recent alerts |
| POST | `/scan` | Trigger immediate scan |
| WS | `/ws/alerts` | Real-time alert stream |

## Monitors

| Monitor | Source | API Key? |
|---------|--------|----------|
| Wayback CDX | Internet Archive | No |
| Google Autocomplete | Google Suggest | No |
| Google Trends | pytrends | No |
| CrUX | Chrome UX Report | Yes (free) |
| Page Change | HTTP headers | No |
| Social Signals | Reddit, HN | No |
| Google Cache | Search engines | No |
| Chrome Extension | Crowdsourced | No |
