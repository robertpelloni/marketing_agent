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
