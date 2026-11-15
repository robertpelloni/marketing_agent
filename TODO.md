# TODO — marketing_agent

Immediate, granular tasks. Long-term direction lives in `ROADMAP.md`.

## Now (Operator / Account Auth Required)
- [ ] Refresh rotated/revoked API keys in prod `/opt/marketing_agent/.env` (and workstation env): `GITHUB_TOKEN`, `DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `APOLLO_API_KEY`, `HUNTER_API_KEY` (unauthenticated GitHub fallback + template outreach already active in `v0.7.4` meanwhile)
- [ ] Live real-charge card test: `$1.99` per-team 14-day trial today -> `$5/seat/mo` day 15 recurring on `https://hypernexus.site/checkout.html` (server-side Stripe session + line items + 14-day trial verified in `v0.7.5`; operator real card charge remaining)
- [ ] Decide on Bobbi's two optional site flags: `live.html` trial CTA target (`/contact.php` vs `/checkout.html`) and archiving/updating repo-only `$19.99` pages (`cloud_index_b`, `cloud_login_a/b`, `testimonials.html`)
- [ ] Manual posts: LinkedIn (6) / X (tweets in `POSTING_GUIDE.md`) / Hashnode / Medium; tick sync checklist
- [ ] Marketing bot secrets: `DEVTO_API_KEY`, `BLUESKY_TN_PASSWORD`, `BLUESKY_HN_PASSWORD`, funded X API credits
- [ ] Resubmit sitemap in Google Search Console after `lastmod` corrections

## Completed in v0.7.5 (2026-11-14)
- [x] Sync Bobbi's live pricing copy commit `e5c43b8c` cleanly into `main` without losing a single word of copy
- [x] Implement server-side `CreateHyperNexusTrialCheckout` (`$1.99` one-time team trial today + `$5/seat/month` recurring with `TrialPeriodDays = 14` + optional `$200` Founding Team Rate Lock line item, seats clamped `1–5`, `$3.01` coupon removed, founding grandfathering persisted on `checkout.session.completed`)
- [x] Close `$0.01` forged-amount hole (`HTTP 400 Bad Request` on any client-supplied `price_amount` / `first_month_amount`, plus return-URL host allowlist) and unify homepage, pricing, and checkout buttons on plan fields (`internal/billing/billing_test.go`, `internal/web/checkout_test.go`)
- [x] Commit cleanly BEFORE building Linux binary so deployed `vcs.revision` matches `git rev-parse HEAD` (`vcs.modified=false`), deploy `v0.7.5` + updated HTML to Hetzner `5.161.250.43`, and verify unpaid Stripe test sessions across all entry points

## Completed in v0.7.4 (2026-11-14)
- [x] Fix `CadenceTracker.GetNextStep` Step 2 ↔ Step 3 infinite loop so Steps 4 (`linkedin-connect`) and 5 (`breakup-email`) are reachable (`internal/communication/cadence.go`)
- [x] Enforce delivery-before-persistence in `CadenceAwareManager.checkCadence` and `Manager.ProcessInbound("START_OUTREACH")`, recording `Skipped` when optional channels lack targets and syncing `outreach_log`
- [x] Stop recording synthetic `START_OUTREACH` as fake `Inbound` / premature `StateEngaged`, remove duplicate 30m retry loop, add `401` LLM classifier fallback, and narrow `"spam"` unsubscribe false positive (`internal/communication/manager.go`)
- [x] Enable unauthenticated `401` GitHub API fallback + `directRepoFullName` commits lookup across `GitHubEnrichSource`, `TargetDiscoveryWorker`, `GitHubIssueSource`, and `GitHubCITracker` so 17,953 discovered deals enrich even while `GITHUB_TOKEN` is revoked
- [x] Fix lowercase `"mcp"`/`"llm"` bug in `github_sender.go` and blank-field placeholder defaults in `responder.go`
- [x] Bound `go-rod` headless Chromium in `linkedin_source.go` and `linkedin_sender.go` (1-slot semaphore, 45s timeout, `NoSandbox(true)`, `Kill()` + `Cleanup()`) to protect shared Hetzner VPS RAM/disk
- [x] Stamp and expose `version` (`0.7.4`) in `/health/detailed`, `/api/v1/metrics`, and `build.bat`; add Hetzner-only runtime guard to `start.bat`; merge Dependabot Go security updates (`echo/v4`, `golang.org/x/*`); refresh `README.md`, `CHANGELOG.md`, `MEMORY.md`, `ROADMAP.md`, and `HANDOFF.md`

## Completed Earlier (v0.7.1 – v0.7.3)
- [x] Merge `backup/wip-leader-sync-20261107` -> `main`, push (`main == origin/main`)
- [x] Confirm leader-sync FF + merge `feature/v0.7.1-consolidated-dashboard` (`github/main` + GitLab `origin/main` synced)
- [x] Decide one-time `$1.99` vs recurring (`$1.99` first month, `$5/mo` after; `checkout.html` & `pricing.html` reconciled)
- [x] Integration & CI workflows all green (`dockerignore`, `golangci-lint` Go 1.25, migrations `000013`/`000014`, schema-drift sentinel `internal/db/schema_test.go`, enrichment query fix)
- [x] Single-page dashboard (`/`, `/#outreach` 301 redirect, Operator Tools `6D`, Provider Credential panel `6A`, `/api/v1/metrics` per-source counters, Apollo `401/403` failure classification)
- [x] Nondestructive archive of legacy scripts (`scripts/archive/LEGACY_RUNNERS.md`), stale A/B pages, and unreferenced media (`hypernexus_site/archive/`)
