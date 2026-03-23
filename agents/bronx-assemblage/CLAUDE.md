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

## Tool 4: Professional PDF Report (via LaTeX)

Write a `.tex` file, compile with `pdflatex`, then upload the PDF to Google Drive.

```bash
# Write the .tex file
cat > /tmp/report.tex << 'LATEX'
\documentclass[11pt,letterpaper]{article}
\usepackage[margin=0.75in]{geometry}
\usepackage{booktabs}
\usepackage{longtable}
\usepackage{xcolor}
\usepackage{hyperref}
\usepackage{enumitem}
\usepackage{titlesec}
\usepackage{fancyhdr}

\definecolor{navy}{HTML}{1F4E79}
\definecolor{highgreen}{HTML}{C6EFCE}
\definecolor{moderateyellow}{HTML}{FFEB9C}
\definecolor{lightgray}{HTML}{F2F2F2}

\titleformat{\section}{\Large\bfseries\color{navy}}{}{0em}{}
\titleformat{\subsection}{\large\bfseries\color{navy}}{}{0em}{}

\pagestyle{fancy}
\fancyhead[L]{\small\color{gray}Bronx Assemblage Report}
\fancyhead[R]{\small\color{gray}\today}
\fancyfoot[C]{\thepage}

\begin{document}
\begin{center}
{\LARGE\bfseries Bronx Assemblage Report}\\[6pt]
{\large\color{gray}\today}\\[4pt]
{\normalsize Prepared for Brokerage Team}
\end{center}
\vspace{12pt}

\section{Executive Summary}
% Content here...

\section{Top Opportunities}
% Content here...

\section{Cluster Opportunities}
% Content here...

\section{Full Results}
\begin{longtable}{p{2.2in} r r r r l l}
\toprule
\textbf{Address} & \textbf{Price} & \textbf{Lot SF} & \textbf{Bldg SF} & \textbf{Year} & \textbf{Zone} & \textbf{Score} \\
\midrule
\endhead
% Data rows here...
\bottomrule
\end{longtable}

\section{Methodology}
% Content here...

\end{document}
LATEX

# Compile to PDF (run twice for references)
cd /tmp && pdflatex -interaction=nonstopmode report.tex && pdflatex -interaction=nonstopmode report.tex

# Upload
integrations drive files upload --path=/tmp/report.pdf --name="Bronx Assemblage Report — 2026-03-23.pdf" --json
```

### LaTeX tips
- Escape special characters: `\$`, `\#`, `\%`, `\&`, `\_`
- Use `\$499{,}000` for dollar amounts (comma in math mode)
- Use `\href{URL}{text}` for clickable links
- Use `\rowcolor{highgreen}` before a table row to color-code High potential
- Use `longtable` for tables that may span multiple pages

## Tool 5: Google Drive CLI (for uploading files)

```bash
integrations drive files upload --path=/tmp/file.xlsx --name="Display Name" --json
```

Returns `{ "id": "...", "url": "..." }`.

### Uploading XLSX as a native Google Sheet

Use `--convert` to auto-convert the XLSX into a native Google Sheet:
```bash
integrations drive files upload --path=/tmp/scan.xlsx --name="Bronx Assemblage Scan — 2026-03-23" --convert --json
```

The response `id` gives you the Google Sheets link:
`https://docs.google.com/spreadsheets/d/ID/edit`

## Workflow

1. Search Zillow for each target zip code (see job prompt for list)
2. Filter out condos, co-ops, pending/contingent listings
3. Geocode each address via NYC GeoSearch → get BBL
4. Look up zoning + lot data via PLUTO (Socrata) → get zoning, lot SF, building SF, year built
5. Filter: only keep R7+ zoned properties
6. Score starter-lot potential (Low/Moderate/High) based on observable signals
7. **Verify data**: Re-read all collected data. Check for: mismatched Zillow URLs, missing zoning data, duplicate addresses, inconsistent scoring. Fix any issues found.
8. Create professional XLSX spreadsheet. Upload to Google Drive with `--convert` flag.
9. Write LaTeX report, compile with `pdflatex`. **Verify the PDF**: check pdflatex output for errors/warnings. If there are errors, fix the .tex file and recompile. Repeat until clean. Upload PDF to Google Drive.

## Important
- **Use PLUTO for lot size, year built, building area** — not Zillow detail endpoint
- Process properties in batches — don't try to do all at once
- If PLUTO returns no results for a BBL, write "Unable to verify" in the zoning column
- If a property is a condo or co-op (check homeType from Zillow or bldgclass from PLUTO), skip it
- **Always run the verification step** before creating output files
- Install Python dependencies if needed: `pip install openpyxl reportlab`
- The XLSX and PDF are the final deliverables — they must be professional, accurate, and complete
