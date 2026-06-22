# TODO

<<<<<<< HEAD
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
- [x] Implement autonomous sales-feature code generation
    - [x] Implement the Inbound Communication background worker
    - [x] Add negotiation & pricing engine bounds (Tiered Pricing)
=======
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
<<<<<<< HEAD
- [ ] Replace all `log.Printf` with a leveled, structured logger (`slog` or `zerolog`)
- [ ] Add Prometheus metrics endpoint (`/metrics`) with business and system counters
=======
- [x] Replace all `log.Printf` with a leveled, structured logger (`slog` or `zerolog`)
- [x] Add Prometheus metrics endpoint (`/metrics`) with business and system counters
>>>>>>> origin/main
- [x] Add correlation/request IDs to all log lines
- [x] Add `pprof` endpoint for production debugging

### Error Handling & Resilience
- [x] Add retry with exponential backoff to all external API calls
- [x] Add circuit breaker for external integrations (CRM, Stripe, GitHub)
- [x] Add dead-letter tracking for failed interactions/updates
- [x] Add per-worker health status with last successful run timestamp

## Phase 7 — Real Integrations & Multi-Channel Outreach

### Real Enrichment Providers
<<<<<<< HEAD
- [ ] Implement Apollo.io API enrichment source (replace mock)
- [ ] Implement Hunter.io email finder as secondary source
- [ ] Implement LinkedIn Sales Navigator scraper for contact discovery
- [ ] Add enrichment source fallback chain
=======
- [x] Implement Apollo.io API enrichment source (replace mock)
- [x] Implement Hunter.io email finder as secondary source
- [x] Implement LinkedIn Sales Navigator scraper for contact discovery
- [x] Add enrichment source fallback chain
>>>>>>> origin/main

### Real Communication Channels
- [x] Implement SMTP email sender for outbound outreach
<<<<<<< HEAD
- [x] Implement IMAP/POP3 email polling for inbound message ingestion
- [x] Implement LinkedIn message sending via headless automation
- [x] Implement GitHub Issue/PR comment outreach as a technical hook
=======
- [x] Implement IMAP/POP3 email polling for inbound ingestion
<<<<<<< HEAD
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
=======
- [x] Implement Apollo.io API enrichment source (replace mock)
- [x] Implement Hunter.io email finder as secondary source
- [x] Implement LinkedIn Sales Navigator scraper for contact discovery
- [x] Add enrichment source fallback chain

### Real Communication Channels
- [x] Implement SMTP email sender for outbound outreach
- [x] Implement IMAP/POP3 email polling for inbound ingestion
- [x] Implement LinkedIn message sending via headless automation
- [x] Implement GitHub Issue/PR comment outreach as technical hook
>>>>>>> origin/main
- [x] Add channel preference logic per contact
- [x] Add outreach cadence management (configurable follow-up schedule)

### Real LLM Integration
<<<<<<< HEAD
- [x] Implement OpenAI/Anthropic LLM provider (replace mock)
- [x] Add provider fallback chain for LLM calls (primary → secondary → tertiary)
- [x] Add token budget tracking per deal/contact to control costs
- [x] Add prompt versioning — store and track prompt templates with A/B testing capability
- [x] Add response quality scoring — auto-evaluate generated responses before sending

### Real CRM Integration
- [x] Implement Salesforce CRM adapter (replace generic REST mock)
- [x] Implement HubSpot CRM adapter as alternative
- [x] Add CRM field mapping configuration — map local fields to CRM-specific schema
=======
- [x] Implement Hermes Agent LLM provider (replace mock) — routes through local Hermes gateway with 200+ model support
- [x] Add provider fallback chain for LLM calls - Hermes handles NVIDIA - OpenRouter - LM Studio/Ollama waterfall natively
- [x] Add token budget tracking per deal/contact
- [x] Add prompt versioning with A/B testing capability
- [x] Add response quality scoring before sending
>>>>>>> origin/main

### Real CRM Integration
- [x] Implement Salesforce CRM adapter
- [x] Implement HubSpot CRM adapter
- [x] Add CRM field mapping configuration

## Phase 8 — Intelligence & Autonomous Evolution

### Advanced Lead Intelligence
- [x] Implement real GitHub repository analysis for tech stack and bottleneck detection
- [x] Implement real technical blog/RSS ingestion for hiring signals
<<<<<<< HEAD
<<<<<<< HEAD
- [x] Add competitor intelligence tracking
=======
- [ ] Add competitor intelligence tracking
>>>>>>> origin/main
- [ ] Add unified intent signal aggregation

### Autonomous Development Improvements
- [x] Replace hardcoded `LocalAgent.ProposeSolution` with LLM-powered code generation
- [x] Add rollback mechanism for failed verification
- [x] Add PR feedback loop using `GetPRComments`
- [x] Add task dependency resolution
- [x] Add concurrent task execution for independent tasks

### Advanced Sales Strategy
<<<<<<< HEAD
- [ ] Add multi-touch outreach sequences across channels
- [ ] Add A/B testing for outreach templates
- [x] Add objection handling library with success rates
- [x] Add human-in-the-loop approval workflow for high-value deals
=======
=======
- [ ] Add competitor intelligence tracking
- [ ] Add unified intent signal aggregation

### Autonomous Development Improvements
- [ ] Replace hardcoded `LocalAgent.ProposeSolution` with LLM-powered code generation
- [ ] Add rollback mechanism for failed verification
- [ ] Add PR feedback loop using `GetPRComments`
>>>>>>> origin/main
- [ ] Add task dependency resolution
- [ ] Add concurrent task execution for independent tasks

### Advanced Sales Strategy
>>>>>>> origin/main
- [x] Add multi-touch outreach sequences across channels
- [x] Add A/B testing infrastructure for outreach templates (metrics tracking, impression recording)
- [x] Add template selection algorithm for A/B testing (conversion-based ranking via GetTopTemplate)
- [x] Add template success tracking when interactions convert (full integration complete)
- [x] Add objection handling library with success rates (full integration with outcome tracking)
- [x] Add human-in-the-loop approval workflow for high-value deals (auto-flag Enterprise/>$100k deals, ApproveDeal method)
<<<<<<< HEAD
>>>>>>> origin/main
=======
>>>>>>> origin/main
- [x] Add deal forecasting using historical patterns

### Self-Improving Prompts v2
- [ ] Add A/B prompt testing with vs. without successful examples
- [x] Add interaction sentiment analysis
<<<<<<< HEAD
- [x] Add prompt performance tracking over time
- [x] Add negative example injection from failed outreach
=======
- [ ] Add prompt performance tracking over time
- [ ] Add negative example injection from failed outreach
>>>>>>> origin/main

## Phase 9 — Security, Compliance & Scale

### Security
<<<<<<< HEAD
- [x] Add rate limiting on all HTTP endpoints
- [x] Add authentication to web dashboard (Session-based)
- [x] Add CSRF protection for dashboard form submissions
- [x] Add input sanitization for webhook payloads and form inputs
- [x] Add secrets encryption at rest
- [x] Add GDPR data export endpoint
- [x] Add GDPR data deletion endpoint
- [x] Add webhook IP allowlisting
=======
- [ ] Add rate limiting on all HTTP endpoints
- [x] Add authentication to web dashboard (Session-based)
- [ ] Add CSRF protection for dashboard form submissions
- [ ] Add input sanitization for webhook payloads and form inputs
>>>>>>> origin/main
- [ ] Add secrets encryption at rest
- [ ] Add GDPR data export endpoint
- [ ] Add GDPR data deletion endpoint
- [ ] Add webhook IP allowlisting
>>>>>>> origin/main

### Scale & Performance
<<<<<<< HEAD
- [ ] Add PostgreSQL connection pooling with configurable limits
- [ ] Add Redis caching layer for frequently accessed data (company lookups, performance metrics)
- [ ] Add horizontal scaling support — make workers stateless so multiple instances can run
- [ ] Add message queue (NATS/RabbitMQ) to decouple workers from direct DB polling
=======
- [ ] Add Redis caching layer for frequently accessed data
- [ ] Add horizontal scaling support (stateless workers)
- [ ] Add message queue (NATS/RabbitMQ) to decouple workers
>>>>>>> origin/main
- [ ] Add database read replicas for dashboard queries
<<<<<<< HEAD
- [x] Add pagination to dashboard deal list
- [x] Add worker performance profiling
=======
- [ ] Add pagination to dashboard deal list
- [ ] Add worker performance profiling
>>>>>>> origin/main

### Deployment & Operations
- [ ] Add Kubernetes manifests (Deployment, Service, ConfigMap, Secret)
- [ ] Add Helm chart for one-command cluster deployment
<<<<<<< HEAD
- [ ] Add Terraform modules for cloud infrastructure provisioning
- [ ] Add blue-green deployment strategy with automatic rollback
- [ ] Add database backup automation — periodic `pg_dump` with S3/blob storage upload
- [ ] Add log aggregation — ship structured logs to ELK/Datadog/CloudWatch
=======
- [ ] Add Terraform modules for cloud infrastructure
- [ ] Add blue-green deployment with automatic rollback
- [ ] Add database backup automation with cloud storage
- [ ] Add log aggregation to ELK/Datadog/CloudWatch
>>>>>>> origin/main

## Phase 10 — Platform & Ecosystem

### API & Extensibility
<<<<<<< HEAD
- [x] Add REST API for external pipeline management
- [x] Add outbound webhooks on deal state changes
=======
- [ ] Add REST API for external pipeline management
- [ ] Add outbound webhooks on deal state changes
>>>>>>> origin/main
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
>>>>>>> origin/main
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
<<<<<<< HEAD
=======
- [x] Rebrand from TormentNexus to TormentNexus across all product-facing references
>>>>>>> origin/main
