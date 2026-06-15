# Memory: Architectural Observations & Design Preferences

## Current State

- The project is at **v0.5.0**.
- **LLM provider is REAL** via Hermes Agent gateway.
- **Multi-channel outreach is functional** (SMTP, GitHub Comments).
- **CRM integration is REAL** (Salesforce, HubSpot) with dynamic field mapping.
- Core modules: scraper, enricher, researcher, communication (cadence-aware), CRM, billing, deploy, autodev.

## Architectural Traits

- **Event-Driven & Worker-Based:** Background goroutines handle all periodic tasks with configurable intervals.
- **Interface-Based Integration:** All external systems (CRM, Billing, LLM, Outreach) are behind Go interfaces.
- **Dynamic Mapping:** CRM clients support `FieldMapping` to allow alignment with diverse enterprise schemas without code changes.
- **Cadence-Aware Outreach:** Outreach is not just a single message but a sequence of touches across channels.
- **Self-Development Loop:** The bot autonomously selects and implements tasks from `TODO.md` using the `autodev` orchestrator.
- **Executive Protocol:** Strict git synchronization and submodule management for repository integrity.

## Key Discovered Heuristics

- **GitHub Tech Hooks:** Commenting on open issues/PRs related to AI infrastructure is a high-conversion outreach strategy for TormentNexus.
- **CRM Schema Diversity:** Enterprise Salesforce/HubSpot instances rarely use default field names; hence, dynamic mapping is critical for adoption.
- **Context Harvesting:** Injecting successful past interactions (flagged upon deal win) into LLM prompts significantly improves response quality.

## Known Technical Debt

- **LinkedIn Automation:** LinkedInSender currently uses simulation; requires browser automation (rod/chromedp) for real message sending.
- **Unstructured Logging:** Migration to `slog` is complete, but some legacy `log.Printf` may remain in internal packages.
- **No Rate Limiting:** Web dashboard and API endpoints lack global rate limiting.
- **Missing Indices:** `interactions.success` and `deals.current_state` have indices in ROADMAP but need verification in schema.

## Design Preferences

- **Local-First Native Go:** The orchestration layer must remain dependency-light and fast.
- **Atomic State Transitions:** All lead state changes must be atomic DB operations.
- **Mock Fallbacks:** Every real integration must have a corresponding mock for testing and offline development.
- **Safety First:** Outreach requires explicit opt-out disclaimers and tone guardrails.

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (Target Discovery) | ✅ Real | `pkg/agents/discovery.go` |
| GitHub API (Outreach) | ✅ Real | `internal/communication/github_sender.go` |
| GitHub API (PRs/CI) | ✅ Real | `internal/gitcheck/`, `internal/deploy/` |
| Stripe Billing | ✅ Real | `internal/billing/billing.go` |
| Salesforce CRM | ✅ Real | `internal/crm/salesforce.go` (Dynamic Mapping) |
| HubSpot CRM | ✅ Real | `internal/crm/hubspot.go` (Dynamic Mapping) |
| Hermes LLM | ✅ Real | `internal/llm/hermes.go` |
| SMTP Email | ✅ Real | `internal/communication/smtp_sender.go` |
| IMAP Receiving | ✅ Real | `internal/communication/imap_receiver.go` |
| LinkedIn Outreach | ⚠️ Simulated | `internal/communication/linkedin_sender.go` |
