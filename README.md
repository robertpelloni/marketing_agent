# TormentNexus Autonomous Sales Pipeline Architecture

An asynchronous, event-driven orchestration layer in Go for automated B2B lead generation, enrichment, hyper-personalized outreach, and billing. This repository implements multi-agent orchestration, background worker engines, and strict state management to run reliable autonomous sales pipelines for enterprise workflows.

## Table of Contents

- Project Overview
- Features
- Architecture & Conventions
- Getting Started
- Configuration
- Database & Migrations
- Development Guidelines
- Testing
- Repository Management — EXECUTIVE PROTOCOL
- CI / Validation
- Contributing
- License & Contact

## Project Overview

TormentNexus is a modular Go-based system that runs concurrent agent workers to:

- Discover and scrape target companies and contacts (headless scraper engines).
- Enrich contact and company data with third-party providers.
- Run hyper-personalized outreach workflows across channels.
- Track interactions, deals, and billing with strict relational state persisted in PostgreSQL.
- Autonomously develop, test, and merge its own code changes via CI-gated PRs.
- Self-improve outreach quality by learning from successful past interactions.

Key design goals:

- Concurrency-first: goroutines and channels for background workers.
- Deterministic state machine: atomic state transitions for leads and deals.
- Testability: abstract external integrations behind interfaces and include explicit mock endpoints.
- Autonomous development: tools & scripts that maintain sync with upstream forks and validate merges.
- Self-learning: feedback loops that improve sales strategy and LLM prompt quality over time.

## Features

- Multi-agent orchestration and pluggable worker engines.
- Atomic 7-state machine for lead lifecycle (Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost).
- Headless browser scraper profiles for deterministic scraping.
- External integrations via interfaces to allow mocking in tests.
- Comprehensive logs and target histories persisted to PostgreSQL.
- Defensive execution loops and mock testing endpoints for each agent.
- RAG-powered technical Q&A using TormentNexus documentation as grounding context.
- Self-improving prompts: successful past interactions injected as few-shot examples.
- Tiered pricing engine (Enterprise $50K, Mid-Market $15K, SMB $5K).
- Lead scoring, qualification, and intelligent routing.
- Stripe billing integration with automated invoicing for won deals.
- Autonomous development loop: reads TODO.md, proposes code, verifies, creates PRs, and merges after CI passes.
- Dual-direction intelligent merge engine for branch reconciliation.
- Self-service deployment dashboard with sync and build triggers.
- GitHub webhook integration for automatic deployment on push.
- Real-time CI status tracking and deployment health monitoring.

## Architecture & Conventions

- Language: Go 1.24 (primary), PL/pgSQL for DB functions.
- Agents live in `/pkg/agents/` or `/internal/`.
- New engines must implement the `Agent` interface (or the module-specific equivalent).
- All background routines and workers must:
  - Implement graceful shutdown and health probes.
  - Include defensive execution loops (backoff, jitter, circuit-breakers).
  - Provide explicit mock endpoints for testing.
- Database mapping follows strict relational models:
  - Companies -> Contacts -> Interactions -> Deals with full foreign keys.
  - All updates to lead state must be performed via atomic transactions to enforce state machine invariants.

### Recommended Repo Layout

```
/cmd/           - CLI binaries / worker entrypoints
/pkg/agents/    - Agent implementations and worker engines
/pkg/config/    - Safety configuration and guardrails
/internal/      - Internal frameworks: autodev, billing, communication, crm, db, deploy, enrichment, gitcheck, gitres, llm, researcher, sales, scraper, web
/migrations/    - SQL migrations (strict relational schemas)
/scripts/       - Automation scripts (sync_repo.sh, smoke_test.go)
/tests/e2e/     - End-to-end integration tests
```

### Background Workers

| Worker | Package | Interval | Purpose |
|---|---|---|---|
| Scraper | `internal/scraper` | 1h | Discover new leads from job boards & GitHub |
| Enricher | `internal/enrichment` | 1h | Enrich companies with contact data |
| Researcher | `internal/researcher` | 1h | Build technical dossiers via crawling |
| CRM Sync | `internal/crm` | 30m | Bidirectional CRM reconciliation |
| Target Discovery | `pkg/agents` | 2h | GitHub MCP server discovery |
| Communication | `internal/communication` | 30m | Process inbound + trigger outbound |
| AutoDev | `internal/autodev` | 1h | Self-code, PR, and merge cycle |
| Deploy Sync | `internal/deploy` | Configurable | Background repo synchronization |
| Deploy Monitor | `internal/deploy` | Configurable | Health check monitoring |

### State Machine

```
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won
                                                                    ↘ Closed_Lost
```

All transitions are atomic database updates enforced via the `lead_state` enum in PostgreSQL.

## Getting Started

### Prerequisites

- Go 1.24+ (or the version pinned in go.mod)
- PostgreSQL 13+
- (Optional) Docker & Docker Compose for local stacks
- (Optional) GitHub Personal Access Token with `repo` permissions

### Quick Start (Docker)

```bash
docker compose up --build
```

The dashboard will be available at `http://localhost:8080`.

### Quick Start (Local)

1. Set environment variables (example):
   - `DATABASE_URL=postgres://user:pass@localhost:5432/tormentnexus?sslmode=disable`
   - `PORT=8080`
   - `GITHUB_TOKEN=ghp_xxxx` (optional, enables real GitHub integration)
   - `GITHUB_REPOSITORY=owner/repo` (optional, enables CI tracking and AutoDev PRs)
   - `DEPLOY_SYNC_INTERVAL=1h` (optional, enables background sync)

2. Run migrations:
   ```bash
   migrate -path migrations/ -database "$DATABASE_URL" up
   ```

3. Initialize submodules:
   ```bash
   git submodule update --init --recursive
   ```

4. Build and run:
   ```bash
   go build -o bin/sales_bot ./cmd/sales_bot
   ./bin/sales_bot
   ```

   Or use the provided scripts:
   ```batch
   build.bat
   start.bat
   ```

## Configuration

Use environment variables to configure runtime behavior:

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | `postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable` | PostgreSQL connection string |
| `GITHUB_TOKEN` | No | — | GitHub PAT for API access (enrichment, CI, PRs) |
| `GITHUB_REPOSITORY` | No | — | `owner/repo` for CI tracking and AutoDev |
| `GITHUB_WEBHOOK_SECRET` | No | — | HMAC secret for webhook verification |
| `CRM_BASE_URL` | No | — | REST CRM API base URL (enables real CRM) |
| `CRM_API_KEY` | No | — | REST CRM API key |
| `DEPLOY_SYNC_INTERVAL` | No | — | Duration string (e.g., `1h`, `15m`) for background sync |
| `GO_TEST_MODE` | No | — | Set to `true` to skip git operations in tests |
| `SKIP_AUTODEV_SYNC` | No | — | Set to `true` to skip git sync in AutoDev |
| `SKIP_AUTODEV_TESTS` | No | — | Set to `true` to skip test verification in AutoDev |

Agents and integrations are configured through typed configuration structs and dependency injection; avoid hard-coded credentials.

## Database & Migrations

- All schema changes MUST be delivered through the `migrations/` folder and reviewed as part of PRs.
- Enforce relational integrity with explicit foreign keys:
  - `companies(id)` → `contacts(company_id)`
  - `contacts(id)` → `interactions(contact_id)`
  - `companies(id)` → `deals(company_id)`
- Use explicit database transactions for state transitions to ensure atomicity.
- Avoid nullable foreign keys unless justified and documented.
- Include test fixtures and rollback scripts for each migration.

### Current Schema (4 migrations)

| Migration | Description |
|---|---|
| `000001` | Initial schema: companies, contacts, interactions, deals + state enum + updated_at triggers |
| `000002` | Add `technical_dossier` column to deals |
| `000003` | Create `pull_requests` table for AutoDev tracking |
| `000004` | Add `success` boolean to interactions (prompt optimization loop) |

## Development Guidelines

### Agent & Engine Development

- New agents must implement the internal `Agent` interface.
- Place new agents under `/pkg/agents/` or `/internal/`.
- Agents must:
  - Expose health and metrics endpoints.
  - Provide mock testing endpoints that simulate external providers.
  - Use defensive loops with exponential backoff and circuit-breaker behavior.

### Integrations

- Abstract all external communication using interfaces to allow injection of mock implementations in tests.
- Scraper engines must support headless configuration profiles (no GUI dependency).
- All external API calls should implement retry with exponential backoff.

### State Machine

- Enforce rigid, atomic state updates for all leads via transactional functions.
- Log state transitions in an audit table for compliance and debugging.

### Testing

- Unit tests for pure logic and plumbing.
- Integration tests that run against the database (use ephemeral DB instances).
- End-to-end tests must be possible using mock endpoints without calling external providers.

### Logging & Observability

- Structured logging (JSON) and correlation IDs for requests/tasks.
- Emit metrics (counts, latencies, error rates) for each agent and worker queue.

### Security

- Secrets must never be stored in plaintext in the repo; use environment variables or secret stores.
- Validate and sanitize all inputs from scraping and third-party sources.
- Webhook endpoints must verify HMAC signatures before processing.

## Testing

- Run unit tests:
  ```bash
  go test ./... -v
  ```

- Integration tests require `DATABASE_URL`:
  ```bash
  DATABASE_URL=postgres://user:pass@localhost:5432/sales_bot go test ./... -v -tags=integration
  ```

- E2E tests:
  ```bash
  DATABASE_URL=postgres://user:pass@localhost:5432/sales_bot go test ./tests/e2e/... -v
  ```

- Smoke test against a running instance:
  ```bash
  TARGET_URL="https://your-instance.com" go run scripts/smoke_test.go
  ```

## Repository Management — EXECUTIVE PROTOCOL

To preserve autonomous development and safe merges, follow the EXECUTIVE PROTOCOL:

- **Upstream Tracking:**
  - Always sync with the parent fork and update submodules recursively.
  - Use `scripts/sync_repo.sh` for automated synchronization.

- **Intelligent Merge:**
  - Use the dual-direction merge engine to reconcile feature branches with `main`.
  - Forward merge: Feature → Main
  - Reverse merge: Main → Feature (prevents drift)

- **Validation:**
  - Builds must pass merge integrity tests defined under `internal/gitcheck`.
  - Run validation before merging:
    ```bash
    ./scripts/sync_repo.sh --fetch-upstream
    go test ./internal/gitcheck/... ./internal/gitres/...
    ```

- **Automation:**
  - CI must run the full validation, linters, unit and integration tests.

## CI / Validation

The CI pipeline (`.github/workflows/ci.yml`) runs:

- Version consistency check (`VERSION` vs `VERSION.md`)
- Integrity tests (`internal/gitcheck`)
- Conflict resolution tests (`internal/gitres`)
- Full project test suite (`go test ./...`)
- Build verification (`go build ./cmd/sales_bot`)

## Known Issues

- `TestResolveConflictTheirs` in `internal/gitres/resolve_test.go` fails on Windows due to CRLF line endings. The test expects `\n` but Git writes `\r\n` on Windows.

## Contributing

- Follow the coding and branch conventions documented in `AGENTS.md`.
- All PRs must:
  - Target a feature branch and include migration files (if DB changes).
  - Include tests (unit + integration where applicable).
  - Pass `internal/gitcheck` and CI.
  - Include a clear description of agent interfaces added/changed.

## Useful Scripts & Tools

- `scripts/sync_repo.sh` — upstream sync and submodule updater (use per EXECUTIVE PROTOCOL).
- `scripts/smoke_test.go` — production health verification.
- `build.bat` / `start.bat` — Windows build and start scripts.
- `--reconcile` flag — run branch reconciliation standalone.
- `--inventory` flag — generate submodule inventory table.

## License & Contact

- License: (Insert license here)
- Maintainer: robertpelloni
- For security issues or urgent problems, create an issue in this repo and mark it as high-priority.

---

This repository is intended for production-grade autonomous orchestration. Follow the conventions above to ensure agent implementations remain safe, testable, and auditable.
