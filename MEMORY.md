# MEMORY — Internal Architectural Observations & Runtime Knowledge

## Architecture & Framework Rules
- **Repository Identity**: `marketing_agent` (formerly `enterprise_sales_bot`), written in Go 1.22 modular monolith architecture with PostgreSQL for persistent state management.
- **Dual-Brand Split**: Dynamic routing between `hypernexus.site` (corporate/enterprise) and `tormentnexus.site` (open-source developers).
- **Single Page Dashboard Philosophy**: Keep the web dashboard consolidated on a single responsive view (`/`), avoiding scattered subpages and keeping high-value conversion indicators at the top.
- **Headless Unix vs Desktop Windows Systray**: Windows desktop runs `getlantern/systray` for taskbar notifications and interactive launchers. Unix/Linux runners operate headlessly in background mode controlled by OS signals.
- **Cadence Engine Max Attempts**: Configured to `MaxAttempts: 10` across multi-channel touches (Email -> GitHub -> Email -> LinkedIn -> Email) to ensure leads receive sustained, high-value touches before exhaustion.
- **Testing Guardrails**: Target specific internal packages (`go test ./internal/... ./cmd/...`) to avoid executing unconfigured submodule tests.
