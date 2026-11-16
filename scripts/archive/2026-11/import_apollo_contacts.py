#!/usr/bin/env python3
"""
Import Apollo CSV contacts into PostgreSQL marketing_agent database.
Links contacts to companies/deals by matching company domain.

Usage:
    python3 import_apollo_contacts.py /opt/marketing_agent/data/apollo_exports/latest_apollo_export.csv
"""

import csv
import sys
import psycopg2

DB_CONFIG = {
    "dbname": "sales_bot",
    "user": "sales_bot",
    "host": "/var/run/postgresql",
}


def normalize_domain(domain):
    """Normalize a domain for matching."""
    domain = domain.lower().strip()
    domain = domain.replace("http://", "").replace("https://", "")
    domain = domain.replace("www.", "")
    domain = domain.split("/")[0]
    return domain


def extract_domain_from_linkedin(linkedin_url):
    """Try to extract company domain from LinkedIn URL."""
    if not linkedin_url:
        return ""
    # LinkedIn company URLs are like /company/some-company
    if "/company/" in linkedin_url:
        parts = linkedin_url.split("/company/")
        if len(parts) > 1:
            company_slug = parts[1].split("/")[0].split("?")[0]
            return company_slug
    return ""


def import_contacts(csv_path):
    """Import contacts from CSV into PostgreSQL."""
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()

    imported = 0
    linked = 0
    skipped = 0

    with open(csv_path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)

        for row in reader:
            name = row.get("name", "").strip()
            email = row.get("email", "").strip()
            title = row.get("title", "").strip()
            company = row.get("company", "").strip()
            linkedin = row.get("linkedin", "").strip()

            if not email or not name:
                skipped += 1
                continue

            # Check if contact already exists
            cur.execute("SELECT id FROM contacts WHERE email = %s", (email,))
            existing = cur.fetchone()

            if existing:
                skipped += 1
                continue

            # Find or create company
            company_id = None
            if company:
                # Try to match by name
                cur.execute(
                    "SELECT id FROM companies WHERE LOWER(name) = LOWER(%s)", (company,)
                )
                company_row = cur.fetchone()

                if company_row:
                    company_id = company_row[0]
                else:
                    # Create new company
                    domain = (
                        normalize_domain(extract_domain_from_linkedin(linkedin)) or ""
                    )
                    cur.execute(
                        "INSERT INTO companies (name, domain) VALUES (%s, %s) RETURNING id",
                        (company, domain),
                    )
                    company_id = cur.fetchone()[0]

            # Insert contact
            cur.execute(
                """INSERT INTO contacts (company_id, name, role, email, linkedin_url)
                   VALUES (%s, %s, %s, %s, %s)
                   ON CONFLICT (email) DO NOTHING
                   RETURNING id""",
                (company_id, name, title, email, linkedin),
            )

            result = cur.fetchone()
            if result:
                imported += 1
                contact_id = result[0]

                # Create a deal if company exists and no deal yet
                if company_id:
                    cur.execute(
                        "SELECT id FROM deals WHERE company_id = %s LIMIT 1",
                        (company_id,),
                    )
                    deal = cur.fetchone()

                    if not deal:
                        cur.execute(
                            """INSERT INTO deals (company_id, current_state, cadence_step)
                               VALUES (%s, 'Researched', 0)""",
                            (company_id,),
                        )
                        linked += 1

    conn.commit()
    cur.close()
    conn.close()

    print(f"[+] Imported: {imported}, Linked deals: {linked}, Skipped: {skipped}")


def main():
    if len(sys.argv) < 2:
        # Default path
        csv_path = "/opt/marketing_agent/data/apollo_exports/latest_apollo_export.csv"
    else:
        csv_path = sys.argv[1]

    import_contacts(csv_path)


if __name__ == "__main__":
    main()
