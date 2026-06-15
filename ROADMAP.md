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
- [x] Add graceful shutdown for background workers
- [x] Add integration tests with ephemeral DB
- [x] Add web dashboard handler tests
- [x] Add negative/error-path unit tests for `db/repository.go`
- [x] Add test coverage reporting to CI pipeline
- [x] Fix `contacts.email` NULL constraint
- [x] Add database indices for `interactions.success` and `deals.current_state`
- [x] Add migration for `audit_log` table
- [x] Add `deleted_at` soft-delete column
- [x] Add database migration runner to startup
- [x] Consolidate configuration into a typed struct
- [x] Add structured JSON logging (`slog`)
- [x] Add Prometheus metrics endpoint (`/metrics`)
- [x] Add health check dependencies for workers

---

## Phase 7 — Real Integrations & Multi-Channel Outreach (v0.5.0)

### 7.1 Multi-Channel Outreach (In Progress)
- [x] **Implement GitHub Issue/PR comment outreach** as a technical hook channel
- [ ] **Implement LinkedIn message sending** via headless automation (Scaffolded)
- [x] **Implement SMTP email sender** for outbound outreach
- [x] **Implement IMAP/POP3 email polling** for inbound ingestion
- [ ] **Implement LinkedIn connection request** with personalized note
- [ ] **Add channel preference logic** — route outreach via the contact's preferred channel

### 7.2 Real CRM Integration (In Progress)
- [x] **Implement Salesforce CRM adapter**
- [x] **Implement HubSpot CRM adapter**
- [ ] **Implement dynamic CRM field mapping** — configurable via environment variables (v0.5.0 target)
- [ ] **Add bidirectional sync for custom fields** (Deal Stage, Lead Source, Technical Dossier)

### 7.3 Real LLM Integration
- [x] **Implement Hermes Agent LLM provider**
- [ ] **Add token budget tracking** per deal/contact
- [ ] **Add response quality scoring** — auto-evaluate generated responses before sending

---

## Phase 8 — Intelligence & Autonomous Evolution

### 8.1 Advanced Lead Intelligence
- [x] **Implement real GitHub repository analysis** — detect tech stack and bottlenecks
- [ ] **Implement real technical blog/RSS ingestion** — parse engineering blogs for hiring signals
- [ ] **Add competitor intelligence** — track when targets evaluate or adopt competing solutions

### 8.2 Autonomous Development Improvements
- [ ] **Replace hardcoded `LocalAgent.ProposeSolution`** with LLM-powered code generation
- [ ] **Add rollback mechanism** for failed verification
- [ ] **Add PR feedback loop** using `GetPRComments`

### 8.3 Advanced Sales Strategy
- [ ] **Add multi-touch outreach sequences** across channels
- [ ] **Add A/B testing for outreach templates**
- [ ] **Add human-in-the-loop approval workflow** for high-value deals
- [x] **Add deal forecasting** using historical patterns

---

## Phase 9 — Security, Compliance & Scale

### 9.1 Security
- [ ] **Add rate limiting** on all HTTP endpoints
- [x] **Add authentication** to the web dashboard (Session-based)
- [ ] **Add CSRF protection**
- [ ] **Add secrets encryption at rest**

### 9.2 Scale & Performance
- [ ] **Add Redis caching layer**
- [ ] **Add horizontal scaling support** (stateless workers)
- [ ] **Add message queue** (NATS/RabbitMQ) to decouple workers

---

## Phase 10 — Platform & Ecosystem
- [ ] **Add REST API** for external pipeline management
- [ ] **Add plugin system** for custom sources and responders
- [ ] **Package the sales engine as a reusable service**
