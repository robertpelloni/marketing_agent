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
