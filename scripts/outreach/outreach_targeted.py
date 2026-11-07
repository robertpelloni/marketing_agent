#!/usr/bin/env python3
"""
outreach_targeted.py — Personalized CTO pitches for HyperNexus.

Sends highly personalized outreach to CTOs and technical decision-makers
from the contacts database. Uses role + company + skill targeting to customize the pitch.

Usage:
  python outreach_targeted.py --help
  python outreach_targeted.py --dry-run
  python outreach_targeted.py --limit 5
  python outreach_targeted.py --recent --skills mcp,rag
"""

import argparse
import logging
import sys

logger = logging.getLogger("outreach_targeted")
if not logger.handlers:
    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
    logger.addHandler(handler)
    logger.setLevel(logging.INFO)


def get_db_connection():
    """Import and obtain DB connection with informative errors."""
    try:
        from outreach_smtp import get_db
        return get_db()
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        raise


def get_cto_contacts(limit=None, recent_hours=None, skills=None):
    """Fetch CTO/technical decision-maker contacts from the DB, prioritizing latest enriched."""
    try:
        conn = get_db_connection()
    except Exception:
        return []

    cur = conn.cursor()

    query = """SELECT c.name, c.email, c.role, co.name AS company, co.tech_stack AS tech_stack
           FROM contacts c
           LEFT JOIN companies co ON c.company_id = co.id
           WHERE c.email IS NOT NULL
             AND c.email != ''
             AND (
               LOWER(c.role) LIKE '%%cto%%'
               OR LOWER(c.role) LIKE '%%chief technology%%'
               OR LOWER(c.role) LIKE '%%founder%%'
               OR LOWER(c.role) LIKE '%%head of engineering%%'
               OR LOWER(c.role) LIKE '%%vp engineering%%'
               OR LOWER(c.role) LIKE '%%architect%%'
             )"""
    params = []
    if recent_hours:
        query += " AND c.updated_at >= NOW() - (%s * INTERVAL '1 hour')"
        params.append(int(recent_hours))

    if skills:
        skill_clauses = []
        for s in skills:
            s_clean = s.strip().lower()
            if s_clean:
                skill_clauses.append("(LOWER(co.tech_stack) LIKE %s OR LOWER(c.role) LIKE %s)")
                params.extend([f"%{s_clean}%", f"%{s_clean}%"])
        if skill_clauses:
            query += " AND (" + " OR ".join(skill_clauses) + ")"

    query += " ORDER BY c.updated_at DESC, c.id DESC LIMIT %s"
    params.append(limit if limit else 1000)

    cur.execute(query, tuple(params))
    rows = cur.fetchall()
    cur.close()
    conn.close()
    return [
        {
            "name": r[0],
            "email": r[1],
            "role": r[2],
            "company": r[3] or "",
            "tech_stack": r[4] or "",
        }
        for r in rows
    ]


def subject_for(contact, skills=None):
    if skills and any(s.lower() in ("mcp", "tool") for s in skills):
        return f"HyperNexus for {contact['company']} — local MCP tool router & memory"
    if skills and any(s.lower() in ("rag", "vector") for s in skills):
        return f"HyperNexus for {contact['company']} — local RAG vector memory"
    return f"HyperNexus for {contact['company']} — local AI, no cloud"


def body_for(contact, skills=None):
    name = contact.get("name") or "there"
    role = contact.get("role") or "your team"
    company = contact.get("company") or "your company"

    bullet_point_1 = "  • Persistent L2 vector memory — shared across all AI tools"
    bullet_point_2 = "  • Progressive tool routing — up to 60% token savings"

    if skills and any(s.lower() == "mcp" for s in skills):
        bullet_point_2 = "  • Dynamic MCP server aggregation — cross-harness tool parity with Claude & Codex"
    elif skills and any(s.lower() in ("rag", "vector") for s in skills):
        bullet_point_1 = "  • Local GraphRAG memory vault — sub-millisecond retrieval without cloud leakage"

    return (
        f"Hi {name},\n\n"
        f"As {role} at {company}, you probably deal with the same problem "
        f"we set out to solve: AI tools that send your data to the cloud.\n\n"
        f"HyperNexus is a local-first AI control plane that runs entirely "
        f"on your infrastructure:\n\n"
        f"{bullet_point_1}\n"
        f"{bullet_point_2}\n"
        f"  • Model-agnostic — works with Ollama, vLLM, llama.cpp\n"
        f"  • Zero-downtime self-hosted deployment\n\n"
        f"For a team like {company}, that means full control over your AI "
        f"stack with no vendor lock-in and no data leaving your perimeter.\n\n"
        f"Would a 15-minute demo be useful? https://hypernexus.site\n\n"
        f"Best,\n"
        f"Robert Pelloni\n"
        f"Founder, HyperNexus\n"
        f"https://hypernexus.site"
    )


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description="Targeted outreach campaigns for CTOs and technical decision-makers."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview contacts and messages without sending.",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=None,
        help="Maximum number of contacts to target.",
    )
    parser.add_argument(
        "--recent",
        action="store_true",
        help="Filter to contacts updated or enriched within the last 24 hours.",
    )
    parser.add_argument(
        "--hours",
        type=int,
        default=None,
        help="Filter to contacts updated or enriched within the last N hours.",
    )
    parser.add_argument(
        "--skills",
        type=str,
        default=None,
        help="Comma-separated skill/tech keywords to target (e.g. mcp,rag,agent,vector).",
    )
    return parser.parse_args(argv)


def main():
    args = parse_args()
    recent_hours = 24 if args.recent else args.hours
    skills = [s.strip() for s in args.skills.split(",")] if args.skills else None

    contacts = get_cto_contacts(limit=args.limit, recent_hours=recent_hours, skills=skills)
    print(f"Found {len(contacts)} CTO/technical contacts matching criteria")

    if args.dry_run:
        for c in contacts:
            subject = subject_for(c, skills=skills)
            print(f"\n--- Preview for {c['name']} <{c['email']}> ({c['role']} @ {c['company']}) ---")
            print(f"Subject: {subject}")
            print(body_for(c, skills=skills))
        return

    try:
        from outreach_smtp import send_email
    except ImportError:
        logger.error("outreach_smtp not available for actual sending")
        return

    sent = 0
    for c in contacts:
        subject = subject_for(c, skills=skills)
        body = body_for(c, skills=skills)
        ok = send_email(
            c["email"],
            subject,
            body,
            company_name=c["company"],
            category="targeted-cto",
        )
        if ok:
            sent += 1

    print(f"Targeted outreach complete: sent {sent}/{len(contacts)}")


if __name__ == "__main__":
    main()
