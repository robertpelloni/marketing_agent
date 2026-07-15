# Show HN: TormentNexus — An OS for AI models (local-first, open-source)

I built TormentNexus because existing AI coding tools were missing a fundamental layer: a **control plane** between models and tools.

**The problem:** Every AI coding harness (Claude Code, Cursor, Codex, etc.) handles tools differently. You configure MCP servers once for Claude Code, again for Cursor, again for Codex. Tool schemas get dumped into context (50K+ tokens). Memory evaporates between sessions. Provider rate limits kill your workflow.

**What TormentNexus does:**

1. **Progressive MCP Tool Routing** — A semantic vector search ranks tools by relevance to your active prompt and injects only the top matches. LRU eviction, profile-based routing, lazy binary startup. Your agents never drown in tool schemas.

2. **Cross-Harness Tool Parity** — 27 golden fixtures, 6 L2 platforms. Byte-for-byte identical tool signatures for Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf. Configure once, use everywhere. No vendor lock-in.

3. **LLM Waterfall** — Transparent cascade: Primary APIs → OpenRouter → local LM Studio/Ollama. Catches 429s and 5xx errors, retries with the exact same payload down the chain. Zero downtime.

4. **L1/L2 Memory Architecture** — Session scratchpad (L1) + permanent SQLite vector store (L2) with sqlite-vec. 14,726 memories persisted across restarts. Context harvesting pulls relevant heuristics from past sessions automatically.

5. **Multi-Agent Swarm** — Planner, Implementer, Tester, Critic rotate through shared sessions via the A2A protocol. PairOrchestrator enforces the collaboration cycle. Debate consensus before shipping.

6. **11K+ MCP Server Catalog** — Indexed from Glama, Smithery, MCP.run, npm, and GitHub Topics. Semantic auto-discovery with 5 adapters.

**Stack:** Go 1.26 kernel (232 files, 446 HTTP handlers) + Next.js 16 dashboard (91 pages). SQLite with sqlite-vec for dependency-free vector search.

**Open source:** github.com/MDMAtk/TormentNexus (AGPLv3)
**Commercial:** hypernexus.site (SSO, RBAC, audit trails, container provisioning)

The dashboard shows real SQLite rows and real Go goroutine states — no mocked data. Every widget is backed by actual telemetry.

Would love feedback from anyone building AI tooling or working with multi-agent systems. What am I missing? What would make this indispensable for your workflow?

---

## Posting instructions

1. Go to <https://news.ycombinator.com/submit>
2. Title: `Show HN: TormentNexus — An OS for AI models (local-first, open-source)`
3. URL: `https://tormentnexus.site`
4. Post the above text as the first comment
5. Best time: Tuesday-Thursday, 7-9am ET
