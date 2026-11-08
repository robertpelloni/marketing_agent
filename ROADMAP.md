# ROADMAP — marketing_agent

Source of truth for long-term direction. Short-term tasks live in TODO.md.
Verified against repo + production (5.161.250.43) on 2026-11-07.

## M1 — Repo/prod parity lock (DONE 2026-11-07)
- Trial-funnel pages, sitemap dates, outreach docs merged main <- backup/wip-leader-sync-20261107, pushed.
- Production already served these files; repo now matches. Rule: scp to /var/www/hypernexus.site only from committed main.

## M2 — $1.99 trial verification
- Live click-test: homepage button -> /checkout.html -> Stripe shows $1.99 one-time.
- Decide: keep one-time $1.99 (current) vs $1.99/mo recurring (one-line: interval month).
- Backend already supports ad-hoc price_amount (no new Stripe Price ID needed).

## M3 — Outreach blitz
- Copy drafted for all 11 posts. Remaining is manual: LinkedIn, X, Hashnode, Medium (account auth required).
- Track in POSTING_GUIDE.md sync checklist + marketing/outreach_queue_2026-11-03.md.

## M4 — Integration suite green
- GitHub "Integration Tests" red since 2026-07-11 (needs real Postgres). Unit tests + build pass; service runs in prod.
- Quarantine or fix the integration workflow so the gate means something again.

## M5 — Marketing bot autonomy
- scripts/marketing_bot is the canonical runner (per AGENTS.md). Legacy auto_*/apollo_* one-offs in scripts/ to be archived (see TODO).
- Missing secrets (DEVTO_API_KEY, BLUESKY_*, etc.) block any automated posting. No auto-post until keys exist.

## M6 — Dashboard (no action)
- internal/web serves 3 pages: / dashboard, /dashboard/outreach, /demo + /health. Already consolidated; the "merge into one page" ask is satisfied by current shape. Do not rewrite server.go without tests.

## Non-goals
- System tray icon: N/A (headless Linux systemd service + static sites, no desktop component).
- checkout.html in sitemap: deliberately excluded (transactional page).