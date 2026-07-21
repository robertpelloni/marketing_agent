#!/usr/bin/env python3
"""
Gmail Email Sender for TormentNexus
Sends personalized cold emails via Gmail SMTP
"""

import smtplib
import csv
import time
import json
import os
import sys
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

# Gmail SMTP settings
SMTP_HOST = "smtp.gmail.com"
SMTP_PORT = 587
SMTP_USER = "pelloni.robert@gmail.com"
SMTP_PASS = "jnfy pwri cyam fuul"

# Email settings
FROM_NAME = "Robert Pelloni"
FROM_EMAIL = "pelloni.robert@gmail.com"
DAILY_LIMIT = 200
DELAY_BETWEEN_EMAILS = 30

# Email templates
TEMPLATES = {
    "initial": {
        "subject": "Quick question about AI at {company}",
        "body": """Hi {first_name},

I noticed you're {title} at {company}. Quick question:

Are your AI agents overwhelmed by 50K-token tool dumps?

We built TormentNexus to solve exactly this. Our progressive MCP tool routing injects only the 3 most relevant tools per request instead of dumping everything into context.

Result: 10x faster inference, 90% fewer hallucinations.

Worth a look? https://hypernexus.site

Best,
Robert Pelloni
HyperNexus"""
    },
    "followup": {
        "subject": "Re: Quick question about AI at {company}",
        "body": """Hi {first_name},

Following up on my last email. Are you maintaining separate tool configs for Claude Code, Cursor, Codex, and Gemini CLI?

TormentNexus provides byte-for-byte identical tool signatures across all major AI coding harnesses. One config, universal parity.

Plus: LLM waterfall failover (OpenAI -> OpenRouter -> local Ollama on any 429/5xx error).

Demo: https://hypernexus.site

Best,
Robert Pelloni"""
    }
}

LOG_FILE = "email_send_log.csv"
SENT_FILE = "sent_emails.json"

def load_sent_emails():
    if os.path.exists(SENT_FILE):
        with open(SENT_FILE, "r") as f:
            return json.load(f)
    return {}

def save_sent_emails(sent):
    with open(SENT_FILE, "w") as f:
        json.dump(sent, f, indent=2)

def load_contacts(filename):
    contacts = []
    with open(filename, "r", encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        for row in reader:
            if row.get("email"):
                contacts.append(row)
    return contacts

def send_email(to_email, subject, body):
    try:
        msg = MIMEMultipart()
        msg["From"] = f"{FROM_NAME} <{FROM_EMAIL}>"
        msg["To"] = to_email
        msg["Subject"] = subject
        msg.attach(MIMEText(body, "plain"))
        
        with smtplib.SMTP(SMTP_HOST, SMTP_PORT) as server:
            server.starttls()
            server.login(SMTP_USER, SMTP_PASS)
            server.send_message(msg)
        
        return True
    except Exception as e:
        print(f"  Error: {e}")
        return False

def log_send(email, status, template):
    file_exists = os.path.exists(LOG_FILE)
    with open(LOG_FILE, "a", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        if not file_exists:
            writer.writerow(["timestamp", "email", "status", "template"])
        writer.writerow([datetime.now().isoformat(), email, status, template])

def main():
    print("="*60)
    print("Gmail Email Sender for TormentNexus")
    print("="*60)
    
    sent = load_sent_emails()
    print(f"Already sent: {len(sent)} emails")
    
    contact_files = [
        "combined_export_20260720_183403.csv",
        "hunter_expanded_export_20260720_182551.csv",
        "hunter_target_companies_20260720_171159.csv",
        "apollo_contacts_export_20260720_183217.csv"
    ]
    
    all_contacts = []
    for cf in contact_files:
        if os.path.exists(cf):
            contacts = load_contacts(cf)
            all_contacts.extend(contacts)
            print(f"Loaded {len(contacts)} from {cf}")
    
    unique_contacts = {}
    for c in all_contacts:
        email = (c.get("email") or "").lower().strip()
        if email and email not in unique_contacts and email not in sent:
            unique_contacts[email] = c
    
    print(f"Unique unsent contacts: {len(unique_contacts)}")
    
    sent_today = 0
    template = "initial"
    
    for email, contact in list(unique_contacts.items())[:DAILY_LIMIT]:
        if sent_today >= DAILY_LIMIT:
            print("\nReached daily limit")
            break
        
        first_name = contact.get("first_name", "there")
        title = contact.get("title", "your role")
        company = contact.get("company", "your company")
        
        subject = TEMPLATES[template]["subject"].format(
            first_name=first_name, title=title, company=company
        )
        body = TEMPLATES[template]["body"].format(
            first_name=first_name, title=title, company=company
        )
        
        print(f"[{sent_today+1}] {email} ({first_name} at {company})...")
        success = send_email(email, subject, body)
        
        if success:
            sent[email] = {"sent_at": datetime.now().isoformat(), "template": template}
            save_sent_emails(sent)
            log_send(email, "sent", template)
            sent_today += 1
            print("  Sent")
        else:
            log_send(email, "failed", template)
        
        time.sleep(DELAY_BETWEEN_EMAILS)
    
    print(f"\nSent today: {sent_today}")
    print(f"Total sent: {len(sent)}")

if __name__ == "__main__":
    main()
