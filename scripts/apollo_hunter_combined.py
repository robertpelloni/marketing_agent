#!/usr/bin/env python3
"""
Apollo + Hunter Combined Export
1. Get contacts from Apollo (names, titles, companies)
2. Find emails using Hunter (domain search)
3. Combine into one CSV
"""

import requests
import csv
import time
from datetime import datetime

APOLLO_KEY = "Hk2hKzzYIddTsygrnoufDw"
HUNTER_KEY = "6d2e42cff14189b3c2453e8e4155efa27e1f02b2"

APOLLO_URL = "https://api.apollo.io/api/v1/mixed_people/api_search"
HUNTER_URL = "https://api.hunter.io/v2/domain-search"

HEADERS = {"Content-Type": "application/json", "X-Api-Key": APOLLO_KEY}


def get_apollo_contacts(params, max_contacts=1000):
    """Get contacts from Apollo API."""
    print(f"\n{'=' * 60}")
    print("Step 1: Getting contacts from Apollo")
    print(f"{'=' * 60}")

    all_contacts = []
    page = 1

    while len(all_contacts) < max_contacts:
        try:
            search_params = {**params, "page": page, "per_page": 100}
            resp = requests.post(
                APOLLO_URL, json=search_params, headers=HEADERS, timeout=30
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

    print(f"  Total Apollo contacts: {len(all_contacts):,}")
    return all_contacts


def get_hunter_emails(domain, limit=10):
    """Get emails from Hunter for a domain."""
    try:
        resp = requests.get(
            HUNTER_URL,
            params={
                "domain": domain,
                "api_key": HUNTER_KEY,
                "limit": limit,
                "type": "personal",
            },
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        return data.get("data", {}).get("emails", [])
    except Exception:
        return []


def extract_domain(website):
    """Extract domain from website URL."""
    if not website:
        return None
    website = website.replace("http://", "").replace("https://", "").replace("www.", "")
    return website.split("/")[0].lower()


def enrich_with_hunter(contacts, max_domains=100):
    """Add Hunter emails to Apollo contacts."""
    print(f"\n{'=' * 60}")
    print("Step 2: Finding emails with Hunter")
    print(f"{'=' * 60}")

    # Group contacts by company domain
    domains = {}
    for contact in contacts:
        org = contact.get("organization", {})
        website = org.get("website_url", "")
        domain = extract_domain(website)

        if domain and domain not in domains:
            domains[domain] = []
        if domain:
            domains[domain].append(contact)

    print(f"  Found {len(domains)} unique domains")

    # Get emails for each domain
    enriched = []
    domain_count = 0

    for domain, domain_contacts in domains.items():
        if domain_count >= max_domains:
            break

        emails = get_hunter_emails(domain, limit=10)

        if emails:
            # Match emails to contacts by title/name
            for contact in domain_contacts:
                contact_title = (contact.get("title", "") or "").lower()
                contact_name = (contact.get("first_name", "") or "").lower()

                best_match = None
                best_score = 0

                for email_info in emails:
                    email_title = (email_info.get("position", "") or "").lower()
                    email_first = (email_info.get("first_name", "") or "").lower()

                    # Score based on title match
                    score = 0
                    if contact_title and email_title:
                        if any(word in email_title for word in contact_title.split()):
                            score += 50

                    # Score based on name match
                    if contact_name and email_first:
                        if contact_name == email_first:
                            score += 30

                    # Use confidence as base score
                    score += email_info.get("confidence", 0) * 0.2

                    if score > best_score:
                        best_score = score
                        best_match = email_info

                if best_match and best_score > 20:
                    contact["hunter_email"] = best_match.get("value", "")
                    contact["hunter_confidence"] = best_match.get("confidence", 0)
                    enriched.append(contact)

        domain_count += 1
        if domain_count % 10 == 0:
            print(f"  Processed {domain_count} domains...")
        time.sleep(0.3)  # Rate limiting

    print(f"  Enriched {len(enriched)} contacts with emails")
    return enriched


def export_to_csv(contacts, filename):
    """Export contacts to CSV."""
    if not contacts:
        print("No contacts to export")
        return 0

    fieldnames = [
        "first_name",
        "last_name",
        "title",
        "email",
        "company",
        "website",
        "linkedin_url",
        "hunter_email",
        "hunter_confidence",
    ]

    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()

        for contact in contacts:
            org = contact.get("organization", {})
            writer.writerow(
                {
                    "first_name": contact.get("first_name", ""),
                    "last_name": contact.get(
                        "last_name", contact.get("last_name_obfuscated", "")
                    ),
                    "title": contact.get("title", ""),
                    "email": contact.get("hunter_email", ""),
                    "company": org.get("name", ""),
                    "website": org.get("website_url", ""),
                    "linkedin_url": contact.get("linkedin_url", ""),
                    "hunter_email": contact.get("hunter_email", ""),
                    "hunter_confidence": contact.get("hunter_confidence", 0),
                }
            )

    return len(contacts)


def main():
    """Main workflow."""
    print("=" * 60)
    print("Apollo + Hunter Combined Export")
    print("=" * 60)
    print(f"Apollo API: {APOLLO_KEY[:10]}...")
    print(f"Hunter API: {HUNTER_KEY[:10]}...")

    # Step 1: Get contacts from Apollo
    apollo_contacts = get_apollo_contacts(
        params={"person_titles": ["CTO", "VP Engineering"], "per_page": 100},
        max_contacts=500,
    )

    # Step 2: Enrich with Hunter emails
    enriched_contacts = enrich_with_hunter(apollo_contacts, max_domains=50)

    # Step 3: Export to CSV
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"apollo_hunter_export_{timestamp}.csv"

    count = export_to_csv(enriched_contacts, filename)

    print(f"\n{'=' * 60}")
    print("EXPORT COMPLETE")
    print(f"{'=' * 60}")
    print(f"Total contacts: {len(apollo_contacts):,}")
    print(f"With emails: {count:,}")
    print(f"File: {filename}")

    # Show sample
    print("\nSample contacts:")
    for contact in enriched_contacts[:5]:
        name = f"{contact.get('first_name', '')} {contact.get('last_name', '')}"
        title = contact.get("title", "")
        email = contact.get("hunter_email", "")
        print(f"  {name} | {title} | {email}")


if __name__ == "__main__":
    main()
