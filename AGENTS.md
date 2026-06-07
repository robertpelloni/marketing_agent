# TormentNexus Autonomous Sales Pipeline Architecture

This system is an asynchronous, event-driven orchestration layer written in Go to automate B2B lead generation, enrichment, hyper-personalized outreach, and billing for the TormentNexus repository.

## Tech Stack

- **Language:** Go 1.24 using standard concurrency paradigms (goroutines, channels) for background workers.
- **Database:** PostgreSQL 13+ with strict relational schema and atomic state transitions.
- **External APIs:** GitHub (go-github), Stripe (stripe-go), generic REST CRM.
- **Core Pattern:** Multi-agent autonomous orchestrations, task workers, and state logging.

## Module Architecture

| Package | Purpose | Key Interfaces |
|---|---|---|
| `internal/scraper` | Lead discovery from job boards & GitHub | `LeadSource` |
| `internal/enrichment` | Contact enrichment (Apollo, Hunter) | `EnrichmentSource` |
| `internal/researcher` | Technical dossier building | `Crawler`, `DossierProcessor` |
| `internal/communication` | Inbound/outbound state machine | `IntentClassifier`, `ResponseGenerator`, `SalesStrategy`, `OrderProcessor` |
| `internal/crm` | Bidirectional CRM sync | `CRMClient` |
| `internal/billing` | Stripe invoicing & payment tracking | `BillingClient` |
| `internal/sales` | Order fulfillment for won deals | `OrderDB` |
| `internal/llm` | LLM provider abstraction | `LLMProvider` |
| `internal/autodev` | Autonomous code development | `Agent` |
| `internal/deploy` | CI tracking & deployment | `CITracker`, `WorkflowDispatcher` |
| `internal/gitcheck` | Git operations & PR management | `PRManager` |
| `internal/gitres` | Intelligent merge engine | — |
| `internal/db` | PostgreSQL data layer | — |
| `internal/web` | HTTP dashboard & API | — |
| `pkg/agents` | Target discovery worker | — |
| `pkg/config` | Safety guardrails | — |

## Extension Conventions

- All new worker engines or agent subclasses must implement the internal `Agent` interface (or module-specific equivalent).
- Add new background routines to `/pkg/agents/` or `/internal/`.
- Maintain state, run logs, and target histories inside the existing database configuration layer.
- Always include explicit mock testing endpoints and defensive execution loops.
- All external integrations must be abstracted behind Go interfaces for testability and swappability.

## System Guidelines

- **State Machine:** Enforce rigid, atomic state updates for all leads in the PostgreSQL database. The 7-state lifecycle is: `Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost`.
- **Integrations:** All scraper engines must utilize headless configuration profiles. External communication modules use abstract interfaces to allow mock testing.
- **Configuration:** All environment variables should be consolidated into a typed `Config` struct loaded at startup (see Phase 6 TODO).
- **Logging:** Migrate from `log.Printf` to structured JSON logging with `slog` (see Phase 6 TODO).
- **Error Handling:** External API calls must implement retry with exponential backoff and circuit breakers (see Phase 6 TODO).

## Database Schema Constraints

All data migrations must use strict relational mappings with full foreign key constraints tracking Companies -> Contacts -> Interactions -> Deals.

### Current Tables

| Table | Purpose | Key Columns |
|---|---|---|
| `companies` | Target organizations | `domain` (UNIQUE), `tech_stack[]`, `hiring_signals[]`, `market_cap_tier` |
| `contacts` | Decision-makers | `company_id` (FK), `email` (UNIQUE), `github_handle`, `linkedin_url` |
| `interactions` | Communication log | `contact_id` (FK), `channel`, `direction`, `success` (bool) |
| `deals` | Pipeline tracking | `company_id` (FK), `current_state` (enum), `quoted_pricing`, `technical_dossier` |
| `pull_requests` | AutoDev PR tracking | `id` (PK), `branch`, `status`, `task_description` |

### Known Schema Debt

- `contacts.email` UNIQUE constraint allows multiple NULLs — needs partial index or NOT NULL.
- Missing indices on `interactions.success` and `deals.current_state` for query performance.
- No `audit_log` table for state transition history.
- No `deleted_at` soft-delete columns for GDPR compliance.

## Autonomous Development & Repository Management Protocol

The system follows a strict "EXECUTIVE PROTOCOL" for repository synchronization and intelligent merging:

- **Upstream Tracking:** Always sync with the parent fork and update all submodules recursively.
- **Intelligent Merge:** Use the dual-direction merge engine to reconcile feature branches with `main`.
  - Forward merge: Feature → Main
  - Reverse merge: Main → Feature (prevents drift)
- **Validation:** Every build must pass the merge integrity tests defined in `internal/gitcheck`.
- **Automation:** Utilize `scripts/sync_repo.sh` for automated synchronization.
- **CI Gating:** AutoDev PRs are only merged after CI passes and staging validates successfully.
