"""Seed the database with sample NYC addresses for testing."""

import httpx
import sys

API = "http://localhost:8000"

ADDRESSES = [
    {"address": "350 5th Ave", "borough": "Manhattan"},          # Empire State Building
    {"address": "1 Vanderbilt Ave", "borough": "Manhattan"},     # One Vanderbilt
    {"address": "432 Park Ave", "borough": "Manhattan"},         # 432 Park Avenue
    {"address": "30 Hudson Yards", "borough": "Manhattan"},      # Hudson Yards
    {"address": "270 Park Ave", "borough": "Manhattan"},         # JPMorgan new HQ
]


def main():
    for addr in ADDRESSES:
        try:
            resp = httpx.post(f"{API}/addresses", json=addr)
            if resp.status_code == 201:
                data = resp.json()
                print(f"  Added: {data['address']} ({data['borough']}) → id={data['id']}")
            else:
                print(f"  Failed to add {addr['address']}: {resp.status_code} {resp.text}")
        except httpx.ConnectError:
            print(f"ERROR: Cannot connect to {API}. Is the server running?")
            sys.exit(1)

    print(f"\nSeeded {len(ADDRESSES)} addresses.")


if __name__ == "__main__":
    main()
