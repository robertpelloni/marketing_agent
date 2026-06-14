# Changelog

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2026-06-13

### Added
- **Enterprise CRM Refinement:**
    - Implemented automated record associations: Linked HubSpot Notes to Deals and Salesforce Tasks to Opportunities (via `WhatId`).
    - Enhanced `PushDeal` with upsert logic: The system now uses `PATCH` to update existing records, preventing duplicate entries in HubSpot and Salesforce.
- **Improved Governance:**
    - Standardized version 0.6.0 across all project and module metadata.
    - Synchronized handoff documentation with finalized CRM-to-Outreach cycle verification.

## [0.5.1] - 2026-06-12

### Added
- **Live CRM Inbound Integration:**
    - Integrated HubSpot and Salesforce inbound email polling into the autonomous response loop.
    - Updated `CRM Worker` to automatically trigger the `communication.Manager` for new external interactions.
    - Enhanced CRM clients to extract and associate sender emails from HubSpot Communications and Salesforce EmailMessage objects.
- **Architectural Improvements:**
    - Implemented a base `IMAPPoller` in `internal/mail` for real-time email ingestion.
    - Decoupled `CRM Worker` from `communication.Manager` using an `InboundProcessor` interface to prevent circular imports.
    - Hardened response generation logic with robust nil-checks to support operation during database maintenance or simulation.
- **Verification Suite:**
    - Developed a new End-to-End verification script (`scripts/verify_live_flow/main.go`) to validate the complete CRM-to-Outreach autonomous lifecycle.

## [0.5.0] - 2026-06-08

### Added
- **CRM Field Mapping Configuration:**
    - Implemented `FieldMapping` in `internal/crm` to allow customizable property names for HubSpot and Salesforce.
    - Added `SetFieldMapping` to `CRMClient` interface and implemented it across all clients (HubSpot, Salesforce, REST, Mock).
    - Exposed CRM field mappings via environment variables and updated `internal/config`.
    - Integrated dynamic mapping into `HubSpotCRMClient` and `SalesforceCRMClient` for API requests and response parsing.
- **Deployment & Readiness:**
    - Updated `Dockerfile` and `docker-compose.staging.yml` for staging environment readiness.
    - Enhanced `DEPLOY.md` with CRM field mapping instructions and staging validation steps.
    - Verified full framework state with comprehensive unit and integration tests.

## [0.4.8] - 2026-06-08

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
