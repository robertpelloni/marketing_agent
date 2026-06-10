# Changelog

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.8] - 2026-06-10

### Added
- **Production Hardening (Phase 6):**
    - Upgraded authentication to use `bcrypt` for salted password hashing and secure random session IDs.
    - Implemented structured JSON logging using the Go `slog` package across all worker modules.
    - Exposed Prometheus metrics at `/metrics` and integrated instrumentation for lead discovery, interaction processing, and deals won.
    - Refined `LearningSalesEngine` to prioritize `MeetingRequest` intents for qualified leads.
    - Hardened security by protecting the UAT simulation endpoint with mandatory authentication.
    - Fixed Go environment stability issues by pinning dependencies to Go 1.24 compatible versions.

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
- Rebranded all product-facing references from "Borg" to "TormentNexus" across 14 files (Go source, tests, markdown docs, CI config).
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
