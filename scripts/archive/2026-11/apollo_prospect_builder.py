#!/usr/bin/env python3
"""
Apollo + Hunter Prospector for TormentNexus
Builds targeted prospect lists using Apollo.io and Hunter.io APIs
"""

import os
import csv
import requests
import time
from datetime import datetime

# Load environment
APOLLO_API_KEY = os.getenv("APOLLO_API_KEY", "Hk2hKzzYIddTsygrnoufDw")
HUNTER_API_KEY = os.getenv("HUNTER_API_KEY", "6d2e42cff14189b3c2453e8e4155efa27e1f02b2")

# Target personas
TARGETS = {
    "cto_ai_startups": {
        "title": ["CTO", "Chief Technology Officer"],
        "seniority": ["c_suite"],
        "keywords": "AI LLM machine learning",
        "company_size": ["11,50", "51,200"],
    },
    "vp_engineering": {
        "title": ["VP Engineering", "VP of Engineering", "Engineering VP"],
        "seniority": ["vp"],
        "keywords": "AI developer tools",
        "company_size": ["51,200", "201,500"],
    },
    "ai_leads": {
        "title": ["AI Lead", "ML Lead", "Machine Learning Lead", "AI Manager"],
        "seniority": ["director"],
        "keywords": "LLM AI agents",
        "company_size": ["51,200", "201,500"],
    },
    "platform_engineers": {
        "title": ["Platform Engineer", "DevOps Engineer", "Infrastructure Engineer"],
        "seniority": ["senior"],
        "keywords": "developer tools AI",
        "company_size": ["51,200", "201,500"],
    },
}

# Company lists to search
TARGET_COMPANIES = [
    "cursor.sh",
    "continue.dev",
    "langchain.dev",
    "crewai.com",
    "openrouter.ai",
    "together.ai",
    "anyscale.com",
    "modal.com",
    "replicate.com",
    "huggingface.co",
]


def search_apollo(target_key: str, page: int = 1) -> list:
    """Search Apollo for people matching target persona."""
    target = TARGETS[target_key]

    payload = {
        "api_key": APOLLO_API_KEY,
        "q_keywords": target["keywords"],
        "person_titles": target["title"],
        "person_seniorities": target["seniority"],
        "organization_num_employees_ranges": target["company_size"],
        "page": page,
        "per_page": 25,
    }

    try:
        resp = requests.post(
            "https://api.apollo.io/api/v1/people/search",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()

        people = data.get("people", [])
        print(f"  Found {len(people)} contacts (page {page})")
        return people
    except Exception as e:
        print(f"  Error: {e}")
        return []


def search_hunter(domain: str) -> list:
    """Search Hunter for emails at a domain."""
    try:
        resp = requests.get(
            "https://api.hunter.io/v2/domain-search",
            params={
                "domain": domain,
                "api_key": HUNTER_API_KEY,
                "limit": 10,
                "type": "personal",
            },
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()

        emails = data.get("data", {}).get("emails", [])
        print(f"  Found {len(emails)} emails at {domain}")
        return emails
    except Exception as e:
        print(f"  Error: {e}")
        return []


def verify_email(email: str) -> dict:
    """Verify email deliverability with Hunter."""
    try:
        resp = requests.get(
            "https://api.hunter.io/v2/email-verifier",
            params={"email": email, "api_key": HUNTER_API_KEY},
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        return data.get("data", {})
    except Exception as e:
        print(f"  Verify error for {email}: {e}")
        return {}


def build_prospect_list(target_key: str, max_pages: int = 3) -> list:
    """Build a prospect list for a target persona."""
    print(f"\n{'=' * 60}")
    print(f"Building list: {target_key}")
    print(f"{'=' * 60}")

    all_people = []

    for page in range(1, max_pages + 1):
        people = search_apollo(target_key, page)
        all_people.extend(people)

        if len(people) < 25:
            break

        time.sleep(1)  # Rate limiting

    print(f"\nTotal contacts found: {len(all_people)}")
    return all_people


def enrich_with_hunter(people: list) -> list:
    """Enrich Apollo contacts with Hunter email verification."""
    print(f"\n{'=' * 60}")
    print("Enriching with Hunter verification")
    print(f"{'=' * 60}")

    enriched = []

    for person in people:
        email = person.get("email")
        if not email:
            continue

        verification = verify_email(email)
        status = verification.get("result", "unknown")
        score = verification.get("score", 0)

        # Only keep valid or accept_all emails
        if status in ["valid", "accept_all"] and score >= 80:
            person["hunter_status"] = status
            person["hunter_score"] = score
            enriched.append(person)
            print(f"  ✓ {email} ({status}, {score}%)")
        else:
            print(f"  ✗ {email} ({status}, {score}%)")

        time.sleep(0.5)  # Rate limiting

    print(f"\nValid emails: {len(enriched)}/{len(people)}")
    return enriched


def export_to_csv(prospects: list, filename: str):
    """Export prospects to CSV."""
    if not prospects:
        print("No prospects to export")
        return

    fieldnames = [
        "name",
        "title",
        "email",
        "linkedin_url",
        "organization.name",
        "organization.website_url",
        "hunter_status",
        "hunter_score",
    ]

    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()

        for person in prospects:
            # Flatten organization
            org = person.get("organization", {})
            person["organization.name"] = org.get("name", "")
            person["organization.website_url"] = org.get("website_url", "")
            writer.writerow(person)

    print(f"Exported {len(prospects)} prospects to {filename}")


def export_to_apollo_sequence(prospects: list, sequence_name: str):
    """Export prospects formatted for Apollo sequence import."""
    filename = f"{sequence_name}_apollo_import.csv"

    fieldnames = [
        "First Name",
        "Last Name",
        "Title",
        "Email",
        "Company",
        "Website",
        "LinkedIn URL",
    ]

    with open(filename, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()

        for person in prospects:
            name_parts = person.get("name", "").split(" ", 1)
            first_name = name_parts[0] if name_parts else ""
            last_name = name_parts[1] if len(name_parts) > 1 else ""

            org = person.get("organization", {})

            writer.writerow(
                {
                    "First Name": first_name,
                    "Last Name": last_name,
                    "Title": person.get("title", ""),
                    "Email": person.get("email", ""),
                    "Company": org.get("name", ""),
                    "Website": org.get("website_url", ""),
                    "LinkedIn URL": person.get("linkedin_url", ""),
                }
            )

    print(f"Exported {len(prospects)} prospects to {filename}")
    print("Import this file into Apollo → Sequences → Import")


def main():
    """Main prospecting workflow."""
    print("=" * 60)
    print("TormentNexus Apollo + Hunter Prospector")
    print("=" * 60)
    print(f"Timestamp: {datetime.now().isoformat()}")
    print(f"Apollo API Key: {APOLLO_API_KEY[:10]}...")
    print(f"Hunter API Key: {HUNTER_API_KEY[:10]}...")

    all_prospects = []

    # Build lists for each target persona
    for target_key in TARGETS:
        people = build_prospect_list(target_key, max_pages=2)
        all_prospects.extend(people)
        time.sleep(2)  # Rate limiting between searches

    # Also search Hunter for target company domains
    print(f"\n{'=' * 60}")
    print("Searching Hunter for target company domains")
    print(f"{'=' * 60}")

    hunter_emails = []
    for domain in TARGET_COMPANIES:
        emails = search_hunter(domain)
        hunter_emails.extend(emails)
        time.sleep(1)

    # Enrich Apollo contacts with Hunter verification
    if all_prospects:
        enriched = enrich_with_hunter(all_prospects[:50])  # Limit for rate limiting

        # Export results
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        export_to_csv(enriched, f"prospects_{timestamp}.csv")
        export_to_apollo_sequence(enriched, f"tormentnexus_{timestamp}")

    # Summary
    print(f"\n{'=' * 60}")
    print("PROSPECTING COMPLETE")
    print(f"{'=' * 60}")
    print(f"Total Apollo contacts: {len(all_prospects)}")
    print(f"Total Hunter emails: {len(hunter_emails)}")
    print(f"Valid prospects: {len(enriched) if all_prospects else 0}")
    print("\nNext steps:")
    print("1. Review exported CSV files")
    print("2. Import into Apollo sequences")
    print("3. Launch outreach campaigns")
    print("4. Track open/reply rates")


if __name__ == "__main__":
    main()
