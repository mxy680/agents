#!/usr/bin/env python3
"""
Daily LLC Entity Monitor — Broad Filter
Pulls ALL new LLC formations from NY DOS and applies only minimal filtering
to exclude obvious non-real-estate entities. The agent does the real analysis.
"""
import json
import re
import sys
import os
import subprocess
from datetime import datetime, timedelta

OUT_DIR = "/tmp/llc_monitor"
TODAY = datetime.now()
# Look back 3 days to catch weekend filings
SINCE = (TODAY - timedelta(days=3)).strftime("%Y-%m-%d")

# Exclude patterns — things that are definitely NOT real estate
EXCLUDE_KEYWORDS = [
    "CONSULTING", "CONSULTANT",
    "MARKETING", "MEDIA",
    "TECHNOLOGY", "TECH", "SOFTWARE", "DIGITAL",
    "FASHION", "BEAUTY", "SALON", "SPA", "NAIL",
    "RESTAURANT", "FOOD", "CATERING", "BAKERY", "CAFE", "PIZZA", "GRILL",
    "TRUCKING", "TRANSPORT", "LOGISTICS", "MOVING",
    "CLEANING", "LAUNDRY",
    "MEDICAL", "DENTAL", "HEALTH", "THERAPY", "CLINIC", "PHARMACY",
    "LAW", "LEGAL", "ATTORNEY",
    "ACCOUNTING", "TAX SERVICE", "BOOKKEEPING",
    "INSURANCE",
    "MUSIC", "ENTERTAINMENT", "FILM", "PRODUCTION",
    "FITNESS", "GYM",
    "AUTO", "CAR WASH", "MECHANIC",
    "PLUMBING", "ELECTRIC", "HVAC",
    "LANDSCAPING", "LAWN",
    "PET", "GROOMING", "VETERINARY",
    "DAYCARE", "CHILDCARE", "TUTORING",
    "CHURCH", "MINISTRY", "TEMPLE",
]

# Strong include patterns — definitely real estate
STRONG_INCLUDE_KEYWORDS = [
    "REALTY", "REAL ESTATE",
    "PROPERTIES", "PROPERTY",
    "DEVELOPMENT", "DEVELOPERS",
    "HOLDINGS", "HOLDING",
    "EQUITIES", "EQUITY",
    "ACQUISITIONS",
    "HOUSING",
    "RESIDENTIAL",
    "CONSTRUCTION",
    "BUILDERS",
    "HOMES",
    "APARTMENTS",
    "TENANTS",
    "LANDLORD",
]

# Pattern: starts with a number (address-style LLC like "540 WEST 29 LLC")
ADDRESS_PATTERN = re.compile(r"^\d+\s+\w+", re.IGNORECASE)

# NYC borough indicators
BOROUGH_INDICATORS = ["BRONX", "BX", "BROOKLYN", "BK", "MANHATTAN", "MN",
                      "QUEENS", "QN", "STATEN", "SI"]


def classify(name: str) -> tuple[str, str]:
    """Classify an entity name. Returns (category, reason)."""
    upper = name.upper()

    # Exclude obvious non-real-estate
    for kw in EXCLUDE_KEYWORDS:
        if kw in upper:
            return "exclude", f"excluded: {kw}"

    # Strong include — definitely real estate
    for kw in STRONG_INCLUDE_KEYWORDS:
        if kw in upper:
            return "strong", f"keyword: {kw}"

    # Address pattern — starts with a number (e.g. "540 WEST 29 LLC")
    if ADDRESS_PATTERN.match(name):
        return "strong", "address pattern"

    # Contains a number anywhere (e.g. "PARKCHESTER 1776 LLC")
    if re.search(r"\d{2,}", name):
        # Has a number with 2+ digits — could be an address
        return "possible", "contains number"

    # Borough indicator + LLC
    if "LLC" in upper:
        for bi in BOROUGH_INDICATORS:
            if bi in upper:
                return "possible", f"borough: {bi}"

    # Generic LLC with ambiguous name — include for agent review
    if "LLC" in upper and len(upper.split()) <= 4:
        return "possible", "short LLC name"

    return "exclude", "no signals"


def get_recent_llcs() -> list:
    """Pull new LLC formations from NY DOS."""
    print(f"Pulling new LLC formations since {SINCE}...", file=sys.stderr)
    cmd = [
        "integrations", "nydos", "entities", "recent",
        f"--since={SINCE}", "--type=llc", "--limit=5000", "--json"
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
        if result.returncode != 0:
            print(f"  [WARN] CLI error: {result.stderr[:200]}", file=sys.stderr)
            return []
        data = json.loads(result.stdout)
        return data if isinstance(data, list) else []
    except Exception as e:
        print(f"  [WARN] Failed: {e}", file=sys.stderr)
        return []


def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    # Step 1: Get recent LLCs
    llcs = get_recent_llcs()
    print(f"  Total new LLCs: {len(llcs)}", file=sys.stderr)

    # Step 2: Broad classification
    strong: list = []
    possible: list = []
    excluded = 0

    for entity in llcs:
        name = entity.get("name", "")
        category, reason = classify(name)
        entity["match_category"] = category
        entity["match_reason"] = reason

        if category == "strong":
            strong.append(entity)
        elif category == "possible":
            possible.append(entity)
        else:
            excluded += 1

    # Combine: strong first, then possible
    candidates = strong + possible

    print(f"  Strong matches: {len(strong)}", file=sys.stderr)
    print(f"  Possible matches: {len(possible)}", file=sys.stderr)
    print(f"  Excluded: {excluded}", file=sys.stderr)
    print(f"  Total candidates for agent review: {len(candidates)}", file=sys.stderr)

    # Save ALL candidates — let the agent decide
    output_path = f"{OUT_DIR}/candidates_{TODAY.strftime('%Y-%m-%d')}.json"
    with open(output_path, "w") as f:
        json.dump(candidates, f, indent=2)

    # Print summary for the agent
    print(f"\n=== LLC Monitor Results ===", file=sys.stderr)
    print(f"  Period: {SINCE} to {TODAY.strftime('%Y-%m-%d')}", file=sys.stderr)
    print(f"  New LLCs scanned: {len(llcs)}", file=sys.stderr)
    print(f"  Strong matches: {len(strong)}", file=sys.stderr)
    print(f"  Possible matches: {len(possible)}", file=sys.stderr)
    print(f"  Excluded (non-RE): {excluded}", file=sys.stderr)

    if strong:
        print(f"\n  STRONG MATCHES:", file=sys.stderr)
        for m in strong[:30]:
            name = m.get("name", "")
            filer = m.get("filer_name", "")
            reason = m.get("match_reason", "")
            print(f"    {name} | Filer: {filer} | ({reason})", file=sys.stderr)

    print(f"\nSaved {len(candidates)} candidates to {output_path}", file=sys.stderr)
    print(json.dumps({
        "total_llcs": len(llcs),
        "strong": len(strong),
        "possible": len(possible),
        "excluded": excluded,
        "candidates": len(candidates),
        "path": output_path,
    }))


if __name__ == "__main__":
    main()
