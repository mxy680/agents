# Off-Market Property Scanner

## TL;DR

Instead of only looking at properties listed for sale on Zillow, we scan every small residential property in high-density zones across NYC and check public records for signs the owner might want to sell — tax liens, code violations, foreclosure filings, and more. You get a ranked list of off-market opportunities before anyone else knows the owner is motivated.

---

## What It Does

Right now, our weekly scan starts with Zillow — we only look at properties that are already listed for sale. That means we're competing with every other buyer who can see the same listing.

The Off-Market Scanner flips this. We start with the city's own property database (PLUTO), which has every single lot in NYC. We pull all 1-5 family residential properties in R8 and above zones — the lots where the zoning allows significantly more density than what's currently built. Then we run each one through our full signal engine.

## What Signals We Check

For every qualifying property, we automatically check:

- **Tax liens** — Owner owes multiple years of back taxes
- **Foreclosure filings** — Bank has started legal proceedings
- **Court judgments** — Federal or IRS liens against the property
- **Estate/probate** — Owner has died, heirs are managing the property
- **Building violations** — 10+ open HPD violations means the owner has given up on maintenance
- **Environmental fines** — Unpaid ECB violations piling up
- **Fire department vacate orders** — Building declared uninhabitable
- **Eviction filings** — Landlord trying to clear the building
- **Scaffolding permits** — Stuck for 3+ years means stalled repairs

Each signal adds points to a composite score. A property with a tax lien, 15 HPD violations, and a recent foreclosure filing scores much higher than one with just a long time on market.

## What You Get

A weekly report (spreadsheet + PDF) with:

- Every R8+ property in NYC ranked by distress score
- Owner information from city records
- All signals that fired and why they matter
- Block-level analysis (are developers already buying nearby?)
- Direct links to city zoning maps and public records

## Why It Matters

The best deals in NYC real estate are the ones nobody else knows about. A property owner sitting on a tax lien with 20 HPD violations and a foreclosure filing is highly motivated to sell — but they haven't listed yet. By the time they list on Zillow, you're competing with every buyer in the market. This tool finds them first.

## First Run

We'll start with R8+ zones only. These are the most lucrative development sites, and we want to see how many properties have distress signals before expanding to R7. We expect hundreds of qualifying properties, with a meaningful subset showing strong motivation signals.
