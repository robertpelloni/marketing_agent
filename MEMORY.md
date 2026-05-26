# Memory: Architectural Observations & Design Preferences

## Current State
- The project is in its early implementation phase (Phase 1/2).
- Core database models and migrations are implemented in Go and PostgreSQL.
- A robust merge integrity and conflict resolution testing framework is in place.
- The project uses Go 1.24 and follows standard Golang concurrency patterns.

## Architectural Traits
- **Event-Driven:** Designed to be asynchronous and event-driven.
- **Interface-Based:** External integrations (scrapers, email providers) are abstracted behind interfaces for easier mocking and rotation.
- **Rigid State Management:** Lead transitions are handled via an atomic state machine in the database.
- **Automation First:** Every feature is built with the intent of being fully autonomous.
- **Self-Development Loop:** The system includes an `autodev` module that autonomously selects tasks from `TODO.md`, proposes changes, and verifies them via a branch-push-PR-merge lifecycle.
- **Autonomous Continuous Delivery:** Codebase updates initiated by the bot trigger automated GitHub Action workflows for testing and deployment to ensure system stability.
- **Self-Learning Sales Engine:** The `communication` package features a `LearningSalesEngine` that analyzes interaction history and lead context to decide on autonomous responses, state transitions, or human escalation.

## Design Preferences
- **Go (Golang):** Preferred for the orchestration layer due to its performance and concurrency model.
- **PostgreSQL:** Used for reliable relational data storage and state tracking.
- **Headless Scrapers:** Required for robust data extraction from modern web platforms.
- **Atomic Commits:** Prefer small, descriptive commits that correspond to specific features or fixes.
