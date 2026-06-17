# TODO

## Phase 10 — Platform & Ecosystem
- [ ] Add plugin system for custom sources, classifiers, and responders via Go plugins or WASM
- [ ] Add multi-tenant support with data isolation per organization

## Phase 11 — TormentNexus-as-a-Service
- [ ] Package the sales engine as a reusable SaaS product
- [ ] Add SaaS billing with per-seat and per-outreach tiers
- [ ] Add onboarding wizard for target ICP definition
- [ ] Add community template marketplace

## Unfinished & Polish
- [ ] Add Redis caching layer for frequently accessed data
- [ ] Add horizontal scaling support for stateless workers
- [ ] Add message queue (NATS/RabbitMQ) to decouple workers
- [ ] Add database read replicas for dashboard queries
- [ ] Add Kubernetes manifests and Helm chart for deployment
- [ ] Add Terraform modules for cloud infrastructure
- [ ] Add blue-green deployment with automatic rollback

## Completed (v0.9.0 Integrated)
- [x] Finalize Enterprise Hardening: GDPR, Encryption, Metrics, Security
- [x] Integrate Sophisticated Outreach: A/B testing, Objection Handling, Cadences
- [x] Implement Outbound Webhooks on deal state changes with retry logic
- [x] Resolve all upstream merge conflicts and unify architecture
- [x] Fix redundant code and import cycles in communication/sales
- [x] Export internal worker methods for robust integration testing
- [x] Verify all internal tests pass in the unified environment
