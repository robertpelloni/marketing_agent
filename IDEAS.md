# IDEAS — Autonomous Revenue Engine Innovations & Expansions

## 1. Multi-LLM Shadow Mode Prompt Tournaments
- Run candidate prompt variants through parallel shadow evaluations against historical email responses.
- Dynamically promote prompt templates with highest simulated agreement and win likelihood into active cadence steps.

## 2. Autonomous Social Sentiment Arbitrage
- Scrape Twitter/X and Reddit threads in real time for keywords like "enterprise AI orchestration", "local LLM security", or "agentic workflows".
- Automatically draft personalized contextual technical responses with dual-brand backlink insertion within 3 minutes of thread creation.

## 3. Dynamic Quote Generation via Real-Time Competitor Pricing Scraping
- Headless Rod browser regularly scrapes competitor pricing tables (e.g. LangChain Enterprise, OpenAI Corporate, Weights & Biases).
- Generates hyper-competitive ROI discount quotes directly embedded into the quotation endpoint `/api/v1/quote`.

## 4. Voice-Agent Outbound Qualification Integration
- Integrate an ultra-low latency open-source voice model (like Whisper + Kokoro) to conduct automated qualification calls for enterprise leads that submit quote requests.

## 2026-11-17 Additions (from production findings)
- **Multi-account Gmail pacing**: the daily 550 cap plus a fully disabled OAuth client showed single-account mail is a single point of failure. Rotate sends across 2-3 Workspace accounts with per-account daily budgets and automatic failover when one account's client is disabled (breaker state per account, surfaced on 6A).
- **Breaker/lane telemetry on the dashboard**: lane flags, auth breakers, and provider health are now computed in code (`providerPanelList`, breakers) but only partly visible. Emit them in `/api/v1/metrics` (the 2026-11-13 fallback-telemetry idea, still open) so 6A shows live breaker state instead of env presence.
- **Blog generation resilience**: long posts through the FreeLLM stopgap still occasionally exceed even 300s. Add per-topic retry with backoff + partial-content salvage (publish the completed sections) instead of an all-or-nothing 4h-cycle retry.
- **Tenant session experience**: signed sessions landed; next step is a tenant self-service view (own container status, own usage) behind the same `company_N` subject — the container handlers already extract it.
- **Reddit via official API only**: headless Reddit is retired; if the lane ever reopens, use the official API with the existing rate-limit/circuit-breaker policy wrapper (the 2026-11-13 backoff idea, narrowed to live sources: GitHub only today).
- **Disk guardrail**: the shared VPS hit 91% from other tenants' backups. Add a dashboard disk gauge (statfs in /health/detailed) so the operator sees the cliff before Postgres does.
- Note: §2's Twitter/X scraping premise is moot — the X lane is retired by operator decision (2026-11-16); sentiment work would target Reddit/HN/Bluesky instead.

## 2026-11-13 Additions (from production findings)
- **Schema-drift sentinel**: two production bugs (deals.cadence_step, interactions.template_id/response_id) were columns the code used but no migration created. Add a CI job that boots Postgres, runs migrations, and asserts every column referenced in internal/db queries exists in information_schema.
- **Fallback-source telemetry**: GitHub search PROTOCOL_ERROR bursts were only visible in journal spam. Emit per-source success/failure counters to /api/v1/metrics so the dashboard shows source health instead of log-diving.
- **Credential health panel**: DeepSeek/Apollo/Bluesky/X failures are all account issues discovered via logs. A dashboard card probing each provider's health endpoint (already have /health/detailed llm_provider) would surface 'fund this account' states to the operator immediately.
- **Backoff-aware enrichment scheduling**: with retry/backoff now in GitHubEnrichSource, extend the same policy wrapper to Hunter/Apollo sources and jitter cycle start times to avoid thundering herds after restarts.
