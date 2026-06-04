# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1-dev] - 2026-06-01

### Added
- Implemented "Self-Improving Prompts" feature to optimize outreach using successful past interactions.
- Reconciled repository using the Dual-Direction Intelligent Merge Engine (Executive Protocol Step 2).
- Updated database schema and repository to support interaction success tracking.
- Enhanced RAGResponseGenerator with few-shot learning from successful examples.
- Resolved CI/CD stability issues by correcting Gosec action references and fixing linting errors.

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
