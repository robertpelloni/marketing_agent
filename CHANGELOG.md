# Changelog

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.5] - 2026-06-08

### Added
- **CRM Hardening & Verification:**
    - Implemented asynchronous retry logic with exponential backoff for all CRM synchronization points.
    - Created a new CRM integration verification utility (`scripts/crm_verify/verify_crm_integration.go`) for E2E simulation.
    - Standardized CRM error logging across background workers.
    - Improved Graceful Shutdown in `main.go` with worker drain wait time.
    - Optimized Web Server router by pre-initializing ServeMux.

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

### Documentation
- **ROADMAP.md:** Expanded from flat completed list to 5-phase forward roadmap (Phases 6–10) with ~80 new items.
- **TODO.md:** Rebuilt as actionable task list organized by phase with specific, trackable items.
- **VISION.md:** Added current state assessment, architecture diagram (mermaid), evolution roadmap, and key metrics table.
- **README.md:** Comprehensive rewrite with full feature list, worker table, state machine, config reference, and known issues.
- **DEPLOY.md:** Added env var table, CLI flags, Docker instructions, staging validation, and production checklist.
- **MEMORY.md:** Added technical debt inventory and integration status matrix (real vs. mock).
- **IDEAS.md:** Expanded with inbound lead capture, community intelligence, A/B testing, GDPR, and more.
- **AGENTS.md:** Added module architecture table, schema debt section, and system guidelines.

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
