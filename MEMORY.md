# Project Memory: TormentNexus (v0.9.0)

## 1. Architecture & Design Decisions
TormentNexus is a high-performance **Modular Monolith** built in **Go 1.25.0**. It is designed as an autonomous engine for the entire B2B sales lifecycle.

*   **State-Driven Pipeline:** Lead progress is managed via a strict Finite State Machine (FSM) in PostgreSQL.
*   **Autonomous Development (AutoDev):** A self-correcting loop that generates code from `TODO.md` via LLM, featuring concurrent execution, a rollback mechanism (`git reset --hard`), and a PR feedback loop.
*   **Multi-Channel Outreach:** Integrated support for Email (SMTP/IMAP), LinkedIn (simulated), and GitHub (comment-based outreach). Governed by a `CadenceAwareManager` that respects contact preferences.

## 2. Integrated Enterprise Features
*   **GDPR Compliance:** Native data portability and soft-delete (`deleted_at`) support.
*   **Security Hardening:** AES-GCM encryption for secrets at rest, global rate limiting (5 req/s), CSRF protection, and Slowloris mitigation (`ReadHeaderTimeout`).
*   **Observability:** Structured JSON logging via `slog`, business-focused Prometheus metrics, and worker performance profiling.
*   **Phase 10 Webhooks:** Outbound state change notifications with built-in retry logic and standardized JSON payloads.

## 3. Engineering Patterns
*   **Adapter Pattern:** Used for CRM (Salesforce/HubSpot) and LLM (Hermes Agent) integrations to remain provider-agnostic.
*   **Strategy Pattern:** Powering the `LearningSalesEngine` to determine the best next action based on lead quality and intent.
*   **State Machine:** Enforcing lead transitions: `Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Pending_Approval → Closed_Won/Lost`.

## 4. Conflict Resolution & Merging (v0.9.0)
The v0.9.0 release involved a complex merge of enterprise hardening features with sophisticated upstream outreach logic (A/B testing, objection handling). Import cycles between `communication` and `sales` were resolved by decoupling `SentimentResult`. Redundant code (e.g., `CalculateQuote`) was consolidated to ensure a single source of truth.

## 5. Development Environment
*   **CI/CD:** Enforced integration testing with ephemeral PostgreSQL.
*   **Deployment:** Standardized Docker and Docker Compose workflows.
*   **Executive Protocol:** Strict repo synchronization (Upstream Sync → Branch Merging → Catch-Up Sync → Submodule Cleanup).
