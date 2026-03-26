# Probate Alert System

## TL;DR

When a property owner dies, their real estate almost always changes hands. We monitor NYC public records daily for estate and probate filings, then check if the property is either large enough to develop (15,000+ buildable SF) or has 5+ units. When we find a match, you get an immediate alert with the property details, owner info, and filing date — weeks or months before the property hits the open market.

---

## What It Does

Properties in probate sell more than 50% of the time. The heirs usually don't want to manage a Bronx walk-up or a Brooklyn multifamily — they want cash. The challenge is knowing about these properties early enough to approach the estate before other buyers find out.

The Probate Alert System monitors two data sources every day:

1. **ACRIS (NYC property records)** — We scan all new document filings across every borough for party names containing "ESTATE OF" or "EXECUTOR." When someone dies and their property enters probate, the estate's name appears in ACRIS filings.

2. **Property records (PLUTO)** — For each probate hit, we immediately look up the property's zoning, lot size, building area, and unit count.

## Alert Criteria

You only get notified when BOTH conditions are met:

- The property is in probate or estate proceedings
- AND the property has either:
  - **15,000+ buildable square feet** (based on zoning and lot size), OR
  - **5+ residential units**

This filters out single-family homes and small properties that aren't worth the outreach effort, and focuses on the deals that matter.

## What You Get

A daily alert (email or dashboard notification) with:

- **Property address** and borough
- **Owner / estate name** from ACRIS
- **Filing date** — how recent the probate filing is
- **Property details** — lot size, building area, zoning, year built, number of units
- **ZoLa link** — direct link to the city's zoning map for the lot
- **ACRIS link** — direct link to the filing

## Why Being Early Matters

Probate properties typically follow this timeline:

1. Owner dies → Estate files in Surrogate's Court and ACRIS (we detect it here)
2. Executor is appointed → 2-6 months
3. Estate decides to sell → 3-12 months
4. Property is listed or sold off-market → 6-18 months

By detecting at step 1, you have months of lead time over buyers who wait for a listing. The executor is often overwhelmed, unfamiliar with the property, and receptive to a clean, fast offer — especially if you're the first person to reach out.

## How Often It Runs

Daily. New ACRIS filings appear within days of the actual court filing. The system checks every morning and only alerts on new probate signals it hasn't seen before — no duplicate alerts.
