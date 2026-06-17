# Roadmap

## COMPLETED: Phase 1 — Infrastructure & Data Modeling
- [x] Core Database Migrations & Models (PostgreSQL)
- [x] State Machine Logic Implementation (Initial)
- [x] Base Go Project Structure
- [x] Merge Integrity & Conflict Detection Tests
- [x] Dual-Direction Intelligent Merge Engine (Full Branch Reconciliation)
- [x] Conflict Resolution Simulation Tests
- [x] Dockerization & Standardized Environment

## COMPLETED: Phase 2 — Lead Generation & Enrichment
- [x] Target Discovery Scraper Module (Job Boards, GitHub)
- [x] Engineering Contact Enrichment Engine (Apollo/Hunter integrations)
- [x] CRM API Integration (Mock + REST)

## COMPLETED: Phase 3 — Research & Personalization
- [x] Technical Context Aggregator (GitHub & Technical Blog crawler)
- [x] Automated Technical Bottleneck Detection
- [x] Hyper-Personalization LLM Layer

## COMPLETED: Phase 4 — Conversational Engine & CRM
- [x] Inbound Communication State Machine (Initial)
- [x] Automated Lead Scoring & Prioritization
- [x] Self-Learning Sales Workflow Engine (Initial)
- [x] CRM Routing Metadata Integration
- [x] Order Fulfillment & Billing Orchestration
- [x] RAG-Powered Technical Q&A (Pseudo-RAG)
- [x] Negotiation & Pricing Engine (Tiered Pricing)
- [x] Self-Improving Prompts & Feedback Loop

## COMPLETED: Phase 5 — Automation & Fulfillment
- [x] TormentNexus Outreach System Foundation (Target Discovery, Safety Policies)
- [x] Billing & ERP Integration (Stripe)
- [x] Automated Provisioning
- [x] Autonomous Development Module (Functional)
- [x] Executive Sync Protocol Integration
- [x] Autonomous PR Handling & Merging (Persistent)
- [x] Automated CI/CD for Codebase Updates
- [x] Real-time CI Status Monitoring & Merging Guardrails
- [x] Self-Service Deployment Pipeline (UI & Backend)
- [x] Deployment Health Monitoring

## COMPLETED: Phase 6 — Production Hardening & Reliability
- [x] Fix CRLF line-ending test failure in `internal/gitres/resolve_test.go`
- [x] Add connection pool configuration to `db.NewDB()`
- [x] Add graceful shutdown with drain timeouts for all background workers
- [x] Add integration tests with ephemeral DB for core modules
- [x] Add web dashboard handler tests
- [x] Add database migration runner to application startup

## COMPLETED: Phase 7 — Real Integrations & Multi-Channel Outreach
- [x] Implement Apollo.io API enrichment source
- [x] Implement Hunter.io email finder
- [x] Implement SMTP email sender
- [x] Implement IMAP email polling for inbound ingestion
- [x] Implement multi-channel outreach (Email, GitHub, LinkedIn)
- [x] Implement outreach cadence management (CadenceAwareManager)
- [x] Implement Hermes Agent LLM provider with fallback chain

## COMPLETED: Phase 8 — Intelligence & Autonomous Evolution
- [x] Implement GitHub repository analysis and blog/RSS ingestion
- [x] Implement sentiment analysis and deal forecasting
- [x] Implement objection handling library with success rate tracking
- [x] Advanced AutoDev: Concurrent tasks, PR feedback loop, and rollback mechanism
- [x] Human-in-the-Loop (HITL) approval workflow for high-value deals

## COMPLETED: Phase 9 — Security, Compliance & Scale
- [x] Add rate limiting and CSRF protection
- [x] Add session-based authentication to web dashboard
- [x] Add GDPR data export and soft-delete endpoints
- [x] Add secrets encryption at rest (AES-GCM)
- [x] Add Prometheus metrics and worker performance profiling

## COMPLETED: Phase 10 — Platform & Ecosystem
- [x] Add REST API (v1) for external pipeline management
- [x] Add outbound webhooks on deal state changes with retry logic
- [ ] Add plugin system for custom sources, classifiers, and responders
- [ ] Add multi-tenant support with data isolation

---

## Phase 11 — TormentNexus-as-a-Service
- [ ] Package the sales engine as a reusable SaaS product
- [ ] Add SaaS billing with per-seat and per-outreach tiers
- [ ] Add onboarding wizard for target ICP definition
- [ ] Add community template marketplace
