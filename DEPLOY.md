# DEPLOY — Deployment & Environment Setup Instructions

## Production Server Environment
- **Target Host**: Hetzner VPS (`5.161.250.43`) running Ubuntu 24.04 LTS
- **Service Name**: `marketing-agent.service` (systemd managed)
- **Database**: PostgreSQL (`sales_bot` database with pgvector extension)
- **Web Server & Reverse Proxy**: Nginx with SSL certificates provisioned via Let's Encrypt
- **Port Allocations**:
  - Web Dashboard: `8087` (local staging), proxied via Nginx for production domains
  - TormentNexus Sidecar: `4300` / `7778`

## Build & Release Protocol
1. **Compile Binary**:
   ```bash
   ./build.sh
   ```
   This will run internal package tests, extract the version from `VERSION.md`, and output `bin/marketing_agent`.

2. **Deploy to Production VPS**:
   ```bash
   scp bin/marketing_agent root@5.161.250.43:/opt/marketing_agent/bin/
   ssh root@5.161.250.43 "systemctl restart marketing-agent"
   ```

3. **Verify Service Health**:
   ```bash
   curl http://5.161.250.43:8087/health
   curl http://5.161.250.43:8087/health/detailed
   ```

## Current Deploy Procedure (2026-11-13, v0.7.3) - ALWAYS DEPLOY
Standing rule: every push to main is deployed in the same session. Do not run local servers; runtime lives only on Hetzner.
1. Build: $env:GOOS='linux'; go build -trimpath -ldflags '-X main.version=<VERSION>' -o bin/marketing_agent_linux ./cmd/marketing_agent (VERSION file is the single version source).
2. Backup: ssh root@5.161.250.43 'cp /opt/marketing_agent/bin/marketing_agent /opt/marketing_agent/bin/marketing_agent.bak-<date>'
3. Copy: scp binary to /opt/marketing_agent/bin/marketing_agent.new; scp any new migrations/*.sql to /opt/marketing_agent/migrations/; scp site files to /var/www/hypernexus.site/ when hypernexus_site/ changed.
4. Restart: mv .new over bin/marketing_agent; systemctl restart marketing-agent.
5. Verify: curl http://127.0.0.1:8084/health (expect OK) and /health/detailed (database OK, system_health Success); journalctl -u marketing-agent --since '-5 minutes' for ERRORs.
Notes: prod repo /opt/marketing_agent is a clean checkout of main (synced 2026-11-13; backup of prior stale files in /root/prod-backup-20261113/). Prod DB already has pgvector; fresh local Postgres without pgvector fails migration 000009. GitHub Actions Deploy workflow also exists but its production job is DEPLOY_HOST-gated; manual deploy above is the operative path.

## Operator Lane Flags & LLM Stopgap (2026-11-17)
Prod `/opt/marketing_agent/.env` (backups: `.env.bak-20261116-freellm`, binary `.bak-*` alongside):
- `TWITTER_DISABLED=true` — X lane retired (credits depleted, HTTP 402; user will not fund). Hub + panel honor it.
- `LINKEDIN_DISABLED=true` — LinkedIn retired on all 3 surfaces (social hub, cadence step -> Skipped, discovery source). Headless Chromium times out on this VPS and caused an OOM kill (2026-11-17 02:28 UTC).
- `GITHUB_POSTING_DISABLED=true` — GitHub comment posting off; discovery/enrichment unaffected.
- LLM stopgap: `DEEPSEEK_BASE_URL=http://127.0.0.1:4000`, `DEEPSEEK_MODEL=free-llm-plain`, `DEEPSEEK_TIMEOUT_SECONDS=300`, placeholder key — the DeepSeek-compatible provider talks to the local FreeLLM router (`aimm-freellm.service`) because the MiMo token-plan key returns 401. `free-llm-plain` replaced the reasoning `free-llm` model (slow; leaked thinking into post copy; blog posts timed out at 120s). Health reports `DeepSeek: Connected`. Durable fix: rotate a real `DEEPSEEK_API_KEY`/`HERMES_API_KEY` and restore the upstream base URL.
- Known operator repairs pending: Gmail OAuth client disabled by Google (`disabled_client`; re-enable in Google Cloud Console) and IMAP app password invalid (`Invalid credentials`). Cadence aborts cleanly around both.
- Dependabot is scoped by `.github/dependabot.yml` (gomod/npm/github-actions only); unscoped `uv` jobs against vendored `go/agents/` previously poisoned `system_health`.
- Test-suite note: `go test ./...` passes (20 pkgs) with a clean tree — `finalizeCycle` honors `GO_TEST_MODE` and no longer bumps `internal/autodev/VERSION*` during tests (fixed 2026-11-17).
