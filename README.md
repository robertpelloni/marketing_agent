# marketing_agent

An autonomous B2B sales-pipeline and dual-brand growth service written in Go, built to market **HyperNexus** (`hypernexus.site`) and **TormentNexus** (`tormentnexus.site`). It discovers prospective engineering leads, enriches contacts across multiple sources with automatic unauthenticated GitHub fallback, runs cadence-aware multi-channel outreach, handles inbound replies, and connects prospects directly to Stripe billing. This repository also contains the public sites and launch materials (`hypernexus_site/`, `tormentnexus_site/`, `landing/`, `demo_site/`, `launch_materials/`).

> **Alpha software.** Version **0.7.4** (per `VERSION` and `VERSION.md`).

## Operational Status

- **CI & Test Suite Green:** All 5 GitHub Actions workflows (`CI`, `Docker Build`, `Integration Tests`, `Deploy`, and `CodeQL`) are green on `main`, backed by real containerized PostgreSQL + `pgvector` migrations and a schema-drift sentinel (`internal/db/schema_test.go`).
- **Hetzner-Only Runtime & Always-Deploy Policy:** Production runs exclusively as a headless `systemd` unit (`marketing-agent`) on Hetzner (`5.161.250.43`, `/opt/marketing_agent`). Local builds and unit tests run on developer workstations, while local server launches are blocked by default in `start.bat`. Every push to `main` is deployed directly to Hetzner and verified via `/health`, `/health/detailed`, and `/api/v1/metrics`.
- **Single-Page Operations Hub (`/`):** Consolidated dashboard with six high-value-first sections (Subscription, KPIs, Cold Outreach, Active Deals, Social/DevRel, and Pipeline Ops + Operator Tools + Provider Credential Presence), live WebSocket telemetry, and `/dashboard/outreach` → `/#outreach` 301 redirect (`/demo` serves the standalone TormentNexus Cloud Console view).
- **Pricing Model:** `$1.99` first month, then `$5/mo` recurring (or `$19.99/yr`), implemented via Stripe Checkout ad-hoc pricing + once-coupon (`first_month_amount: 199`, `price_amount: 500`).
- **External Credential Status:** When third-party enrichment or LLM keys (`APOLLO_API_KEY`, `HUNTER_API_KEY`, `DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `GITHUB_TOKEN`) are missing or revoked (`401`), the service surfaces honest per-source failure counters on `/api/v1/metrics` and automatically falls back to unauthenticated public GitHub API lookups and deterministic template/classifier paths rather than stalling or recording phantom sends.

## Core Capabilities

1. **Lead Discovery (`internal/scraper`, `pkg/agents/discovery.go`):** Discovers active engineering organizations from GitHub repositories, GitHub issues, Hacker News "Who is Hiring", Reddit, and bounded headless LinkedIn searches.
2. **Resilient Contact Enrichment (`internal/enrichment`):** Waterfall enrichment across Hunter.io, Apollo.io, and GitHub commit history (`directRepoFullName` fast path + HTTP/1.1 transport + sliding-window circuit breaker + automatic `401` unauthenticated fallback).
3. **Technical Dossier Research (`internal/researcher`):** Compiles prospect tech-stack context for personalized messaging.
4. **Multi-Channel Cadence Outreach (`internal/communication`):** Monotonic 5-step sequence (`initial-email` → `github-issue-comment` → `followup-email` → `linkedin-connect` → `breakup-email`) with adaptive follow-up spacing, template placeholder defaults, delivery-before-persistence invariants, and automatic `Skipped` progression when optional channels lack targets.
5. **CRM & 7-State Deal Lifecycle (`internal/crm`, `internal/db`):** Tracks `Discovered` → `Researched` → `Outreach_Sent` → `Engaged` → `Negotiating` → `Closed_Won` / `Closed_Lost`.
6. **Stripe Billing (`internal/billing`):** Checkout session creation, subscription management, and webhook provisioning.
7. **Content & Social Publishing (`pkg/agents`):** Automated technical blog generation and dual-brand social publishing queues.

## Stack

| Component | Technology |
|---|---|
| Language | Go 1.25 |
| Database | PostgreSQL 13+ (`pgvector` in integration CI) |
| LLM | DeepSeek / Hermes / MiMo (health-gated fallthrough in `internal/llm`) |
| Email | Gmail OAuth2 / SMTP + IMAP reply polling |
| Browser Automation | `go-rod` (1-slot semaphore, 45s hard timeout, `--no-sandbox`, auto-cleanup) |
| Billing | Stripe |
| Websites | Static HTML/CSS/JS served via nginx (`/var/www/hypernexus.site`, `/var/www/tormentnexus.site`) |

## Build & Test

```bash
# Unix / CI build (stamps main.version from VERSION)
./build.sh

# Windows build (stamps main.version from VERSION, skips destructive gitcheck tests)
build.bat
```

> **Warning (`internal/gitcheck` & `internal/gitres`):** Running `go test ./...` at the repository root executes `gitcheck`/`gitres` tests that run `git reset` against the working tree. Always commit first, use an isolated worktree, or test explicit packages:
>
> ```bash
> go test ./cmd/... ./internal/auth/... ./internal/billing/... ./internal/communication/... \
>   ./internal/config/... ./internal/crm/... ./internal/db/... ./internal/deploy/... \
>   ./internal/enrichment/... ./internal/knowledge/... ./internal/llm/... ./internal/logging/... \
>   ./internal/researcher/... ./internal/sales/... ./internal/scraper/... ./internal/web/... ./pkg/...
> ```

## License & Links

- License: See `LICENSE` / `LICENSE.txt`
- Products: https://hypernexus.site · https://tormentnexus.site · https://github.com/HyperNexusLLC/HyperNexus
- Author: https://robertpelloni.com
