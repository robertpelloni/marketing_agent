# TODO

## Phase 7 — Real Integrations & Multi-Channel Outreach [v0.5.0]

### Multi-Channel Outreach
- [ ] **[HIGH]** Implement real GitHub outreach posting in `internal/communication/github_sender.go`
- [ ] **[HIGH]** Implement LinkedIn outreach scaffold in `internal/communication/linkedin_sender.go`
- [ ] **[MEDIUM]** Add LinkedIn connection request logic
- [ ] **[MEDIUM]** Wire LinkedIn/GitHub outreach status into Web Dashboard

### CRM Integration
- [ ] **[HIGH]** Implement dynamic field mapping for Salesforce in `internal/crm/salesforce.go`
- [ ] **[HIGH]** Implement dynamic field mapping for HubSpot in `internal/crm/hubspot.go`
- [ ] **[MEDIUM]** Add environment variables for CRM field configuration to `internal/config/config.go`

### Web Dashboard
- [ ] **[MEDIUM]** Add "Lead Dossier" view to the dashboard
- [ ] **[LOW]** Enhance deal table with outreach channel indicators

## Phase 8 — Intelligence & Autonomous Evolution

### Advanced Lead Intelligence
- [ ] **[MEDIUM]** Implement technical blog/RSS ingestion worker
- [ ] **[LOW]** Add competitor tracking signals to Lead Score

### Autonomous Development
- [ ] **[HIGH]** Integration LLM-powered code generation into `LocalAgent`
- [ ] **[MEDIUM]** Implement rollback mechanism in `Orchestrator`

### Sales Strategy
- [ ] **[MEDIUM]** Implement A/B testing for outreach templates
- [ ] **[LOW]** Add Human-in-the-loop approval gates for deals > $50k

## Release Management
- [ ] **[HIGH]** Bump version to 0.5.0
- [ ] **[HIGH]** Update CHANGELOG.md for v0.5.0 release
- [ ] **[MEDIUM]** Update VISION.md, MEMORY.md, and DEPLOY.md with Phase 6/7 progress

## Completed (Phase 6 Hardening)
- [x] Fix CRLF line-ending test failure
- [x] Add connection pool configuration
- [x] Add graceful shutdown for workers
- [x] Add integration tests with ephemeral DB
- [x] Add web dashboard handler tests
- [x] Add negative/error-path unit tests for repository
- [x] Add test coverage reporting to CI
- [x] Fix contacts.email NULL constraint
- [x] Add database indices for performance
- [x] Add migration for audit_log table
- [x] Add deleted_at soft-delete column
- [x] Add database migration runner to startup
- [x] Consolidate configuration into a typed struct
- [x] Add structured JSON logging (slog)
- [x] Add Prometheus metrics endpoint
- [x] Add health check dependencies for workers
