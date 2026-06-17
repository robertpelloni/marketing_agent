[PROJECT_MEMORY]

# TormentNexus: Architectural Design & Engineering Patterns (v0.8.0)

## 1. System Philosophy: Autonomous Modular Monolith
TormentNexus is architected as a high-performance **Modular Monolith** driven by concurrent background workers. The system manages the entire B2B sales lifecycle autonomously, from discovery and enrichment to technical outreach and fulfillment.

### Core Architectural Layers:
*   **Autonomous Orchestration (\`internal/autodev\`):** The "meta-brain" of the system. As of v0.8.0, it supports **Concurrent Task Execution** via goroutines and **Task Dependency Resolution** (parsing \`DependsOn\` metadata from TODO.md). It handles self-correction through a **PR Feedback Loop** and features an **Automated Rollback** mechanism for failed verifications.
*   **Intelligent Communication (\`internal/communication\`):** A state-aware engine managing a rigid 7-state FSM. It utilizes LLMs for intent classification and technical response generation, grounded by a **RAG-based Technical Dossier**.
*   **Lead Discovery & Intelligence (\`internal/scraper\`, \`internal/enrichment\`):** Multi-source pipeline targeting high-intent signals from HN, GitHub, LinkedIn, and engineering blogs (\`BlogWorker\`). It includes **Competitor Tracking** to refine lead scoring.
*   **Integration & Observability (\`internal/crm\`, \`internal/web\`):** Adapter-based support for Salesforce, HubSpot, and Stripe. v0.8.0 introduces **Outbound Webhooks** for real-time status syncing and a comprehensive **REST API v1** for external management.

## 2. Key Design Patterns
*   **Adapter Pattern:** Ensures provider-agnosticism across all external services.
*   **Finite State Machine (FSM):** Enforces atomic lead progression: \`Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won/Lost\`.
*   **Human-in-the-Loop (HITL) Gate:** A safety pattern for high-value enterprise accounts, requiring manual approval on the dashboard before outreach.
*   **Worker Pattern:** Decoupled background services running with context-aware concurrency and graceful shutdown.

## 3. Security & Robustness Decisions
*   **Layered Defense:** Global rate limiting (5 req/s), CSRF protection on all forms, input sanitization (XSS mitigation), and HMAC webhook verification.
*   **Operational Integrity:** HEARTBEAT monitoring and top-level panic recovery in the entry point. Automated \`git reset --hard\` on verification failure in the dev loop.
*   **Performance Profiling:** Real-time cycle duration tracking for background workers displayed on the dashboard.

## 4. Technology Stack
*   **Language:** Go 1.25 (Leveraging latest performance and concurrency features).
*   **Data:** PostgreSQL (Structured persistence for leads, PRs, and metrics).
*   **LLM Gateway:** Hermes Agent (OpenAI-compatible local gateway).
