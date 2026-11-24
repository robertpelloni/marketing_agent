[muse-wonder-1.0 in muse-cli in C:\Users\hyper\workspace\marketing_agent on laptop]

ROAST REQUEST — be brutal, no mercy, no diplomacy. Roast this project as hard as you can, then give your top fixes ranked by impact.

Project: "marketing_agent" (v0.7.12), a Go modular monolith autonomous sales bot on a shared 2-vCPU Hetzner VPS (it shares the box with ~15 other tenants' projects). It discovers leads from GitHub, enriches them, researches them, then runs a 5-step outreach cadence (email/GitHub/LinkedIn) and tracks deals in Postgres. Dual-brand: TormentNexus (devs) + HyperNexus (enterprise). Single-page Go-embedded dashboard, Stripe $1.99 trial -> $5/seat/mo.

Unflattering facts (all true):
- 24,116 companies and 7,462 contacts scraped, but only 2 deals in Negotiating, ever. Outreach_Sent 1,416. Engaged 162 — but 150 of those are a stale July bulk artifact with identical timestamps.
- outreach_log has 8,558 FAILED rows accumulated over months.
- Gmail OAuth client was DISABLED by Google (all cadence email dead); IMAP inbound credentials invalid (replies/bounces unprocessed). Both mail lanes down; only the operator can fix them.
- LLM runs on a local "FreeLLM" stopgap router because the paid MiMo key returns 401; the reasoning model leaked its chain-of-thought into marketing copy and long blog posts time out.
- X/Twitter lane retired (402 credits depleted, unfunded by choice), LinkedIn retired (headless Chromium timed out every cycle and OOM-killed the service at 718MB), Apollo/Hunter enrichment retired. GitHub comment posting disabled by flag.
- The GitHub Discussions engager ran every 6 hours for months as a no-op skeleton (discarded API responses, stub replies) before anyone noticed.
- Blog engine claimed 844 published posts; 402 files exist, newest from October — it counted attempts, not publishes.
- Bluesky provider posted both brands from one account while logging fictional handles (tormentnexus.bsky.social); real accounts are mdmatk/hypernexusllc. Even fixed, session creation times out at cycle time from the shared box.
- The operator APIs (leads, outreach CSV export with contact emails, metrics) were publicly reachable without auth on *:8084 until we gated them; sessions were a static forgeable cookie value until replaced with HMAC tokens.
- The whole stack was cleanly stopped by the operator on Nov 22 and is currently DOWN (unit disabled).
- Dashboard is one giant Go file (internal/web/server.go, ~105KB, embedded HTML).

Deliver: (1) the most brutal roast you can write about this — the concept, the execution, the funnel math, the ops; (2) your ranked top-10 fix list (impact order, concrete); (3) one sentence: should this live or be killed? Other lanes (Gemini, ChatGPT, Muse) are roasting it too; I will fix the merged findings and report back.
