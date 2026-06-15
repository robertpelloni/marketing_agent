# Changelog

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2026-06-15

### Added
- **LLM-Powered Autonomous Development:**
    - Upgraded `autodev.LocalAgent` to use LLM-powered Go code generation.
    - Integrated `llm.LLMProvider` into the autonomous dev loop for dynamic solution proposing.
- **UAT Simulation Portal:**
    - Integrated an interactive User Acceptance Testing (UAT) portal into the web dashboard (`/api/v1/test/simulate_inbound`).
    - Enables real-time simulation of inbound messages and verification of autonomous brain replies.
- **Multi-Channel Sales Engagement:**
    - Implemented `GitHubSender` for technical outreach via repository issue comments.
    - Scaffolded `LinkedInSender` for cross-platform professional engagement.
- **Lead Discovery Intelligence:**
    - Implemented `BlogWorker` for RSS/Atom ingestion of technical engineering blogs to detect hiring/innovation signals.
    - Enhanced `LearningSalesEngine` with competitor detection scoring (LangChain, LlamaIndex, etc.).
- **Enterprise CRM Field Mapping:**
    - Added `FieldMapping` and `FieldMappingSetter` to CRM clients (Salesforce, HubSpot) to support custom enterprise property schemas.
- **Security & CI Hardening:**
    - Resolved all `gosec` security vulnerabilities including Slowloris mitigations (G112) and Secure/SameSite cookie hardening (G124).
    - Applied explicit error handling (errcheck) and removed all unused code to satisfy strict `golangci-lint` CI requirements.


### Added
- **Enrichment Source Fallback Chain:**
    - Implemented `FallbackSource` in `internal/enrichment/fallback.go`.
    - Wraps multiple `EnrichmentSource` instances (Hunter.io, Apollo.io, Mock) and tries each in order.
    - Logs clear pass/fail indicators per source with structured status reporting.
    - Respects context cancellation mid-chain.
    - Exposes `Status()`, `Sources()`, and `Names()` for observability and testing.
    - Integrated into `cmd/sales_bot/main.go` replacing flat source iteration.
    - Added comprehensive unit tests (8 test cases across 3 test functions) — all passing.

- **GitHub Repository Analysis:**
    - New `scraper.GitHubAnalyzer` analyzes public repos for tech stack, languages, topics, activity, and infrastructure patterns.
    - Detects bottlenecks (high open issues, popular-but-inactive repos).
    - Identifies technologies from topics and descriptions across 12 categories (AI/ML, LLM, Kubernetes, Go, Python, Rust, etc.).
    - Generates `RepoAnalysis` with `InsightSummary` for personalized outreach hooks.
    - Supports both org and user repository analysis.
    - Comprehensive unit tests for insight generation and helper methods.

- **Deal Forecasting:**
    - New `sales.ForecastingEngine` that predicts win probability using historical patterns.
    - Combines source win rate, stage baseline, time-in-stage penalty, interaction sentiment, and engagement quantity.
    - Generates risk factors (stalled progress, negative sentiment, low engagement, low source win rate).
    - `PipelineSummary` aggregates forecasts across all deals with at-risk detection.
    - `PercentileForecast()` computes P10/P50/P90 revenue ranges for pipeline.
    - Learns from closed deals to improve accuracy over time.
    - Comprehensive unit tests for healthy and at-risk scenarios.

- **Interaction Sentiment Analysis:**
    - New `communication.SentimentAnalyzer` with heuristic keyword-based classification.
    - Detects positive/negative/neutral/mixed sentiment with confidence scoring (-100 to +100).
    - Urgency detection for time-sensitive replies.
    - Optional LLM-assisted refinement for deeper semantic analysis.
    - `AggregateDealSentiment()` combines multiple interactions into per-deal trends.
    - Generates context-aware next-action recommendations per sentiment class.
    - Comprehensive unit tests (8 tests) covering all sentiment types, urgency, and aggregation.

- **Prompt Versioning & A/B Testing:**
    - New `llm.PromptRegistry` for managing prompt templates with version tracking.
    - `RegisterVersion()`, `GetActiveVersion()`, `ResolvePrompt()` with `${key}` placeholder interpolation.
    - `AssignExperiment()` configures weighted A/B experiments across prompt versions.
    - `RecordOutcome()` tracks success/failure per variant for analytics.
    - JSON persistence at `data/prompt_registry.json` for state across restarts.
    - Comprehensive unit tests for registration, selection, rendering, outcomes, and persistence.

- **Token Budget Tracking:**
    - New `llm.TokenBudget` with configurable budget, reset interval, warning threshold, and exceeded callback.
    - `RecordUsage()` logs tokens and checks budget; `IsWithinBudget()`, `GetUsage()`, `ShouldWarn()`, `EstimatedCost()` for observability.
    - `DealTokenTracker` tracks per-deal token consumption.
    - `BudgetAwareProvider` wraps any `LLMProvider` with automatic budget enforcement returning `ErrBudgetExceeded` when exhausted.
    - Comprehensive unit tests for budget, exceed, and provider wrapping.

- **GitHub Issue/PR Comment Outreach:**
    - New `communication.GitHubCommentSender` that searches for relevant issues/PRs in a target org and posts a technical hook comment.
    - Includes `SendComment`, `SearchRelevantIssues`, and `FindAndComment` helpers.
    - Simulated placeholder comment generation (`GenerateTechHookComment`).
    - Uses `go‑github` client; respects rate‑limiting; logs actions.

- **LinkedIn Message Sending (simulation placeholder):**
    - New `communication.LinkedInSender` with `Send`, `HealthCheck`, and connection‑request stubs.
    - Works with `LINKEDIN_USERNAME`/`LINKEDIN_PASSWORD` env vars.
    - Currently logs simulated messages; ready for headless‑browser automation (rod/chromedp).

- **Outreach Cadence Management:**
    - Added `cadence.go` defining `CadenceStep`, `CadenceSchedule`, `CadenceTracker`, and `CadenceAwareManager`.
    - Provides a default 5‑touch multi‑channel schedule (email → GitHub → email → LinkedIn → email).
    - `CadenceAwareManager` runs a periodic scheduler that decides when to trigger the next touch based on interaction history.
    - Integrated with the existing `communication.Manager` via composition.

- **Response Quality Scoring:**
    - New `communication.QualityScorer` with heuristic + optional LLM-assisted evaluation.
    - Scores messages 0–100 on personalization, relevance, CTA presence, tone, and length.
    - `Evaluate()` returns `QualityScore` with issues and suggestions.
    - `ScoreAndLog()` logs pass/fail with detailed breakdown.
    - Configurable minimum threshold (default 60) to block low-quality outreach.

- **Salesforce CRM Adapter:**
    - New `crm.SalesforceClient` implementing `CRMClient` (push deals, lead updates, account validation, sync contacts/interactions, fetch deal details).
    - Uses env vars `SALESFORCE_INSTANCE_URL`, `SALESFORCE_ACCESS_TOKEN`, `SALESFORCE_API_VERSION`.
    - Placeholder mapping functions for lead‑state conversions.

- **HubSpot CRM Adapter:**
    - New `crm.HubSpotClient` implementing `CRMClient` (push deals, lead updates, account validation, sync contacts/interactions, fetch deal details).
    - Configured via `HUBSPOT_BASE_URL`, `HUBSPOT_API_KEY` or `HUBSPOT_ACCESS_TOKEN`.
    - Includes helper functions for converting Salesforce‑style states.

- **Channel Preference per Contact:**
    - Added `preferred_channel` column to `contacts` table (migration `000005`).
    - Extended `Contact` model with `PreferredChannel` field and introduced `db.Channel` type with constants (`ChannelEmail`, `ChannelLinkedIn`, `ChannelGitHub`) and helper methods (`DefaultChannel`, `IsValid`, `String`).
    - Updated `CreateContact`, `ListContactsByCompany`, `GetContactByEmail` to include `preferred_channel`; added `UpdateContactPreferredChannel` method.
    - Communication Manager now respects contact channel preference via `DefaultChannelForContact()` helper, using it for inbound/outbound interaction channel tagging and sender routing.
    - Web dashboard displays contact channel preference as an inline dropdown (Email/LinkedIn/GitHub) with auto‑submit on change via new `update_channel` POST handler.

- **LinkedIn Sales Navigator Scraper:**
    - Implemented `LinkedInSource` in `internal/scraper/linkedin_source.go` implementing `LeadSource` interface.
    - Configurable via `LINKEDIN_USERNAME`/`LINKEDIN_PASSWORD` environment variables.
    - Includes `SetTargetTitles()` for configurable job title filtering (CTO, VP Engineering, Lead Developer, etc.).
    - `HealthCheck()` validates credential presence and configuration.
    - Simulation fallback returns high‑value AI/ML targets when credentials are not configured.
    - Integrated into `cmd/sales_bot/main.go` scraper source list alongside HN and GitHub sources.
    - Comprehensive unit tests (discovery, health check, credential configuration, title configuration) — all passing.
    - Designed for future headless browser automation (placeholder for rod/chromedp).

## [0.4.9] - 2026-06-10

### Added
- **Hacker News "Who is Hiring" Lead Discovery:**
    - Implemented `HNWhoIsHiringSource` in `internal/scraper/hn_source.go`.
    - Scrapes HN Algolia API for latest "Who is Hiring" threads.
    - Parses 200+ top-level comments per thread for company name, domain, tech stack.
    - Filters for AI/LLM relevance using 30+ keyword patterns.
    - Deduplicates by domain and classifies market cap tier from posting context.

- **Hunter.io Email Enrichment:**
    - Implemented `HunterSource` in `internal/enrichment/hunter.go`.
    - Calls Hunter.io domain search API to find professional email addresses.
    - Filters results for decision-makers (VP, Director, CTO, Lead, Architect, etc.).
    - Health check verifies API key validity.

- **SMTP Email Sending:**
    - Implemented `SMTPSender` in `internal/communication/smtp_sender.go`.
    - Supports STARTTLS (port 587) and direct SSL (port 465).
    - Builds RFC 5322 compliant messages with proper headers.
    - Health check verifies SMTP connectivity and authentication.
    - `MockEmailSender` for testing without sending.

- **IMAP Email Receiving:**
    - Implemented `EmailReceiver` in `internal/communication/imap_receiver.go`.
    - Polls IMAP inbox for unread messages at configurable interval.
    - Parses inbound emails and matches sender to contacts in database.
    - Feeds matched emails into the Communication Manager's inbound pipeline.
    - Tracks last processed UID to avoid reprocessing.

- **Communication Manager Email Integration:**
    - `Manager` now accepts optional `EmailSender` — sends real emails after persisting outbound interactions.
    - `NewManager()` signature updated to accept `EmailSender` (nil = log-only mode).
    - Added `GetDB()` method for IMAP receiver contact lookup.

- **Config Extensions:**
    - Added `HunterAPIKey`, SMTP fields (`SMTPHost/Port/Username/Password/From/FromName`), IMAP fields (`IMAPHost/Port/Username/Password/Folder/IMAPPollInterval`).

### Changed
- HN scraper now runs as primary lead source alongside mock fallback.
- Hunter.io runs as primary enrichment source when `HUNTER_API_KEY` is set.
- Main.go wires all new components with auto-detection from environment variables.

## [0.4.8] - 2026-06-10

### Added
- **Hermes Agent LLM Integration (Phase 7 foundation):**
    - Implemented `HermesLLMProvider` in `internal/llm/hermes.go` — an OpenAI-compatible client that routes all LLM calls through a local Hermes Agent gateway.
    - Added `HermesConfig` struct with `BaseURL`, `APIKey`, and `Model` fields for flexible configuration.
    - Added `HealthCheck()` method for runtime connectivity verification.
    - Wired Hermes as the primary LLM provider in `cmd/sales_bot/main.go` with automatic fallback to `MockLLMProvider` when `HERMES_API_URL`/`HERMES_API_KEY` are not set.
    - Added `LLMIntentClassifier` integration — when Hermes is available, the bot uses LLM-based intent classification instead of keyword-matching mocks.
    - Added LLM provider health status to the web dashboard (System Health section) and `/health/detailed` JSON endpoint.
    - Extended `Config` struct with `HermesAPIURL`, `HermesAPIKey`, and `HermesModel` fields.
    - Added integration tests that verify end-to-end Hermes connectivity (health check + LLM generation).
    - Configured Hermes API server (`API_SERVER_HOST=0.0.0.0`) for cross-WSL/Windows access.

### Changed
- `web.NewServer()` now accepts an `llm.LLMProvider` parameter for health reporting.
- Dashboard HTML updated with LLM provider status indicator (green for Hermes connected, grey for mock).

## [0.4.7] - 2026-06-08

### Added
- **Final Feature Validation:**
    - Verified end-to-end functionality including secure dashboard, hardened CRM sync, and graceful lifecycle management.
    - Standardized build and test utilities for production deployment.

## [0.4.6] - 2026-06-08

### Added
- **User Authentication:**
    - Implemented a simple session-based authentication module in `internal/auth`.
    - Added middleware to protect the web dashboard and deployment controls.
    - Integrated login page and session cookie management into `internal/web/server.go`.
- **Infrastructure Improvements:**
    - Centralized all environment variables into a typed `Config` struct in `internal/config`.
    - Optimized the web server router with a pre-initialized `ServeMux`.
    - Enhanced Graceful Shutdown with a 2-second worker drain wait time.

## [0.4.5] - 2026-06-08

### Added
- **CRM Hardening & Verification:**
    - Implemented asynchronous retry logic with exponential backoff for all CRM synchronization points.
    - Created a new CRM integration verification utility (`scripts/crm_verify/verify_crm_integration.go`) for E2E simulation.
    - Standardized CRM error logging across background workers.

## [0.4.4] - 2026-06-08

### Added
- **E2E Verified CRM Integration:**
    - Enhanced E2E test suite to verify CRM synchronization across enrichment and research phases.
    - Validated CRM client unit tests for contact and dossier synchronization.

## [0.4.3] - 2026-06-08

### Added
- **Enhanced CRM Integration:**
    - Extended CRMClient with SyncContacts for real-time contact synchronization.
    - Integrated CRM synchronization into Enrichment Worker and Researcher modules.
    - Expanded PushDeal payload to include technical dossiers for better CRM visibility.

### Changed
- Rebranded all product-facing references from "TormentNexus" to "TormentNexus" across 14 files (Go source, tests, markdown docs, CI config).
- Comprehensive documentation overhaul: ROADMAP, TODO, VISION, README, DEPLOY, MEMORY, IDEAS, AGENTS, HANDOFF all updated with gap analysis, forward-looking phases, and technical debt inventory.

## [0.4.1] - 2026-06-05

### Added
- Implemented "Self-Improving Prompts" feature to optimize outreach using successful past interactions.
- Reconciled repository using the Dual-Direction Intelligent Merge Engine (Executive Protocol Step 2).
- Updated database schema and repository to support interaction success tracking.
- Enhanced RAGResponseGenerator with few-shot learning from successful examples.
- Resolved CI/CD stability issues by correcting Gosec action references and fixing linting errors.
- Upgraded system to Go 1.24 across all environments.

## [0.4.0-dev] - 2026-05-31

### Added
- Integrated Phase 5: Automated Provisioning for won deals.
- Consolidated database repository logic and resolved method re-declaration conflicts.
- Completed dual-direction intelligent merge for full branch reconciliation.
- Updated documentation and roadmap to reflect end-to-end sales lifecycle readiness.

## [0.3.0-dev] - 2026-05-26

### Added
- Functional autonomous development loop with self-updating workflows.
- Dual-Direction Intelligent Merge Engine (Forward & Reverse) integrated into sync cycle.
- Production-ready CRM integration and lead state reconciliation.
- Dossier-aware pseudo-RAG response logic for hyper-personalized technical outreach.
- Tiered pricing engine and automated lead scoring/prioritization.
- Persistent PR tracking in PostgreSQL with real-time web dashboard.
- Standardized Dockerized deployment pipeline and health monitoring.
- Enhanced CI/CD with PostgreSQL integration testing and coverage reporting.

## [0.2.0] - 2025-05-25

### Added
- Native Go implementation of the Executive Sync Protocol in `internal/gitcheck`.
- Autonomous Development Module: `internal/autodev` for self-initiated task processing.
- Core project documentation: `VISION.md`, `MEMORY.md`, `DEPLOY.md`, and `IDEAS.md`.
- Initial database models and migrations for lead tracking (merged from previous iteration).
- Phase 1 infrastructure and conflict resolution tests.

### Changed
- Incremented version to 0.2.0.

## [0.1.0] - 2025-05-25

### Added
- Initial repository structure and synchronization protocol.
- `AGENTS.md` for architectural governance.
- `VERSION`, `ROADMAP.md`, and `TODO.md` for project tracking.
- Go module initialization.
- Basic `build.bat` and `start.bat` execution scripts.
- Automated conflict detection and merge integrity tests in `internal/gitcheck`.
- Automated conflict resolution simulation tests in `internal/gitres`.
- CI pipeline configuration in `.github/workflows/ci.yml`.
