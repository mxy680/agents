#!/usr/bin/env python3
"""
Phase 9: Professional LaTeX PDF Report
"""

import json
import sys
import subprocess
import os
from datetime import datetime


def escape_latex(s):
    """Escape special LaTeX characters."""
    if s is None:
        return "Not stated"
    s = str(s)
    replacements = [
        ("\\", "\\textbackslash{}"),
        ("&", "\\&"),
        ("%", "\\%"),
        ("$", "\\$"),
        ("#", "\\#"),
        ("_", "\\_"),
        ("{", "\\{"),
        ("}", "\\}"),
        ("~", "\\textasciitilde{}"),
        ("^", "\\textasciicircum{}"),
        ("<", "\\textless{}"),
        (">", "\\textgreater{}"),
    ]
    for old, new in replacements:
        if old != "\\":
            s = s.replace(old, new)
    return s


def fmt_price(val):
    if not val:
        return "N/A"
    try:
        return f"\\${int(val):,}"
    except:
        return str(val)


def fmt_int(val):
    if val is None or val == "Not stated":
        return "N/A"
    try:
        v = int(float(val))
        return f"{v:,}" if v > 0 else "N/A"
    except:
        return str(val) if val else "N/A"


def fmt_bool(val):
    if val is True:
        return "\\textbf{Yes}"
    if val is False:
        return "No"
    return "N/A"


def generate_property_detail(prop, rank):
    """Generate detailed paragraph for a high-priority property."""
    addr = escape_latex(prop.get("address", "N/A"))
    borough = escape_latex(prop.get("_borough", "N/A"))
    score = prop.get("_score", 0)
    priority = escape_latex(prop.get("_priority", ""))
    price = prop.get("price", 0)
    zoning = escape_latex(prop.get("_zoning", "N/A"))
    lot_area = prop.get("_lot_area", "N/A")
    bldg_area = prop.get("_bldg_area", "N/A")
    year_built = prop.get("_year_built", "N/A")
    bldg_class = prop.get("_bldg_class", "N/A")
    hpd = prop.get("_hpd_violations", 0) or 0
    dom = prop.get("daysOnMarket", 0) or 0
    bbl = prop.get("_bbl", "N/A")
    zillow_url = prop.get("zillowUrl", "")
    zola_url = prop.get("_zola_url", "")

    reasons = prop.get("_score_reasons", [])
    signals_tex = "\n".join(f"  \\item {escape_latex(r)}" for r in reasons if r)

    price_str = f"\\${price:,}" if price else "Not stated"

    out = f"""
\\subsection*{{\\textnormal{{\\small #{rank} —}} {addr}}}
\\begin{{tabular}}{{@{{}}llll@{{}}}}
\\textbf{{Borough:}} & {borough} & \\textbf{{Zoning:}} & {zoning} \\\\
\\textbf{{Asking Price:}} & {price_str} & \\textbf{{Score:}} & \\textbf{{{score}}} ({priority}) \\\\
\\textbf{{Lot Area:}} & {fmt_int(lot_area)} SF & \\textbf{{Building SF:}} & {fmt_int(bldg_area)} SF \\\\
\\textbf{{Year Built:}} & {fmt_int(year_built)} & \\textbf{{Building Class:}} & {escape_latex(bldg_class)} \\\\
\\textbf{{HPD Violations:}} & {hpd} & \\textbf{{Days on Market:}} & {dom} \\\\
\\textbf{{BBL:}} & {escape_latex(bbl)} & & \\\\
\\end{{tabular}}

\\vspace{{4pt}}
\\textbf{{Distress Signals:}}
\\begin{{itemize}}[leftmargin=2em,itemsep=1pt,parsep=0pt]
{signals_tex}
\\end{{itemize}}
"""

    if zillow_url or zola_url:
        out += "\\vspace{2pt}\\small "
        if zillow_url:
            out += f"\\textbf{{Zillow:}} \\url{{{zillow_url}}}\\quad "
        if zola_url:
            out += f"\\textbf{{ZoLa:}} \\url{{{zola_url}}}"
        out += "\n"

    out += "\\vspace{8pt}\\hrule\\vspace{6pt}\n"
    return out


def main():
    with open("/tmp/nyc_assemblage/verified_properties.json") as f:
        properties = json.load(f)

    with open("/tmp/nyc_assemblage/clusters.json") as f:
        clusters = json.load(f)

    # Stats
    total = len(properties)
    immediate = [p for p in properties if p.get("_priority") == "Immediate"]
    high = [p for p in properties if p.get("_priority") == "High"]
    moderate = [p for p in properties if p.get("_priority") == "Moderate"]
    watchlist = [p for p in properties if p.get("_priority") == "Watchlist"]

    n_tax_lien = sum(1 for p in properties if p.get("_tax_lien"))
    n_lis_pendens = sum(1 for p in properties if p.get("_acris_lis_pendens"))
    n_lis_recent = sum(1 for p in properties if p.get("_acris_lis_pendens_recent"))
    n_fed_lien = sum(1 for p in properties if p.get("_acris_federal_lien"))
    n_estate = sum(1 for p in properties if p.get("_acris_estate"))
    n_hpd5 = sum(1 for p in properties if (p.get("_hpd_violations") or 0) >= 5)
    n_hpd10 = sum(1 for p in properties if (p.get("_hpd_violations") or 0) >= 10)
    n_demo = sum(1 for p in properties if p.get("_block_demo"))
    n_nb = sum(1 for p in properties if p.get("_block_new_building"))
    n_llc = sum(1 for p in properties if p.get("_block_llc_deed"))
    n_dom180 = sum(1 for p in properties if (p.get("daysOnMarket") or 0) > 180)
    n_cluster = sum(1 for p in properties if p.get("_in_cluster"))
    n_se_drop = sum(1 for p in properties if (p.get("se_price_drop_pct") or 0) > 10)

    # Borough counts
    from collections import Counter
    borough_counts = Counter(p.get("_borough", "") for p in properties)
    zone_counts = Counter(p.get("_zoning", "") for p in properties)
    top_zones = zone_counts.most_common(8)

    # Full results table — top 50 properties
    table_rows = ""
    for p in properties[:50]:
        addr = escape_latex(p.get("address", "")[:45])
        borough_abbr = {"Bronx": "BX", "Brooklyn": "BK", "Manhattan": "MN", "Queens": "QN"}.get(p.get("_borough", ""), "??")
        price = p.get("price", 0)
        price_str = f"\\${price//1000}K" if price else "N/A"
        score = p.get("_score", 0)
        priority = p.get("_priority", "")[:3].upper()
        zoning = escape_latex(p.get("_zoning", "N/A"))
        lot = fmt_int(p.get("_lot_area"))
        yr = fmt_int(p.get("_year_built"))
        tax = "Y" if p.get("_tax_lien") else "N"
        lis = "Y" if p.get("_acris_lis_pendens") else "N"
        hpd_val = p.get("_hpd_violations", 0) or 0
        hpd_str = str(hpd_val) if hpd_val else "0"
        demo_str = "Y" if p.get("_block_demo") else "N"
        dom = p.get("daysOnMarket", 0) or 0
        dom_str = str(dom) if dom else "N/A"

        # Color coding
        if score >= 20:
            row_color = "\\rowcolor{ImmGreen}"
        elif score >= 15:
            row_color = "\\rowcolor{HighGreen}"
        elif score >= 10:
            row_color = "\\rowcolor{ModYellow}"
        else:
            row_color = ""

        table_rows += f"{row_color}{borough_abbr} & {addr} & {price_str} & {zoning} & {score} & {tax} & {lis} & {hpd_str} & {demo_str} & {dom_str} \\\\\n"

    # Cluster table
    cluster_rows = ""
    for i, cluster in enumerate(clusters[:15], 1):
        borough = escape_latex(cluster.get("borough", ""))
        block = escape_latex(cluster.get("block", ""))
        n_props = cluster.get("property_count", 0)
        addresses = "; ".join(escape_latex(a[:35]) for a in cluster.get("addresses", [])[:3])
        if len(cluster.get("addresses", [])) > 3:
            addresses += " ..."
        combined_ask = cluster.get("combined_asking_price", "")
        ask_str = f"\\${int(combined_ask):,}" if isinstance(combined_ask, (int, float)) else "N/A"
        total_lot = cluster.get("total_lot_area_sf", "")
        lot_str = f"{int(total_lot):,} SF" if isinstance(total_lot, (int, float)) else "N/A"
        zones = escape_latex(", ".join(cluster.get("zones", [])))
        max_score = cluster.get("max_score", 0)
        cluster_rows += f"{borough} & {block} & {n_props} & \\small {addresses} & {ask_str} & {lot_str} & {zones} & {max_score} \\\\\n"

    # Immediate priority details
    immediate_details = ""
    for rank, prop in enumerate(immediate[:15], 1):
        immediate_details += generate_property_detail(prop, rank)

    if not immediate_details:
        immediate_details = "\\textit{No properties reached the Immediate Priority threshold (score 20+) in this scan.}\n"

    # High priority brief table
    high_rows = ""
    for p in high[:20]:
        addr = escape_latex(p.get("address", "")[:45])
        borough_abbr = {"Bronx": "BX", "Brooklyn": "BK", "Manhattan": "MN", "Queens": "QN"}.get(p.get("_borough", ""), "??")
        score = p.get("_score", 0)
        price = p.get("price", 0)
        price_str = f"\\${price//1000}K" if price else "N/A"
        zone = escape_latex(p.get("_zoning", "N/A"))
        key_signals = escape_latex("; ".join(
            r for r in p.get("_score_reasons", [])
            if "Active for-sale" not in r
        )[:80])
        high_rows += f"\\rowcolor{{HighGreen}}{borough_abbr} & {addr} & {price_str} & {zone} & {score} & {key_signals} \\\\\n"

    latex = r"""
\documentclass[11pt,letterpaper]{article}
\usepackage[margin=1in,top=0.8in,bottom=0.8in]{geometry}
\usepackage{booktabs}
\usepackage{longtable}
\usepackage{array}
\usepackage{xcolor}
\usepackage{colortbl}
\usepackage{fancyhdr}
\usepackage{hyperref}
\usepackage{enumitem}
\usepackage{parskip}
\usepackage{microtype}
\usepackage{multicol}
\usepackage{tabularx}
\usepackage{helvet}
\renewcommand{\familydefault}{\sfdefault}

% Color definitions
\definecolor{NavyBlue}{RGB}{31,78,121}
\definecolor{ImmGreen}{RGB}{198,239,206}
\definecolor{HighGreen}{RGB}{226,240,217}
\definecolor{ModYellow}{RGB}{255,242,204}
\definecolor{LightGray}{RGB}{242,242,242}
\definecolor{DarkRed}{RGB}{192,0,0}
\definecolor{SignalOrange}{RGB}{196,89,17}

% Hyperref setup
\hypersetup{
  colorlinks=true,
  linkcolor=NavyBlue,
  urlcolor=NavyBlue,
  pdftitle={NYC Assemblage Intelligence Report},
  pdfauthor={NYC Assemblage Intelligence System}
}

% Header/footer
\pagestyle{fancy}
\fancyhf{}
\fancyhead[L]{\textcolor{NavyBlue}{\textbf{NYC Assemblage Intelligence}}}
\fancyhead[R]{\textcolor{gray}{\small March 24, 2026 --- Confidential}}
\fancyfoot[C]{\textcolor{gray}{\small\thepage}}
\renewcommand{\headrulewidth}{0.4pt}
\renewcommand{\headrule}{\color{NavyBlue}\hrule width\headwidth height\headrulewidth}

\setlength{\parindent}{0pt}

\begin{document}

% ============================================================
% TITLE PAGE
% ============================================================
\begin{center}
\vspace*{1.5cm}
{\color{NavyBlue}\rule{\textwidth}{3pt}}
\vspace{12pt}

{\LARGE\textbf{\color{NavyBlue} NYC Assemblage Intelligence Report}}

\vspace{6pt}
{\large Pre-Market Distress Signal Analysis --- R7+ Residential Development Sites}

\vspace{6pt}
{\large\textbf{March 24, 2026}}

\vspace{12pt}
{\color{NavyBlue}\rule{\textwidth}{1pt}}
\vspace{6pt}

{\small\color{gray}
Bronx \textbullet{} Brooklyn \textbullet{} Manhattan \textbullet{} Queens\\
Data sources: Zillow, NYC PLUTO, ACRIS, DOB, HPD, NYC Finance, StreetEasy\\
\textbf{CONFIDENTIAL --- For Internal Brokerage Use Only}
}
\end{center}

\vspace{0.5cm}

% ============================================================
% SECTION 1: EXECUTIVE SUMMARY
% ============================================================
\section*{\color{NavyBlue}\rule[0.5ex]{0.3\textwidth}{1pt}\quad Executive Summary \quad\rule[0.5ex]{0.3\textwidth}{1pt}}

This report presents the results of a systematic scan of all residential for-sale listings across four NYC boroughs (Bronx, Brooklyn, Manhattan, and Queens), cross-referenced against municipal databases to identify development site opportunities with pre-market distress signals.

\begin{multicols}{2}
\subsection*{Search Coverage}
\begin{itemize}[leftmargin=1.5em,itemsep=2pt]
  \item \textbf{137 zip codes} searched across 4 boroughs
  \item \textbf{Staten Island excluded} --- minimal R7+ zoning
  \item Listing types: 1--5 family residential
  \item Excluded: condos, co-ops, pending/contingent
\end{itemize}

\subsection*{Pipeline Results}
\begin{itemize}[leftmargin=1.5em,itemsep=2pt]
  \item Active listings reviewed: All for-sale 1--5 family
  \item R7+ qualifying properties: \textbf{""" + str(total) + r"""}
  \item Immediate Priority (20+): \textbf{\textcolor{DarkRed}{""" + str(len(immediate)) + r"""}}
  \item High Priority (15--19): \textbf{""" + str(len(high)) + r"""}
  \item Moderate Priority (10--14): \textbf{""" + str(len(moderate)) + r"""}
  \item Watchlist (5--9): \textbf{""" + str(len(watchlist)) + r"""}
  \item Block clusters detected: \textbf{""" + str(len(clusters)) + r"""}
\end{itemize}
\end{multicols}

\subsection*{Borough Distribution of R7+ Qualifying Properties}
\begin{tabular}{lrr}
\toprule
\textbf{Borough} & \textbf{R7+ Properties} & \textbf{\% of Total} \\
\midrule
""" + "\n".join(
    f"{b} & {c} & {100*c//total if total else 0}\\% \\\\"
    for b, c in sorted(borough_counts.items())
) + r"""
\midrule
\textbf{Total} & \textbf{""" + str(total) + r"""} & \textbf{100\%} \\
\bottomrule
\end{tabular}

% ============================================================
% SECTION 2: PRE-MARKET SIGNAL SUMMARY
% ============================================================
\section*{\color{NavyBlue}\rule[0.5ex]{0.25\textwidth}{1pt}\quad Pre-Market Signal Summary \quad\rule[0.5ex]{0.25\textwidth}{1pt}}

\begin{multicols}{2}
\subsection*{Legal \& Financial Distress}
\begin{tabular}{lr}
\toprule
\textbf{Signal} & \textbf{Count} \\
\midrule
Tax Liens (NYC Finance) & \textbf{""" + str(n_tax_lien) + r"""} \\
Lis Pendens / Judgments & \textbf{""" + str(n_lis_pendens) + r"""} \\
Lis Pendens $<$ 90 Days & \textbf{\textcolor{DarkRed}{""" + str(n_lis_recent) + r"""}} \\
Federal / IRS Liens & \textbf{""" + str(n_fed_lien) + r"""} \\
Estate / Probate Signals & \textbf{""" + str(n_estate) + r"""} \\
\bottomrule
\end{tabular}

\subsection*{Property \& Market Signals}
\begin{tabular}{lr}
\toprule
\textbf{Signal} & \textbf{Count} \\
\midrule
HPD Violations 5+ & \textbf{""" + str(n_hpd5) + r"""} \\
HPD Violations 10+ & \textbf{""" + str(n_hpd10) + r"""} \\
Demo Permit on Block & \textbf{""" + str(n_demo) + r"""} \\
New Building on Block & \textbf{""" + str(n_nb) + r"""} \\
LLC Deed on Block & \textbf{""" + str(n_llc) + r"""} \\
DOM $>$ 180 Days & \textbf{""" + str(n_dom180) + r"""} \\
In Block Cluster & \textbf{""" + str(n_cluster) + r"""} \\
SE Price Drop $>$10\% & \textbf{""" + str(n_se_drop) + r"""} \\
\bottomrule
\end{tabular}
\end{multicols}

\textbf{Key finding:} """ + str(n_tax_lien + n_lis_pendens + n_estate) + r""" properties carry at least one strong legal or financial distress signal, representing actionable off-market outreach opportunities.

% ============================================================
% SECTION 3: IMMEDIATE PRIORITY PROPERTIES
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.2\textwidth}{1pt}\quad Immediate Priority Properties (Score 20+) \quad\rule[0.5ex]{0.2\textwidth}{1pt}}

These properties carry multiple converging distress signals and represent the highest-urgency outreach targets.

""" + immediate_details + r"""

% ============================================================
% SECTION 4: HIGH PRIORITY PROPERTIES
% ============================================================
\section*{\color{NavyBlue}\rule[0.5ex]{0.25\textwidth}{1pt}\quad High Priority Properties (Score 15--19) \quad\rule[0.5ex]{0.25\textwidth}{1pt}}

\small
\begin{longtable}{>{\raggedright}p{0.5cm}>{\raggedright}p{5.5cm}p{1.3cm}p{1.0cm}r>{\raggedright}p{5.0cm}}
\toprule
\textbf{Bor.} & \textbf{Address} & \textbf{Price} & \textbf{Zone} & \textbf{Score} & \textbf{Key Signals} \\
\midrule
\endhead
""" + high_rows + r"""
\bottomrule
\end{longtable}
\normalsize

% ============================================================
% SECTION 5: BLOCK CLUSTER OPPORTUNITIES
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.2\textwidth}{1pt}\quad Top Block Cluster Opportunities \quad\rule[0.5ex]{0.2\textwidth}{1pt}}

Block clusters are groups of 2 or more qualifying R7+ properties located on the same block. These represent the highest-potential assemblage targets, where a single acquirer could approach multiple owners simultaneously.

\small
\begin{longtable}{p{1.2cm}p{1.0cm}c>{\raggedright}p{4.5cm}p{1.6cm}p{1.4cm}p{1.2cm}c}
\toprule
\textbf{Borough} & \textbf{Block} & \textbf{\#} & \textbf{Addresses} & \textbf{Combined Ask} & \textbf{Total Lot} & \textbf{Zone} & \textbf{Max Score} \\
\midrule
\endhead
""" + cluster_rows + r"""
\bottomrule
\end{longtable}
\normalsize

% ============================================================
% SECTION 6: FULL RESULTS TABLE (TOP 50)
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.2\textwidth}{1pt}\quad Full Results Table (Top 50 by Score) \quad\rule[0.5ex]{0.2\textwidth}{1pt}}

\textit{Color key: \colorbox{ImmGreen}{Green = Immediate (20+)} \quad \colorbox{HighGreen}{Light Green = High (15--19)} \quad \colorbox{ModYellow}{Yellow = Moderate (10--14)}}

\vspace{6pt}
\tiny
\begin{longtable}{>{\raggedright}p{0.4cm}>{\raggedright}p{4.8cm}p{1.1cm}p{0.9cm}cp{0.3cm}p{0.3cm}p{0.3cm}p{0.3cm}p{0.7cm}}
\toprule
\textbf{Bor} & \textbf{Address} & \textbf{Price} & \textbf{Zone} & \textbf{Score} & \textbf{TL} & \textbf{LP} & \textbf{HPD} & \textbf{DM} & \textbf{DOM} \\
\midrule
\endhead
""" + table_rows + r"""
\bottomrule
\multicolumn{10}{l}{\tiny TL=Tax Lien, LP=Lis Pendens, HPD=Violations, DM=Demo Permit on Block, DOM=Days on Market} \\
\end{longtable}
\normalsize

% ============================================================
% SECTION 7: METHODOLOGY
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.3\textwidth}{1pt}\quad Methodology \quad\rule[0.5ex]{0.3\textwidth}{1pt}}

\subsection*{Data Sources}

\begin{description}[leftmargin=2em,labelindent=0em,itemsep=4pt]
  \item[\textbf{Zillow}] Active for-sale listings searched by zip code. All 137 zip codes across Bronx, Brooklyn, Manhattan, and Queens were searched individually. Results deduped by ZPID.
  \item[\textbf{NYC PLUTO}] Authoritative lot data via NYC Planning Department. Used for: zoning district, lot area, building area, year built, building class, and residential unit count. Accessed via NYC GeoSearch geocoder and Socrata API.
  \item[\textbf{ACRIS}] NYC's official property records system (NYC Open Data). Checked for: judgment filings (lis pendens), federal/IRS liens, tax lien sale certificates, estate/probate party names, and LLC deed transfers on the same block.
  \item[\textbf{DOB}] NYC Department of Buildings permit data. Checked for demolition (DM) and new building (NB) permits filed on the same block in the last 6 months.
  \item[\textbf{HPD}] NYC Housing Preservation \& Development open violations database (dataset csn4-vhvf). Counted active violations per property as a proxy for owner neglect.
  \item[\textbf{NYC Finance}] NYC tax lien sale list. Properties appearing on this list have delinquent property taxes, indicating financial distress.
  \item[\textbf{StreetEasy}] Price history for properties with preliminary score $\geq$3. Checked for price drops, relisting cycles, and recent price reductions.
\end{description}

\subsection*{Zoning Filter}

Properties were retained only if PLUTO reports a primary zoning district of R7 or higher: R7, R7-1, R7-2, R7A, R7B, R7D, R7X, R8, R8A, R8B, R8X, R9, R9A, R9X, R10, R10A, R10X, C4-4, C4-5, or MX zones with R7+ residential component.

\subsection*{Composite Scoring Model}

\begin{tabular}{p{9cm}rp{3cm}}
\toprule
\textbf{Signal} & \textbf{Points} & \textbf{Source} \\
\midrule
Active for-sale listing & +3 & Zillow \\
R8+ zoning & +3 & PLUTO \\
Pre-war construction (before 1945) & +2 & PLUTO \\
Lot under 2,000 SF & +2 & PLUTO \\
Days on market $>$ 180 & +2 & Zillow \\
Tax lien / delinquency & +4 & NYC Finance \\
Judgment/lis pendens (last 90 days) & +5 & ACRIS \\
Judgment/lis pendens (older) & +2 & ACRIS \\
Federal/IRS lien & +3 & ACRIS \\
Tax lien sale certificate & +3 & ACRIS \\
Estate/probate signal & +5 & ACRIS \\
LLC deed on same block (last 12mo) & +3 & ACRIS \\
Demolition permit on same block (last 6mo) & +3 & DOB \\
New building permit on same block (last 6mo) & +2 & DOB \\
HPD violations 5--9 & +2 & HPD \\
HPD violations 10+ & +4 & HPD \\
Adjacent lot also for sale (cluster) & +4 & Zillow + PLUTO \\
Price drop $>$10\% from original & +3 & StreetEasy \\
3+ listing/delisting cycles & +4 & StreetEasy \\
Price drop in last 30 days & +2 & StreetEasy \\
\bottomrule
\end{tabular}

\subsection*{Priority Tiers}
\begin{description}[leftmargin=2em,itemsep=2pt]
  \item[\textbf{Immediate (20+)}] Multiple strong signals converging. Pursue immediate outreach.
  \item[\textbf{High (15--19)}] Significant distress indicators. Schedule outreach within 2 weeks.
  \item[\textbf{Moderate (10--14)}] Notable signals. Monitor and include in regular outreach rotation.
  \item[\textbf{Watchlist (5--9)}] Active R7+ listing. Monitor for signal changes.
\end{description}

\subsection*{Limitations}
\begin{itemize}[leftmargin=2em,itemsep=2pt]
  \item Zillow search results are capped at 40 per query. Price range splits were used for saturated zip codes, but some listings in highly active markets may have been missed.
  \item PLUTO zoning reflects the current primary zoning district. Overlay districts, special purpose areas, and recent rezonings may not be captured.
  \item ACRIS documents are filed asynchronously. There may be a lag between court filings and ACRIS recording.
  \item HPD violation counts reflect open violations only. Closed violations (resolved issues) are excluded.
  \item This report does not calculate FAR, development yield, or as-of-right buildable square footage.
  \item All data is as of the report date (March 24, 2026). Conditions may change rapidly.
\end{itemize}

\vspace{1cm}
\begin{center}
{\small\color{gray}
\textit{This report is for internal brokerage and acquisition planning use only.}\\
\textit{All data sourced from publicly available NYC municipal databases.}\\
\textit{Generated by NYC Assemblage Intelligence System --- March 24, 2026}
}
\end{center}

\end{document}
"""

    tex_path = "/tmp/nyc_assemblage/report.tex"
    with open(tex_path, "w") as f:
        f.write(latex)

    print(f"LaTeX source written to {tex_path}", file=sys.stderr)

    # Compile PDF
    for attempt in range(2):
        result = subprocess.run(
            ["pdflatex", "-interaction=nonstopmode", "-output-directory=/tmp/nyc_assemblage", tex_path],
            capture_output=True, text=True, timeout=120
        )
        print(f"  pdflatex attempt {attempt+1} exit code: {result.returncode}", file=sys.stderr)

        if result.returncode != 0:
            # Show errors
            lines = result.stdout.split("\n")
            errors = [l for l in lines if l.startswith("!") or "Error" in l]
            for e in errors[:20]:
                print(f"  LaTeX error: {e}", file=sys.stderr)
            if attempt == 0:
                print("  Retrying compilation...", file=sys.stderr)
        else:
            print("  ✓ PDF compiled successfully", file=sys.stderr)
            break

    pdf_path = "/tmp/nyc_assemblage/report.pdf"
    if os.path.exists(pdf_path):
        size = os.path.getsize(pdf_path)
        print(f"\n✓ PDF ready: {pdf_path} ({size:,} bytes)", file=sys.stderr)
        print(json.dumps({"path": pdf_path, "size": size}))
    else:
        print("  ✗ PDF not generated — check LaTeX errors above", file=sys.stderr)
        print(json.dumps({"path": None, "error": "compilation failed"}))


if __name__ == "__main__":
    main()
