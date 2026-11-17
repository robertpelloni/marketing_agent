## [Unreleased]

### Fixed

- **Gmail Auth Circuit Breaker (`internal/communication/gmail_oauth2_sender.go`)**: After a provider-side credential rejection (`disabled_client` etc.), `Send` trips a 15-minute breaker: one ERROR trip log per window, then every caller (cadence, comm manager) fails fast with an "auth breaker open" error — which cadence still classifies as provider-auth and aborts on — instead of a token fetch + identical ERROR per deal. Success clears it. Tests: `gmail_oauth2_sender_test.go` (trip/suppress/retry-after-window, non-auth errors ignored, clear-on-success).
- **Dashboard Leads Paging Controls (`internal/web/server.go`)**: Operator Tools Leads gained Prev/Next page buttons with range display (`N–M of total`) walking the paginated `/api/v1/leads` 10 at a time (was: first 10 only, no way to page). JS syntax-verified; API bounds covered by `leads_pagination_test.go`.
- **Signed, Expiring Sessions (`internal/auth/auth.go`)**: Replaced the forgeable static session values (`authorized_admin_session`, bare `company_<id>` — valid for anyone who read the source) with HMAC-SHA256 signed tokens `v1.<subject>.<expiryUnix>.<sig>` (24h TTL). Signing secret from `SESSION_SECRET`, falling back to the admin password so existing deployments need no new env (rotating either invalidates all sessions). All four raw cookie parses in the web layer now go through validated `auth.CompanyID(r)`/`Subject(r)`. Existing browser sessions must log in once more. Tests: `auth_test.go` (roundtrip, legacy rejection, subject/signature tamper, expiry, wrong-secret, middleware) + web route tests minting real tokens.
- **Leads Pagination (`internal/db/repository.go`, `internal/web/server.go`)**: `GET /api/v1/leads` now paginates — `?limit` (default 100, max 500) and `?offset` (floor 0), stable `ORDER BY id`, total in `X-Total-Count` — instead of returning all ~24k companies (6.7MB). `ListCompaniesPage`/`CountCompanies` added; `ListAllCompanies` delegates unbounded for internal callers. Dashboard List Leads fetches the first 10 and shows the total. Tests: `leads_pagination_test.go` (9 bound cases).
- **Operator Endpoints Auth-Gated (`internal/web/server.go`)**: `/api/v1/leads`, `/api/v1/outreach/export`, `/api/v1/metrics`, and `/metrics` now require the dashboard session (`s.auth.Middleware`). The service binds `*:8084` and nginx proxies `/api/v1/` publicly, so the outreach CSV (contact emails), the full company list, lead creation, and DB metrics were internet-facing without auth. Public funnel endpoints (health, quote, checkout, newsletter, outreach stats aggregates) stay open. Tests: `routes_auth_test.go`; staging script now expects 303 for export.
- **Leads 500 Root Cause (`internal/db/repository.go`, `internal/web/server.go`)**: `ListAllCompanies` failed scanning NULL `tech_stack`/`hiring_signals` arrays (2,075 prod rows) via `pq.Array`; query now `COALESCE`s arrays to `'{}'` and checks `rows.Err()`. The handler also logs the real error instead of swallowing it behind a bare 500. `handleDetailedHealth` nil-guards DB/tracker so health reports instead of panicking.

## [0.7.11] - 2026-11-17

### Added

- **Dashboard Lane Truth & Leads UI (`internal/web/server.go`, `dashboard_panel_test.go`)**: Extracted Section 6A provider list into `providerPanelList`, which now reports `X/Twitter (social posting)` and `GitHub comment posting` lane state from `TWITTER_DISABLED` / `GITHUB_POSTING_DISABLED` (amber `disabled via …`) instead of only credential presence; wired `/api/v1/leads` into Operator Tools 6D with a List Leads control (count + first entries, honest non-array error state).
- **Scripts Index (`scripts/SCRIPTS_INDEX.md`)**: Canonical vs archived inventory for `scripts/`; 20 retired Apollo/Hunter and superseded social-poster scripts moved to `scripts/archive/2026-11/` via `git mv`.

### Fixed

- **Autodev Test Pollution Fixed (`internal/autodev/orchestrator.go`)**: `finalizeCycle` wrote `VERSION`/`VERSION.md`/`CHANGELOG.md` relative to CWD, so every `go test ./...` bumped the committed files under `internal/autodev/` (observed as phantom modifications after full-suite runs). It now returns early under `GO_TEST_MODE=true` (the flag the tests already set, same convention as `SKIP_AUTODEV_*`). Tests: `finalize_cycle_test.go` (skips in test mode, still writes outside it); verified by a full-suite run leaving a clean tree.
- **GitHub Discussions Engager Implemented (`pkg/agents/github_discussions.go`)**: The 6h prod worker was a skeleton — `searchDiscussions` discarded every GraphQL response and returned 0 results, `postReply` was a no-op returning nil. Now: proper JSON envelope decoding (nodes mapped with owner/repo split, GraphQL errors surfaced, injectable `apiURL`), a real `addDiscussionComment` mutation, and an operator gate — `engage()` skips the whole cycle when `GITHUB_POSTING_DISABLED=true` (prod stance; replies are real public comments). Tests: `github_discussions_test.go` (parse, GraphQL error, mutation body, disabled-gate).
- **Configurable LLM Generation Timeout (`internal/llm/deepseek.go`, `internal/config/config.go`, `cmd/marketing_agent/main.go`)**: `DEEPSEEK_TIMEOUT_SECONDS` (default 120) now bounds DeepSeek-compatible generation calls; prod sets 300 because long blog posts through the FreeLLM router exceeded 120s. Prod also switched `DEEPSEEK_MODEL` from `free-llm` (reasoning model: slow, leaked "The user wants…" thinking into post copy) to `free-llm-plain` (23s vs 33s on probe, clean copy) — direct marketing generation verified succeeding after the switch. Test: `TestDeepSeekTimeoutConfigurable`.
- **LinkedIn Discovery Gate (`internal/scraper/linkedin_source.go`)**: `Discover` honors `LINKEDIN_DISABLED=true` by returning no companies immediately — previously it burned ~47s on a headless scrape that always timed out on the VPS, then injected 3 fabricated "simulated" companies into the CRM. A retired lane now produces neither scrape load nor fake leads. Test: `TestLinkedInSourceDisabledSkipsScrape` (fails on the old code with the timeout + fabrication observed).
- **Cadence LinkedIn Step Gate (`internal/communication/cadence.go`)**: The cadence `linkedin-connect` step now honors `LINKEDIN_DISABLED=true` (via `linkedInLaneDisabled`), recording the step as `Skipped` and advancing instead of launching the headless sender that times out on the VPS (completes the LinkedIn lane retirement started in the social hub). Test: `TestLinkedInLaneDisabled`. Round-2 verification also confirmed `system_health` back to `Success` after the Dependabot scoping.
- **Cadence Abort on Provider Auth Rejection (`internal/communication/cadence.go`)**: Added `isProviderAuthError` (`disabled_client`, `invalid_grant`, `invalid_client`, token-fetch failures) alongside the Gmail daily-limit abort: when the Gmail OAuth client is disabled provider-side, the cadence cycle now aborts with one operator-action log line instead of recording a failed send per deal (observed live 2026-11-17 after Google disabled the OAuth client). Test: `TestIsProviderAuthError`.
- **LinkedIn Lane Gate (`pkg/agents/social_platform.go`)**: `LINKEDIN_DISABLED=true` skips the headless LinkedIn provider at hub registration (same pattern as `TWITTER_DISABLED`); the lane timed out every cycle on the VPS and Chromium is the service's main OOM risk (OOM kill 2026-11-17 02:28 UTC, 718M peak). Panel 6A now also reports the LinkedIn lane state. Test: `TestHubSkipsLinkedInWhenDisabled`.
- **Dependabot Scope (`.github/dependabot.yml`)**: New config limiting updates to `gomod` (`/`, `/go`), `npm` (`/`), and `github-actions` (`/`). Unscoped Dependabot was attempting failing `uv` updates against the vendored `go/agents/software-agent-sdk` subtree on every push, poisoning combined commit status (`system_health=Failure` while all real workflows were green).

- **X/Twitter 402 Failure Spam (`pkg/agents/social_platform.go`, `social_platform_bluesky.go`, `social_poster.go`)**: Added `TWITTER_DISABLED=true` hub gate (mirrors `GITHUB_POSTING_DISABLED`), an `ErrTwitterCreditsDepleted` sentinel for HTTP 402 so a credits-depleted lane logs once as disabled instead of WARN + simulated DB rows every 4h social cycle, and a poster-level skip for platforms with no registered provider (no LLM generation, no simulated rows). Focused tests in `pkg/agents/social_platform_test.go`.
- **Nginx Site Patch Script (`scripts/patch-nginx-hypernexus.py`)**: Tracked the already-applied live hypernexus.site patch (extensionless `.html` via `try_files $uri $uri.html`, real 404, `/features` 301) — verified live: `/features` 301, `/download.html` 200, unknown path 404.

## [0.7.10] - 2026-11-15

### Fixed

- **Staticcheck `QF1002` Tagged Switch (`internal/communication/imap_receiver.go`)**: Converted `extractEntityParts` `mediaType` dispatch to a tagged `switch mediaType` so GitHub Actions `golangci-lint` (`lint` job) passes cleanly.
- **Root-Anchored Binary Ignore & Memory Tracking (`.gitignore`)**: Anchored `/marketing_agent` and `/marketing_agent_linux` to the repo root (so `cmd/marketing_agent/` is never shadowed by `.gitignore`) and unignored `.memory/`.

## [0.7.9] - 2026-11-15

### Added

- **Idempotent IMAP MIME Body Parser & Persistent UID/Message-ID Deduplication (`internal/communication/imap_receiver.go`, `imap_receiver_test.go`, `internal/db/db.go`, `repository.go`)**:
  - Replaced `imap.FetchBody` (`BODYSTRUCTURE` metadata, which left `email.Body` empty on 100% of inbound emails) with `BODY.PEEK[]` (`&imap.BodySectionName{Peek: true}`) and full RFC 5322 `multipart/*` + `quoted-printable` / `base64` / HTML fallback decoding via `ParseRawMIMEBody`.
  - Added `processed_inbound_emails` table + `GetMaxProcessedInboundUID` / `IsInboundEmailProcessed` / `MarkInboundEmailProcessed` and strict `uid > r.lastUID` filtering to eliminate the RFC 3501 `UID <lastUID+1>:*` high-water wrap-around loop and prevent duplicate `Inbound` rows across service restarts.
- **Real Mailer-Daemon / DSN Bounce Recipient Processing (`internal/communication/imap_receiver.go`)**: Added `ExtractPermanentBounceRecipients` to parse `mailer-daemon@` / `postmaster@` DSNs before sender lookup, distinguishing temporary delivery delays (`(Delay)` / `will retry for`) from permanent `5xx` bounces and automatically clearing bounced `contacts.email` (`NULLIF($1, '')` so `contacts_email_key` `UNIQUE` never collides) + transitioning deals to `Closed_Lost`.
- **Graceful Polite-Decline Handling & Outbound Placeholder Sanitizer (`internal/communication/manager.go`, `responder.go`)**:
  - Added `ExtractFreshInboundText`, `StripQuotedReply`, and `isPoliteDecline` so quoted outreach blocks never pollute classification and polite non-commercial/decline replies transition deals cleanly to `Closed_Lost` without pushy rebuttals.
  - Added `SanitizeOutboundText` across `Generate`, `GenerateFromTemplate`, and `Manager.ProcessInbound` to strip any `[Your Name]` / `[My Name]` / `{{...}}` placeholders before delivery, and updated `seedTemplates` (`internal/db/db.go`) to auto-upgrade legacy `[Your Name]` / LLM-waterfall rows in Postgres on startup.

### Fixed

- **OpenAI-Compatible `/v1` Gateway URL Normalization & Honest Error Diagnostics (`internal/llm/deepseek.go`, `hermes.go`, `llm_test.go`, `cmd/marketing_agent/main.go`)**: Normalized `BaseURL` (`normalizeOpenAIBaseRoot`) across `DeepSeekLLMProvider` and `HermesLLMProvider` so both root and `/v1` gateway URLs resolve `/v1/models` and `/v1/chat/completions` cleanly (surfacing accurate `401 (invalid_key)` status instead of misleading `404`), and enabled `LLMIntentClassifier` when `DEEPSEEK_API_KEY` is configured.
- **Expanded Bot Author & Malformed TLD Filtering (`internal/db/contact_hygiene.go`, `contact_hygiene_test.go`, `internal/communication/cadence.go`)**: Blocked `action@github.com`, malformed numeric/unlinked TLDs (`*.unlinked2026...`, raw IP literals), and `-bot` / ` Bot` commit author names (`ellmos-multi-bot`, `MCP Scanner Bot`) in `SelectOutreachContact`.

## [0.7.8] - 2026-11-15

### Fixed

- **Staticcheck `QF1003` Cleanup (`internal/researcher/crawlers.go`)**: Replaced `if/else if/else` on `resp.StatusCode` with a tagged `switch` in `GitHubCrawler.Crawl` so GitHub Actions `golangci-lint` (`lint` job) passes cleanly.
- **Foundation-Model & Self-Repo Discovery Exclusion (`internal/scraper/github_issue_source.go`, `pkg/agents/discovery.go`)**: Added `anthropic`, `anthropics`, `openai`, `modelcontextprotocol`, `cloudflare`, `vercel`, `huggingface`, `hypernexussoft`, and `robertpelloni` to GitHub issue/repository discovery exclusion filters so official vendor SDK repos (e.g. `anthropics/claude-agent-sdk-python`) and self-owned repos are never ingested as cold-outreach prospects.

## [0.7.7] - 2026-11-15

### Added

- **Shared Contact & Bot Email Hygiene (`internal/db/contact_hygiene.go`, `contact_hygiene_test.go`)**: Added `IsAutomatedOrUnroutableEmail`, `IsAutomatedCommitAuthor`, `IsPersonalEmailDomain`, and `SelectOutreachContact` to reject AI coding agents (`codex@openai.com`, `claude@anthropic.com`, `cursoragent@cursor.com`, `devin@cognition.ai`), CI/dependency bots (`*[bot]*`, `bot@renovateapp.com`, `actions@github.com`, `snyk-bot`), and non-MX/placeholder domains (`*.github.io`, `*.local`, `localhost`, `example.com`) while preserving human surnames ending in `bot` (`talbot`, `abbott`, `cabot`).
- **In-Memory RSS Feed Caching & 429 Backoff (`internal/researcher/rss_crawler.go`)**: Added per-URL parsed-entry caching (`30m` default TTL) plus `15m` negative backoff caching on `HTTP 429` / `5xx` responses with an explicit `HyperNexus-ResearchBot/1.0` `User-Agent`, eliminating repeated per-lead RSS downloads that triggered `HTTP 429` on external engineering blogs.
- **Multi-Signal Intent Scoring (`internal/researcher/intent.go`)**: Expanded `IntentAggregator.Aggregate` across hiring/issue signals, tech stack (`Go`/`Rust`/`TypeScript`/`Python`), verified corporate domains, market-cap tier, MCP/multi-agent product-fit keywords, and verified GitHub/RSS/competitor findings instead of collapsing leads onto a constant score of `25`.

### Fixed

- **Enrichment FIFO Queue Rotation & GitHub Rate-Limit Circuit Tripping (`internal/enrichment/github_enrich.go`, `fallback.go`, `worker.go`, `internal/db/repository.go`)**:
  - `GitHubEnrichSource` now detects `HTTP 403`/`429` rate-limit exhaustion (`X-RateLimit-Remaining: 0` / `"rate limit"`), honors `Retry-After` / `X-RateLimit-Reset` to open the circuit breaker immediately without futile retries, treats `404`/`409` empty/deleted repos as clean empties, and propagates API errors instead of misreporting rate-limited commit fetches as `empty`.
  - `FallbackSource` skips domain-only enrichment lanes (`Hunter.io`, `Apollo.io`) on shared code-host leads (`github.com/<owner>/<repo>`) without inflating `empty` telemetry counters, and `Enricher` calls `TouchDealUpdatedAt` on clean zero-contact checks so `ListDealsForEnrichment` (`ORDER BY updated_at ASC`) rotates through all `17,900+` discovered leads while leaving rate-limited leads queued for the next cycle.
- **Outreach Bot Guardrails (`internal/communication/cadence.go`, `manager.go`)**: `CadenceAwareManager` and `Manager` now select the first valid human contact per company (`SelectOutreachContact`) and skip/refuse outbound email touches to automated or unroutable addresses already stored in Postgres.
- **Canonical GitHub Issue Lead Domains & Core API Lookup (`internal/scraper/github_issue_source.go`)**: `GitHubIssueSource.Discover` now resolves verified profile websites via `/users/{owner}`, falls back to canonical `github.com/{owner}/{repo}` instead of fabricating `<org>.com`/`.io` NXDOMAINs, fetches repo metadata via Core API `/repos/{owner}/{repo}` instead of burning Search API quota, and exits the keyword loop early on `HTTP 429`/`403`.
- **Researcher Target Hygiene & Deduplication (`internal/researcher/processor.go`, `crawlers.go`)**: `Researcher` now skips leads that already have a compiled `TechnicalDossier`, strips raw email addresses and consumer webmail domains (`gmail.com`, `icloud.com`, etc.) from crawler targets, and disables `401` `GITHUB_TOKEN` headers in `GitHubCrawler` after the first rejection.

## [0.7.6] - 2026-11-14

### Added

- **Stripe Billing Key Health & Expiry Observability (`internal/billing/billing.go`, `internal/web/server.go`)**: Added cached `StripeBillingClient.HealthCheck` and `KeyMode` (`Stripe (live)` / `Stripe (test)`), surfacing live `STRIPE_API_KEY` validity (`api_key_expired` / `HTTP 401` detection, write-only restricted-key awareness, and zero secret-byte exposure) in `/health/detailed` (`billing` field) and Dashboard Section `6A` (`Stripe (checkout/billing)`).
- **Wire-Level Stripe Checkout Form Verification (`internal/billing/billing_test.go`, `internal/web/checkout_test.go`)**: Added end-to-end `httptest.Server` tests verifying the exact `application/x-www-form-urlencoded` `/v1/checkout/sessions` payloads serialized by `stripe-go/v81` across all 4 entry flows (`homepage`, `pricing 3 seats`, `checkout 1 seat + email`, `founding 9->5 clamped seats`) as well as expired-key health and redaction behavior.

### Fixed

- **Redacted Public Checkout Error Responses (`internal/web/server.go`)**: `handleCreateCheckout` no longer echoes raw upstream `stripe.Error` strings (which include masked key prefixes/suffixes on `api_key_expired`) to unauthenticated callers; returns `HTTP 503 Service Unavailable` with a redacted diagnostic on Stripe auth/config errors and `HTTP 502 Bad Gateway` on upstream failures while logging full details server-side.
- **CI Staticcheck Cleanup (`internal/communication/cadence_test.go`)**: Fixed `SA4006` unused `prog` assignment at line 323 so GitHub Actions `golangci-lint` (`lint` job) passes cleanly.

## [0.7.5] - 2026-11-14

### Fixed

- **Gate 2 Server-Side Subscription & Trial Checkout (`internal/billing/billing.go`, `internal/web/server.go`)**: Enforced the finalized HyperNexus Professional License pricing strictly server-side inside a single Stripe Checkout `subscription` session (`CreateHyperNexusTrialCheckout` / `BuildHyperNexusTrialSessionParams`):
  - `$1.99` (`199` cents) one-time per-team 14-day trial fee charged today (`quantity: 1`, non-adjustable).
  - `$5.00/seat/month` (`500` cents) recurring subscription starting day 15 (`TrialPeriodDays: 14`, `seats` clamped `1–5`).
  - Optional `$200.00` (`20000` cents) one-time Founding Team Rate Lock line item when `founding: true` / `plan: "founding"`, persisting the grandfathered `$5.00/seat` 12-month rate lock on `checkout.session.completed`.
  - Removed the legacy `$3.01` once-coupon mechanism and closed the `$0.01` forged-amount hole: `POST /api/v1/billing/checkout` now deterministically rejects any client-supplied `price_amount` or `first_month_amount` with `HTTP 400 Bad Request` (and validates `success_url`/`cancel_url` hosts).
- **Unified Plan-Field Checkout Payloads (`hypernexus_site/index.html`, `pricing.html`, `checkout.html`, `cloud_login.html`)**: Preserved commit `e5c43b8c` pricing copy verbatim while updating all entry buttons to send server-validated plan fields (`tier`, `plan`, `seats`, `founding`) instead of raw amounts, adding buyer-selectable seat count (`1–5`) and optional `$200` Founding Team Rate Lock controls on `pricing.html` and `checkout.html`.

## [0.7.4] - 2026-11-14

### Fixed

- **Cadence Step 2 ↔ Step 3 Infinite Loop (`internal/communication/cadence.go`)**: Fixed `CadenceTracker.GetNextStep`, which previously matched `lastChannel == "email"` against Step 1 after Step 3 (`followup-email`), trapping leads in an infinite Step 2/3 loop and making Step 4 (`linkedin-connect`) and Step 5 (`breakup-email`) unreachable. Progression now resolves monotonically via `TemplateID` and `Direction: "Skipped"` history.
- **Delivery-Before-Persistence Outreach Invariant (`internal/communication/cadence.go`, `manager.go`)**: `CadenceAwareManager.checkCadence` and `Manager.ProcessInbound("START_OUTREACH")` now execute channel delivery (`sender.Send`, `linkedinSender.Send`, `githubSender.SendComment`) *before* recording `Outbound` interactions or advancing deal state/cadence steps. Unavailable channels (missing email, missing `LinkedInURL`, or zero relevant GitHub issues) record `Direction: "Skipped"` so deals advance cleanly to subsequent touches instead of stalling forever.
- **Synthetic `START_OUTREACH` State Corruption & Duplicate Retry Loop (`internal/communication/manager.go`)**: Stopped persisting synthetic `"START_OUTREACH"` triggers as fake `Inbound` rows, stopped prematurely marking newly contacted leads as `StateEngaged` before any prospect reply, removed the competing 30-minute 3x retry loop in `pollAndProcess`, added heuristic `MockIntentClassifier` fallback when LLM keys return `401`, and narrowed `isUnsubscribe` so benign `"check your spam folder"` replies no longer trigger GDPR erasure (`mark as spam`, `report spam`, and `stop spamming` remain active).
- **Unauthenticated GitHub Fallback & Shared Code-Host Guard (`internal/enrichment/github_enrich.go`, `hunter.go`, `apollo.go`, `pkg/agents/discovery.go`, `internal/scraper/github_issue_source.go`, `internal/deploy/github_tracker.go`, `cmd/marketing_agent/main.go`)**: Always enables `GitHubEnrichSource` regardless of `GITHUB_TOKEN` configuration, automatically disables a `401 Unauthorized` token in memory and retries unauthenticated against public GitHub endpoints, adds a `directRepoFullName` fast path (`github.com/<owner>/<repo>` → `/repos/<owner>/<repo>/commits`) that fixes live prod's `org:<repo> -> HTTP 400` failure, and skips `HunterSource` / `ApolloSource` on shared code-host domains (`github.com`, `gitlab.com`, etc.) so repository leads never query corporate GitHub Inc. contacts or burn paid API credits.
- **GitHub Issue Relevance & Template Fallbacks (`internal/communication/github_sender.go`, `responder.go`)**: Fixed lowercase-comparison bug where uppercase `"MCP"`/`"LLM"` keywords in `CalculateRelevance` could never match lowercased issue text; fixed `extractOrgFromDomain` for `github.com/<owner>/<repo>`; added clean fallback defaults (`there`, `your team`, `AI and modern engineering tooling`) for blank template fields in `GenerateFromTemplate`.
- **Bounded Headless Chromium Lifecycle (`internal/scraper/linkedin_source.go`, `internal/communication/linkedin_sender.go`)**: Added 1-slot concurrency semaphores, strict 45s context / 40s page timeouts, Linux root-safe launcher flags (`NoSandbox(true)` including `connectViaRod`, `Headless(true)`, `disable-gpu`, `disable-dev-shm-usage`), and mandatory `lnchr.Kill()` + `lnchr.Cleanup()` to prevent orphaned Chromium processes and `/tmp/rod` disk leaks on shared VPS infrastructure.

### Changed

- **Observability & Version Stamping (`internal/web/server.go`, `cmd/marketing_agent/main.go`, `build.bat`, `start.bat`)**: Exposed `version` (`0.7.4`) in `/health/detailed` and `/api/v1/metrics`, fixed DeepSeek vs. Hermes provider labeling on `/` and `/health/detailed`, synchronized `outreach_log` writes from `CadenceAwareManager` so `/api/v1/outreach/stats` stays current, aligned `build.bat` with `build.sh` (version stamping + no working-tree-resetting `gitcheck` test), and enforced the Hetzner-only runtime guard in `start.bat`.
- **Post-0.7.3 Production Hardening Included**: Incorporates enrichment HTTP/1.1 transport fix, sliding-window circuit breaker, per-source `/api/v1/metrics` counters, Apollo `401/403` failure classification, schema-drift sentinel (`internal/db/schema_test.go`), dashboard Operator Tools (`6D`) and Provider Credential panel (`6A`), DeepSeek → Hermes health-gated fallthrough, and Dependabot Go module security updates (`echo/v4`, `golang.org/x/*`).

## [0.7.3] - 2026-11-13

### Changed

- **Dashboard consolidation completed**: merged `github/feature/v0.7.1-consolidated-dashboard`
  (24 commits: outreach export, Prometheus/JSON metrics, telemetry reconnect, cadence and
  discovery improvements) into `main`. `/` is now the single operations page with six
  high-value-first sections (Subscription, KPIs, Outreach, Deals, Social, Pipeline Ops),
  quick-nav anchors, and tooltips. `/dashboard/outreach` now 301-redirects to `/#outreach`;
  the legacy page is archived at `web/dashboard/archive/outreach-legacy.html`.
  `/demo` remains as the separate TormentNexus Cloud Console subproject view, linked from
  the dashboard nav along with `/health` and `/api/v1/metrics`.
- **Systray**: Windows tray (`cmd/marketing_agent/systray.go`) Blog item now opens
  `https://hypernexus.site/blog/` instead of a nonexistent local `/blog` route.
  Tray menu: Open Dashboard, View Cold Outreach Tracking, HyperNexus/TormentNexus sites,
  Blog, Restart Service, Quit Servers & Exit.
- **Tests**: fixed `internal/sales` billing mocks for the seats/interval/price checkout
  interface (pre-existing failure). Full suite passes except
  `TestApolloSource_HealthCheck`, which calls the live Apollo API and fails with HTTP 401
  without a valid `APOLLO_API_KEY`.
- **Version**: `VERSION` and `VERSION.md` aligned as the single version source (0.7.3).

### Notes

- Warning: running `go test ./...` in the repo root resets the working tree
  (`internal/gitcheck`/`gitres` tests run `git reset` against the checkout) and bumps
  `internal/autodev/VERSION*`. Commit before running the full suite, or run it in a
  separate worktree/clone.

## [Unreleased]

### Changed

- **1.99 first month then 5-mo recurring**: checkout posts price_amount 500 plus interval month plus first_month_amount 199. Backend creates the 5-mo price with a once-coupon for 301 off and skips the 7-day trial so the first charge fires immediately. Deployed to prod, service restarted with zero errors, live page verified.

- **$1.99 trial funnel**: all buy buttons (index nav/hero/pricing card, pricing page) go straight
  to the card on hypernexus.site (`/checkout.html`); signup page bypassed. Checkout card shows
  `$1.99 Trial`, sends `price_amount: 199` (one-time payment via ad-hoc Stripe price_data, no new
  Price ID needed). Pricing copy notes the trial; monthly/yearly toggle untouched.
- **Sitemap**: corrected 8 `lastmod` dates (/ and /pricing 2026-11-05, /blog/ 2026-11-04,
  5 security posts 2026-11-03). All 11 posts + blog index covered.
- **Media canonical**: `hypernexus_extreme.jpg` drives og:image, twitter:image, both live.html
  video posters; no dangling refs. Deleted byte-identical dupes (chunky_poster_v1, mega_poster_final).

### Added

- **ROADMAP.md / TODO.md**: verified milestones + granular task list (replaces ad-hoc tracking).
- **Outreach**: LinkedIn/X/Reddit copy for all 11 posts in marketing/outreach_queue_2026-11-03.md;
  dev.to TLS + CISO checklists (canonical URL + tags).

### Ops

- **WIP preserved**: backup/wip-leader-sync-20261107 merged to main 2026-11-07, pushed (main == origin/main).
- Production verified: marketing-agent, nginx, postgresql active; no error-log entries in 24h.

## [0.7.1] - 2026-11-04

### Added
- **Consolidated Unified Operations Dashboard**: Integrated cold outreach analytics and live delivery logs directly into the primary dashboard (`internal/web/server.go`), removing subpage fragmentation.
- **CSV Data Export**: Added direct `/api/v1/outreach/export` CSV streaming endpoint and one-click "Export CSV ↓" action on the main dashboard for offline CRM ingestion and audits.
- **Resilient Telemetry Stream**: Implemented exponential backoff auto-reconnect logic for dashboard WebSocket client to maintain live event streaming across network interruptions.
- **Rich Contextual Tooltips**: Added intuitive `?` marker popovers with CSS hover details across all pipeline stages, MEDDPICC funnel metrics, manual deal actions, and contact channels.
- **Enhanced Systray Control**: Added direct launch options for HyperNexus and TormentNexus sites, unified dashboard, and outreach monitoring in `cmd/marketing_agent/systray.go`.
- **Streamlined Unix Build Script**: Added executable flags, version stamping from `VERSION.md`, and clean compilation output in `build.sh`.

### Fixed
- **Cadence Unit Test Failure**: Corrected `TestDefaultCadence_MaxAttempts` and `TestCadenceTracker_GetNextStep_ExhaustedAfterMaxAttempts` to reflect the configured `MaxAttempts: 10` sequence.
- **Headless Unix Graceful Exit**: Improved `cmd/marketing_agent/systray_unix.go` signal handling and context cancellation.

### Changed
- **Script Housekeeping**: Nondestructively moved redundant legacy scripts (`*_old.py`) into `scripts/archive/`.
- **LinkedIn Outbound Safety Indicators**: Added `CheckDailyConnectionLimit` threshold validation in `internal/communication/linkedin_sender.go` with unit tests and dynamic safety badge popovers on dashboard deal cards.
- **Devvit Reddit Integration & Preview Thumbnails**: Added `/api/social/reddit` alias endpoint and enriched JSON responses with cached post preview thumbnails (`thumbnail_url`, `preview_url`), brand routing, author, and timestamp metadata.
- **Automated Staging Health Checks**: Added `scripts/staging_health_check.sh` and wired syntax/executable verification into `.github/workflows/ci.yml`.
- **24-Hour Recency Target Discovery**: Added `BuildDiscoveryQuery` in `pkg/agents/discovery.go` with dynamic 24-hour lookback filter (`pushed:>YYYY-MM-DD`) and unit tests in `pkg/agents/discovery_test.go`.
- **Docker Compose Smoke Testing**: Added `scripts/docker_smoke_test.sh` to validate `docker-compose.staging.yml`, `docker-compose.yml`, and `deploy/docker-compose.yml`, integrated into `.github/workflows/ci.yml`.
- **Database Connection Pool Metrics**: Added `Stats()` on `*db.DB` and implemented Prometheus (`/metrics`) and JSON (`/api/v1/metrics`) telemetry endpoints for active, idle, and wait duration connection pool monitoring.
- **MCP & Skill-Based Response Optimization**: Enriched `RAGResponseGenerator.Generate` in `internal/communication/responder.go` with Model Context Protocol (MCP) and modular skill-based targeting heuristics, progressive tool routing guidance, and unit tests in `internal/communication/llm_responder_test.go`.
- **Targeted Outreach Prioritization**: Updated `scripts/outreach/outreach_targeted.py` and `scripts/outreach/run_outreach.py` to prioritize contacts by latest update timestamp (`c.updated_at DESC`) with `--recent` and `--hours` flags to target newly enriched prospects first.
- **Integration Test CI & Migration Runner**: Added dedicated migration command in `cmd/migrate/main.go` and updated `.github/workflows/integration-tests.yml` to cleanly execute schema migrations against containerized Postgres with pgvector.
- **Prometheus Alert Rules**: Added `deploy/prometheus-rules.yml` defining alert rules for connection pool exhaustion (`DatabasePoolExhausted`), wait queue contention, and health check probe failures, validated in CI.
- **Outreach Engagement Rates**: Added computed `open_rate` and `reply_rate` percentages to `/api/v1/outreach/stats` and displayed real-time rates on dashboard KPI cards with tooltip markers.
- **Channel-Adaptive Cadence Delays**: Implemented `CalculateAdaptiveDelay` in `internal/communication/cadence.go` with progressive spacing (24h -> 48h -> 72h) for repeated email follow-up touches to optimize response rates and reduce spam flags.
- **Cadence Follow-Up Personalization**: Added support for `{{research_insight}}` and `{{technical_dossier}}` placeholders in `GenerateFromTemplate` (`internal/communication/responder.go`) to incorporate prospect-specific research in multi-touch sequences.
- **Governance Updates**: Synchronized `ROADMAP.md`, `TODO.md`, `VISION.md`, `MEMORY.md`, `DEPLOY.md`, `IDEAS.md`, `HANDOFF.md`, and bumped version to `0.7.1`.

## [0.7.2] - 2026-08-20

### Repository Sync

- **Comprehensive sync**: Scanned and analyzed 90+ repositories for feature branches
- **Branch analysis**: All feature branches were either empty, stale, or upstream-only
- **HyperNexus sync**: Updated submodules, committed formatting changes
- **Server sync**: All services verified active, database at 1,786 contacts
- **Download links**: Updated to point to HyperNexusLLC/HyperNexus releases
- **Documentation**: Updated HANDOFF.md, VERSION.md, CHANGELOG.md

## [0.7.1] - 2026-07-13

### Fixed

- **Email template variables**: Added `stripTemplateVars()` regex filter that removes any bracketed
  placeholders (`[Your Name]`, `[Company Name]`, etc.) from LLM-generated email content.
  Even when the system prompt says "never include these", LLMs sometimes slip them in.
  This post-processing filter catches them all.
- **HN scraper broken**: HN's Algolia search API started returning HTML instead of JSON
  (rate-limiting). Rewrote to use Firebase user endpoint (`/user/whoishiring.json`).
  Added Algolia fallback with Content-Type check to avoid JSON parse errors.

### Added

- **Scraper constructors**: `NewLinkedInSource()` and `NewRedditSource()` restore build
  after partial merge. Reddit stubs search across r/LocalLLaMA, r/MachineLearning, etc.

### Changed

- **Blog engine rewritten**: 28 diverse topics (was 8), each with 3 angle variations.
  JSON state persistence prevents duplicates. Max tokens 3000.
- **Dashboard**: Pipeline funnel bar replaces text summary.
- **Systray**: Added Open Blog, Open Site, Restart Service menu items.
- **Project**: Scripts archived to `scripts/archive/`.

## [0.7.0] - 2026-07-13

### Fixed

- Template variable placeholder stripping in LLM emails
- Scraper constructors restored after branch merge
- TormentNexus TTY crash via script wrapper
- Orphaned `GetLastProvisionedCompany` call in server

### Added

- CDP poster for Twitter/Reddit silent posting
- Blog cross-post publisher (dev.to, hashnode)
- GitHub Discussions engager
- Launch materials (Product Hunt, Show HN)
- Stripe checkout with seat quantities

## [0.6.1] - 2026-07-07

### Added

- Secrets encryption at rest (`pkg/crypto` + secrets DB migration)
- GraphRAG memory vault integration tests
- Telemetry WebSocket integration tests
- Web dashboard handler tests
- GDPR endpoint unit tests
- Feedback loop integration tests
- Negative DB tests with nil checking

### Changed

- Dead `.gitmodules` entry removed (tormentnexus submodule never committed)
- `build.bat` and `start.bat` updated (submodule graceful handling, .exe extension)

## [0.6.0] - 2026-07-06

### Added

- Stripe subscription billing system with full subscription lifecycle
- Grandfathered rate support — auto-freezes old rates on price increases
- Stripe webhook handler (checkout, invoice, subscription events)
- Billing API endpoints (checkout, subscription, cancel, portal)
- HyperNexus $5 sale pricing with slashed $100 + pulsing SALE ENDING... SOON!
- HyperNexus Stripe checkout integration on pricing buttons
- Dashboard: real-time telemetry WebSocket, audit log stream, Hermes latency gauge
- Dashboard: self-service deploy (sync repo, trigger build) buttons
- GraphRAG memory vault schema and access layer
- Webhook IP allowlisting
- Database migration: subscriptions, price history, billing events tables

### Changed

- Robotic scream and heartbeat audio removed from TormentNexus canvas
- SpookySkull nose cavity moved between eyes and mouth, made angular and on top layer
- Both sites: copy updated with real technical specs (384-dim embeddings, BM25, L1/L2/L3 capacities)
- Version bumped to 0.6.0

## [0.5.3] - 2026-06-25

### Changed

- HyperNexus site: light theme, neon gradients, frosted glass cards, animated splashes
- TormentNexus site: skull jaw/nose adjustments, smaller nasal cavity
- Pricing: removed $0 tier, added self-host redirect to TormentNexus.site
- Video: NotebookLM video downloads and embedded players

### Fixed

- HyperNexus: removed body::before grid overlay causing frosted glass effect
- HyperNexus: fixed body overflow/scroll issues, cleaned CSS
