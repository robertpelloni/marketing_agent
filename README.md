# marketing_agent

An autonomous B2B sales-pipeline service written in Go, built to market **HyperNexus**: it discovers prospective customers, enriches contacts, runs LLM-personalized outreach cadences over email, handles replies, and connects interested prospects to Stripe billing. This repo also hosts the public websites and launch materials for HyperNexus / TormentNexus / AI Money Machine (see `hypernexus_site/`, `tormentnexus_site/`, `landing/`, `demo_site/`, `launch_materials/`).

> **Alpha software.** Version 0.7.0. Expect breaking changes; not ready for unattended production use.

## Status — the honest part

- The GitHub **"Integration Tests" workflow has failed on every run since 2026-07-11**. Unit tests exist and the service builds, but the integration suite (which self-migrates against a real Postgres) is red and has been for months. Treat "it works" claims accordingly.
- The service runs on the production server as a systemd unit and does send real outreach when configured — it is a live system with a broken test gate, not a demo.
- The Go module path is `gitlab.com/robertpelloni/marketing_agent`; the repo is mirrored to GitLab, where additional pipelines run.

## What it does

1. **Discovers leads** from GitHub repos, job boards, and HN hiring threads (`internal/scraper`)
2. **Enriches contacts** via Apollo.io, Hunter.io, and GitHub activity (`internal/enrichment`)
3. **Researches prospects** into technical dossiers (`internal/researcher`)
4. **Runs outreach** — multi-touch, cadence-aware email with LLM-generated personalization and autonomous reply handling (`internal/communication`, `internal/llm`)
5. **Syncs a CRM** bidirectionally (`internal/crm`) and tracks a 7-state lead lifecycle: Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost
6. **Bills** via Stripe checkout and subscriptions (`internal/billing`)
7. **Publishes content** — generated blog posts cross-posted to dev.to/Hashnode (`pkg/agents`, posting scripts at repo root)

Supporting packages: `internal/db` (PostgreSQL layer), `internal/auth`, `internal/config`, `internal/web` (HTTP API + dashboard), `internal/webhook`, `internal/autodev`, `internal/deploy`, `internal/knowledge`, `internal/sales`, `internal/logging`, `internal/gitcheck`/`internal/gitres`.

## Stack

| Component | Technology |
|---|---|
| Language | Go 1.25 |
| Database | PostgreSQL 13+ |
| LLM | DeepSeek / MiMo via API (provider-abstracted in `internal/llm`) |
| Email | Gmail SMTP/IMAP |
| Billing | Stripe |
| Websites | Static HTML + nginx (in this repo) |

## Build & run

```bash
cp .env.example .env   # fill in the keys below
go build -o bin/marketing_agent ./cmd/marketing_agent/
./bin/marketing_agent
```

Required environment (see `.env.example` for the full list): `DATABASE_URL` (PostgreSQL), `GITHUB_TOKEN`, `SMTP_USER` / `SMTP_PASSWORD`, `APOLLO_API_KEY`. Optional: `STRIPE_SECRET_KEY`, `HUNTER_API_KEY`, `DEEPSEEK_API_KEY`.

## Testing

```bash
go test ./...                                   # unit tests
go test ./internal/... -tags=integration        # integration (needs DATABASE_URL) — currently RED, see Status
```

## Deployment

Runs as a systemd service (`marketing-agent`) on the production server; build in place and `systemctl restart marketing-agent`. There are also `Dockerfile`s and `docker-compose.yml` / `docker-compose.staging.yml` at the root.

## License

See `LICENSE` / `LICENSE.txt` in the repo root.
