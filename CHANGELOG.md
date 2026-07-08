
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
