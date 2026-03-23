# Bronx Assemblage Scout — Tool Documentation

## Authentication
All credentials are pre-configured via environment variables. Run commands directly — do not check for missing tokens.

## Tool 1: Zillow CLI (search only)

Search for properties:
```bash
integrations zillow properties search --location="Bronx, NY 10451" --limit=40 --json
```

The search returns: zpid, address, price, beds, baths, sqft, homeType, status, zillowUrl, latitude, longitude, daysOnMarket.

**Do NOT use `integrations zillow properties get`** — the detail endpoint has CSRF issues. Use NYC PLUTO (Tool 2) for lot size, year built, and building area instead.

## Tool 2: NYC PLUTO + GeoSearch (zoning, lot data, year built)

PLUTO is the authoritative source for lot data. It returns zoning, lot area, building area, year built, and building class — all in one call.

### Step A: Geocode the address to get BBL

```bash
curl -s "https://geosearch.planninglabs.nyc/v2/search?text=1776+Seminole+Ave+Bronx+NY"
```

The BBL is in `.features[0].properties.addendum.pad.bbl`. Example: `2037620044` where `2` = Bronx.

### Step B: Look up lot data via PLUTO (Socrata)

```bash
curl -s "https://data.cityofnewyork.us/resource/64uk-42ks.json?bbl=2037620044"
```

Returns: `zonedist1` (zoning), `lotarea` (lot SF), `bldgarea` (building SF), `yearbuilt`, `bldgclass`, `numbldgs`, `numfloors`, `unitsres`, `unitstotal`, `address`.

This is more reliable than Zillow for lot size and year built.

### Step C: Build the ZoLa URL

```
https://zola.planning.nyc.gov/lot/2/03762/0044
```

Format: `/lot/{borough}/{block}/{lot}` — parse from BBL where:
- Digit 1: borough (2 = Bronx)
- Digits 2-6: block (5 digits, zero-padded)
- Digits 7-10: lot (4 digits, zero-padded)

### Zoning classification guide

Only include properties zoned R7 or higher:
- **R7, R7-1, R7-2** — medium-density residential
- **R7A, R7B, R7D, R7X** — contextual medium-density
- **R8, R8A, R8B, R8X** — high-density residential
- **R9, R9A, R9X** — very high-density
- **R10, R10A, R10X** — highest density
- **C4-4, C4-5** and similar commercial zones with R7+ residential equivalent — include these too
- **M1-4/R7A** and similar MX zones — include if the R-component is R7+

Exclude: R1 through R6B (low/medium density — not worth assembling).

## Tool 3: Professional XLSX Spreadsheet (via Python openpyxl)

Create the spreadsheet as a styled .xlsx file using Python, then upload to Google Drive.

```python
import openpyxl
from openpyxl.styles import Font, PatternFill, Alignment, Border, Side

wb = openpyxl.Workbook()
ws = wb.active
ws.title = "Assemblage Scan"

# Header styling
header_font = Font(bold=True, color="FFFFFF", size=11, name="Arial")
header_fill = PatternFill(start_color="1F4E79", end_color="1F4E79", fill_type="solid")
thin_border = Border(
    left=Side(style='thin'), right=Side(style='thin'),
    top=Side(style='thin'), bottom=Side(style='thin')
)

# Write headers
headers = ["Property Address", "Asking Price", "Units", "Lot Size (SF)", "Building SF", "Year Built", "Zoning", "Starter Lot Potential", "Block Context Note", "Why This Could Be a Starting Point", "Zillow Link", "ZoLa Link", "Notes"]
for col, header in enumerate(headers, 1):
    cell = ws.cell(row=1, column=col, value=header)
    cell.font = header_font
    cell.fill = header_fill
    cell.alignment = Alignment(horizontal='center', wrap_text=True)
    cell.border = thin_border

# Color coding for Starter Lot Potential
high_fill = PatternFill(start_color="C6EFCE", end_color="C6EFCE", fill_type="solid")  # green
moderate_fill = PatternFill(start_color="FFEB9C", end_color="FFEB9C", fill_type="solid")  # yellow
low_fill = PatternFill(start_color="FFC7CE", end_color="FFC7CE", fill_type="solid")  # red

# Write data rows with formatting...
# Set column widths
ws.column_dimensions['A'].width = 35  # Address
ws.column_dimensions['B'].width = 14  # Price
# etc.

wb.save("/tmp/bronx_assemblage_scan.xlsx")
```

Upload to Google Drive:
```bash
integrations drive files upload --path=/tmp/bronx_assemblage_scan.xlsx --name="Bronx Assemblage Scan — 2026-03-23.xlsx" --json
```

## Tool 4: Professional PDF Report (via Python reportlab)

Create the report as a styled PDF, then upload to Google Drive.

```python
from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.lib.colors import HexColor
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle
from reportlab.lib import colors

doc = SimpleDocTemplate("/tmp/bronx_assemblage_report.pdf", pagesize=letter,
    topMargin=0.75*inch, bottomMargin=0.75*inch,
    leftMargin=0.75*inch, rightMargin=0.75*inch)

styles = getSampleStyleSheet()
title_style = ParagraphStyle('Title', parent=styles['Title'], fontSize=18, spaceAfter=12)
heading_style = ParagraphStyle('Heading', parent=styles['Heading2'], fontSize=14, spaceAfter=8, textColor=HexColor('#1F4E79'))
body_style = ParagraphStyle('Body', parent=styles['Normal'], fontSize=10, spaceAfter=6, leading=14)

elements = []
elements.append(Paragraph("Bronx Assemblage Report — 2026-03-23", title_style))
elements.append(Paragraph("Prepared for Brokerage Team", body_style))
elements.append(Spacer(1, 12))

# Add sections: Executive Summary, Top Opportunities, Cluster Analysis, Full Table, Methodology
elements.append(Paragraph("1. Executive Summary", heading_style))
elements.append(Paragraph("...", body_style))

# Add a formatted table
table_data = [["Address", "Price", "Zone", "Potential"], ...]
t = Table(table_data, colWidths=[2.5*inch, 1*inch, 0.8*inch, 1*inch])
t.setStyle(TableStyle([
    ('BACKGROUND', (0, 0), (-1, 0), HexColor('#1F4E79')),
    ('TEXTCOLOR', (0, 0), (-1, 0), colors.white),
    ('FONTSIZE', (0, 0), (-1, -1), 8),
    ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
    ('ROWBACKGROUNDS', (0, 1), (-1, -1), [colors.white, HexColor('#F2F2F2')]),
]))
elements.append(t)

doc.build(elements)
```

Upload to Google Drive:
```bash
integrations drive files upload --path=/tmp/bronx_assemblage_report.pdf --name="Bronx Assemblage Report — 2026-03-23.pdf" --json
```

## Tool 5: Google Drive CLI (for uploading files)

```bash
integrations drive files upload --path=/tmp/file.xlsx --name="Display Name" --json
```

Returns `{ "id": "...", "url": "..." }`.

## Workflow

1. Search Zillow for each target zip code (see job prompt for list)
2. Filter out condos, co-ops, pending/contingent listings
3. Geocode each address via NYC GeoSearch → get BBL
4. Look up zoning + lot data via PLUTO (Socrata) → get zoning, lot SF, building SF, year built
5. Filter: only keep R7+ zoned properties
6. Score starter-lot potential (Low/Moderate/High) based on observable signals
7. **Verify**: Re-read all collected data. Check for: mismatched Zillow URLs, missing zoning data, duplicate addresses, inconsistent scoring. Fix any issues found.
8. Create professional XLSX spreadsheet with styled headers, color-coded potential scores, and proper column widths. Upload to Google Drive.
9. Create professional PDF report with executive summary, top opportunities, cluster analysis, full results table, and methodology. Upload to Google Drive.

## Important
- **Use PLUTO for lot size, year built, building area** — not Zillow detail endpoint
- Process properties in batches — don't try to do all at once
- If PLUTO returns no results for a BBL, write "Unable to verify" in the zoning column
- If a property is a condo or co-op (check homeType from Zillow or bldgclass from PLUTO), skip it
- **Always run the verification step** before creating output files
- Install Python dependencies if needed: `pip install openpyxl reportlab`
- The XLSX and PDF are the final deliverables — they must be professional, accurate, and complete
