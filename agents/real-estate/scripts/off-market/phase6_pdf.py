#!/usr/bin/env python3
"""
Phase 6: Professional LaTeX PDF Report (Off-Market R8+ Properties)

Input:  /tmp/off_market_scan/verified_properties.json
        /tmp/off_market_scan/clusters.json
Output: /tmp/off_market_scan/report.pdf
"""

import json
import sys
import subprocess
import os
from collections import Counter
from datetime import datetime

DATE_DISPLAY = datetime.now().strftime("%B %d, %Y")
DATE_SHORT = datetime.now().strftime("%Y-%m-%d")


def escape_latex(s) -> str:
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


def fmt_int(val) -> str:
    if val is None or val == "Not stated":
        return "N/A"
    try:
        v = int(float(str(val).replace(",", "").replace("$", "")))
        return f"{v:,}" if v > 0 else "N/A"
    except Exception:
        return str(val) if val else "N/A"


def fmt_bool(val) -> str:
    if val is True:
        return "\\textbf{Yes}"
    if val is False:
        return "No"
    return "N/A"


def generate_property_detail(prop: dict, rank: int) -> str:
    """Generate detailed analysis block for a high-priority property."""
    addr = escape_latex(prop.get("address", "N/A"))
    borough = escape_latex(prop.get("_borough", "N/A"))
    score = prop.get("_score", 0)
    priority = escape_latex(prop.get("_priority", ""))
    zoning = escape_latex(prop.get("_zoning", "N/A"))
    lot_area = prop.get("_lot_area", "N/A")
    bldg_area = prop.get("_bldg_area", "N/A")
    year_built = prop.get("_year_built", "N/A")
    bldg_class = prop.get("_bldg_class", "N/A")
    hpd = prop.get("_hpd_violations", 0) or 0
    bbl = prop.get("_bbl", "N/A")
    zola_url = prop.get("_zola_url", "")
    tax_lien = prop.get("_tax_lien", False)
    lis_pendens = prop.get("_acris_lis_pendens", False)
    estate = prop.get("_acris_estate", False)
    llc_block = prop.get("_block_llc_deed", False)
    in_cluster = prop.get("_in_cluster", False)

    reasons = prop.get("_score_reasons", [])
    signals_tex = "\n".join(f"  \\item {escape_latex(r)}" for r in reasons if r)

    try:
        lot_sf = int(float(str(lot_area).replace(",", "").replace("$", ""))) if lot_area and lot_area != "N/A" else 0
        bldg_sf = int(float(str(bldg_area).replace(",", "").replace("$", ""))) if bldg_area and bldg_area != "N/A" else 0
        yr = int(float(str(year_built).replace(",", "").replace("$", ""))) if year_built and year_built != "N/A" else 0
    except (ValueError, TypeError):
        lot_sf, bldg_sf, yr = 0, 0, 0

    # Build "WHY" analysis
    why_parts = []

    zone_upper = (prop.get("_zoning") or "").upper()
    if any(zone_upper.startswith(p) for p in ["R8", "R9", "R10"]):
        why_parts.append(
            f"This property sits in a \\textbf{{{zoning}}} zone, one of the highest-density residential "
            f"designations in NYC. The zoning allows substantially more density than the current building "
            f"provides, making the land far more valuable than the structure."
        )
    elif zone_upper.startswith("R7"):
        why_parts.append(
            f"Zoned \\textbf{{{zoning}}}, this property is in a medium-to-high density residential district "
            f"where the city permits significantly larger buildings than what currently exists on the lot."
        )

    if yr > 0 and yr < 1945:
        why_parts.append(
            f"Built in {yr}, this is a {datetime.now().year - yr}-year-old structure. Pre-war buildings "
            f"in this condition typically have high maintenance costs and declining structural integrity, "
            f"making owners more receptive to acquisition offers --- especially when coupled with other distress signals."
        )

    if hpd >= 50:
        why_parts.append(
            f"\\textbf{{The building has {hpd} open HPD violations}} --- an extraordinary number indicating "
            f"severe, sustained neglect. Buildings at this violation level often face HPD emergency repair "
            f"orders, which add liens to the property. This is one of the strongest indicators of an owner ready to sell."
        )
    elif hpd >= 10:
        why_parts.append(
            f"With \\textbf{{{hpd} open HPD violations}}, this building shows significant owner neglect. "
            f"The violations likely include hazardous conditions the owner is not addressing --- a classic signal "
            f"of an owner who has lost the ability or willingness to maintain the property."
        )
    elif hpd >= 5:
        why_parts.append(
            f"The property has {hpd} open HPD violations, indicating moderate neglect. This suggests the owner "
            f"may be experiencing financial strain or disengagement from the property."
        )

    if tax_lien:
        why_parts.append(
            "\\textbf{This property appears on the NYC tax lien sale list}, meaning the owner owes multiple "
            "years of delinquent property taxes. NYC will auction the lien if unpaid. This is a direct indicator "
            "of financial distress --- an off-market acquisition offer may be welcomed as a way to resolve the debt."
        )

    if lis_pendens:
        why_parts.append(
            "\\textbf{A lis pendens (foreclosure filing) has been recorded against this property in ACRIS.} "
            "The bank or lender has initiated legal proceedings to take the property. The owner is under severe "
            "time pressure and is among the most motivated sellers in this dataset."
        )

    if estate:
        why_parts.append(
            "\\textbf{ACRIS party records indicate estate or probate involvement} (party names contain ``ESTATE OF'' "
            "or ``EXECUTOR''). The original owner likely died and heirs are managing disposition. Estate-held "
            "properties frequently sell below market value because heirs want liquidity, not long-term ownership."
        )

    if llc_block:
        why_parts.append(
            "Recent ACRIS records show \\textbf{deed transfers to LLCs on this block}, indicating that developer "
            "acquisition activity is already underway in the immediate area. This is both an opportunity (the block "
            "is ``in play'') and a competitive signal."
        )

    if in_cluster:
        cluster_size = prop.get("_cluster_size", 0)
        why_parts.append(
            f"This property is part of a \\textbf{{cluster of {cluster_size} qualifying properties on the same block}}. "
            f"Multiple off-market opportunities on one block create a rare assemblage entry point where 2+ lots "
            f"can be approached simultaneously without any being publicly listed."
        )

    if prop.get("_fdny_vacate"):
        why_parts.append(
            "\\textbf{An FDNY vacate order has been issued on this block}, meaning at least one building has been "
            "deemed uninhabitable. The affected owner faces costly remediation or demolition --- a strong motivation to sell."
        )

    complaints_311 = prop.get("_311_complaints", 0) or 0
    if complaints_311 >= 10:
        why_parts.append(
            f"The address has received \\textbf{{{complaints_311} 311 complaints}} in the last 12 months, "
            f"indicating significant quality-of-life issues and building neglect."
        )

    ecb = prop.get("_ecb_violations", 0) or 0
    if ecb > 0:
        why_parts.append(
            f"The property has {ecb} defaulted ECB/OATH violations --- unpaid environmental fines that add "
            f"to the owner's financial burden."
        )

    if lot_sf > 0 and lot_sf < 2000:
        why_parts.append(
            f"The lot is only {lot_sf:,} SF --- too small for significant standalone development but ideal as "
            f"an assemblage \\textit{{starter lot}}. Acquiring this parcel and then approaching adjacent owners "
            f"creates the foundation for a developable site."
        )

    why_text = "\n\n".join(why_parts) if why_parts else (
        "This property qualifies based on zoning and lot characteristics. While no strong individual distress "
        "signals are present at this time, the R8+ zoning designation makes it a worthwhile off-market outreach candidate."
    )

    out = f"""
\\subsection*{{\\textnormal{{\\small \\#{rank} ---}} {addr}}}

\\begin{{tabular}}{{@{{}}p{{3.5cm}}p{{3.5cm}}p{{3.5cm}}p{{3.5cm}}@{{}}}}
\\textbf{{Borough:}} {borough} & \\textbf{{Zoning:}} {zoning} & \\textbf{{Lot:}} {fmt_int(lot_area)} SF & \\textbf{{Score:}} \\textbf{{{score}}} ({priority}) \\\\
\\textbf{{Year Built:}} {fmt_int(year_built)} & \\textbf{{HPD Violations:}} {hpd} & \\textbf{{BBL:}} {escape_latex(bbl)} & \\textbf{{Class:}} {escape_latex(bldg_class)} \\\\
\\end{{tabular}}

\\vspace{{6pt}}
\\textbf{{Why This Property Stands Out:}}

{why_text}

\\vspace{{4pt}}
\\textbf{{Scoring Breakdown:}}
\\begin{{itemize}}[leftmargin=2em,itemsep=1pt,parsep=0pt]
{signals_tex}
\\end{{itemize}}
"""

    if zola_url:
        out += f"\\vspace{{2pt}}\\small \\href{{{zola_url}}}{{ZoLa Zoning Map}}\n"

    out += "\\vspace{10pt}\\hrule\\vspace{8pt}\n"
    return out


def main():
    with open("/tmp/off_market_scan/verified_properties.json") as f:
        properties = json.load(f)

    with open("/tmp/off_market_scan/clusters.json") as f:
        clusters = json.load(f)

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
    n_cluster = sum(1 for p in properties if p.get("_in_cluster"))
    n_311 = sum(1 for p in properties if (p.get("_311_complaints") or 0) >= 10)
    n_ecb = sum(1 for p in properties if (p.get("_ecb_violations") or 0) > 0)
    n_fdny = sum(1 for p in properties if p.get("_fdny_vacate"))
    n_dob_complaints = sum(1 for p in properties if (p.get("_dob_complaints") or 0) >= 3)
    n_block_co = sum(1 for p in properties if p.get("_block_co"))
    n_citibike = sum(1 for p in properties if (p.get("_citibike_stations") or 0) >= 5)

    borough_counts = Counter(p.get("_borough", "") for p in properties)

    # Full results table — top 50 properties
    table_rows = ""
    for p in properties[:50]:
        addr = escape_latex(p.get("address", "")[:45])
        borough_abbr = {"Bronx": "BX", "Brooklyn": "BK", "Manhattan": "MN", "Queens": "QN"}.get(
            p.get("_borough", ""), "??"
        )
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

        if score >= 15:
            row_color = "\\rowcolor{ImmGreen}"
        elif score >= 10:
            row_color = "\\rowcolor{HighGreen}"
        elif score >= 6:
            row_color = "\\rowcolor{ModYellow}"
        else:
            row_color = ""

        table_rows += f"{row_color}{borough_abbr} & {addr} & {zoning} & {score} & {tax} & {lis} & {hpd_str} & {demo_str} \\\\\n"

    # Cluster table
    cluster_rows = ""
    for i, cluster in enumerate(clusters[:15], 1):
        borough = escape_latex(cluster.get("borough", ""))
        block = escape_latex(cluster.get("block", ""))
        n_props = cluster.get("property_count", 0)
        addresses = "; ".join(escape_latex(a[:35]) for a in cluster.get("addresses", [])[:3])
        if len(cluster.get("addresses", [])) > 3:
            addresses += " ..."
        total_lot = cluster.get("total_lot_area_sf", "")
        lot_str = f"{int(total_lot):,} SF" if isinstance(total_lot, (int, float)) else "N/A"
        zones = escape_latex(", ".join(cluster.get("zones", [])))
        max_score = cluster.get("max_score", 0)
        cluster_rows += f"{borough} & {block} & {n_props} & \\small {addresses} & {lot_str} & {zones} & {max_score} \\\\\n"

    # Top 25 property details (Immediate + High)
    top_details = ""
    top_props = immediate + high
    for rank, prop in enumerate(top_props[:25], 1):
        top_details += generate_property_detail(prop, rank)

    if not top_details:
        top_details = "\\textit{No properties reached High Priority or above in this scan.}\n"

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

% Hyperref setup
\hypersetup{
  colorlinks=true,
  linkcolor=NavyBlue,
  urlcolor=NavyBlue,
  pdftitle={NYC Off-Market R8+ Intelligence Report},
  pdfauthor={NYC Off-Market Intelligence System}
}

% Header/footer
\pagestyle{fancy}
\fancyhf{}
\fancyhead[L]{\textcolor{NavyBlue}{\textbf{NYC Off-Market R8+ Intelligence}}}
\fancyhead[R]{\textcolor{gray}{\small """ + DATE_DISPLAY + r""" --- Confidential}}
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

{\LARGE\textbf{\color{NavyBlue} NYC Off-Market R8+ Intelligence Report}}

\vspace{6pt}
{\large Pre-Market Distress Signal Analysis --- Properties Not Listed For Sale}

\vspace{6pt}
{\large\textbf{""" + DATE_DISPLAY + r"""}}

\vspace{12pt}
{\color{NavyBlue}\rule{\textwidth}{1pt}}
\vspace{6pt}

{\small\color{gray}
Bronx \textbullet{} Brooklyn \textbullet{} Manhattan \textbullet{} Queens\\
Data sources: PLUTO, ACRIS, DOB, HPD, NYC Finance, 311, ECB, FDNY, Citi Bike\\
\textbf{CONFIDENTIAL --- For Internal Brokerage Use Only}
}

\vspace{12pt}
\fbox{\parbox{0.85\textwidth}{\centering\small\textbf{Important:} These properties are NOT currently listed for sale.
Distress signals indicate the owner may be motivated to sell. This report is designed for
proactive off-market outreach to owners before properties enter the market.}}
\end{center}

\vspace{0.5cm}

% ============================================================
% SECTION 1: EXECUTIVE SUMMARY
% ============================================================
\section*{\color{NavyBlue}\rule[0.5ex]{0.3\textwidth}{1pt}\quad Executive Summary \quad\rule[0.5ex]{0.3\textwidth}{1pt}}

This report presents the results of a systematic analysis of all R8+ zoned properties across four NYC boroughs,
cross-referenced against municipal distress databases to identify off-market acquisition opportunities.
\textbf{None of these properties are currently listed for sale.} Distress signals indicate owners who may
be financially motivated to entertain off-market offers.

\begin{multicols}{2}
\subsection*{Pipeline Results}
\begin{itemize}[leftmargin=1.5em,itemsep=2pt]
  \item R8+ off-market properties analyzed: \textbf{""" + str(total) + r"""}
  \item Immediate Priority (15+): \textbf{\textcolor{DarkRed}{""" + str(len(immediate)) + r"""}}
  \item High Priority (10--14): \textbf{""" + str(len(high)) + r"""}
  \item Moderate Priority (6--9): \textbf{""" + str(len(moderate)) + r"""}
  \item Watchlist (1--5): \textbf{""" + str(len(watchlist)) + r"""}
  \item Block clusters detected: \textbf{""" + str(len(clusters)) + r"""}
\end{itemize}

\subsection*{Top Distress Signals}
\begin{itemize}[leftmargin=1.5em,itemsep=2pt]
  \item Tax liens: \textbf{""" + str(n_tax_lien) + r"""}
  \item Lis pendens (any): \textbf{""" + str(n_lis_pendens) + r"""}
  \item Lis pendens $<$90 days: \textbf{\textcolor{DarkRed}{""" + str(n_lis_recent) + r"""}}
  \item Estate/probate signals: \textbf{""" + str(n_estate) + r"""}
  \item Federal/IRS liens: \textbf{""" + str(n_fed_lien) + r"""}
  \item HPD violations 10+: \textbf{""" + str(n_hpd10) + r"""}
  \item FDNY vacate orders: \textbf{\textcolor{DarkRed}{""" + str(n_fdny) + r"""}}
\end{itemize}
\end{multicols}

\subsection*{Borough Distribution}
\begin{tabular}{lrr}
\toprule
\textbf{Borough} & \textbf{R8+ Properties} & \textbf{\% of Total} \\
\midrule
""" + "\n".join(
    f"{b} & {c} & {100*c//total if total else 0}\\% \\\\"
    for b, c in sorted(borough_counts.items())
) + r"""
\midrule
\textbf{Total} & \textbf{""" + str(total) + r"""} & \textbf{100\%} \\
\bottomrule
\end{tabular}

\textbf{Key finding:} """ + str(n_tax_lien + n_lis_pendens + n_estate) + r""" properties carry at least one strong legal or financial distress signal,
representing actionable off-market outreach opportunities.

% ============================================================
% SECTION 2: SIGNAL SUMMARY
% ============================================================
\section*{\color{NavyBlue}\rule[0.5ex]{0.25\textwidth}{1pt}\quad Distress Signal Summary \quad\rule[0.5ex]{0.25\textwidth}{1pt}}

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

\subsection*{Property \& Block Signals}
\begin{tabular}{lr}
\toprule
\textbf{Signal} & \textbf{Count} \\
\midrule
HPD Violations 5+ & \textbf{""" + str(n_hpd5) + r"""} \\
HPD Violations 10+ & \textbf{""" + str(n_hpd10) + r"""} \\
Demo Permit on Block & \textbf{""" + str(n_demo) + r"""} \\
New Building on Block & \textbf{""" + str(n_nb) + r"""} \\
LLC Deed on Block & \textbf{""" + str(n_llc) + r"""} \\
In Block Cluster & \textbf{""" + str(n_cluster) + r"""} \\
311 Complaints 10+ & \textbf{""" + str(n_311) + r"""} \\
ECB Defaulted Violations & \textbf{""" + str(n_ecb) + r"""} \\
FDNY Vacate Orders & \textbf{\textcolor{DarkRed}{""" + str(n_fdny) + r"""}} \\
DOB Complaints 3+ & \textbf{""" + str(n_dob_complaints) + r"""} \\
New CO on Block & \textbf{""" + str(n_block_co) + r"""} \\
CitiBike 5+ Stations & \textbf{""" + str(n_citibike) + r"""} \\
\bottomrule
\end{tabular}
\end{multicols}

% ============================================================
% SECTION 3: TOP 25 PROPERTY DETAILS
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.15\textwidth}{1pt}\quad Priority Properties --- Detailed Analysis \quad\rule[0.5ex]{0.15\textwidth}{1pt}}

The following properties scored 10 or above in the composite model, indicating multiple converging distress signals.
\textbf{These properties are not listed for sale.} Each entry explains why the owner may be motivated to sell.

""" + top_details + r"""

% ============================================================
% SECTION 4: BLOCK CLUSTER OPPORTUNITIES
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.2\textwidth}{1pt}\quad Block Cluster Opportunities \quad\rule[0.5ex]{0.2\textwidth}{1pt}}

Block clusters are groups of 2 or more qualifying R8+ properties located on the same block. These represent the
highest-potential assemblage targets, where a single acquirer can approach multiple off-market owners on the same block.

\small
\begin{longtable}{p{1.2cm}p{1.0cm}cp{4.5cm}p{1.4cm}p{1.2cm}c}
\toprule
\textbf{Borough} & \textbf{Block} & \textbf{\#} & \textbf{Addresses} & \textbf{Total Lot} & \textbf{Zone} & \textbf{Max Score} \\
\midrule
\endhead
""" + cluster_rows + r"""
\bottomrule
\end{longtable}
\normalsize

% ============================================================
% SECTION 5: FULL RESULTS TABLE (TOP 50)
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.2\textwidth}{1pt}\quad Full Results Table (Top 50 by Score) \quad\rule[0.5ex]{0.2\textwidth}{1pt}}

\textit{Color key: \colorbox{ImmGreen}{Green = Immediate (15+)} \quad \colorbox{HighGreen}{Light Green = High (10--14)} \quad \colorbox{ModYellow}{Yellow = Moderate (6--9)}}

\vspace{6pt}
\tiny
\begin{longtable}{>{\raggedright}p{0.4cm}>{\raggedright}p{5.2cm}p{0.9cm}cp{0.3cm}p{0.3cm}p{0.3cm}p{0.3cm}}
\toprule
\textbf{Bor} & \textbf{Address} & \textbf{Zone} & \textbf{Score} & \textbf{TL} & \textbf{LP} & \textbf{HPD} & \textbf{DM} \\
\midrule
\endhead
""" + table_rows + r"""
\bottomrule
\multicolumn{8}{l}{\tiny TL=Tax Lien, LP=Lis Pendens, HPD=Violations (count), DM=Demo Permit on Block} \\
\end{longtable}
\normalsize

% ============================================================
% SECTION 6: METHODOLOGY
% ============================================================
\newpage
\section*{\color{NavyBlue}\rule[0.5ex]{0.3\textwidth}{1pt}\quad Methodology \quad\rule[0.5ex]{0.3\textwidth}{1pt}}

\subsection*{Data Sources}

\begin{description}[leftmargin=2em,labelindent=0em,itemsep=4pt]
  \item[\textbf{NYC PLUTO}] Authoritative lot data via NYC Planning Department. Used for: zoning district, lot area, building area, year built, building class, and unit count. R8+ filter applied (R8, R8A, R8B, R8X, R9, R9A, R9X, R10, R10A, R10X, and equivalent commercial/mixed districts).
  \item[\textbf{ACRIS}] NYC's official property records system (NYC Open Data). Checked for: judgment filings (lis pendens), federal/IRS liens, tax lien sale certificates, estate/probate party names, and LLC deed transfers on the same block.
  \item[\textbf{DOB}] NYC Department of Buildings permit data. Checked for demolition (DM) and new building (NB) permits filed on the same block in the last 6 months.
  \item[\textbf{HPD}] NYC Housing Preservation \& Development open violations database (dataset csn4-vhvf). Counted active violations per property as a proxy for owner neglect.
  \item[\textbf{NYC Finance}] NYC tax lien sale list. Properties appearing on this list have delinquent property taxes, indicating financial distress.
  \item[\textbf{311}] NYC 311 complaint data (Socrata). Complaint volume at property address in last 12 months.
  \item[\textbf{ECB/OATH}] Environmental Control Board defaulted violations. Unpaid environmental fines signal owner disinvestment.
  \item[\textbf{FDNY}] Fire Department vacate orders. Buildings ordered vacated are uninhabitable --- strong motivated seller signal.
  \item[\textbf{DOB Complaints}] Department of Buildings open complaints (unsafe structure, illegal conversion).
  \item[\textbf{Citi Bike}] Station density within 1km via GBFS feed. Transit accessibility proxy for land value.
\end{description}

\subsection*{Composite Scoring Model}

\begin{tabular}{p{9cm}rp{3cm}}
\toprule
\textbf{Signal} & \textbf{Points} & \textbf{Source} \\
\midrule
R8+ zoning & +3 & PLUTO \\
Pre-war construction (before 1945) & +2 & PLUTO \\
Lot under 2,000 SF & +2 & PLUTO \\
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
Adjacent off-market lot on same block (cluster) & +4 & PLUTO \\
311 complaints 10+ in 12 months & +3 & 311 \\
Defaulted ECB/OATH violations & +2 & ECB \\
FDNY vacate order on block & +5 & FDNY \\
Open DOB complaints 3+ & +2 & DOB \\
New CO on block (last 12mo) & +2 & DOB CO \\
Citi Bike 5+ stations within 1km & +2 & Citi Bike \\
\textbf{Off-market with any distress signal (bonus)} & \textbf{+3} & \textbf{All} \\
\bottomrule
\end{tabular}

\subsection*{Priority Tiers}
\begin{description}[leftmargin=2em,itemsep=2pt]
  \item[\textbf{Immediate (15+)}] Multiple strong signals. Pursue immediate off-market outreach.
  \item[\textbf{High (10--14)}] Significant distress indicators. Schedule outreach within 2 weeks.
  \item[\textbf{Moderate (6--9)}] Notable signals. Include in regular outreach rotation.
  \item[\textbf{Watchlist (1--5)}] R8+ property with minimal signals. Monitor for changes.
\end{description}

\subsection*{Limitations}
\begin{itemize}[leftmargin=2em,itemsep=2pt]
  \item PLUTO zoning reflects the current primary zoning district. Overlay districts and recent rezonings may not be captured.
  \item ACRIS documents are filed asynchronously. There may be a lag between court filings and ACRIS recording.
  \item HPD violation counts reflect open violations only.
  \item This report does not calculate FAR, development yield, or as-of-right buildable square footage.
  \item All data is as of the report date (""" + DATE_DISPLAY + r"""). Conditions may change rapidly.
  \item Owner contact information is not included. Outreach requires owner name lookup via ACRIS/DOF records.
\end{itemize}

\vspace{1cm}
\begin{center}
{\small\color{gray}
\textit{This report is for internal brokerage and acquisition planning use only.}\\
\textit{All data sourced from publicly available NYC municipal databases.}\\
\textit{Generated by NYC Off-Market Intelligence System --- """ + DATE_DISPLAY + r"""}
}
\end{center}

\end{document}
"""

    tex_path = "/tmp/off_market_scan/report.tex"
    with open(tex_path, "w") as f:
        f.write(latex)

    print(f"LaTeX source written to {tex_path}", file=sys.stderr)

    # Compile PDF (run twice for cross-references)
    for attempt in range(2):
        result = subprocess.run(
            ["pdflatex", "-interaction=nonstopmode", "-output-directory=/tmp/off_market_scan", tex_path],
            capture_output=True, text=True, timeout=120
        )
        print(f"  pdflatex attempt {attempt+1} exit code: {result.returncode}", file=sys.stderr)

        if result.returncode != 0:
            lines = result.stdout.split("\n")
            errors = [l for l in lines if l.startswith("!") or "Error" in l]
            for e in errors[:20]:
                print(f"  LaTeX error: {e}", file=sys.stderr)
            if attempt == 0:
                print("  Retrying compilation...", file=sys.stderr)
        else:
            print("  PDF compiled successfully", file=sys.stderr)
            break

    pdf_path = "/tmp/off_market_scan/report.pdf"
    if os.path.exists(pdf_path):
        size = os.path.getsize(pdf_path)
        print(f"\nPDF ready: {pdf_path} ({size:,} bytes)", file=sys.stderr)
        print(json.dumps({"path": pdf_path, "size": size}))
    else:
        print("  PDF not generated --- check LaTeX errors above", file=sys.stderr)
        print(json.dumps({"path": None, "error": "compilation failed"}))


if __name__ == "__main__":
    main()
