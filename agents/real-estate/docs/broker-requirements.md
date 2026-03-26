# Broker Feature Requirements

Received 2026-03-26 from Jonah's team.

## 1. Off-Market Scanner

Remove the Zillow listing requirement. Start from PLUTO (all R8+ zoned properties), run distress signals against them regardless of whether they're listed for sale. This finds motivated sellers before they list.

**First pass:** R8+ only (more lucrative, unknown volume).

**Data flow:**
1. Query PLUTO for all 1-5 family residential properties in R8+ zones across 4 boroughs
2. Run signal checks (ACRIS, DOB, HPD, NYC Finance, 311, ECB, FDNY, etc.)
3. Score and rank
4. Deliver XLSX + PDF

**Difference from current pipeline:** Current pipeline starts with Zillow active listings → PLUTO filter. This inverts it: start with PLUTO → signal enrichment. No Zillow dependency at all.

---

## 2. Probate Alert System

Standalone daily monitor for probate/estate filings on high-value properties.

**Criteria:**
- Property is in probate or estate proceedings
- AND either:
  - 15,000+ buildable SF, OR
  - 5+ residential units

**Why:** These properties sell >50% of the time. Being first to know is the edge.

**Data sources:**
- ACRIS party names containing "ESTATE OF" or "EXECUTOR"
- NYSCEF Surrogate's Court filings (probate cases)
- PLUTO for buildable SF and unit counts

**Output:** Daily alert (email or dashboard notification) with property address, owner info, buildable SF, unit count, and ACRIS filing date.

---

## 3. LLC Entity Monitor

Daily intelligence on new LLC formations that are likely real estate transactions.

### 3a. Pattern Learning (NY DOF)

Use NY Department of Finance (DOF) tax records to learn what entity naming patterns are used for real estate. DOF tax bills have entity names (e.g., "1776 SEMINOLE AVE LLC", "GRAND CONCOURSE REALTY LLC"). Build a pattern corpus from this data.

### 3b. Daily New Entity Scan (NY DOS)

NY Department of State (DOS) publishes new business entity formations. Daily:
1. Pull all new LLC/Corp formations from NY DOS
2. Pattern-match against learned real estate naming conventions
3. Flag entities that likely reference a NYC property
4. For each flagged entity:
   - Extract mailing address from the DOS filing
   - Search DOS for other entities at the same mailing address (reveals the buyer/developer)
5. Alert with: entity name, formation date, mailing address, related entities

### 3c. Manual Entity Search

User inputs a property address that's "in contract." Tool searches for LLCs formed in the last 90 days that are remotely similar to that address. This reveals who the buyer is before the deal closes publicly in ACRIS.

**Example:** Property at 1776 Seminole Ave is in contract. Search DOS for entities formed in last 90 days containing "1776", "SEMINOLE", or "1776 SEMINOLE". Results might show "1776 SEMINOLE LLC" formed 2 weeks ago by "JOHN DOE" at "123 Park Ave" — now you know the buyer.

**Data sources:**
- NY DOS Corporation & Business Entity Database (https://appext20.dos.ny.gov/corp_public/CORPSEARCH.ENTITY_SEARCH_ENTRY)
- NY DOF Property Tax Bills (for learning entity naming patterns)
- ACRIS (for cross-referencing entity names with property transactions)

---

## Priority

1. **Off-Market Scanner** — easiest, all pieces exist, just invert the pipeline
2. **Probate Alert System** — moderate, needs daily ACRIS + PLUTO cross-ref
3. **LLC Entity Monitor** — most complex, highest value, most differentiated
