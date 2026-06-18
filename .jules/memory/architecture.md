# TormentNexus: Authoritative Architectural & Engineering Memory (v0.9.0)

## 1. System Vision & Architecture
TormentNexus is a high-performance **Modular Monolith** for autonomous B2B sales. It uses a strict state machine in PostgreSQL: `Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Pending_Approval → Closed_Won/Lost`.

## 2. Core Modules
*   **AutoDev:** Self-correcting LLM loop with verification suites and PR feedback integration. Supports concurrent task execution and task dependency resolution.
*   **Communication:** `CadenceAwareManager` with multi-channel touches (Email, LinkedIn, GitHub), objection handling, and A/B template testing.
*   **Intelligence:** `SentimentAnalyzer` and `ForecastingEngine` for win-probability and revenue predictions.
*   **Compliance:** GDPR-ready (export/delete endpoints) with soft-delete logic.

## 3. Engineering Patterns
*   **Security:** AES-GCM encryption, global rate limiting (5 req/s), CSRF protection, and Slowloris mitigation.
*   **Resilience:** Exponential backoff retries and circuit breakers for external APIs.
*   **Observability:** Structured JSON logging (`slog`), Prometheus metrics, and `pprof`.

---

I have resolved the import cycle and am now fixing a build error in `responder.go` where `CalculateQuote` was being referenced via the `sales` package but is actually declared in the `communication` package. I will also clean up unused imports in `engine.go` to ensure a clean build.