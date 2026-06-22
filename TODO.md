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
- [x] Implement Apollo.io API enrichment source (replace mock)
- [x] Implement Hunter.io email finder as secondary source
- [x] Implement LinkedIn Sales Navigator scraper for contact discovery
- [x] Add enrichment source fallback chain

### Real Communication Channels
- [x] Implement SMTP email sender for outbound outreach
- [x] Implement IMAP/POP3 email polling for inbound message ingestion
- [x] Implement LinkedIn message sending via headless automation
- [x] Implement GitHub Issue/PR comment outreach as a technical hook
- [x] Add channel preference logic per contact
- [x] Add outreach cadence management (configurable follow-up schedule)

### Real LLM Integration
- [x] Implement OpenAI/Anthropic LLM provider (replace mock)
- [x] Add provider fallback chain for LLM calls (primary → secondary → tertiary)
- [x] Add token budget tracking per deal/contact to control costs
- [x] Add prompt versioning — store and track prompt templates with A/B testing capability
- [x] Add response quality scoring — auto-evaluate generated responses before sending

### Real CRM Integration
- [x] Implement Salesforce CRM adapter (replace generic REST mock)
- [x] Implement HubSpot CRM adapter as alternative
- [x] Add CRM field mapping configuration — map local fields to CRM-specific schema

## Phase 8 — Intelligence & Autonomous Evolution

### Advanced Lead Intelligence
- [x] Implement real GitHub repository analysis — detect tech stack, architecture patterns, and bottlenecks from actual source code
- [x] Implement real technical blog/RSS ingestion — parse engineering blogs for hiring signals and pain points
- [x] Add competitor intelligence tracking
- [x] Add unified intent signal aggregation

### Autonomous Development Improvements
- [ ] Replace hardcoded `LocalAgent.ProposeSolution` with LLM-powered code generation
- [ ] Add rollback mechanism — if verification fails, revert to pre-change state
- [ ] Add PR feedback loop — use `GetPRComments` to refine the agent's code generation accuracy
- [ ] Add task dependency resolution
- [ ] Add concurrent task execution for independent tasks

### Advanced Sales Strategy
- [x] Add multi-touch outreach sequences across channels
- [x] Add A/B testing for outreach templates — track conversion per template variant
- [x] Add objection handling library — curated rebuttals indexed by objection type with success rates
- [x] Add human-in-the-loop approval workflow — require explicit approval for deals above a configurable threshold
- [x] Add deal forecasting — predict close probability and expected revenue using historical patterns

### Self-Improving Prompts v2
- [ ] Add A/B prompt testing — compare outreach generated with vs. without successful examples
- [x] Add interaction sentiment analysis — auto-classify sentiment of inbound messages to refine strategy
- [ ] Add prompt performance tracking — measure response quality over time as few-shot examples accumulate
- [ ] Add negative example injection — learn from failed outreach (flagged `success=false`) to avoid repeated patterns

## Phase 9 — Security, Compliance & Scale

### Security
- [ ] Add rate limiting on all HTTP endpoints (dashboard, webhook, health)
- [x] Add authentication to the web dashboard (OAuth2 or API key)
- [ ] Add CSRF protection for dashboard form submissions
- [ ] Add input sanitization for all user-supplied data (webhook payloads, form inputs)
- [ ] Add secrets encryption at rest
- [ ] Add GDPR data export endpoint
- [ ] Add GDPR data deletion endpoint
- [ ] Add webhook IP allowlisting

### Scale & Performance
- [ ] Add PostgreSQL connection pooling with configurable limits
- [ ] Add Redis caching layer for frequently accessed data (company lookups, performance metrics)
- [ ] Add horizontal scaling support — make workers stateless so multiple instances can run
- [ ] Add message queue (NATS/RabbitMQ) to decouple workers from direct DB polling
- [ ] Add database read replicas for dashboard queries
- [ ] Add pagination to dashboard deal list
- [ ] Add worker performance profiling

### Deployment & Operations
- [ ] Add Kubernetes manifests (Deployment, Service, ConfigMap, Secret)
- [ ] Add Helm chart for one-command cluster deployment
- [ ] Add Terraform modules for cloud infrastructure provisioning
- [ ] Add blue-green deployment strategy with automatic rollback
- [ ] Add database backup automation — periodic `pg_dump` with S3/blob storage upload
- [ ] Add log aggregation — ship structured logs to ELK/Datadog/CloudWatch

## Phase 10 — Platform & Ecosystem

### API & Extensibility
- [x] Add REST API for external pipeline management (`/api/v1/leads`, `/api/v1/deals`, `/api/v1/interactions`)
- [x] Add webhook outbound — notify external systems on deal state changes
- [x] Add plugin system — allow custom enrichment sources, classifiers, and responders
- [ ] Add multi-tenant support — isolate data and config per organization

### TormentNexus-as-a-Service
- [ ] Package the sales engine as a reusable service
- [ ] Add SaaS billing with per-seat and per-outreach pricing tiers
- [ ] Add onboarding wizard
- [ ] Add community template marketplace
