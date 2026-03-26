# LLC Entity Monitor

## TL;DR

In NYC real estate, buyers almost always form a new LLC before closing a deal — and the LLC name usually references the property address. We monitor the NY Secretary of State's database for new LLCs formed every day, use pattern matching to identify the ones that look like real estate entities, and alert you with the entity name, the person behind it, and their contact information. When you hear a property is "in contract," we can also search for LLCs matching that address to tell you who the buyer is — before the deal closes publicly.

---

## What It Does

### The Pattern

When a developer or investor is buying a property in NYC, they almost always:

1. Form a new LLC named after the property address (e.g., "1776 SEMINOLE AVE LLC" or "45 SEMINOLE LLC")
2. File the LLC with the NY Department of State
3. Close the deal weeks or months later
4. The deed transfer appears in ACRIS (public record) — but by then, it's old news

The LLC formation happens before the closing. It's public record. Nobody is systematically watching it.

### What We Monitor

**Every day**, we pull all new LLC and corporate formations from the NY Department of State's public database. For NYC, this is typically 200-500 new entities per day. Most of them are restaurants, tech startups, consulting firms — not relevant.

We filter for entities that look like real estate transactions by matching against patterns we learn from NYC Department of Finance tax records. Property-owning entities have distinctive naming patterns:

- Street addresses in the name ("1776 SEMINOLE AVE LLC")
- Borough references ("BX REALTY LLC", "GRAND CONCOURSE HOLDINGS")
- Common real estate suffixes ("REALTY", "HOLDINGS", "PROPERTIES", "DEVELOPMENT")
- Numbered entities that match NYC address formats

### What You Get

**Daily digest** with the handful of new entities that look like real estate deals:

- **Entity name** — The LLC name (often contains the property address)
- **Formation date** — When it was filed
- **Filer name** — The attorney or formation service (often reveals the buyer's law firm)
- **Mailing address** — The entity's registered address
- **Related entities** — Other LLCs at the same mailing address (reveals the buyer's full portfolio)

### Manual Search: "Who's Buying This Property?"

When you hear that a specific property is in contract, you can search for it:

> "Search for any entities formed in the last 90 days that match 1776 Seminole Ave"

The tool searches the NY DOS database for LLCs containing "1776" or "SEMINOLE" formed recently. If someone formed "1776 SEMINOLE AVE LLC" three weeks ago, you now know the buyer — and you can look up the filer's other entities to identify the person or firm behind the deal.

We tested this today. Searching for "1776 Seminole Ave" returned:
- **45 SEMINOLE LLC** — filed by Nafatli Unger, address: 1303 53rd Street, Brooklyn
- **ARISA 1776 OWNER LLC** — filed by a formation service for an entity at 285 West End Avenue

These are real results from today's database. Each one tells a story about who is actively transacting around that address.

## Why This Is Valuable

**Competitive intelligence.** When you know who's buying before the deal closes, you can:

- Identify which developers are active in your target areas
- Approach adjacent property owners before the developer does
- Understand market activity in real time, not 3 months after closing
- Track specific developers' acquisition patterns across the city

**Deal origination.** When you see a new LLC formed for an address near your target area, it confirms market momentum — someone else believes this block is worth investing in. That's a signal to accelerate your own outreach to nearby owners.

**Speed.** LLC formations appear in the Department of State database within days of filing. ACRIS deed transfers can lag by weeks or months. This gives you a significant time advantage.

## How It Works

1. The NY Department of State publishes all new entity filings daily through a public data portal
2. Our system pulls the filings every morning
3. Pattern matching identifies real estate entities (using patterns learned from tax records)
4. Matches are enriched with related entities at the same mailing address
5. Results are delivered as a daily alert
6. Manual searches are available on-demand for specific addresses
