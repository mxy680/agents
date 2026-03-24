# Detecting Traffic/Interest Spikes for a Specific Address

## Problem
Detect when a specific property address is getting unusual attention on: PropertyShark, ZoLa, DOB, Actovia, NYC Department of Finance, or CoStar.

## TL;DR — 3 Working Methods

| Method | What It Detects | Latency | Cost |
|--------|----------------|---------|------|
| **Filing Activity Monitoring** (NYC Open Data APIs) | New permits, complaints, deeds, mortgages, rezoning apps | 1–30 days | Free |
| **CoStar Marketing Center** (broker access required) | Actual page views, unique visitors, traffic sources per listing | Real-time | CoStar subscription |
| **CrUX API Tripwire** (Google) | When a specific URL crosses ~200+ monthly Chrome visits | 28-day rolling | Free |

---

## Method 1: Filing Activity Monitoring (Best — Free, Reliable)

**Rationale:** Traffic spikes on property research sites are *caused* by real-world events (new permits, sales, complaints, rezoning). Monitoring the filings themselves catches the signal before or simultaneously with the traffic spike.

### Data Sources (All Free, All Have APIs)

#### DOB — NYC Open Data (Socrata SODA API)

| Dataset | ID | Signal | Update |
|---------|----|--------|--------|
| DOB NOW Approved Permits | `rbx6-tga4` | Construction activity | Daily |
| DOB NOW Job Filings | `w9ak-ipjd` | Pre-permit interest | Daily |
| DOB Permit Issuance (legacy) | `ipu4-2q9a` | Permits issued | Daily |
| DOB Complaints | `eabe-havv` | Neighbor/tenant activity | Daily |
| DOB Violations | `3h2n-5cm9` | Enforcement activity | Daily |

**Example query — new permits for a specific BIN:**
```
GET https://data.cityofnewyork.us/resource/rbx6-tga4.json?
  $where=bin='1234567' AND issued_date > '2026-03-01'
  &$order=issued_date DESC
  &$limit=50
```

**Example query — complaints by address:**
```
GET https://data.cityofnewyork.us/resource/eabe-havv.json?
  $where=house_number='123' AND house_street='MAIN ST'
  &$order=date_entered DESC
```

No auth required. Free app token = unlimited requests.

#### ACRIS — NYC DOF (Socrata SODA API)

| Dataset | ID | Signal | Update |
|---------|----|--------|--------|
| Real Property Master | `bnx9-e6tj` | All recorded documents (deeds, mortgages, liens) | Monthly |
| Real Property Legals | `8h5j-fqxa` | BBL→document mapping | Monthly |
| Real Property Parties | `636b-3b5g` | Who's involved in transactions | Monthly |

**Example query — all filings for a specific BBL:**
```
GET https://data.cityofnewyork.us/resource/8h5j-fqxa.json?
  borough=1&block=1000&lot=23
  &$order=good_through_date DESC
```

Then join `document_id` to the Master table for dates and document types.

#### ACRIS NRD — Same-Day Email Alerts (FREE)

**URL:** https://a836-acrissds.nyc.gov/NRD/

Register any BBL to receive email notification **the day after** any deed or mortgage is recorded. This is the fastest free alert for ownership/financing changes. No bulk API — register each BBL via the web form (automatable via Playwright).

#### ZAP — Zoning Application Portal (Socrata)

| Dataset | ID | Signal |
|---------|-----|--------|
| ZAP Project Data | `hgx4-8ukb` | New ULURP/land use applications |
| ZAP BBL | `2iga-a6mk` | Links projects → BBLs |

A new ULURP filing against a BBL is a strong signal that a property is being considered for rezoning — this drives massive ZoLa lookup spikes.

#### Address → BIN/BBL Resolution

Use the NYC Geoclient API to convert addresses to BIN/BBL:
```
GET https://api.cityofnewyork.us/geoclient/v2/address?
  houseNumber=123&street=Main+St&borough=Brooklyn
```

### Spike Detection Logic

1. For each monitored address, query all 5+ DOB datasets + ACRIS monthly
2. Count filings per rolling 30-day window
3. Compare to historical baseline (most residential properties: 0–1 filings/year)
4. Alert when filings/month > 2× historical average

### Commercial Alternatives (Faster)

| Service | Latency | Cost | URL |
|---------|---------|------|-----|
| DOB.Watch | Minutes | ~$10–15/mo/property | https://dob.watch |
| RegWatch | Near real-time | Paid | https://regwatch.nyc |
| DOBGuard | Near real-time | Paid | https://dobguard.com |

These scrape DOB NOW directly (bypassing the 24h Open Data lag).

---

## Method 2: CoStar Marketing Center (Direct Traffic Data)

**Requirement:** You must be the listing broker/owner on CoStar or LoopNet.

The **Marketing Center** dashboard (MarketingCenter.CoStar.com) shows per-listing:

- **Total Views** (impressions in search results)
- **Detail Page Views** (clicks into listing)
- **Unique Prospects** (unique viewers)
- **Top Visitors** — company name, location, view count, return visits
- **New vs. Returning Visitors**
- **In-Market vs. Out-of-Market** breakdown
- **Traffic Sources** (CoStar Suite, LoopNet, Cityfeet, Showcase, newsletters)
- **Average Time on Page**
- **Visitor Map** (geographic origin)

Data is exportable (CSV/PDF), customizable 30-day window. This is the **only platform that directly exposes per-property traffic metrics**.

**Limitation:** Only works for properties YOU have listed on CoStar/LoopNet. No access to traffic data for other brokers' listings.

**CREXi** (crexi.com) has a similar dashboard with pageviews, impressions, unique visitors, and a real-time lead feed. CREXi also has a Listing API for partners.

---

## Method 3: CrUX API Tripwire (Free, URL-Level)

The **Chrome User Experience Report** (CrUX) API returns performance metrics for URLs visited by enough Chrome users. You can use it as a crude traffic tripwire.

**How it works:**
```bash
curl "https://chromeuxreport.googleapis.com/v1/records:queryRecord?key=YOUR_API_KEY" \
  -d '{
    "url": "https://www.propertyshark.com/mason/Property/12345/123-Main-St/",
    "formFactor": "DESKTOP"
  }'
```

- If the response has `"error": "NOT_FOUND"` → page has fewer than ~200 monthly Chrome visits
- If the response returns metrics → page crossed the traffic threshold
- **Monitor weekly:** transition from NOT_FOUND → has data = traffic spike detected

**Limitations:**
- 28-day rolling window (not real-time)
- Only detects threshold crossing, not exact view counts
- Only Chrome desktop/mobile users counted

---

## Method 4: Ahrefs Organic Search Monitoring ($499/mo)

If someone is **Googling a specific address** more frequently, Ahrefs will detect it:

1. Enter the exact PropertyShark/DOB URL in Ahrefs Site Explorer
2. View the "Organic Traffic" chart — shows estimated monthly search-driven visits
3. Set up an alert for traffic changes on that URL

**Best for:** Detecting when a property address becomes publicly discussed (news coverage, controversy, viral posts) and people start searching for it.

**Limitation:** Only sees organic search traffic. Misses direct visits, bookmarks, and social referrals.

---

## Method 5: Social Listening (For High-Profile Properties)

Use **Mention.com**, **Awario**, or **BrandMentions** to monitor when a specific property URL is shared on:
- Reddit, Twitter/X, Facebook
- Real estate forums, news sites, blogs

Set alert for: `"propertyshark.com/property/[your-address]"` OR `"123 Main St Brooklyn"`

**Cost:** $50–500/mo depending on platform.

---

## Recommended Implementation

### For a specific known property (1–10 addresses):
1. Register all BBLs in **ACRIS NRD** (free, same-day email alerts)
2. Set up a cron job polling **NYC Open Data** (DOB + ACRIS + ZAP) daily
3. Use **CrUX API** as a weekly tripwire for the PropertyShark/DOB page URLs
4. If you have broker access, check **CoStar Marketing Center** weekly

### For portfolio monitoring (10+ addresses):
1. Use **DOB.Watch** or **RegWatch** for near-real-time filing alerts
2. Build automated Socrata API polling with filing velocity detection
3. Automate ACRIS NRD registration via Playwright
4. Set up Ahrefs alerts for the top property URLs

### For detecting institutional/quiet interest:
This is the hardest case. No platform exposes who is researching a property quietly. The best proxy is:
1. **CoStar Marketing Center** (if you're the listing broker — you'll see visitor companies)
2. **ACRIS filings** (mortgages, options, assignments appear when deals progress)
3. **DOB permit applications** (new filings indicate someone is planning to develop)

---

## Platform-Specific Findings

| Platform | Exposes Traffic Data? | Best Alternative Signal |
|----------|----------------------|------------------------|
| **PropertyShark** | No — no view counts, no API, no trending indicators | Monitor page content changes (new filings update the page) |
| **ZoLa** | No — no analytics exposed | Monitor ZAP dataset for new ULURP filings by BBL |
| **DOB** | No — no view counts on DOB NOW or BIS | Monitor DOB Open Data for new permits/complaints/violations |
| **Actovia** | No — static ownership/debt data only | Use for identifying who owns a property after detecting interest elsewhere |
| **NYC DOF** | No — government sites don't expose analytics | ACRIS NRD alerts + Open Data API polling |
| **CoStar** | **YES** — Marketing Center shows full traffic analytics | Only available to listing broker/owner |

---

## Key API Endpoints Reference

```
# NYC Geoclient (address → BBL/BIN)
https://api.cityofnewyork.us/geoclient/v2/address

# DOB Permits (daily updates)
https://data.cityofnewyork.us/resource/rbx6-tga4.json

# DOB Complaints (daily updates)
https://data.cityofnewyork.us/resource/eabe-havv.json

# DOB Violations (daily updates)
https://data.cityofnewyork.us/resource/3h2n-5cm9.json

# ACRIS Legals (monthly updates)
https://data.cityofnewyork.us/resource/8h5j-fqxa.json

# ACRIS Master (monthly updates)
https://data.cityofnewyork.us/resource/bnx9-e6tj.json

# ZAP Projects (rezoning applications)
https://data.cityofnewyork.us/resource/hgx4-8ukb.json

# ZAP BBL mapping
https://data.cityofnewyork.us/resource/2iga-a6mk.json

# ACRIS NRD (same-day email alerts)
https://a836-acrissds.nyc.gov/NRD/

# CrUX API (traffic threshold detection)
https://chromeuxreport.googleapis.com/v1/records:queryRecord
```
