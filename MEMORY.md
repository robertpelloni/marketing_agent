# MEMORY — Internal Architectural Observations & Runtime Knowledge

## Architecture & Framework Rules
- **Repository Identity**: `marketing_agent` (formerly `enterprise_sales_bot`), written in Go 1.25 modular monolith architecture with PostgreSQL for persistent state management.
- **Dual-Brand Split**: Dynamic routing between `hypernexus.site` (corporate/enterprise) and `tormentnexus.site` (open-source developers).
- **Single Page Dashboard Philosophy**: Keep the web dashboard consolidated on a single responsive view (`/`), avoiding scattered subpages and keeping high-value conversion indicators at the top.
- **Headless Unix vs Desktop Windows Systray**: Windows desktop runs `getlantern/systray` for taskbar notifications and interactive launchers. Unix/Linux runners operate headlessly in background mode controlled by OS signals.
- **Cadence Engine Max Attempts**: Configured to `MaxAttempts: 10` across multi-channel touches (Email -> GitHub -> Email -> LinkedIn -> Email) to ensure leads receive sustained, high-value touches before exhaustion.
- **Testing Guardrails**: Target specific internal packages (`go test ./internal/... ./cmd/...`) to avoid executing unconfigured submodule tests.

## Standing Constraints & Lessons (2026-11-13)
- **Hetzner only**: never run local servers (no local binary/Postgres/dashboard). Build and unit-test locally; deploy and watch logs only on 5.161.250.43.
- **Always deploy**: every push is followed by a Hetzner deploy + health check in the same session.
- **Test-suite hazard**: go test ./... in the repo root runs git reset against the checkout (internal/gitcheck & gitres tests) and bumps internal/autodev/VERSION*. Commit before full-suite runs or use a worktree.
- **Migration discipline**: code-referenced columns MUST ship with a migration in the same commit (000013 cadence_step, 000014 template_id/response_id were retrofits for exactly this failure).
- **Enrichment invariant**: enrichment queries must not require the contacts they exist to create (ListDealsForEnrichment vs the contact-joined ListDealsByState).
- **Pricing truth**: $1.99 first month, then $5/mo (yearly $19.99). No free-trial copy anywhere.
- **Systray**: Windows-only (cmd/marketing_agent/systray.go): Open Dashboard, Outreach, sites, Blog, Restart, Quit Servers & Exit. Linux runs headless under systemd.
