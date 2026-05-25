# Borg Autonomous Sales Pipeline Architecture

This system is an asynchronous, event-driven orchestration layer written in Go to automate B2B lead generation, enrichment, hyper-personalized outreach, and billing for the Borg repository.

## System Guidelines
- **Language:** Go (Golang) using standard concurrency paradigms (goroutines, channels) for background workers.
- **State Machine:** Enforce rigid, atomic state updates for all leads in the PostgreSQL database.
- **Integrations:** All scraper engines must utilize headless configuration profiles. External communication modules use abstract interfaces to allow mock testing.

## Database Schema Constraints
All data migrations must use strict relational mappings with full foreign key constraints tracking Companies -> Contacts -> Interactions -> Deals.
