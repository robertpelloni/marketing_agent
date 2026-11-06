# Granular Short-Term Tasks & Implementation TODOs

## Immediate Completed Items (v0.7.1)
- [x] Fix `internal/communication/cadence_test.go` failures (`MaxAttempts: 10`).
- [x] Consolidate `/dashboard/outreach` metrics and logs directly into `/` dashboard in `internal/web/server.go`.
- [x] Integrate tooltips (`?` markers with hover guidance) for deals, metrics, controls, and channels.
- [x] Upgrade `cmd/marketing_agent/systray.go` and `systray_unix.go` for TormentNexus & HyperNexus support.
- [x] Archive obsolete scripts (`scripts/*_old.py`) non-destructively into `scripts/archive/`.
- [x] Improve `build.sh` with executable permissions and version stamping.

- [x] Connect WebSocket telemetry auto-reconnect backoff in dashboard frontend script.
- [x] Implement export action for cold outreach metrics to CSV directly from main dashboard.

- [x] Add LinkedIn connection request limit warning threshold indicator on deal cards.
- [x] Enhance Devvit Reddit app integration endpoint with cached post preview thumbnails.
- [x] Implement automated staging health check verification script in CI/CD pipeline.
- [x] Add 24-hour recency qualification (`pushed:>YYYY-MM-DD`) to Target Discovery lead generation queries.

- [x] Add automated smoke testing for Docker Compose staging configurations.
- [x] Add PostgreSQL connection pool health metrics to Prometheus endpoint.
- [x] Optimize LLM-powered replies for newly routed leads utilizing MCP and skill-based targeting.
- [x] Prioritize recently enriched contacts in targeted outreach automation (`--recent` / `--hours`).

- [x] Add Prometheus alert rule templates for DB pool exhaustion and high error rates.

## Upcoming Short-Term Tasks
- [ ] Integrate OpenTelemetry trace instrumentation for multi-agent LLM waterfall calls.
- [ ] Implement Redis-backed rate limiting for external platform webhooks.
