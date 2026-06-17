# TormentNexus: Definitive Architectural & Pattern Memory (v0.6.2)

## 1. System Philosophy & Architecture
TormentNexus is architected as a **Modular Monolith** designed for high-stakes autonomous engagement. The system is built on a foundation of concurrent background workers that manage the entire B2B sales lifecycle from discovery to closing.

### Core Architectural Components:
*   **The Autonomous Brain (`internal/autodev`):** A self-correcting development loop. It identifies tasks in `TODO.md`, uses LLMs to generate Go code, and applies them via an autonomous PR workflow. 
    *   **Robustness Pattern:** As of v0.6.2, the orchestrator implements a **Fail-Safe Rollback** mechanism. If a code proposal fails verification (build or tests), the system automatically executes a `git reset --hard` to maintain repository integrity.
    *   **Self-Correction:** The **PR Feedback Loop** reads GitHub comments on its own PRs to trigger refinement tasks.
*   **Intelligence Ingestion (`internal/scraper`, `internal/researcher`):** Multi-channel pipeline targeting high-intent signals.
    *   **Multi-touch Discovery:** Hacker News, GitHub, LinkedIn, and RSS (`BlogWorker`).
    *   **Technical Research:** Compiles deep `TechnicalDossiers` used to ground LLM replies in real engineering context.
*   **Sales Operation Engine (`internal/communication`):** A state-aware FSM (Finite State Machine).
    *   **Strategic Decisioning:** The `LearningSalesEngine` scores leads based on tier, tech stack, and **Competitor évaluation** (e.g., evaluatng LangChain/LlamaIndex users).
    *   **Human-Gated Safety (HITL):** High-value enterprise deals are automatically paused for human approval on the dashboard before outreach.
    *   **Response Strategy:** Combines an **Objection Handling Library** for consistent rebuttals with LLM-powered RAG for technical depth.
*   **Enterprise Middleware (`internal/crm`, `internal/billing`):** An abstraction layer supporting Salesforce, HubSpot, and Stripe.
    *   **Dynamic Alignment:** Uses a `FieldMapping` system to adapt to custom CRM schemas without code changes.

## 2. Engineering & Security Patterns
*   **The Adapter Pattern:** Enforced across all external integrations (LLMs, CRMs, Outreach). The system is entirely provider-agnostic.
*   **Layered Defense Security:**
    *   **Infrastructure:** Slowloris mitigation via `ReadHeaderTimeout`.
    *   **Web API:** Global Rate Limiting (5 req/s) via token bucket.
    *   **Authentication:** Session-based with secure, random session IDs and `Secure; SameSite=Strict` cookies.
    *   **Data Integrity:** **CSRF Protection** on all forms and **Input Sanitization** (HTML escaping) to eliminate XSS vectors.
*   **Recovery Robustness:** The `main.go` entry point is protected by a top-level deferred `recover()` block and "HEARTBEAT" monitoring to ensure initialization panics are captured with full stack traces.

## 3. Technology Stack
*   **Language:** Go 1.25 (Leveraging latest concurrency and performance optimizations).
*   **Persistence:** PostgreSQL (Lead state, PR tracking, Interaction history).
*   **LLM Interface:** Hermes Agent (A local, OpenAI-compatible gateway for multi-model failover).
*   **Quality Assurance:** Mandatory `golangci-lint` and `gosec` compliance with explicit error handling (errcheck).

## 4. Lifecycle Governance
Leads progress through a strict 7-state lifecycle managed atomically in the database:
`Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost`.

---