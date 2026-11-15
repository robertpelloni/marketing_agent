# HANDOFF.md — marketing_agent (`v0.7.7`, 2026-11-15)

## 1. Current State Summary
- **Version:** `0.7.7` (`VERSION` and `VERSION.md` aligned)
- **Remotes & Production:** Local `main`, GitHub (`github/main`), GitLab (`origin/main`), and Hetzner production (`5.161.250.43:/opt/marketing_agent`) synchronized on `v0.7.7` with clean working trees and exact deployed binary parity (`go version -m /opt/marketing_agent/bin/marketing_agent` -> `vcs.revision == git rev-parse HEAD`, `vcs.modified=false`).
- **Live Billing Status (Gate 2 Complete):** Live `STRIPE_API_KEY` rotated in `/opt/marketing_agent/.env` (`chmod 600`, LF-normalized); `http://127.0.0.1:8084/health/detailed` reports `"billing": "Stripe (live): Connected"`. All 4 live unpaid `cs_live_...` sessions (`homepage`, `pricing 3 seats`, `checkout 1 seat + email`, `founding 9->5 clamped seats`) passed both Antigravity wire/API checks and Bobbi's independent live browser walk on Muse.
- **CDP Coordination Thread:** `https://muse.ai/thread/8f3e8782-3a1b-4afe-8520-2c81819e195b` (`HyperNexus main side chat`), prefixed with `[gemini in antigravity in c:\Users\hyper\workspace\marketing_agent on BOBPC]`.

## 2. What Shipped in `v0.7.7`
1. **Bot / AI-Agent / Unroutable Email Filtering (`internal/db/contact_hygiene.go`, `internal/enrichment/`, `internal/communication/`):**
   - Added `db.IsAutomatedOrUnroutableEmail`, `db.IsAutomatedCommitAuthor`, `db.IsPersonalEmailDomain`, and `db.SelectOutreachContact`.
   - Blocks AI coding agents (`codex@openai.com`, `claude@anthropic.com`, `cursoragent@cursor.com`, `devin@cognition.ai`), CI/dependency bots (`bot@renovateapp.com`, `*[bot]*`, `actions@github.com`, `snyk-bot`), and non-MX domains (`*.github.io`, `*.local`, `localhost`, `example.com`) during enrichment AND skips/blocks them in `CadenceAwareManager` and `Manager` so existing bot rows in Postgres are never emailed again.
2. **GitHub Rate-Limit Circuit Breaker & FIFO Enrichment Queue Rotation (`internal/enrichment/github_enrich.go`, `fallback.go`, `worker.go`, `internal/db/repository.go`):**
   - Fixed `GitHubEnrichSource` swallowing `HTTP 403` rate-limit exhaustion (`X-RateLimit-Remaining: 0`) into false `empty` results: now immediately opens the circuit until `Retry-After` / `X-RateLimit-Reset`, classifies rate-limited calls as `Failures`, treats `404`/`409` deleted/empty repos as clean empties, and skips domain-only `Apollo.io` / `Hunter.io` lanes on `github.com/` repo leads without inflating `empty` counters.
   - Added `db.TouchDealUpdatedAt` in `Enricher`: clean zero-contact leads bump `updated_at` so `ListDealsForEnrichment` (`ORDER BY updated_at ASC LIMIT 100`) rotates across all `17,900+` discovered leads instead of repeating the same 100 unenrichable leads every hour, while rate-limited leads stay at the head of the queue.
3. **Canonical GitHub Issue Lead Domains (`internal/scraper/github_issue_source.go`):**
   - Stopped `GitHubIssueSource.Discover` from fabricating `<org>.com`/`.io` NXDOMAINs (`velella-sh.com`, etc.): checks `/users/{owner}` for a verified non-social website and otherwise stores canonical `github.com/{owner}/{repo}`, fetches repo metadata via Core API `/repos/{owner}/{repo}` instead of Search API `/search/repositories`, and breaks out of the keyword loop immediately on `429`/`403`.
4. **Researcher Feed Caching, Target Hygiene & Dynamic Intent Scoring (`internal/researcher/`):**
   - Added `30m` in-memory caching + `15m` `429`/`5xx` backoff to `RSSFeedCrawler` (eliminating `netflixtechblog.com/feed/ -> HTTP 429` spam).
   - Skipped already-compiled `TechnicalDossier` rows in `Researcher`, stripped raw email addresses and consumer webmail domains (`gmail.com`, `icloud.com`) from crawler targets, disabled dead `401` `GITHUB_TOKEN` in `GitHubCrawler`, and upgraded `IntentAggregator.Aggregate` to score tech stack, corporate domains, market-cap tier, MCP/agent product fit, and live RSS/GitHub signals dynamically.

## 3. Dependabot Audit Notes
- `github/dependabot/go_modules/go_modules-20443bdb86` (`749d85d0`): **Skipped intentionally** — bumps `go.mod` toolchain directive from `go 1.25.0` to `go 1.26.0`, conflicting with `.github/workflows/ci.yml` (`go-version: '1.25'`).
- `github/dependabot/npm_and_yarn/npm_and_yarn-e9c900b0fc` (`61a41a74`): **Deferred intentionally** — bundles breaking major upgrades (`zod 3 -> 4`, `@hono/node-server 1 -> 2`).

## 4. Open Items for Operator / Next Cycle
1. **Optional Real-Card Charge Check:** Complete a real `$1.99` trial checkout on `https://hypernexus.site/checkout.html` whenever Robert chooses to verify live settlement + webhook tenant provisioning end-to-end.
2. **Provider API Keys (`/opt/marketing_agent/.env`):** Refresh `GITHUB_TOKEN` (unlocks 5,000 req/hr GitHub enrichment & issue comments), `DEEPSEEK_API_KEY` / `HERMES_API_KEY` (unlocks live LLM replies & blog generation), and optional `APOLLO_API_KEY` / `HUNTER_API_KEY` / `DEVTO_API_KEY` / `BLUESKY_*`.
