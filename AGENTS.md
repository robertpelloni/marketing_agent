# Borg Autonomous Sales Pipeline Architecture

This system is an asynchronous, event-driven orchestration layer written in Go to automate B2B lead generation, enrichment, hyper-personalized outreach, and billing for the Borg repository.

## System Guidelines
- **Language:** Go (Golang) using standard concurrency paradigms (goroutines, channels) for background workers.
- **State Machine:** Enforce rigid, atomic state updates for all leads in the PostgreSQL database.
- **Integrations:** All scraper engines must utilize headless configuration profiles. External communication modules use abstract interfaces to allow mock testing.

## Database Schema Constraints
All data migrations must use strict relational mappings with full foreign key constraints tracking Companies -> Contacts -> Interactions -> Deals.

## Autonomous Development & Repository Management Protocol
The system follows a strict "EXECUTIVE PROTOCOL" for repository synchronization and intelligent merging:
- **Upstream Tracking:** Always sync with the parent fork and update all submodules recursively.
- **Intelligent Merge:** Use the dual-direction merge engine to reconcile feature branches with `main`.
- **Validation:** Every build must pass the merge integrity tests defined in `internal/gitcheck`.
- **Automation:** Utilize `scripts/sync_repo.sh` for automated synchronization.
