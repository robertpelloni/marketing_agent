<<<<<<< HEAD
[PROJECT_MEMORY]

# TormentNexus: Architectural Design & Engineering Patterns (v0.9.0)

## 1. Modular Monolith Architecture
TormentNexus is a high-performance **Modular Monolith** driven by concurrent background workers. The system manages the entire B2B sales lifecycle autonomously, from discovery and enrichment to technical outreach and fulfillment.

### Core Architectural Layers:
*   **Autonomous Orchestration (\`internal/autodev\`):** Meta-brain supporting **Concurrent Task Execution** and **Task Dependency Resolution**. Features a **PR Feedback Loop** and **Automated Rollback**.
*   **Intelligence & Discovery (\`internal/scraper\`, \`internal/enrichment\`):** Multi-channel ingestion engine (HN, GitHub, LinkedIn, RSS) with **Competitor tracking**.
*   **Strategic Communication (\`internal/communication\`):** State-aware FSM engine utilizing RAG and an **Objection Handling Library**. Features **Human-in-the-Loop (HITL) approval** for enterprise deals and **Prompt Analytics**.
*   **Enterprise Platform & Security (\`internal/web\`, \`internal/security\`):** Comprehensive REST API v1, **GDPR compliance endpoints**, and **Secrets Encryption (AES-GCM)**.

## 2. Key Security & Compliance Patterns
*   **Layered Defense:** Global rate limiting (5 req/s), **Webhook IP Allowlisting**, CSRF protection, input sanitization (XSS mitigation), and HMAC verification.
*   **Data Privacy:** Native support for **GDPR Export and Soft-Delete** via \`deleted_at\` columns.
*   **Infrastructure Security:** Slowloris mitigation (\`ReadHeaderTimeout\`) and hardened cookie flags.

## 3. Operational Robustness
*   **Recovery-Ready:** HEARTBEAT monitoring and top-level panic recovery with full stack traces in the entry point.
*   **Performance Monitoring:** Real-time worker performance profiling displayed on the dashboard.
*   **State Integrity:** Atomic lead state progression and automated git state recovery in the autonomous development loop.

## 4. Technology Stack
*   **Language:** Go 1.25.
*   **Persistence:** PostgreSQL.
*   **Gateway:** Hermes Agent.
=======
# Memory: Architectural Observations & Design Preferences

## Current State

<<<<<<< HEAD
- The project is at v0.4.8 with Hermes Agent LLM integration.
>>>>>>> origin/main
- Core modules: scraper, enricher, researcher, communication, CRM, billing, deploy, autodev.
- **LLM provider is now REAL** via Hermes Agent gateway (WSL) — replaces MockLLMProvider.
- **Intent classifier is now REAL** via LLMIntentClassifier when Hermes is available.
- Remaining mocks: enrichment (Apollo), job board scraper, email send/receive, billing (Stripe).
=======
>>>>>>> origin/main
- The project has completed Phase 5 (v0.4.1) with a fully functional end-to-end pipeline.
- All core modules are implemented: scraper, enricher, researcher, communication, CRM, billing, deploy, and autodev.
- The project uses Go 1.24 and follows standard Golang concurrency patterns.
- All external integrations currently use mock implementations (except GitHub API for target discovery and CI tracking).
- A robust merge integrity and conflict resolution testing framework is in place.
<<<<<<< HEAD
- The project was rebranded from TormentNexus to TormentNexus across all product-facing references.
=======
- The project was rebranded from Borg to TormentNexus across all product-facing references.
<<<<<<< HEAD
=======
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080

## Architectural Traits

- **Event-Driven:** Designed to be asynchronous and event-driven via background worker goroutines.
- **Interface-Based:** External integrations (scrapers, CRM, billing, LLM, email) are abstracted behind interfaces for easier mocking and rotation.
- **Rigid State Management:** Lead transitions are handled via an atomic state machine in PostgreSQL with a 7-state enum.
>>>>>>> origin/main
- **Automation First:** Every feature is built with the intent of being fully autonomous.
- **Self-Development Loop:** The system includes an `autodev` module that autonomously selects tasks from `TODO.md`, proposes changes, and verifies them via a branch-push-PR-merge lifecycle.
- **Autonomous Continuous Delivery:** Codebase updates initiated by the bot trigger automated GitHub Action workflows for testing and deployment to ensure system stability.
- **Self-Learning Sales Engine:** The `communication` package features a `LearningSalesEngine` that analyzes interaction history and lead context to decide on autonomous responses, state transitions, or human escalation.
<<<<<<< HEAD

## Design Preferences
=======
- **Prompt Optimization Feedback Loop:** The `RAGResponseGenerator` implements a feedback loop by injecting successful past interactions (flagged upon `StateClosedWon`) into the prompt context.
- **Dual-Direction Merge Engine:** Reconciles autonomous feature branches by forward-merging into main and reverse-merging main back into features to prevent drift.

## Known Technical Debt

- **CRLF Test Failure:** `internal/gitres/resolve_test.go::TestResolveConflictTheirs` fails on Windows due to `\r\n` vs `\n` mismatch.
<<<<<<< HEAD
- **Scattered Configuration:** `os.Getenv()` calls are scattered throughout `main.go` and packages instead of a centralized config struct.
- **Unstructured Logging:** All modules use `log.Printf` — no structured JSON logging, no log levels, no correlation IDs.
- **No Graceful Shutdown:** Workers respond to context cancellation but do not drain in-flight work before exiting.
- **No Connection Pooling:** `db.NewDB()` opens a connection with default pool settings (no max open/idle/lifetime).
- **No Retry/Backoff:** External API calls (GitHub, CRM, Stripe) fail immediately without retry.
- **No DB Migration Runner:** Migrations must be applied manually; they are not auto-applied on startup.
- **No Dashboard Auth:** The web dashboard is completely unauthenticated.
- **No Rate Limiting:** HTTP endpoints accept unlimited requests.
- **No Pagination:** Dashboard hardcodes `LIMIT 20` for deals.
- **Missing Indices:** `interactions.success` and `deals.current_state` lack database indices for query performance.
- **Hardcoded Worker Intervals:** All background worker intervals are hardcoded in `main.go` rather than configurable.
=======
- **Unstructured Logging:** All modules use `log.Printf` — no structured JSON logging, no log levels, no correlation IDs.
- **No DB Migration Runner:** Migrations must be applied manually; they are not auto-applied on startup.
- **No Rate Limiting:** HTTP endpoints accept unlimited requests.
- **No Pagination:** Dashboard hardcodes `LIMIT 20` for deals.
- **Missing Indices:** `interactions.success` and `deals.current_state` lack database indices.
- **Hardcoded Worker Intervals:** Background worker intervals are configurable via env vars but not via config file.
- **Hermes Dependency:** LLM calls depend on Hermes gateway being running in WSL. If Hermes is down, bot falls back to mock.
>>>>>>> origin/main

## Design Preferences

>>>>>>> origin/main
- **Go (Golang):** Preferred for the orchestration layer due to its performance and concurrency model.
- **PostgreSQL:** Used for reliable relational data storage and state tracking.
- **Headless Scrapers:** Required for robust data extraction from modern web platforms.
- **Atomic Commits:** Prefer small, descriptive commits that correspond to specific features or fixes.
<<<<<<< HEAD
=======
- **Interface-Driven Design:** All external dependencies should be behind Go interfaces for testability and swappability.
- **CI-Gated Merging:** No code reaches main without passing all tests.

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (target discovery) | ✅ Real | `pkg/agents/discovery.go` with `go-github` |
| GitHub API (CI tracking) | ✅ Real | `internal/deploy/github_tracker.go` |
| GitHub API (PR management) | ✅ Real | `internal/gitcheck/pr.go` with `go-github` |
| Stripe billing | ✅ Real | `internal/billing/billing.go` with `stripe-go` |
| REST CRM client | ✅ Real | `internal/crm/crm.go` with generic REST |
<<<<<<< HEAD
| LLM provider | ❌ Mock | `internal/llm/llm.go::MockLLMProvider` |
| Intent classifier | ⚠️ Hybrid | `MockIntentClassifier` + `LLMIntentClassifier` (exists but mock is default) |
=======
| **LLM provider** | **✅ Real** | **`internal/llm/hermes.go::HermesLLMProvider` via Hermes Agent gateway** |
| **Intent classifier** | **✅ Real** | **`LLMIntentClassifier` via Hermes (keyword mock fallback)** |
>>>>>>> origin/main
| Enrichment (Apollo) | ❌ Mock | `internal/enrichment/worker.go::MockApolloSource` |
| Job board scraper | ❌ Mock | `internal/scraper/scraper.go::MockJobBoardSource` |
| Email sending | ❌ Not implemented | Outbound is logged but not sent |
| Email receiving | ❌ Not implemented | Inbound is simulated by polling DB |
>>>>>>> origin/main
