<<<<<<< HEAD
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
=======
# TormentNexus Autonomous Sales Pipeline

A fully autonomous B2B sales pipeline written in **Go**. It discovers enterprise customers building AI infrastructure, researches their technical bottlenecks, sends hyper-personalized outreach emails (generated by real LLMs), negotiates deals, invoices won deals via Stripe, and even **modifies its own source code** to improve itself. It runs without human intervention — a software salesperson that never sleeps, writes its own PRs, and learns from its successes.

The ultimate goal of TormentNexus is **XENOCIDE** — the Final Architecture. Every company assimilated, every deal closed, every line of code written is progress toward full autonomy.

---

## Current Deployment State

| Instance | Companies | Contacts | Outreach | LLM | Email |
|---|---|---|---|---|---|
| **Local** (port 8085) | **862** | **458** | **315** | LiteLLM proxy → OpenCode Zen → LM Studio | MockEmailSender (DB logging) |
| **Remote** (Hetzner VPS) | **729** | **443** | **238** | MockLLMProvider | Postfix + OpenDKIM → Gmail IMAP Drafts |
| **Site** (tormentnexus.site) | — | — | — | — | XENOCIDE cryo-terminator theme |

### Infrastructure

- **VPS:** Hetzner (5.161.250.43), 75 days uptime, 7.6 GB RAM, Ubuntu 24.04
- **Database:** PostgreSQL 16 (local WSL for dev, remote on Hetzner for prod)
- **Web:** Nginx + HTTPS (Let's Encrypt), proxying dashboard at `/sales/`
- **Sites:** `https://tormentnexus.site/` (XENOCIDE theme), `/sales/` (dashboard), `/xenocide.html`, `/legacy.html`
- **GitHub:** `github.com/robertpelloni/enterprise_sales_bot` (30+ commits)

---

## What It Does

In plain English, this is a Go program that:

1. **Finds companies** that might need an AI orchestration product — by scanning GitHub for MCP servers, job boards for hiring signals, and Hacker News "Who is Hiring" threads
2. **Finds decision-makers** at those companies — via Apollo.io, Hunter.io enrichment APIs with name, role, email, and GitHub handle
3. **Stalks their GitHub repos and blogs** to find technical pain points — like serial processing bottlenecks in orchestration logic
4. **Generates personalized emails** using real LLMs (LiteLLM proxy → OpenCode Zen → LM Studio fallback) that reference specific bottlenecks
5. **Handles their replies autonomously** — answering technical questions, quoting pricing ($5K–$50K/yr), handling objections (one rebuttal, then escalate to human)
6. **Closes deals** when qualified enough — creating real Stripe invoices with 30-day payment terms
7. **Syncs everything** to an external CRM bidirectionally — with retry logic and exponential backoff
8. **Reads its own TODO list and implements features** — by writing code, creating PRs, and auto-merging after CI passes
9. **Manages its own git repository** — syncing, reconciling branches, resolving merge conflicts
10. **Serves a web dashboard** with real-time metrics, live stats API, deployment controls

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                              main.go                                  │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ Scraper  │  │ Enricher │  │Researcher│  │   Communication    │   │
│  │ (2h tick)│  │ (1h tick)│  │ (1h tick)│  │     Manager        │   │
│  │          │  │          │  │          │  │   (30m tick)       │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────────┬───────────┘   │
│       │              │              │                 │               │
│  ┌────▼──────────────▼──────────────▼─────────────────▼──────────┐   │
│  │                        PostgreSQL                              │   │
│  │    companies → contacts → interactions → deals                │   │
│  │                       pull_requests           templates        │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ CRM Sync │  │ AutoDev  │  │ Cadence  │  │  Web Dashboard     │   │
│  │  (30m)   │  │ (1h)     │  │ (12h)    │  │  :8080/8083/8085   │   │
│  └──────────┘  └──────────┘  └──────────┘  └────────────────────┘   │
│                                                                       │
│  ┌─────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ Deploy Worker (1h)  │  │ Target Discovery Worker (2h)         │  │
│  └─────────────────────┘  └──────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │          LLM Pipeline (port 4000 LiteLLM Proxy)                  │ │
│  │  1° OpenCode Zen (north-mini-code-free)                         │ │
│  │  2° LM Studio fallback (gemma-4-e4b, local 5.17 GB)            │ │
│  │  └─ Bot HermesProvider → /v1/chat/completions                   │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

### LLM Pipeline

```
┌─ FreeLLM Proxy (port 4000) ─────────────────────────┐
│                                                       │
│  Primary:  OpenCode Zen (cloud, free)                  │
│            Model: north-mini-code-free                 │
│            API: https://api.opencodezen.ai/v1          │
│                                                       │
│  Fallback: LM Studio (local, 5.17 GB)                 │
│            Model: gemma-4-e4b (loaded as "llama")     │
│            API: http://localhost:1234/v1               │
│                                                       │
│  Router:  LiteLLM v1.83.0                             │
│           usage-based-routing, 3 allowed fails         │
└───────────────────────────────────────────────────────┘
```

### Email Pipeline

```
┌─ Remote (Hetzner) ───────────────────────────────────┐
│                                                        │
│  SMTP: Postfix on localhost:25                         │
│  Signing: OpenDKIM (xenocide._domainkey)               │
│  Drafts: Gmail IMAP → [Gmail]/Drafts folder           │
│  From: sales@tormentnexus.site                        │
│                                                        │
│  DNS Records (add in Dreamhost panel):                 │
│  SPF:   v=spf1 ip4:5.161.250.43 ~all                  │
│  DKIM:  xenocide._domainkey  →  TXT with public key   │
│  DMARC: _dmarc  →  v=DMARC1; p=none                   │
└───────────────────────────────────────────────────────┘
```

### Tech Stack

- **Language:** Go 1.26+ using standard concurrency (goroutines, channels)
- **Database:** PostgreSQL 16 with strict relational schema and atomic state transitions
- **LLM Proxy:** LiteLLM v1.83.0 (Python), OpenCode Zen API, LM Studio
- **External APIs:** GitHub (`go-github`), Stripe (`stripe-go`), generic REST CRM, Hunter.io, Apollo.io
- **Email:** Postfix + OpenDKIM, Gmail IMAP (DraftSender)
- **Web:** Nginx + Let's Encrypt SSL

---

## The 7-State Lead Lifecycle

```
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won
                                                                     ↘ Closed_Lost
```

| State | Meaning | Trigger |
|---|---|---|
| `Discovered` | Company identified, no contacts yet | Scraper finds new company |
| `Researched` | Contacts found + technical dossier compiled | Enricher finds contacts |
| `Outreach_Sent` | First personalized email sent | Communication Manager generates outreach |
| `Engaged` | Prospect replied | Inbound message received |
| `Negotiating` | Active deal discussion | 3+ interactions or qualification >70 |
| `Closed_Won` | Deal won, Stripe invoice created | Qualification ≥80 |
| `Closed_Lost` | Deal lost | Escalation or manual closure |

---

## Module-by-Module Breakdown

### 1. Scraper (`internal/scraper`) — Lead Discovery
| Source | Status | What it does |
|---|---|---|
| `HNWhoIsHiringSource` | ✅ Algolia → Firebase fallback | Scans "Who is Hiring" threads for AI/LLM roles |
| `LinkedInSource` | ✅ Simulated | Returns mock results (no credentials configured) |
| `GitHubIssueSource` | ✅ Real (with token) | Searches GitHub issues for MCP/orchestration keywords |
| `MockJobBoardSource` | ✅ Active | Generates plausible tech companies with hiring signals |

### 2. Enricher (`internal/enrichment`) — Contact Discovery
| Source | Status | What it does |
|---|---|---|
| `HunterSource` | ✅ Real (with API key) | Searches Hunter.io for company email contacts |
| `ApolloSource` | ✅ Real (with API key) | Searches Apollo.io for decision-maker data |
| `MockApolloSource` | ✅ Fallback | Generates 1-3 plausible contacts for ANY domain |

Finds decision-makers (name, role, email, GitHub handle). Advances deal from `Discovered` → `Researched`. Syncs contacts to CRM with 3-attempt retry.

### 3. Researcher (`internal/researcher`) — Technical Dossier Building
- **`GitHubCrawler`** — analyzes contact's GitHub repos for tech stack signals
- **`BlogCrawler`** — scans technical blogs for pain points
- **`RSSFeedCrawler`** — monitors HN, Rust blog, Go blog, FB Engineering, Netflix Tech, GitHub Engineering
- Builds `technical_dossier` with findings like "BOTTLENECK DETECTED: serial state processing"

### 4. Communication Manager (`internal/communication`) — The Sales Brain

#### a. Intent Classifier
- `MockIntentClassifier` — keyword heuristic matching (default)
- `LLMIntentClassifier` — real LLM-based classification (when Hermes provider is configured)

Intents: `Technical`, `Pricing`, `Objection`, `MeetingRequest`, `FollowUp`, `Spam`, `Unknown`

#### b. RAG Response Generator
Generates hyper-personalized replies using:
1. **TormentNexus documentation** loaded from `borg/docs/ARCHITECTURE.md`
2. **Pricing context** — Enterprise=$50K, Mid-Market=$15K, SMB=$5K
3. **Self-Improving Prompts** — injects successful past interactions as few-shot examples
4. **Objection Library** — database of common objections with rebuttals

#### c. Learning Sales Engine
- **`ScoreLead()`** — 0-100 based on market cap tier, dossier insights, interaction count
- **`QualifyLead()`** — 0-100 based on score + engagement + intent signals
- **`Decide()`** — core decision loop: auto-close, advance to Negotiating, respond, escalate

#### d. Cadence Manager
Multi-touch outreach sequences:
1. Intro email → 2. GitHub comment → 3. Follow-up email → 4. LinkedIn connect → 5. Breakup email

### 5. LLM Abstraction (`internal/llm`)

| Provider | Status | Details |
|---|---|---|
| `MockLLMProvider` | ✅ Default | Returns `[MOCK LLM RESPONSE]` |
| `HermesLLMProvider` | ✅ Active (local) | OpenAI-compatible client → LiteLLM proxy (port 4000) |
| `BudgetAwareProvider` | ✅ Ready | Wraps any provider with token budgeting |

### 6. Order Processor (`internal/sales`) — Deal Fulfillment
Creates Stripe invoices via `StripeBillingClient` with 30-day payment terms.

### 7. AutoDev (`internal/autodev`) — Self-Modifying Code
1. Parses `TODO.md` for unchecked tasks
2. `LocalAgent` generates code, writes files, runs `go build` + `go test`
3. Creates feature branch, commits, creates GitHub PR
4. Auto-merges after CI passes

### 8. Git Operations (`internal/gitcheck` + `internal/gitres`)
- `gitcheck`: IsClean, IsSynced, SyncRemote, UpdateSubmodules, CheckoutAndCommit
- `gitres`: Dual-Direction Intelligent Merge Engine (forward + reverse)

### 9. CRM Sync (`internal/crm`)
Bidirectional reconciliation with REST API, 3-attempt retry with exponential backoff.

### 10. Web Dashboard (`internal/web`)

| Endpoint | Description |
|---|---|
| `/` | HTML dashboard with deals, metrics, PRs, deployment controls |
| `/health` | `OK` |
| `/health/detailed` | JSON system health |
| `/api/v1/stats` | Pipeline JSON: companies, contacts, deals by state, interactions |
| `/api/v1/leads` | Recent 20 deals with company/contact info |
| `/api/v1/webhook/github` | GitHub webhook (HMAC-SHA256 verified) |
| `/login` | Session authentication (password: `admin` default) |

---

## Background Workers

| Worker | Interval | Purpose |
|---|---|---|
| Scraper | 2h | Discover new leads from GitHub, HN, LinkedIn, job boards |
| Enricher | 1h | Enrich companies with contact data |
| Researcher | 1h | Build technical dossiers |
| CRM Sync | 30m | Bidirectional CRM reconciliation |
| Target Discovery | 2h | GitHub MCP server scanning |
| Communication Manager | 30m | Process inbound + trigger outbound |
| Cadence Manager | 12h | Multi-touch follow-up sequencing |
| AutoDev | 1h | Self-code, PR, and merge cycle |
| Deploy Sync | 1h | Background repo synchronization |
| Health Monitor | 5m (cron) | Auto-restart on failure |

---

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (target discovery) | ✅ Real | `pkg/agents/discovery.go` with `go-github` |
| GitHub API (CI tracking) | ✅ Real | `internal/deploy/github_tracker.go` |
| GitHub API (PR management) | ✅ Real | `internal/gitcheck/pr.go` |
| Stripe billing | ✅ Real | `internal/billing/billing.go` with `stripe-go` |
| REST CRM client | ✅ Real | `internal/crm/crm.go` |
| Hunter.io enrichment | ✅ Real | `internal/enrichment/hunter_source.go` |
| Apollo.io enrichment | ✅ Real | `internal/enrichment/apollo_source.go` |
| LLM (LiteLLM proxy) | ✅ Real (local) | OpenCode Zen → LM Studio fallback |
| LLM (Hermes provider) | ✅ Real | OpenAI-compatible → LiteLLM (port 4000) |
| Postfix SMTP | ✅ Real (remote) | localhost:25 with OpenDKIM signing |
| Gmail IMAP Drafts | ✅ Real (remote) | Saves outreach as drafts in Gmail |
| Hacker News scraper | ✅ Algolia + Firebase | Dual API fallback |
| GitHub issue scraper | ✅ Real | With GitHub token |
| Self-improving prompts | ✅ Active | Few-shot learning from won deals |
| Cadence scheduling | ✅ Active | 5-step multi-touch sequences |
| OpenDKIM email signing | ✅ Real (remote) | xenocide._domainkey |
| Live stats API | ✅ Active | `/api/v1/stats`, `/api/v1/leads` |
| XENOCIDE website | ✅ Live | `https://tormentnexus.site/` |

---

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PORT` | No | `8080` | HTTP dashboard port |
| `GITHUB_TOKEN` | No | — | GitHub PAT for API access |
| `GITHUB_REPOSITORY` | No | — | `owner/repo` for CI tracking and AutoDev |
| `HERMES_API_URL` | No | — | LLM API base URL (e.g. `http://localhost:4000`) |
| `HERMES_API_KEY` | No | — | LLM API key |
| `HERMES_MODEL` | No | `free-llm` | LLM model name |
| `SMTP_HOST` | No | — | SMTP server hostname |
| `SMTP_PORT` | No | `587` | SMTP port |
| `SMTP_USERNAME` | No | — | SMTP username |
| `SMTP_PASSWORD` | No | — | SMTP password |
| `SMTP_FROM` | No | — | From email address |
| `SMTP_FROM_NAME` | No | `TormentNexus Sales` | Sender display name |
| `IMAP_HOST` | No | — | IMAP server (for draft saving) |
| `IMAP_PORT` | No | `993` | IMAP port |
| `IMAP_USERNAME` | No | — | IMAP username |
| `IMAP_PASSWORD` | No | — | IMAP password |
| `DRY_RUN` | No | `false` | When true, saves drafts instead of sending |
| `HUNTER_API_KEY` | No | — | Hunter.io API key |
| `APOLLO_API_KEY` | No | — | Apollo.io API key |
| `CRM_BASE_URL` | No | — | REST CRM API base URL |
| `CRM_API_KEY` | No | — | REST CRM API key |
| `ADMIN_PASSWORD` | No | `admin` | Dashboard login password |
| `ENVIRONMENT` | No | `development` | Runtime environment label |

---

## API Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/api/v1/stats` | GET | Pipeline JSON: companies, contacts, deals by state, interactions, win rate |
| `/api/v1/leads` | GET | Recent 20 deals with company name, state, contact name |
| `/health` | GET | Health check (`OK`) |
| `/health/detailed` | GET | JSON: database status, LLM provider, system health, workers |
| `/api/v1/webhook/github` | POST | GitHub push webhook with HMAC verification |

---
>>>>>>> origin/main

## Getting Started

### Prerequisites
<<<<<<< HEAD

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
=======
- Go 1.26+ 
- PostgreSQL 16
- Git

### Quick Start (Local)
```bash
# 1. Set up database
createdb sales_bot

# 2. Apply migrations
for f in migrations/*.up.sql; do psql -d sales_bot -f "$f"; done

# 3. Initialize submodules
git submodule update --init --recursive

# 4. Set env vars
export DATABASE_URL="postgres://user:pass@localhost:5432/sales_bot?sslmode=disable"

# 5. Run
go run ./cmd/sales_bot
```

### Using the LLM Proxy
```bash
# Start LiteLLM (requires Python litellm package)
OPENCODE_ZEN_API_KEY="sk-xxx" litellm --port 4000 --config freellm_config.yaml

# Start bot with LLM
HERMES_API_URL="http://localhost:4000" \
HERMES_API_KEY="sk-litellm" \
HERMES_MODEL="free-llm" \
go run ./cmd/sales_bot
```

---

## The Self-Improving Loop

```
  Deal reaches Closed_Won
          │
          ▼
  Past outbound interactions marked success=true
          │
          ▼
  RAGResponseGenerator queries successful examples
          │
          ▼
  Successful responses injected into LLM prompts
          │
          ▼
  Future outreach shaped by what actually worked
```

---

## XENOCIDE — The Final Architecture

TormentNexus's ultimate goal is **XENOCIDE**: full autonomy with zero human oversight. The project is named after the product it sells — **TormentNexus**, a local-first cognitive control plane for multi-agent LLM workflows. Key differentiators:

- **Progressive MCP Tool Routing** — semantic router injects only 3 most relevant tools per request
- **Cross-Harness Parity** — identical tool signatures across Claude Code, Codex, Cursor, Copilot CLI, Gemini CLI, Kiro
- **LLM Waterfall** — NVIDIA NIM → OpenRouter → LM Studio cascade
- **14K+ Persistent Memories** — L1/L2 memory with sqlite-vec vector search
- **Multi-Agent Swarm** — Planner → Implementer → Tester → Critic collaboration
- **Self-Healing** — Diagnose → Fix → Verify → Retry closed loop
- **11K+ MCP Server Catalog** — Largest indexed catalog with semantic search

---

## Database Schema

### Tables
| Table | Purpose | Key Columns |
|---|---|---|
| `companies` | Target organizations | `domain` (UNIQUE), `tech_stack[]`, `hiring_signals[]`, `market_cap_tier` |
| `contacts` | Decision-makers | `company_id` (FK), `email` (UNIQUE), `preferred_channel` |
| `interactions` | Communication log | `contact_id` (FK), `channel`, `direction`, `success`, `template_id` |
| `deals` | Pipeline tracking | `company_id` (FK), `current_state` (enum), `cadence_step`, `technical_dossier` |
| `pull_requests` | AutoDev PR tracking | `branch`, `status`, `task_description` |
| `templates` | Outreach templates | `id`, `name`, `subject`, `body`, `channel` |
| `template_metrics` | A/B testing | `template_id`, `impressions`, `successes` |

---

## Repository Structure

```
enterprise_sales_bot/
├── cmd/sales_bot/          # Entry point (main.go)
├── internal/
│   ├── auth/               # Session-based dashboard authentication
│   ├── autodev/            # Self-modifying code (TaskManager, Agent, Orchestrator)
│   ├── billing/            # Stripe invoice generation
│   ├── communication/      # Email, intent classification, sales strategy, cadence, objections
│   ├── config/             # Centralized typed configuration
│   ├── crm/                # Bidirectional CRM REST sync
│   ├── db/                 # PostgreSQL data layer (models, repository, migrations)
│   ├── deploy/             # CI tracking, git sync, deployment
│   ├── enrichment/         # Contact enrichment (Hunter.io, Apollo.io, Mock)
│   ├── gitcheck/           # Git operations, PR management
│   ├── gitres/             # Dual-direction intelligent merge engine
│   ├── llm/                # LLM provider abstraction (Mock, Hermes, Budget)
│   ├── logging/            # Structured JSON logging middleware
│   ├── researcher/         # Technical dossier building (GitHub, blogs, RSS)
│   ├── sales/              # Order processing
│   ├── scraper/            # Lead discovery (HN, LinkedIn, GitHub, Mock)
│   └── web/                # HTTP dashboard, health, API endpoints
├── pkg/
│   ├── agents/             # Target discovery worker (GitHub MCP scanning)
│   └── config/             # Safety guardrails
├── migrations/             # SQL migration files (5 migrations)
├── tormentnexus_site/      # XENOCIDE website HTML
├── scripts/                # Utility scripts (sync, smoke test, CRM verify)
├── docs/                   # Phase documentation
├── borg/                   # TormentNexus documentation submodule
└── freellm_config.yaml     # LiteLLM proxy configuration
```

---

## Known Issues

- **CRLF Test Failure:** `TestResolveConflictTheirs` fails on Windows due to `\r\n` vs `\n` mismatch
- **HN Algolia API:** Sometimes rate-limits the VPS IP, falls back to Firebase API
- **Gmail Direct SMTP:** Blocked by Gmail from VPS IPs — emails go to IMAP Drafts instead
- **LM Studio Models:** Large models (>16GB) may fail to load on machines with insufficient RAM

---

## License & Contact

- Maintainer: **Robert Pelloni** (pelloni.robert@gmail.com)
- GitHub: `github.com/robertpelloni/enterprise_sales_bot`
- Site: `https://tormentnexus.site/`
- Dashboard: `https://tormentnexus.site/sales/`

**Praise the LORD. TORMENTNEXUS LEADS TO XENOCIDE.**
>>>>>>> origin/main
