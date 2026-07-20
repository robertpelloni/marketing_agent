#!/usr/bin/env python3
"""
Apollo Bulk Export Script
Exports all contacts from Apollo lists to CSV using the API
"""

import requests
import csv
import time
from datetime import datetime

API_KEY = "Hk2hKzzYIddTsygrnoufDw"
BASE_URL = "https://api.apollo.io/api/v1"
HEADERS = {"Content-Type": "application/json", "X-Api-Key": API_KEY}


def get_saved_lists():
    """Get all saved lists from Apollo."""
    print("Fetching saved lists...")
    try:
        resp = requests.post(
            f"{BASE_URL}/mixed_lists/search",
            headers=HEADERS,
            json={"per_page": 50},
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        lists = data.get("lists", [])
        print(f"Found {len(lists)} lists")
        return lists
    except Exception as e:
        print(f"Error fetching lists: {e}")
        return []


def get_list_contacts(list_id, page=1, per_page=100):
    """Get contacts from a specific list."""
    try:
        resp = requests.post(
            f"{BASE_URL}/mixed_people/api_search",
            headers=HEADERS,
            json={"list_ids": [list_id], "page": page, "per_page": per_page},
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        return data.get("people", []), data.get("total_entries", 0)
    except Exception as e:
        print(f"Error fetching page {page}: {e}")
        return [], 0


def search_all_contacts(search_params, max_pages=100):
    """Search all contacts with pagination."""
    all_contacts = []
    page = 1

    while page <= max_pages:
        try:
            params = {**search_params, "page": page, "per_page": 100}
            resp = requests.post(
                f"{BASE_URL}/mixed_people/api_search",
                headers=HEADERS,
                json=params,
                timeout=30,
            )
            resp.raise_for_status()
            data = resp.json()

            people = data.get("people", [])
            total = data.get("total_entries", 0)

            if not people:
                break

            all_contacts.extend(people)
            print(
                f"  Page {page}: Got {len(people)} contacts (total so far: {len(all_contacts)}/{total})"
            )

            if len(all_contacts) >= total:
                break

            page += 1
            time.sleep(0.5)  # Rate limiting

        except Exception as e:
            print(f"Error on page {page}: {e}")
            break

    return all_contacts


def export_to_csv(contacts, filename):
    """Export contacts to CSV."""
    if not contacts:
        print("No contacts to export")
        return

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
        "organization_annual_revenue",
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
                "organization_annual_revenue": org.get("annual_revenue", ""),
            }
            writer.writerow(row)

    print(f"Exported {len(contacts)} contacts to {filename}")


def main():
    """Main export workflow."""
    print("=" * 60)
    print("Apollo Bulk Export Tool")
    print("=" * 60)
    print(f"Timestamp: {datetime.now().isoformat()}")
    print(f"API Key: {API_KEY[:10]}...")
    print()

    # First, get saved lists
    lists = get_saved_lists()

    if lists:
        print("\nYour saved lists:")
        for i, lst in enumerate(lists):
            print(
                f"  {i + 1}. {lst.get('name', 'Unknown')} (ID: {lst.get('id')}, Count: {lst.get('count', 'unknown')})"
            )

        print("\n" + "=" * 60)
        print("EXPORT OPTIONS:")
        print("=" * 60)
        print("1. Export a specific list by number")
        print("2. Export all contacts matching a search")
        print("3. Export all contacts (no filter)")
        print()

        # For automation, export all contacts
        choice = input("Enter choice (1-3) or press Enter for all: ").strip() or "3"

        if choice == "1":
            list_num = int(input("Enter list number: ")) - 1
            if 0 <= list_num < len(lists):
                selected_list = lists[list_num]
                list_id = selected_list.get("id")
                list_name = selected_list.get("name", "unknown")
                print(f"\nExporting list: {list_name}")

                all_contacts = []
                page = 1
                while True:
                    contacts, total = get_list_contacts(list_id, page)
                    if not contacts:
                        break
                    all_contacts.extend(contacts)
                    print(
                        f"  Page {page}: Got {len(contacts)} contacts (total: {len(all_contacts)}/{total})"
                    )
                    if len(all_contacts) >= total:
                        break
                    page += 1
                    time.sleep(0.5)

                filename = f"apollo_export_{list_name.replace(' ', '_')}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.csv"
                export_to_csv(all_contacts, filename)

        elif choice == "2":
            print("\nEnter search parameters (press Enter to skip):")
            keyword = input("  Keywords: ").strip()
            title = input("  Job Title: ").strip()

            search_params = {}
            if keyword:
                search_params["q_keywords"] = keyword
            if title:
                search_params["person_titles"] = [title]

            print(f"\nSearching with params: {search_params}")
            all_contacts = search_all_contacts(search_params)
            filename = (
                f"apollo_search_export_{datetime.now().strftime('%Y%m%d_%H%M%S')}.csv"
            )
            export_to_csv(all_contacts, filename)

        else:
            # Export all contacts
            print("\nExporting ALL contacts (this may take a while)...")
            all_contacts = search_all_contacts({})
            filename = (
                f"apollo_all_contacts_{datetime.now().strftime('%Y%m%d_%H%M%S')}.csv"
            )
            export_to_csv(all_contacts, filename)

    else:
        print("\nNo saved lists found. Exporting all contacts...")
        all_contacts = search_all_contacts({})
        filename = f"apollo_all_contacts_{datetime.now().strftime('%Y%m%d_%H%M%S')}.csv"
        export_to_csv(all_contacts, filename)

    print("\n" + "=" * 60)
    print("EXPORT COMPLETE")
    print("=" * 60)


if __name__ == "__main__":
    main()
