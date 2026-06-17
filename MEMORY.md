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
