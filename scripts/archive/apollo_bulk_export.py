#!/usr/bin/env python3
"""
Apollo Bulk Export Tool
Exports contacts from Apollo API to CSV with pagination
"""

import requests
import json
import csv
import time
from datetime import datetime

API_KEY = "Hk2hKzzYIddTsygrnoufDw"
BASE_URL = "https://api.apollo.io/api/v1/mixed_people/api_search"
HEADERS = {"Content-Type": "application/json", "X-Api-Key": API_KEY}


def search_contacts(params, max_contacts=50000):
    """Search Apollo contacts with pagination."""
    all_contacts = []
    page = 1
    total = 0

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
                f"  Page {page}: {len(people)} contacts (total: {len(all_contacts):,}/{total:,})"
            )

            if len(all_contacts) >= total or len(all_contacts) >= max_contacts:
                break

            page += 1
            time.sleep(0.3)  # Rate limiting

        except Exception as e:
            print(f"  Error on page {page}: {e}")
            time.sleep(1)
            continue

    return all_contacts, total


def export_to_csv(contacts, filename):
    """Export contacts to CSV."""
    if not contacts:
        print("No contacts to export")
        return 0

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
        "organization_num_employees",
    ]

    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()

        for person in contacts:
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
                "organization_num_employees": org.get("estimated_num_employees", ""),
            }
            writer.writerow(row)

    return len(contacts)


def main():
    """Main export function."""
    print("=" * 60)
    print("Apollo Bulk Export Tool")
    print("=" * 60)
    print(f"Timestamp: {datetime.now().isoformat()}")
    print(f"API Key: {API_KEY[:10]}...")
    print()

    # Define searches to run
    searches = [
        {
            "name": "CTOs at Software Companies",
            "params": {
                "person_titles": ["CTO", "Chief Technology Officer"],
                "q_keywords": "software technology",
                "per_page": 100,
            },
        },
        {
            "name": "VP Engineering",
            "params": {
                "person_titles": [
                    "VP Engineering",
                    "VP of Engineering",
                    "Head of Engineering",
                ],
                "per_page": 100,
            },
        },
        {
            "name": "AI/ML Engineers",
            "params": {
                "q_keywords": "AI machine learning artificial intelligence",
                "person_titles": ["Engineer", "Developer", "Architect"],
                "per_page": 100,
            },
        },
        {
            "name": "Startup Founders",
            "params": {
                "person_titles": ["Founder", "Co-Founder", "CEO"],
                "organization_num_employees_ranges": ["1,50", "51,200"],
                "per_page": 100,
            },
        },
    ]

    all_contacts = []

    for search_item in searches:
        print(f"\n{'=' * 60}")
        print(f"Search: {search_item['name']}")
        print(f"{'=' * 60}")

        contacts, total = search_contacts(search_item["params"], max_contacts=10000)
        all_contacts.extend(contacts)
        print(f"  Found: {len(contacts):,} (total available: {total:,})")

        time.sleep(1)  # Rate limiting between searches

    # Remove duplicates by ID
    seen_ids = set()
    unique_contacts = []
    for contact in all_contacts:
        cid = contact.get("id")
        if cid and cid not in seen_ids:
            seen_ids.add(cid)
            unique_contacts.append(contact)

    print(f"\n{'=' * 60}")
    print(f"TOTAL UNIQUE CONTACTS: {len(unique_contacts):,}")
    print(f"{'=' * 60}")

    # Export to CSV
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"apollo_export_{timestamp}.csv"

    count = export_to_csv(unique_contacts, filename)
    print(f"\nExported {count:,} contacts to {filename}")

    # Also save as JSON for further processing
    json_filename = f"apollo_export_{timestamp}.json"
    with open(json_filename, "w", encoding="utf-8") as f:
        json.dump(unique_contacts, f, indent=2)
    print(f"Also saved to {json_filename}")

    print(f"\n{'=' * 60}")
    print("EXPORT COMPLETE")
    print(f"{'=' * 60}")
    print(f"CSV file: {filename}")
    print(f"JSON file: {json_filename}")
    print(f"Total contacts: {count:,}")


if __name__ == "__main__":
    main()
