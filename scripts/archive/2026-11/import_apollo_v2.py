#!/usr/bin/env python3
"""Import Apollo CSV contacts into PostgreSQL."""

import csv
import sys
import psycopg2

DB_CONFIG = {
    "dbname": "sales_bot",
    "user": "sales_bot",
    "password": "tormentnexus2026",
    "host": "localhost",
}


def import_contacts(csv_path):
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()

    imported = 0
    skipped = 0
    deals_created = 0

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
            if cur.fetchone():
                skipped += 1
                continue

            # Find or create company
            company_id = None
            if company:
                cur.execute(
                    "SELECT id FROM companies WHERE LOWER(name) = LOWER(%s) LIMIT 1",
                    (company,),
                )
                company_row = cur.fetchone()

                if company_row:
                    company_id = company_row[0]
                else:
                    # Use company name slug as domain to avoid empty domain uniqueness issue
                    domain = (
                        company.lower().replace(" ", "-").replace(".", "") + ".local"
                    )
                    try:
                        cur.execute(
                            "INSERT INTO companies (name, domain) VALUES (%s, %s) RETURNING id",
                            (company, domain),
                        )
                        company_id = cur.fetchone()[0]
                    except psycopg2.errors.UniqueViolation:
                        conn.rollback()
                        cur.execute(
                            "SELECT id FROM companies WHERE LOWER(name) = LOWER(%s) LIMIT 1",
                            (company,),
                        )
                        company_row = cur.fetchone()
                        if company_row:
                            company_id = company_row[0]

            # Insert contact
            try:
                cur.execute(
                    """INSERT INTO contacts (company_id, name, role, email, linkedin_url)
                       VALUES (%s, %s, %s, %s, %s)
                       RETURNING id""",
                    (company_id, name, title, email, linkedin),
                )
                contact_id = cur.fetchone()[0]
                imported += 1

                # Create deal if company exists and no deal yet
                if company_id:
                    cur.execute(
                        "SELECT id FROM deals WHERE company_id = %s LIMIT 1",
                        (company_id,),
                    )
                    if not cur.fetchone():
                        cur.execute(
                            "INSERT INTO deals (company_id, current_state, cadence_step) VALUES (%s, 'Researched', 0)",
                            (company_id,),
                        )
                        deals_created += 1
            except Exception:
                skipped += 1
                conn.rollback()
                continue

    conn.commit()
    cur.close()
    conn.close()

    print(
        f"[+] Imported: {imported}, Deals created: {deals_created}, Skipped: {skipped}"
    )


def main():
    csv_path = (
        sys.argv[1]
        if len(sys.argv) > 1
        else "/opt/marketing_agent/combined_export_20260720_183403.csv"
    )
    import_contacts(csv_path)


if __name__ == "__main__":
    main()
