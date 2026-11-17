# TODO — marketing_agent

Immediate, granular tasks. Long-term direction lives in `ROADMAP.md`.

## Now (Operator / Account Auth Required)
- [ ] Live real-charge card test: `$1.99` per-team 14-day trial today -> `$5/seat/mo` day 15 recurring on `https://hypernexus.site/checkout.html` (live `STRIPE_API_KEY` rotated and all 4 unpaid `cs_live_...` entry flows verified by both Antigravity & Bobbi on `2026-11-15`; final completed `$1.99` card charge is Robert's call)
- [ ] Refresh remaining rotated/revoked third-party API keys in prod `/opt/marketing_agent/.env` (and workstation env): `GITHUB_TOKEN`, `DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `APOLLO_API_KEY`, `HUNTER_API_KEY` (unauthenticated GitHub fallback, rate-limit circuit breaker, FIFO enrichment rotation, and template outreach active in `v0.7.7` meanwhile)
- [ ] Decide on Bobbi's two optional site flags: `live.html` trial CTA target (`/contact.php` vs `/checkout.html`) and archiving/updating repo-only `$19.99` pages (`cloud_index_b`, `cloud_login_a/b`, `testimonials.html`)
- [ ] Manual posts: LinkedIn (6) / X (tweets in `POSTING_GUIDE.md`) / Hashnode / Medium; tick sync checklist
- [ ] Marketing bot secrets: `DEVTO_API_KEY`, `BLUESKY_TN_PASSWORD`, `BLUESKY_HN_PASSWORD`, funded X API credits
- [ ] Resubmit sitemap in Google Search Console after `lastmod` corrections

## Done 2026-11-16 (autopilot loop)
- [x] Retire X lane: `TWITTER_DISABLED=true` (prod env) + hub gate + `ErrTwitterCreditsDepleted` sentinel + poster skip for unregistered platforms (`pkg/agents/social_platform*.go`, `social_poster.go`, tests green, deployed)
- [x] FreeLLM LLM stopgap on prod (`DEEPSEEK_BASE_URL=http://127.0.0.1:4000`, `free-llm`); `/health/detailed` -> `DeepSeek: Connected`. Durable fix still needs a real key rotation below.
- [x] Dashboard truth: Section 6A `providerPanelList` shows `TWITTER_DISABLED` / `GITHUB_POSTING_DISABLED` lane state; Leads API wired into Operator Tools 6D (`internal/web/server.go`, `dashboard_panel_test.go`)
- [x] Archive 20 retired/superseded scripts to `scripts/archive/2026-11/` + `scripts/SCRIPTS_INDEX.md`; track `scripts/patch-nginx-hypernexus.py` (already live)
- [x] Git: `main` 0/0 vs GitHub+GitLab; `borg`-pointer branches (`fix/jules-borg-pin`, `backup/fdafa92-pre-borg-repin`) deliberately unmerged (submodule removed by design)

## Watch
- [ ] **Gmail OAuth client DISABLED (2026-11-17, operator action):** Google disabled the OAuth client (`disabled_client` on token fetch) — ALL cadence email fails until the client is re-enabled/replaced in Google Cloud Console. Code now aborts the cycle with one clear line (`isProviderAuthError`) instead of per-deal failures. After re-enabling, also confirm the refresh token still works.
- [ ] Hetzner disk 86% (11G free, worsening): growth is other tenants' `/tmp` backups (ercm/foreclosure/contentplanner, 3.3G) + `/root/.cache/ms-playwright` (701M) — NOT this project's data; owner decision before any cleanup. `/opt/marketing_agent` itself is ~500M.
- [ ] `system_health` Failure root cause found: failing Dependabot `uv` dynamic jobs (vendored `go/agents/software-agent-sdk`) poison combined status while CI/Integration/Deploy/Docker/CodeQL are green on HEAD. `.github/dependabot.yml` now scopes ecosystems — confirm `system_health` returns to Success after the next push cycle.
- [ ] OOM kill 2026-11-17 02:28 UTC (718M peak, systemd auto-restarted): headless Chromium lanes are the prime suspect; LinkedIn social lane now gated off. Consider a systemd `MemoryHigh` guard for marketing-agent if OOM recurs. Cadence `linkedin-connect` still uses the headless `LinkedInSender` — watch for repeats.

## Completed in v0.7.7 (2026-11-15)
- [x] Rotate live `STRIPE_API_KEY` / `STRIPE_PUBLISHABLE_KEY` / `STRIPE_MANAGEMENT_KEY` in local `.env` and Hetzner `/opt/marketing_agent/.env`, verify `/health/detailed` -> `Stripe (live): Connected`, and verify all 4 live unpaid `cs_live_...` Gate 2 checkout sessions (confirmed by Bobbi's independent live walk on Muse)
- [x] Filter AI-agent, CI-bot, and unroutable email addresses (`codex@openai.com`, `claude@anthropic.com`, `bot@renovateapp.com`, `*.github.io`, `*[bot]*`) across `GitHubEnrichSource`, `ApolloSource`, `HunterSource`, `CadenceAwareManager`, and `Manager` (`internal/db/contact_hygiene.go`)
- [x] Fix `GitHubEnrichSource` `HTTP 403`/`429` rate-limit handling (`X-RateLimit-Remaining: 0` circuit trip + honest `Failures` accounting) and add `TouchDealUpdatedAt` FIFO queue rotation in `Enricher` so clean zero-contact leads rotate instead of blocking all `17,900+` discovered leads (`internal/enrichment/`, `internal/db/repository.go`)
- [x] Stop `GitHubIssueSource` from fabricating `<org>.com`/`.io` NXDOMAINs (use verified `/users/{owner}` website or canonical `github.com/{owner}/{repo}`) and switch repo lookup from Search API to Core API `/repos/{owner}/{repo}` (`internal/scraper/github_issue_source.go`)
- [x] Add `30m` in-memory cache + `15m` `429` backoff to `RSSFeedCrawler`, strip raw email/webmail targets from `Researcher`, skip already-compiled `TechnicalDossier` rows, and upgrade `IntentAggregator` multi-signal scoring (`internal/researcher/`)

## Completed in v0.7.6 (2026-11-14)
- [x] Add cached `StripeBillingClient.HealthCheck` + `KeyMode` (`internal/billing/billing.go`) and surface live Stripe key validity (`api_key_expired` / `HTTP 401` vs write-only `rk_` vs `Connected`) in `/health/detailed` (`billing`) and Dashboard Section `6A`
- [x] Redact upstream `stripe.Error` details in `handleCreateCheckout` (`HTTP 503` with safe diagnostic on `api_key_expired`/`401`, zero secret key prefixes/suffixes leaked to public HTTP callers)
- [x] Add wire-level `httptest.Server` Stripe form-payload tests across all 4 checkout entry flows (`homepage`, `pricing 3 seats`, `checkout 1 seat + email`, `founding 9->5 clamped seats`) + expired-key redaction tests (`internal/billing/billing_test.go`, `internal/web/checkout_test.go`)
- [x] Fix `golangci-lint` `SA4006` warning in `internal/communication/cadence_test.go:323` so GitHub Actions `lint` check run passes cleanly

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
