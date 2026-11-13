# VISION — Autonomous Marketing Agent & Cognitive Revenue Engine

## Core Purpose & North Star
The Marketing Agent is an autonomous, event-driven cognitive revenue machine engineered to execute full-funnel customer acquisition, lead enrichment, hyper-personalized multi-touch cadence communication, and enterprise closing without requiring manual human intervention.

## The Dual-Brand Marketing Strategy
The engine serves two complementary product ecosystems through dynamic audience classification:
1. **TormentNexus (`tormentnexus.site`)**:
   - **Target Audience**: Independent software developers, hacker news hackers, open-source maintainers, AI enthusiasts.
   - **Tone & Persona**: Authentic indie builder, technical, transparent, highly focused on local-first sovereignty, open LLM inference, and native tool execution.
2. **HyperNexus (`hypernexus.site`)**:
   - **Target Audience**: Enterprise CTOs, VP Engineering, Procurement Heads, Corporate IT Decision-Makers.
   - **Tone & Persona**: Ruthless Enterprise Account Executive, ROI-driven, focused on air-gapped security, SOC2/GDPR compliance, tenant containerization, and seat scalability.

## Core Pillars & Architectural Principles
- **Event-Driven Modular Monolith**: High performance, zero unnecessary network hops, written in Go 1.22 for stability and compatibility.
- **GraphRAG Memory & Few-Shot Learning**: Stores historical email responses, objection counters, and deals won using pgvector to inject proven winning examples into real-time LLM prompts.
- **MEDDPICC Pipeline State Machine**: Every deal progresses systematically through Discovered, Researched, Outreach Sent, Engaged, Negotiating, and Closed Won states.
- **Consolidated Dashboard Cockpit**: A single unified dashboard displaying mission-critical metrics, deals, cold outreach logs, and telemetry without fragmented subpages.

## Status Addendum (2026-11-13, v0.7.3)
The north star is unchanged. Current reality check against the vision: the full funnel (discover -> enrich -> research -> outreach cadence -> deals) runs autonomously on Hetzner; the single-page dashboard exposes every stage with tooltips; billing (Stripe), GDPR, containers, quote, metrics, and telemetry are all wired into that one page. Gaps to the vision are credential-bound, not code-bound: LLM generation (DeepSeek key invalid), Apollo enrichment (key 401), Bluesky (401), and X (402 credits) pause their lanes until accounts are funded. Prompt tournaments, A/B auto-selection, and multi-tenant isolation (IDEAS.md) remain the next frontier once the credential lanes reopen.
