#!/usr/bin/env python3
"""
Apollo People Search Scraper via Chrome DevTools Protocol
Searches Apollo website directly, bypassing the broken API.

Usage:
    python3 apollo_cdp_scraper.py --query "AI startups" --pages 5
    python3 apollo_cdp_scraper.py --company "openai.com" --pages 3
"""

import json
import time
import subprocess
import argparse
import csv
import os
from websocket import create_connection

CHROME_PATH = "/home/sales_bot/.cache/rod/browser/chromium-1321438/chrome"
CDP_PORT = 9222
OUTPUT_DIR = "/opt/marketing_agent/data/apollo_exports"
APOLLO_URL = "https://app.apollo.io/#/people"


def launch_chrome():
    """Launch Chrome with CDP enabled."""
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    proc = subprocess.Popen(
        [
            CHROME_PATH,
            "--headless=new",
            "--no-sandbox",
            "--disable-gpu",
            "--disable-dev-shm-usage",
            "--remote-debugging-port=" + str(CDP_PORT),
            "--user-data-dir=/tmp/apollo-scraper-profile",
            "--disable-background-networking",
            "--disable-sync",
            "--no-first-run",
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    time.sleep(3)
    return proc


def get_ws_url():
    """Get the WebSocket debugger URL for the first page."""
    import urllib.request

    resp = urllib.request.urlopen(f"http://127.0.0.1:{CDP_PORT}/json")
    targets = json.loads(resp.read())
    for t in targets:
        if t.get("type") == "page":
            return t["webSocketDebuggerUrl"]
    return None


def cdp_send(ws, method, params=None):
    """Send a CDP command and get result."""
    msg = {"id": 1, "method": method}
    if params:
        msg["params"] = params
    ws.send(json.dumps(msg))
    while True:
        resp = json.loads(ws.recv())
        if resp.get("id") == 1:
            return resp.get("result", {})


def navigate(ws, url):
    """Navigate to a URL."""
    cdp_send(ws, "Page.navigate", {"url": url})
    time.sleep(5)


def evaluate(ws, expression):
    """Evaluate JavaScript in the page."""
    result = cdp_send(
        ws,
        "Runtime.evaluate",
        {"expression": expression, "returnByValue": True, "awaitPromise": True},
    )
    return result.get("result", {}).get("value")


def wait_for_selector(ws, selector, timeout=30):
    """Wait for an element to appear."""
    start = time.time()
    while time.time() - start < timeout:
        result = evaluate(ws, f'document.querySelector("{selector}") !== null')
        if result:
            return True
        time.sleep(1)
    return False


def login_apollo(ws, email, password):
    """Login to Apollo if needed."""
    navigate(ws, "https://app.apollo.io/#/login")
    time.sleep(3)

    # Check if already logged in
    current_url = evaluate(ws, "window.location.href")
    if "people" in str(current_url).lower() or "dashboard" in str(current_url).lower():
        print("[+] Already logged in")
        return True

    print("[*] Logging in to Apollo...")
    evaluate(
        ws,
        f'''
        (function() {{
            var emailInput = document.querySelector('input[name="email"], input[type="email"], #email');
            var passInput = document.querySelector('input[name="password"], input[type="password"], #password');
            if (emailInput && passInput) {{
                emailInput.value = "{email}";
                emailInput.dispatchEvent(new Event('input', {{ bubbles: true }}));
                passInput.value = "{password}";
                passInput.dispatchEvent(new Event('input', {{ bubbles: true }}));
                setTimeout(function() {{
                    var btn = document.querySelector('button[type="submit"], .login-btn, button:has(span)');
                    if (btn) btn.click();
                }}, 500);
            }}
        }})()
    ''',
    )
    time.sleep(5)
    return True


def search_people(ws, query="", company_domain="", title="", pages=5):
    """Search Apollo people and extract contacts."""
    contacts = []

    # Build search URL with filters
    if company_domain:
        search_url = (
            f"https://app.apollo.io/#/people?contactCompanyUrl={company_domain}"
        )
    else:
        search_url = "https://app.apollo.io/#/people"

    navigate(ws, search_url)
    time.sleep(5)

    # If we have a query, type it into search
    if query:
        evaluate(
            ws,
            f'''
            (function() {{
                var searchInput = document.querySelector('input[placeholder*="search"], input[placeholder*="Search"], input[data-cy="search-input"]');
                if (searchInput) {{
                    searchInput.focus();
                    searchInput.value = "{query}";
                    searchInput.dispatchEvent(new Event('input', {{ bubbles: true }}));
                    searchInput.dispatchEvent(new Event('change', {{ bubbles: true }}));
                }}
            }})()
        ''',
        )
        time.sleep(3)

        # Click search button or press Enter
        evaluate(
            ws,
            """
            (function() {
                var btn = document.querySelector('button[type="submit"], [data-cy="search-button"]');
                if (btn) btn.click();
            })()
        """,
        )
        time.sleep(5)

    # Add title filter if specified
    if title:
        evaluate(
            ws,
            """
            (function() {
                // Try to find and click title filter
                var titleFilter = document.querySelector('[data-cy="title-filter"], .title-filter');
                if (titleFilter) titleFilter.click();
            })()
        """,
        )
        time.sleep(2)

    # Extract contacts from current page
    for page in range(pages):
        print(f"[*] Extracting page {page + 1}...")

        page_contacts = evaluate(
            ws,
            """
            (function() {
                var results = [];
                // Try multiple selectors for contact cards
                var rows = document.querySelectorAll('[data-cy="contact-row"], .people-table-row, tr.ant-table-row, .contact-card, [class*="PersonRow"], [class*="person-row"]');
                
                if (rows.length === 0) {
                    // Try table rows
                    rows = document.querySelectorAll('table tbody tr');
                }
                
                rows.forEach(function(row) {
                    try {
                        var nameEl = row.querySelector('[data-cy="contact-name"], .contact-name, a[href*="/people/"], td:nth-child(2) a, td:nth-child(1) a');
                        var titleEl = row.querySelector('[data-cy="contact-title"], .contact-title, td:nth-child(3), td:nth-child(4)');
                        var companyEl = row.querySelector('[data-cy="contact-company"], .contact-company, td:nth-child(4) a, td:nth-child(5) a');
                        var emailEl = row.querySelector('[data-cy="contact-email"], .contact-email, a[href*="mailto:"], td:nth-child(5)');
                        
                        var name = nameEl ? nameEl.textContent.trim() : '';
                        var title = titleEl ? titleEl.textContent.trim() : '';
                        var company = companyEl ? companyEl.textContent.trim() : '';
                        var email = emailEl ? emailEl.textContent.trim() : '';
                        
                        // Try to get email from data attribute
                        if (!email) {
                            var emailBtn = row.querySelector('[data-cy="reveal-email"], button:has(svg)');
                            if (emailBtn) {
                                email = emailBtn.getAttribute('data-email') || '';
                            }
                        }
                        
                        // Get LinkedIn URL
                        var linkedinEl = row.querySelector('a[href*="linkedin.com"]');
                        var linkedin = linkedinEl ? linkedinEl.href : '';
                        
                        if (name && (email || company)) {
                            results.push({
                                name: name,
                                title: title,
                                company: company,
                                email: email,
                                linkedin: linkedin
                            });
                        }
                    } catch(e) {}
                });
                
                return JSON.stringify(results);
            })()
        """,
        )

        if page_contacts:
            try:
                parsed = json.loads(page_contacts)
                contacts.extend(parsed)
                print(f"[+] Found {len(parsed)} contacts on page {page + 1}")
            except:
                pass

        # Click next page
        if page < pages - 1:
            evaluate(
                ws,
                """
                (function() {
                    var nextBtn = document.querySelector('[aria-label="Next"], .ant-pagination-next, button:has(.next-icon), li.ant-pagination-next:not(.ant-pagination-disabled) button');
                    if (nextBtn) nextBtn.click();
                })()
            """,
            )
            time.sleep(3)

    return contacts


def extract_visible_emails(ws):
    """Try to reveal hidden emails by clicking reveal buttons."""
    evaluate(
        ws,
        """
        (function() {
            var revealBtns = document.querySelectorAll('[data-cy="reveal-email"], button:contains("Reveal"), .reveal-email-btn');
            revealBtns.forEach(function(btn) { btn.click(); });
        })()
    """,
    )
    time.sleep(3)


def save_contacts(contacts, filename):
    """Save contacts to CSV and JSON."""
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    # Save CSV
    csv_path = os.path.join(OUTPUT_DIR, filename + ".csv")
    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(
            f, fieldnames=["name", "title", "company", "email", "linkedin"]
        )
        writer.writeheader()
        writer.writerows(contacts)

    # Save JSON
    json_path = os.path.join(OUTPUT_DIR, filename + ".json")
    with open(json_path, "w", encoding="utf-8") as f:
        json.dump(contacts, f, indent=2)

    print(f"[+] Saved {len(contacts)} contacts to {csv_path}")
    return csv_path


def main():
    parser = argparse.ArgumentParser(description="Apollo People Search Scraper via CDP")
    parser.add_argument("--query", default="", help="Search query")
    parser.add_argument("--company", default="", help="Company domain to search")
    parser.add_argument(
        "--title", default="", help='Job title filter (e.g., "CTO,VP,Director")'
    )
    parser.add_argument(
        "--pages", type=int, default=5, help="Number of pages to scrape"
    )
    parser.add_argument("--email", default="", help="Apollo login email")
    parser.add_argument("--password", default="", help="Apollo login password")
    parser.add_argument(
        "--output", default="apollo_contacts", help="Output filename prefix"
    )
    parser.add_argument("--cdp-port", type=int, default=9222, help="CDP port")
    args = parser.parse_args()

    global CDP_PORT
    CDP_PORT = args.cdp_port

    print("[*] Launching Chrome with CDP...")
    chrome_proc = launch_chrome()

    try:
        ws_url = get_ws_url()
        if not ws_url:
            print("[-] Could not connect to Chrome CDP")
            return

        print(f"[+] Connected to CDP: {ws_url}")
        ws = create_connection(ws_url)

        # Enable necessary CDP domains
        cdp_send(ws, "Page.enable")
        cdp_send(ws, "Runtime.enable")
        cdp_send(ws, "Network.enable")

        # Login if credentials provided
        if args.email and args.password:
            login_apollo(ws, args.email, args.password)
        else:
            print(
                "[*] No login credentials provided - using existing session or public access"
            )
            navigate(ws, "https://app.apollo.io/#/people")
            time.sleep(3)

        # Search and extract
        print(
            f"[*] Searching: query='{args.query}', company='{args.company}', pages={args.pages}"
        )
        contacts = search_people(
            ws,
            query=args.query,
            company_domain=args.company,
            title=args.title,
            pages=args.pages,
        )

        # Try to reveal emails
        if contacts:
            extract_visible_emails(ws)
            # Re-extract after revealing
            contacts2 = search_people(
                ws,
                query=args.query,
                company_domain=args.company,
                title=args.title,
                pages=1,
            )
            if contacts2:
                # Merge, preferring entries with emails
                seen = set()
                merged = []
                for c in contacts2 + contacts:
                    key = c.get("name", "") + c.get("company", "")
                    if key not in seen:
                        seen.add(key)
                        merged.append(c)
                contacts = merged

        if contacts:
            save_contacts(contacts, args.output)

            # Also save to combined export for import into marketing agent
            combined_path = os.path.join(OUTPUT_DIR, "latest_apollo_export.csv")
            save_contacts(contacts, "latest_apollo_export")
        else:
            print("[-] No contacts found")

        ws.close()

    finally:
        chrome_proc.terminate()
        chrome_proc.wait()
        print("[*] Chrome closed")


if __name__ == "__main__":
    main()
