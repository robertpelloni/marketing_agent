# ROADMAP — marketing_agent

Source of truth for long-term direction. Short-term tasks live in `TODO.md`.
Verified against repository and Hetzner production (`5.161.250.43`) through `2026-11-14` (`v0.7.5`).

## M1 — Repo/prod parity lock (DONE 2026-11-07 / 2026-11-14)
- Trial-funnel pages, sitemap dates, outreach docs, and `/opt/marketing_agent` checkout synced cleanly to `main` across both GitHub (`github`) and GitLab (`origin`), with deployed binary `vcs.revision` locked to committed `HEAD` (`vcs.modified=false`).
- Rule: publish static site updates to `/var/www/hypernexus.site` and `/var/www/tormentnexus.site` only from committed `main`.

## M2 — Finalized Gate 2 Pricing & Checkout Alignment (CODE + STRIPE SESSIONS DONE `v0.7.5`)
- Copy aligned (`e5c43b8c`) and server-side checkout enforced (`v0.7.5`): one Stripe `subscription` session with `$1.99` one-time per-team 14-day trial fee today + `$5/seat/month` recurring (`TrialPeriodDays = 14`, seats clamped `1–5`) + optional `$200` Founding Team Rate Lock (`12` months), coupon mechanism removed, and client `price_amount` rejected with `HTTP 400`.
- Remaining gate: live operator card real-charge + webhook tenant provisioning check.

## M3 — Outreach blitz & autonomous multi-channel cadence (CORE ENGINE READY `v0.7.4`)
- Copy drafted for all 11 launch posts (`POSTING_GUIDE.md` + `marketing/outreach_queue_2026-11-03.md`); manual account-authenticated posts (LinkedIn, X, Hashnode, Medium) remain in `TODO.md`.
- Autonomous 5-step cadence engine hardened in `v0.7.4`: monotonic `TemplateID` progression (fixing Step 2/3 loop), delivery-before-persistence invariants, unauthenticated GitHub commit enrichment fallback for the 17,953+ discovered leads, and bounded `go-rod` browser lifecycle.

## M4 — Integration & CI suite green (DONE 2026-11-13, `v0.7.3`)
- All 5 GitHub Actions workflows green on `main`: `CI`, `Docker Build`, `Integration Tests`, `Deploy`, and `CodeQL`, guarded by `internal/db/schema_test.go`.

## M5 — Marketing bot & external provider credentials
- `scripts/marketing_bot` is the canonical social runner (legacy `auto_*`/`apollo_*` scripts archived under `scripts/archive/`).
- Dashboard Section `6A` and `/api/v1/metrics` (`enrichment_sources`) expose real-time provider status; rotating valid keys in `/opt/marketing_agent/.env` (`DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `APOLLO_API_KEY`, `HUNTER_API_KEY`, `GITHUB_TOKEN`, `DEVTO_API_KEY`, `BLUESKY_*`) immediately unlocks full-rate enrichment, LLM blog generation, and social auto-posting without code changes.

## M6 — Single-page Operations Hub & Observability (DONE `v0.7.3`–`v0.7.4`)
- `/` is the unified single-page operations hub (Subscription, KPIs, Cold Outreach, Active Deals, Social/DevRel, Pipeline Ops, Operator Tools `6D`, Provider Credentials `6A`); `/dashboard/outreach` 301-redirects to `/#outreach`; `/demo` serves the TormentNexus Cloud Console view.
- `/health`, `/health/detailed` (with stamped `version` and accurate DeepSeek/Hermes label), `/metrics` (Prometheus), and `/api/v1/metrics` (JSON pool + enrichment counters + `version`) provide headless production observability.

## Non-goals
- `checkout.html` in `sitemap.xml`: deliberately excluded (transactional page).
- Workstation server execution: prohibited per standing Hetzner-only policy (`start.bat` guard).
