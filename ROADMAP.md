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

## M6 - Dashboard (DONE 2026-11-13, v0.7.3)
- Merged feature/v0.7.1-consolidated-dashboard into main. / is the single operations page: six high-value-first sections (Subscription, KPIs, Outreach, Deals, Social, Pipeline Ops) with quick-nav anchors and tooltips. /dashboard/outreach 301-redirects to /#outreach (legacy page archived under web/dashboard/archive/). /demo stays as the separate TormentNexus Cloud Console subproject view; /health, /health/detailed, /metrics, and /api/v1/metrics are linked from the nav.
- System tray (Windows desktop runs): cmd/marketing_agent/systray.go provides Open Dashboard, Outreach, both sites, Blog, Restart Service, and Quit Servers & Exit. On Linux the service is headless under systemd (systray_unix.go blocks without a tray).

## Non-goals
- checkout.html in sitemap: deliberately excluded (transactional page).
