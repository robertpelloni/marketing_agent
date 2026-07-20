#!/usr/bin/env python3
"""
Apollo Contact Exporter
Exports contacts from Apollo API to CSV
"""

import requests
import csv
import time
from datetime import datetime

API_KEY = "Hk2hKzzYIddTsygrnoufDw"
BASE_URL = "https://api.apollo.io/api/v1/mixed_people/api_search"
HEADERS = {"Content-Type": "application/json", "X-Api-Key": API_KEY}


def export_contacts(search_name, params, max_contacts=10000):
    """Export contacts matching search parameters."""
    print(f"\n{'=' * 60}")
    print(f"Exporting: {search_name}")
    print(f"{'=' * 60}")

    all_contacts = []
    page = 1

    while len(all_contacts) < max_contacts:
        try:
            search_params = {**params, "page": page, "per_page": 100}
            resp = requests.post(
                BASE_URL, json=search_params, headers=HEADERS, timeout=30
            )
            resp.raise_for_status()
            data = resp.json()

            people = data.get("people", [])
            total = data.get("total_entries", 0)

            if not people:
                break

            all_contacts.extend(people)
            print(
                f"  Page {page}: {len(people)} contacts ({len(all_contacts):,}/{total:,})"
            )

            if len(all_contacts) >= total or len(all_contacts) >= max_contacts:
                break

            page += 1
            time.sleep(0.5)

        except Exception as e:
            print(f"  Error on page {page}: {e}")
            time.sleep(1)
            continue

    # Export to CSV
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"apollo_{search_name.replace(' ', '_')}_{timestamp}.csv"

    fieldnames = [
        "id",
        "first_name",
        "last_name",
        "title",
        "email",
        "linkedin_url",
        "phone",
        "city",
        "state",
        "country",
        "organization_name",
        "organization_website",
        "organization_industry",
    ]

    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()

        for person in all_contacts:
            org = person.get("organization", {})
            row = {
                "id": person.get("id", ""),
                "first_name": person.get("first_name", ""),
                "last_name": person.get(
                    "last_name", person.get("last_name_obfuscated", "")
                ),
                "title": person.get("title", ""),
                "email": person.get("email", ""),
                "linkedin_url": person.get("linkedin_url", ""),
                "phone": person.get("phone", ""),
                "city": person.get("city", ""),
                "state": person.get("state", ""),
                "country": person.get("country", ""),
                "organization_name": org.get("name", ""),
                "organization_website": org.get("website_url", ""),
                "organization_industry": org.get("industry", ""),
            }
            writer.writerow(row)

    print(f"\nExported {len(all_contacts):,} contacts to {filename}")
    return filename, len(all_contacts)


def main():
    """Main function."""
    print("=" * 60)
    print("Apollo Contact Exporter")
    print("=" * 60)
    print(f"API Key: {API_KEY[:10]}...")

    # Define export categories
    exports = [
        {
            "name": "CTOs",
            "params": {
                "person_titles": ["CTO", "Chief Technology Officer"],
                "per_page": 100,
            },
        },
        {
            "name": "VP_Engineering",
            "params": {
                "person_titles": ["VP Engineering", "VP of Engineering"],
                "per_page": 100,
            },
        },
        {
            "name": "AI_ML_Engineers",
            "params": {
                "q_keywords": "AI machine learning",
                "person_titles": ["Engineer", "Developer"],
                "per_page": 100,
            },
        },
        {
            "name": "Founders",
            "params": {
                "person_titles": ["Founder", "Co-Founder", "CEO"],
                "organization_num_employees_ranges": ["1,50"],
                "per_page": 100,
            },
        },
    ]

    total_exported = 0

    for export in exports:
        filename, count = export_contacts(
            export["name"], export["params"], max_contacts=5000
        )
        total_exported += count
        time.sleep(2)  # Rate limiting between exports

    print(f"\n{'=' * 60}")
    print(f"TOTAL EXPORTED: {total_exported:,} contacts")
    print(f"{'=' * 60}")
    print("\nFiles created:")
    for export in exports:
        print(f"  - apollo_{export['name']}_*.csv")


if __name__ == "__main__":
    main()
