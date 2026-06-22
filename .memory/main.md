# TormentNexus / XENOCIDE — Autonomous Sales Pipeline

## Project Purpose

A fully autonomous B2B sales pipeline written in Go that sells the **TormentNexus AI Hypervisor** — a local-first cognitive control plane for multi-agent LLM workflows. It discovers companies, researches technical bottlenecks, generates hyper-personalized outreach, and even modifies its own source code. The ultimate goal is **XENOCIDE**: The Final Architecture.

## Current State

### Running on Hetzner VPS (5.161.250.43)
- **Sales Bot** (port 8080) — systemd service, uses FreeLLM for all LLM calls
- **FreeLLM** (port 4000) — Go binary, routes to OpenCode Zen / LM Studio fallback
- **PostgreSQL 16** (port 5432) — production database with 1,094+ companies
- **Nginx** — serves tormentnexus.site, hypernexus.site, blog, sales dashboard
- **Postfix + OpenDKIM** — email infrastructure with signed emails

### Running Locally (Windows)
- **Sales Bot** (port 8085) — development instance with FreeLLM
- **FreeLLM Proxy** (port 4000) — LiteLLM-based proxy
- **LM Studio** (port 1234) — local inference
- **PostgreSQL 18** (port 5433) — WSL instance

### Pipeline Stats
- 1,094 companies discovered (GitHub, HN, Twitter/X, LinkedIn)
- 493 contacts enriched (Hunter.io + Apollo.io)
- 406 outreach emails sent (LLM-generated via FreeLLM)
- 684 deals researched
- 396 interactions logged
- 15 blog posts at 340K+ chars each

### Websites
- **tormentnexus.site** — XENOCIDE theme (Borg + Terminator + cyborg spiders)
- **hypernexus.site** — Sleek enterprise design
- **https://tormentnexus.site/blog/** — 15 autonomous blog posts
- **https://tormentnexus.site/sales/** — Dashboard (port 8080, removed from nginx)

## Key Decisions Made

1. **No mock data** — All lead sources and enrichment return empty instead of simulated results
2. **FreeLLM for all LLM calls** — No MockLLMProvider in production, waits indefinitely
3. **Gmail IMAP Drafts** — Outreach saved as drafts for manual review, not sent directly
4. **PostgreSQL on both systems** — Windows PG18 (local) + PG16 (Hetzner)
5. **Systemd on Hetzner** — sales-bot.service with EnvironmentFile
6. **Cross-harness parity** — identical tool signatures across 6 platforms
7. **Challenger/SPIN sales methodology** — LLM prompts engineered for maximum persuasion
8. **Lightweight Go proxy (FreeLLM)** — replaces LiteLLM for production

## Milestones

### Completed
- ✅ Autonomous sales pipeline fully operational
- ✅ 1,000+ companies discovered from real sources
- ✅ 400+ personalized outreach emails generated
- ✅ 15 blog posts at 340K+ chars each
- ✅ Email infrastructure (Postfix + OpenDKIM + IMAP Drafts)
- ✅ All mock data removed
- ✅ Twitter/X API v2 search integrated
- ✅ LinkedIn scraping with credentials
- ✅ FreeLLM proxy running on Hetzner
- ✅ Windows + Hetzner PostgreSQL sync
- ✅ Stats API, watchdog, blog generator removed (user request)
- ✅ /etc/services on both systems updated
- ✅ Ports documented in services files

### Planned
- 🎯 FreeLLM-linux native port (in progress)
- 🎯 Autonomously fix TODO items via AutoDev
- 🎯 Real-time stats on XENOCIDE website
- 🎯 Blog subdomain (blog.tormentnexus.site)
- 🎯 Better LLM (GPT-4/Claude) for higher quality content

## Architecture

```
Scraper (GitHub, HN, Twitter/X, LinkedIn) → PostgreSQL
       ↓
Enricher (Hunter.io, Apollo.io) → PostgreSQL
       ↓
Researcher (GitHub, Blog, RSS crawlers) → PostgreSQL
       ↓
Communication Manager → FreeLLM → LLM-generated outreach → Gmail Drafts
       ↓
Sales Engine (Challenger/SPIN/MEDDIC scoring) → Deal advancement
       ↓
AutoDev → Self-modifying code via TODO.md
```

## Notes

- The blog generator expanded posts past 100K chars but content quality suffered (MockLLMProvider was used during earlier cycles)
- Blog content has been cleaned of mock text and placeholder messages added
- FreeLLM is slow (OpenCode Zen model) but the bot is configured to wait indefinitely
- The sales.fwber.me domain self-signed cert needs updating to Let's Encrypt
