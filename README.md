# Borg Autonomous Sales Pipeline Architecture

An asynchronous, event-driven orchestration layer in Go for automated B2B lead generation, enrichment, hyper-personalized outreach, and billing.

This repository implements multi-agent orchestration, background worker engines, and strict state management to run reliable autonomous sales pipelines for enterprise workflows.

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
Borg is a modular Go-based system that runs concurrent agent workers to:
- Discover and scrape target companies and contacts (headless scraper engines).
- Enrich contact and company data with third-party providers.
- Run hyper-personalized outreach workflows across channels.
- Track interactions, deals, and billing with strict relational state persisted in PostgreSQL.

Key design goals:
- Concurrency-first: goroutines and channels for background workers.
- Deterministic state machine: atomic state transitions for leads and deals.
- Testability: abstract external integrations behind interfaces and include explicit mock endpoints.
- Autonomous development: tools & scripts that maintain sync with upstream forks and validate merges.

## Features
- Multi-agent orchestration and pluggable worker engines.
- Atomic state machine for lead lifecycle (Company -> Contact -> Interaction -> Deal).
- Headless browser scraper profiles for deterministic scraping.
- External integrations via interfaces to allow mocking in tests.
- Comprehensive logs and target histories persisted to PostgreSQL.
- Defensive execution loops and mock testing endpoints for each agent.

## Architecture & Conventions
- Language: Go (primary), small PL/pgSQL for DB functions.
- Agents live in `/pkg/agents/` or `/internal/`.
- New engines must implement the `Agent` interface (or the module-specific equivalent).
- All background routines and workers must:
  - Implement graceful shutdown and health probes.
  - Include defensive execution loops (backoff, jitter, circuit-breakers).
  - Provide explicit mock endpoints for testing.
- Database mapping follows strict relational models:
  - Companies -> Contacts -> Interactions -> Deals with full foreign keys.
  - All updates to lead state must be performed via atomic transactions to enforce state machine invariants.

Recommended repo layout
- /cmd/               - CLI binaries / worker entrypoints
- /pkg/agents/        - Agent implementations and worker engines
- /internal/          - Internal frameworks, `gitcheck`, helpers
- /migrations/        - SQL migrations (strict relational schemas)
- /scripts/           - automation scripts (eg. `sync_repo.sh`)
- /docs/              - design docs and runbooks
- /test/              - integration & mock servers

## Getting Started

Prerequisites
- Go 1.20+ (or the version pinned in go.mod)
- PostgreSQL
- (Optional) Docker & Docker Compose for local stacks
- Headless browser runtime (Chromium/Chrome) for scraper engines

Quick start (local)
1. Set environment variables (example):
   - `DATABASE_URL=postgres://user:pass@localhost:5432/borg?sslmode=disable`
   - `PORT=8080`
   - `HEADLESS_CHROME_PATH=/usr/bin/chromium`
   - `GIT_REMOTE_UPSTREAM=git@github.com:parent/fork.git`

2. Run migrations:
   - Example (if using plain SQL migrations):
     psql $DATABASE_URL -f migrations/001_init.sql
   - Or run your migration tool of choice pointing to the `migrations/` folder.

3. Build and run:
   - go build ./cmd/...
   - ./bin/your-worker-binary

Or with Docker Compose (if provided):
- docker compose up --build

## Configuration
Use environment variables to configure runtime behavior. Minimal set:
- DATABASE_URL
- PORT
- HEADLESS_CHROME_PATH
- LOG_LEVEL (info/debug)
- SCRAPING_PROFILE (path or name of headless profile)
- MOCK_MODE (true/false) — enables mock endpoints for integration tests

Agents and integrations are configured through typed configuration structs and dependency injection; avoid hard-coded credentials.

## Database & Migrations
- All schema changes MUST be delivered through the `migrations/` folder and reviewed as part of PRs.
- Enforce relational integrity with explicit foreign keys:
  - `companies(id)` -> `contacts(company_id)`
  - `contacts(id)` -> `interactions(contact_id)`
  - `interactions(id)` -> `deals(interaction_id)`
- Use explicit database transactions for state transitions to ensure atomicity.
- Avoid nullable foreign keys unless justified and documented.
- Include test fixtures and rollback scripts for each migration.

## Development Guidelines

Agent & Engine Development
- New agents must implement the internal `Agent` interface.
- Place new agents under `/pkg/agents/` or `/internal/`.
- Agents must:
  - Expose health and metrics endpoints.
  - Provide mock testing endpoints (e.g., `/internal/mock/agent-name`) that simulate external providers.
  - Use defensive loops with exponential backoff and circuit-breaker behavior.

Integrations
- Abstract all external communication using interfaces to allow injection of mock implementations in tests.
- Scraper engines must support headless configuration profiles (no GUI dependency) and be configurable via SCRAPING_PROFILE.

State Machine
- Enforce rigid, atomic state updates for all leads via transactional functions.
- Log state transitions in a `lead_state_changes` or similar audit table.

Testing
- Unit tests for pure logic and plumbing.
- Integration tests that run against the database (use ephemeral DB instances).
- End-to-end tests must be possible using mock endpoints without calling external providers.

Logging & Observability
- Structured logging (JSON) and correlation IDs for requests/tasks.
- Emit metrics (counts, latencies, error rates) for each agent and worker queue.

Security
- Secrets must never be stored in plaintext in the repo; use environment variables or secret stores.
- Validate and sanitize all inputs from scraping and third-party sources.

## Testing
- Run unit tests:
  - go test ./... -v
- Integration tests:
  - Use `MOCK_MODE=true` to route integrations to local mock endpoints.
- Provide explicit mock endpoints under `/internal/mock/` for each external provider and agent.

## Repository Management — EXECUTIVE PROTOCOL
To preserve autonomous development and safe merges, follow the EXECUTIVE PROTOCOL:
- Upstream Tracking:
  - Always sync with the parent fork and update submodules recursively.
  - Use `scripts/sync_repo.sh` for automated synchronization.
- Intelligent Merge:
  - Use the dual-direction merge engine to reconcile feature branches with `main`.
- Validation:
  - Builds must pass merge integrity tests defined under `internal/gitcheck`.
  - Run validation before merging:
    - ./scripts/sync_repo.sh --fetch-upstream
    - make validate (or run `internal/gitcheck` tooling)
- Automation:
  - CI must run the full validation, linters, unit and integration tests.

Example sync command
```bash
# update fork, submodules, and run gitcheck
scripts/sync_repo.sh && go test ./... && internal/gitcheck
```

## CI / Validation
- CI pipeline must run:
  - Static analysis / linters
  - go vet / go fmt check
  - Unit tests and integration tests
  - `internal/gitcheck` validations and merge integrity checks
- Fail early on schema drift or migration conflicts.

## Contributing
- Follow the coding and branch conventions documented in `CONTRIBUTING.md` (if present).
- All PRs must:
  - Target a feature branch and include migration files (if DB changes).
  - Include tests (unit + integration where applicable).
  - Pass `internal/gitcheck` and CI.
  - Include a clear description of agent interfaces added/changed.

## Useful scripts & tools
- scripts/sync_repo.sh — upstream sync and submodule updater (use per EXECUTIVE PROTOCOL).
- internal/gitcheck — merge integrity and repository validation tool.
- migrations/ — SQL migrations folder.

## Troubleshooting
- If a merge fails integrity checks, run:
  - scripts/sync_repo.sh --debug
  - Inspect `internal/gitcheck` output and resolve conflicts locally.
- For DB migration issues, run migrations against a staging DB and review foreign key failures.

## License & Contact
- License: (Insert license here)
- Maintainer: robertpelloni
- For security issues or urgent problems, create an issue in this repo and mark it as high-priority.

---

This repository is intended for production-grade autonomous orchestration. Follow the conventions above to ensure agent implementations remain safe, testable, and auditable.
