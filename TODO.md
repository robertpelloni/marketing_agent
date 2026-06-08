# TODO

## Phase 6 — Production Hardening & Reliability

### Test Coverage & Quality
- [x] Fix CRLF line-ending test failure in `internal/gitres/resolve_test.go` (`TestResolveConflictTheirs`)
- [x] Add connection pool configuration to `db.NewDB()` (max open/idle conns, lifetime)
- [x] Add graceful shutdown with drain timeouts for all background workers
- [x] Add integration tests with ephemeral DB for `enrichment/worker`, `researcher`, `crm/worker`, and `communication/manager`
- [x] Add web dashboard handler tests for `/` route, webhook endpoint, and form actions
- [x] Add negative/error-path unit tests for `db/repository.go`
- [x] Add test coverage reporting to CI pipeline

### Database & Data Integrity
- [x] Fix `contacts.email` NULL constraint — add `NOT NULL` or partial unique index
- [x] Add `interactions.success` index for efficient `ListSuccessfulInteractions` queries
- [x] Add `deals(current_state)` index for efficient `ListDealsByState` queries
- [x] Add migration for `audit_log` table to track state transitions with metadata
- [x] Add `deleted_at` soft-delete columns for GDPR compliance
- [x] Add database migration runner to application startup

### Configuration & Environment
- [x] Replace scattered `os.Getenv()` calls with a typed config struct
- [x] Add `.env` file loading for local development
- [x] Add config validation at startup with clear error messages
- [x] Add configurable worker intervals via environment variables

### Logging & Observability
- [x] Replace all `log.Printf` with a leveled, structured logger (`slog` or `zerolog`)
- [x] Add Prometheus metrics endpoint (`/metrics`) with business and system counters
- [x] Add correlation/request IDs to all log lines
- [x] Add `pprof` endpoint for production debugging

### Error Handling & Resilience
- [x] Add retry with exponential backoff to all external API calls
- [x] Add circuit breaker for external integrations (CRM, Stripe, GitHub)
- [x] Add dead-letter tracking for failed interactions/updates
- [x] Add per-worker health status with last successful run timestamp

## Phase 7 — Real Integrations & Multi-Channel Outreach

### Real Enrichment Providers
- [ ] Implement Apollo.io API enrichment source (replace mock)
- [ ] Implement Hunter.io email finder as secondary source
- [ ] Implement LinkedIn Sales Navigator scraper for contact discovery
- [ ] Add enrichment source fallback chain

### Real Communication Channels
- [ ] Implement SMTP email sender for outbound outreach
- [ ] Implement IMAP/POP3 email polling for inbound ingestion
- [ ] Implement LinkedIn message sending via headless automation
- [ ] Implement GitHub Issue/PR comment outreach as technical hook
- [ ] Add channel preference logic per contact
- [ ] Add outreach cadence management (configurable follow-up schedule)

### Real LLM Integration
- [ ] Implement OpenAI/Anthropic LLM provider (replace mock)
- [ ] Add provider fallback chain for LLM calls
- [ ] Add token budget tracking per deal/contact
- [ ] Add prompt versioning with A/B testing capability
- [ ] Add response quality scoring before sending

### Real CRM Integration
- [ ] Implement Salesforce CRM adapter
- [ ] Implement HubSpot CRM adapter
- [ ] Add CRM field mapping configuration

## Phase 8 — Intelligence & Autonomous Evolution

### Advanced Lead Intelligence
- [ ] Implement real GitHub repository analysis for tech stack and bottleneck detection
- [ ] Implement real technical blog/RSS ingestion for hiring signals
- [ ] Add competitor intelligence tracking
- [ ] Add unified intent signal aggregation

### Autonomous Development Improvements
- [ ] Replace hardcoded `LocalAgent.ProposeSolution` with LLM-powered code generation
- [ ] Add rollback mechanism for failed verification
- [ ] Add PR feedback loop using `GetPRComments`
- [ ] Add task dependency resolution
- [ ] Add concurrent task execution for independent tasks

### Advanced Sales Strategy
- [ ] Add multi-touch outreach sequences across channels
- [ ] Add A/B testing for outreach templates
- [ ] Add objection handling library with success rates
- [ ] Add human-in-the-loop approval workflow for high-value deals
- [ ] Add deal forecasting using historical patterns

### Self-Improving Prompts v2
- [ ] Add A/B prompt testing with vs. without successful examples
- [ ] Add interaction sentiment analysis
- [ ] Add prompt performance tracking over time
- [ ] Add negative example injection from failed outreach

## Phase 9 — Security, Compliance & Scale

### Security
- [ ] Add rate limiting on all HTTP endpoints
- [x] Add authentication to web dashboard (Session-based)
- [ ] Add CSRF protection for dashboard form submissions
- [ ] Add input sanitization for webhook payloads and form inputs
- [ ] Add secrets encryption at rest
- [ ] Add GDPR data export endpoint
- [ ] Add GDPR data deletion endpoint
- [ ] Add webhook IP allowlisting

### Scale & Performance
- [ ] Add Redis caching layer for frequently accessed data
- [ ] Add horizontal scaling support (stateless workers)
- [ ] Add message queue (NATS/RabbitMQ) to decouple workers
- [ ] Add database read replicas for dashboard queries
- [ ] Add pagination to dashboard deal list
- [ ] Add worker performance profiling

### Deployment & Operations
- [ ] Add Kubernetes manifests (Deployment, Service, ConfigMap, Secret)
- [ ] Add Helm chart for one-command cluster deployment
- [ ] Add Terraform modules for cloud infrastructure
- [ ] Add blue-green deployment with automatic rollback
- [ ] Add database backup automation with cloud storage
- [ ] Add log aggregation to ELK/Datadog/CloudWatch

## Phase 10 — Platform & Ecosystem

### API & Extensibility
- [ ] Add REST API for external pipeline management
- [ ] Add outbound webhooks on deal state changes
- [ ] Add plugin system for custom sources, classifiers, and responders
- [ ] Add multi-tenant support with data isolation

### TormentNexus-as-a-Service
- [ ] Package the sales engine as a reusable SaaS product
- [ ] Add SaaS billing with per-seat and per-outreach tiers
- [ ] Add onboarding wizard for target ICP definition
- [ ] Add community template marketplace

## Completed (Historical)
- [x] Implement robust CI/CD pipeline with automated PostgreSQL integration testing
- [x] Mature CI/CD infrastructure for autonomous provisioning
- [x] Dockerize application for consistent deployment
- [x] Add Docker build step to deployment workflow
- [x] Implement CI status tracking interfaces
- [x] Implement Task 1: Core Database Migrations & Models
- [x] Implement Task 2: The Target Discovery Scraper Module
- [x] Implement Task 3: Engineering Contact Enrichment Engine
- [x] Define Enrichment interfaces
- [x] Implement Mock Enrichment Source
- [x] Implement Enricher background worker
- [x] Add DB persistence for Contacts
- [x] Implement Task 4: Technical Context Aggregator & Prompt Formatter
- [x] GitHub crawler for target engineers
- [x] Technical blog scraper
- [x] Prompt construction engine
- [x] Implement Task 5: The Inbound Communication State Machine
- [x] Define communication interfaces (Classifier, Responder)
- [x] Implement Mock Intent Classifier
- [x] Implement pseudo-RAG Response Generator (Dossier-aware)
- [x] Create Interaction database handlers
- [x] Implement "Self-Improving Prompts" feedback loop
- [x] Resolve CI failures related to Gosec and linting
- [x] Implement autonomous sales-feature code generation
- [x] Implement the Inbound Communication background worker
- [x] Add negotiation & pricing engine bounds (Tiered Pricing)
- [x] Verify pathing in `build.bat` and `start.bat`
- [x] Implement automated conflict detection tests
- [x] Implement Dual-Direction Intelligent Merge Engine (Full Branch Reconciliation)
- [x] Implement automated conflict resolution simulation tests
- [x] Configure Standardized CI/CD pipeline
- [x] Implement Post-Deployment Health Checks
- [x] Implement Autonomous PR generation and merging logic (Mock)
- [x] Implement Persistent PR tracking and dynamic PR dashboard
- [x] Implement Real-time CI status tracking and monitoring
- [x] Implement Self-Service Deployment Pipeline (Sync & Build triggers)
- [x] Add `borg` submodule for technical documentation reference
- [x] Implement Task 6: Automated Provisioning for won deals
- [x] Rebrand from Borg to TormentNexus across all product-facing references
