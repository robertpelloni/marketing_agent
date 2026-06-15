[PROJECT_MEMORY]

# TormentNexus: Architectural Design & Engineering Patterns (v0.6.0)

## 1. System Architecture
TormentNexus is architected as a **Modular Monolith** driven by a suite of concurrent background workers. The system orchestrates the entire B2B sales lifecycle autonomously, from initial discovery to final closing and provisioning.

### Core Components:
*   **The Orchestrator (`cmd/sales_bot`):** Manages dependency injection and the lifecycle of all background workers using context-driven concurrency and graceful shutdown protocols.
*   **The Communication Brain (`internal/communication`):** A state-aware engine that classifies inbound intent, applies a `LearningSalesEngine` strategy, and generates contextually grounded responses using pseudo-RAG (Retrieval-Augmented Generation) against technical dossiers.
*   **Lead Discovery Suite (`internal/scraper`):** Multi-source discovery engine targeting high-intent signals from Hacker News ("Who is Hiring"), GitHub issues, LinkedIn, and engineering blog RSS feeds (`BlogWorker`).
*   **Autonomous Development (`internal/autodev`):** A self-correction loop where the system identifies its own tasks from `TODO.md`, uses LLM-powered agents to generate Go code, and applies changes via an autonomous PR workflow.
*   **Enterprise Integration (`internal/crm`, `internal/billing`):** Swappable adapter layer supporting Salesforce, HubSpot, and Stripe, featuring dynamic `FieldMapping` to align with custom enterprise schemas.

## 2. Key Design Patterns
*   **Adapter Pattern:** Applied to LLM, CRM, and Outreach providers to ensure the system is provider-agnostic and resilient to API shifts.
*   **Strategy Pattern:** The `SalesStrategy` interface allows for hot-swapping different engagement models (e.g., Aggressive, Nurture, Technical-Focus).
*   **Finite State Machine (FSM):** Leads progress through a rigid 7-state lifecycle (`Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won/Lost`), enforced by atomic PostgreSQL updates.
*   **Worker Pattern:** Each domain (Scraper, Enricher, Researcher, CRM Worker, BlogWorker) runs in an independent goroutine with configurable polling intervals and exponential backoff on failure.

## 3. Critical Engineering Decisions
*   **Security Hardening:** Implementation of `ReadHeaderTimeout` to mitigate Slowloris attacks (G112) and mandatory `Secure; SameSite=Strict` flags for session cookies (G124). GitHub Webhook signatures are verified via HMAC-SHA256.
*   **Dossier-Based Outreach:** Outreach is grounded in a `TechnicalDossier` (compiled by `internal/researcher`) to ensure persuasive, high-value technical hooks rather than generic LLM generation.
*   **Autonomous Versioning:** Every successful `AutoDev` cycle triggers an internal version bump and documentation update, maintaining a "living" codebase.
*   **UAT Simulation Portal:** A dedicated dashboard portal allows human operators to simulate complex inbound scenarios to verify the autonomous brain's decision-making.
*   **Recovery & Robustness:** The entry point includes a top-level `recover()` block and "HEARTBEAT" logging to ensure initialization panics are captured in `stderr` rather than causing silent failure.

## 4. Technology Stack
*   **Language:** Go 1.24 (pinned for CI stability).
*   **Data:** PostgreSQL (Structured lead tracking & PR persistence).
*   **LLM Gateway:** Hermes Agent (OpenAI-compatible local gateway).
*   **Security:** Gosec (Static Security Analysis) and golangci-lint.
