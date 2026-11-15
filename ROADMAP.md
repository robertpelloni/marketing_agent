# ROADMAP — marketing_agent

Source of truth for long-term direction. Short-term tasks live in `TODO.md`.
Verified against repository and Hetzner production (`5.161.250.43`) through `2026-11-15` (`v0.7.7`).

## M1 — Repo/prod parity lock (DONE 2026-11-07 / 2026-11-15)
- Trial-funnel pages, sitemap dates, outreach docs, and `/opt/marketing_agent` checkout synced cleanly to `main` across both GitHub (`github`) and GitLab (`origin`), with deployed binary `vcs.revision` locked to committed `HEAD` (`vcs.modified=false`).
- Rule: publish static site updates to `/var/www/hypernexus.site` and `/var/www/tormentnexus.site` only from committed `main`.

## M2 — Finalized Gate 2 Pricing & Checkout Alignment (LIVE VERIFIED `v0.7.5`–`v0.7.7`)
- Copy aligned (`e5c43b8c`) and server-side checkout enforced (`v0.7.5`–`v0.7.6`): one Stripe `subscription` session with `$1.99` one-time per-team 14-day trial fee today + `$5/seat/month` recurring (`TrialPeriodDays = 14`, seats clamped `1–5`) + optional `$200` Founding Team Rate Lock (`12` months), coupon mechanism removed, client `price_amount` rejected with `HTTP 400`, and `STRIPE_API_KEY` health surfaced in `/health/detailed` + Dashboard `6A`.
- Live `STRIPE_API_KEY` rotated on `2026-11-15`: `/health/detailed` reports `Stripe (live): Connected`, and all 4 live unpaid `cs_live_...` checkout entry flows passed independent verification. Only an optional real-card completed `$1.99` charge remains at operator discretion.

## M3 — Outreach blitz, lead hygiene & autonomous multi-channel cadence (ACTIVE IN PROD `v0.7.4`–`v0.7.7`)
- Copy drafted for all 11 launch posts (`POSTING_GUIDE.md` + `marketing/outreach_queue_2026-11-03.md`); manual account-authenticated posts (LinkedIn, X, Hashnode, Medium) remain in `TODO.md`.
- Autonomous 5-step cadence & enrichment pipeline hardened through `v0.7.7`: monotonic `TemplateID` progression, delivery-before-persistence invariants, shared AI-agent/bot/unroutable email filtering (`internal/db/contact_hygiene.go`), GitHub `403`/`429` rate-limit circuit tripping + FIFO enrichment queue rotation (`TouchDealUpdatedAt`), canonical `github.com/<owner>/<repo>` issue lead domains, cached RSS research (`30m` TTL + `15m` `429` backoff), and multi-signal intent scoring.

## M4 — Integration & CI suite green (DONE 2026-11-13, `v0.7.3`)
- All 5 GitHub Actions workflows green on `main`: `CI`, `Docker Build`, `Integration Tests`, `Deploy`, and `CodeQL`, guarded by `internal/db/schema_test.go`.

## M5 — Marketing bot & external provider credentials
- `scripts/marketing_bot` is the canonical social runner (legacy `auto_*`/`apollo_*` scripts archived under `scripts/archive/`).
- Dashboard Section `6A` and `/api/v1/metrics` (`enrichment_sources`) expose real-time provider status; rotating valid keys in `/opt/marketing_agent/.env` (`DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `APOLLO_API_KEY`, `HUNTER_API_KEY`, `GITHUB_TOKEN`, `DEVTO_API_KEY`, `BLUESKY_*`) immediately unlocks 5,000 req/hr GitHub enrichment, LLM blog generation, and social auto-posting without code changes.

## M6 — Single-page Operations Hub & Observability (DONE `v0.7.3`–`v0.7.6`)
- `/` is the unified single-page operations hub (Subscription, KPIs, Cold Outreach, Active Deals, Social/DevRel, Pipeline Ops, Operator Tools `6D`, Provider Credentials `6A`); `/dashboard/outreach` 301-redirects to `/#outreach`; `/demo` serves the TormentNexus Cloud Console view.
- `/health`, `/health/detailed` (with stamped `version`, live Stripe health, and accurate DeepSeek/Hermes label), `/metrics` (Prometheus), and `/api/v1/metrics` (JSON pool + accurate enrichment counters + `version`) provide headless production observability.

## Non-goals
- `checkout.html` in `sitemap.xml`: deliberately excluded (transactional page).
- Workstation server execution: prohibited per standing Hetzner-only policy (`start.bat` guard).
