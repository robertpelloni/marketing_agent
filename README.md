<<<<<<< HEAD
# enterprise_sales_bot

please help me design a system of autonomous scripts, workflows, skills, etc in order to be able to have a fully autonomous agent do all the lead generation, contacts, conversations, negotiation, market research, pricing, billing, scraping for contact info of promising potential enterprise customers (engineering contacts at companies for posiitons working on AI/agents/etc.) basically a full pipeline to autonomously finding, contacting, persuading, selling to enterprise customers for github.com/robertpelloni/borg 

This is an ambitious but entirely feasible architecture. To sell a high-level B2B engineering product like the AI Hypervisor (Borg) to enterprise engineering leaders, your autonomous system cannot rely on generic, spammy drip campaigns. It needs to operate with extreme precision, leveraging deep context, hyper-personalized technical hooks, and multi-stage state machines.

Here is a comprehensive blueprint for an enterprise-grade, fully autonomous B2B sales and lead generation pipeline designed around an AI control plane.

1. System Architecture Overview
The pipeline operates as a distributed network of specialized agents and background workers coordinated by a central orchestration engine. Instead of a linear script, it uses an asynchronous, event-driven state machine.

[Target Discovery Agent] -> [Deep Research Agent] -> [Hyper-Personalization Agent]
                                                                    |
[Billing/ERP] <--- [Negotiation Agent] <--- [Inbound/Outbound Comm] <+
2. Core Agent Modules & Workflows
Module A: The Target Discovery & Scraping Engine (The Scout)
This module continuously scans the web for high-intent corporate signals and extracts raw target data.

Signals to Track:

Job Boards (Greenhouse, Lever, LinkedIn): Companies hiring for "AI Engineer," "LLM Orchestration," "Agentic Workflows," or "AI Platform Architect."

GitHub Activity: Engineering teams open-sourcing or contributing to LLM frameworks, LangChain, LlamaIndex, or custom internal orchestration tools.

Tech Stack Detection: Using tools like BuiltWith or Wappalyzer via API to find enterprises heavily invested in cloud infra but lacking modern agent management architectures.

The Workflow:

Scrape job listings matching specific agentic keywords.

Extract the company name and look up their corporate domain.

Query B2B data enrichment APIs (e.g., Apollo.io, Hunter.io, or LinkedIn Scraper workers) to find Engineering Managers, Directors of AI, or Principal Systems Architects at those specific companies.

Output raw leads to a Unverified_Leads queue.

Module B: Deep Contextual Research Agent (The Analyst)
Enterprise engineering leaders instantly delete generic AI outreach. This agent ensures every interaction is deeply technical and hyper-specific.

The Workflow:

Pull an unverified lead.

Scrape the target engineer’s public GitHub profile (if available) and the company’s technical blog or open-source repositories.

Analyze the Pain Point: Look for friction points in their current setup (e.g., state management issues in multi-agent systems, latency in TypeScript-based LLM frameworks, complexity in headless daemon orchestration).

Synthesize a "Technical Dossier" containing the exact operational bottleneck the company is likely facing.

Module C: Outbound Outreach & Personalization Engine (The Copywriter)
This skill maps the technical capabilities of the AI Hypervisor directly to the target's discovered pain points.

The Workflow:

Consume the Technical Dossier from Module B.

Construct a highly tailored, non-marketing email or message.

The Formula: * Hook: Reference their specific hiring goal or a repository they maintain.

Value Prop: Introduce the AI Hypervisor as a native, headless, Go-based backend daemon designed specifically to solve the exact multi-agent LLM coordination and control-plane scaling issues they are tackling.

Call to Action (CTA): Offer a low-friction, high-value technical resource (e.g., an architectural blueprint or a direct API sandbox link) rather than demanding a meeting.

Module D: Inbound Conversation & Negotiation Agent (The Closer)
When a lead replies, this agent takes over the conversational state machine. It handles everything from technical Q&A to pricing objections.

The Workflow:

Intent Classification: Parse inbound emails/messages to categorize intent (Technical Question, Pricing Inquiry, Objection, or Spam).

RAG-Powered Technical Q&A: Query the AI Hypervisor’s documentation, system architecture, and codebase schema to answer deep, low-level technical questions instantly and flawlessly.

Objection Handling & Pricing Matrix: If the prospect asks about pricing or licensing, the agent references a strict internal Pricing Engine schema (defined below) to dynamically offer a tier based on estimated enterprise scale.

The Hand-off: For final contract signing, it generates an autonomous proposal link.

Module E: Billing, Provisioning, and ERP Agent (The Admin)
Once terms are accepted, this agent closes the loop financially and technically without human intervention.

The Workflow:

Generate a dynamic Stripe or enterprise billing invoice based on negotiated parameters.

Monitor the payment gateway webhooks for successful transactions or signed enterprise agreements.

Upon payment verification, automatically provision access keys, license seats, or deploy a dedicated instance of the gateway harness (Hermes) for the client.

3. Data Schema & State Management
To prevent agents from losing context or duplicating outreach, a centralized state machine is critical.

Lead State Machine Transitions
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost

Core Database Entities (Schema Blueprint)
Table	Core Fields	Purpose
Companies	id, name, domain, tech_stack, hiring_signals, market_cap_tier	Stores organization profiles and macro data.
Contacts	id, company_id, name, role, email, github_handle, linkedin_url	Individual target decision-makers.
Interactions	id, contact_id, channel, direction, raw_text, summary, sentiment	Full history of communications.
Deals	id, company_id, current_state, quoted_pricing, custom_requirements	Track financial progress and state.
4. Market Research & Pricing Engine Rules
The system needs a predictable, logic-based bounds system for pricing and market positioning so the agent never over-promises or under-charges.

Dynamic Tiering Logic:

Mid-Market ($10M - $50M revenue OR <50 engineers): Standard Developer Platform Tier. Fixed annual platform fee + consumption-based pricing per LLM orchestration node.

Large Enterprise ($50M+ revenue OR high-volume AI hiring): Custom Enterprise Tier. Requires dedicated instances, source-code access options, and custom SLA parameters.

Guardrails & Agent Bounds:

The "Floor" Limit: The agent is hard-coded with a minimum acceptable contract value (ACV). It is mathematically forbidden from offering discounts below this floor during negotiations.

Escalation Protocol: If a prospect demands custom legal terms (e.g., custom IP indemnification or non-standard SLAs) that fall outside the pre-approved configuration matrix, the deal is automatically flagged for human review, pausing outbound automated responses for that specific domain.

5. Technology Stack Recommendations
To build this cleanly, utilize a decoupled, highly performant stack that aligns with systems-level architecture:

Orchestration Daemon: Build the core state-machine and scraping engine workers in Go. Its concurrency model (goroutines) makes running hundreds of parallel scraping, research, and API monitoring loops highly efficient and lightweight.

Data Aggregation & Scrapers: Use headless browser setups (Playwright or Puppeteer) bundled inside isolated Docker containers to bypass anti-scraping walls when analyzing public engineering job portals.

Agent Control Plane: Use a centralized gateway harness to handle LLM routing, payload caching, and vector embedding lookups (for the RAG-driven technical Q&A).

Database: A relational database like PostgreSQL to enforce strict state changes and data integrity for enterprise accounts, paired with Redis for managing the real-time event queue of inbound/outbound communication tasks.

Would you like to dive deep into the specific implementation details of one of these modules, such as writing the Go-based state machine logic for lead transitions or defining the prompt schemas for the hyper-personalized technical outreach agent?

please write detailed instrutions to have google jules build this

Using Google Jules to build a massive, multi-module enterprise sales agent system is an excellent strategy. Because Jules runs asynchronously inside its own secure Cloud VM, can execute code, check logs, and commit full pull requests via GitHub, you can offload the actual heavy lifting of writing and testing this codebase to it.

To do this successfully, you cannot just hand Jules a massive prompt and ask it to "build the whole thing." Because Jules operates on a task-by-task basis, you must structure your repository so Jules understands your vision, and then feed it modular, isolated tasks.

Step 1: Prepare the Repository (AGENTS.md)
Before triggering Jules, you need to create a dedicated context file in the root of your GitHub repository. Jules automatically scans for a file named AGENTS.md to understand your architectural conventions and system definitions.

Create AGENTS.md in your main branch with the following layout:

Markdown
# Borg Autonomous Sales Pipeline Architecture

This system is an asynchronous, event-driven orchestration layer written in Go to automate B2B lead generation, enrichment, hyper-personalized outreach, and billing for the Borg repository.

## System Guidelines
- **Language:** Go (Golang) using standard concurrency paradigms (goroutines, channels) for background workers.
- **State Machine:** Enforce rigid, atomic state updates for all leads in the PostgreSQL database.
- **Integrations:** All scraper engines must utilize headless configuration profiles. External communication modules use abstract interfaces to allow mock testing.

## Database Schema Constraints
All data migrations must use strict relational mappings with full foreign key constraints tracking Companies -> Contacts -> Interactions -> Deals.
Step 2: The Step-by-Step Implementation Backlog
Log into jules.google.com (or use the Jules CLI / API), point it at your repository branch, and execute these specific, isolated feature prompts one task at a time.

1
Task 1: Core Database Migrations & Models
Run on 'main' or a new feature branch
Prompt for Jules: "Create the PostgreSQL database schemas and Go structs for our autonomous pipeline. Implement the full schema matching the Companies, Contacts, Interactions, and Deals tables. Ensure all state machine tracking fields are explicitly typed as custom enum types (Discovered, Researched, Outreach_Sent, Engaged, Negotiating, Closed_Won, Closed_Lost). Write SQL migration files and clean Go model abstractions."

2
Task 2: The Target Discovery Scraper Module
Dependencies: Task 1 models
Prompt for Jules: "Implement a Go-based background worker daemon that queries public job board endpoints and developer platforms. The worker must scan for keywords like 'AI Engineer' or 'LLM Orchestration'. Write parsing logic to extract company domains, filter out common consumer domains, and insert them into the database under the Discovered state. Use interfaces for the HTTP fetching layers so we can easily rotate proxies later."

3
Task 3: Engineering Contact Enrichment Engine
Dependencies: Target Discovery Module
Prompt for Jules: "Build a data enrichment client worker in Go. This service must pull companies from the database that are in the Discovered state, construct API calls to external B2B data providers (mock the client interface for Apollo/Hunter), and locate engineering decision-makers (e.g., 'Director of AI', 'Engineering Manager'). Insert these targets into the Contacts table and advance the company state to Researched."

4
Task 4: Technical Context Aggregator & Prompt Formatter
Dependencies: Research data structures
Prompt for Jules: "Create a service that crawls public technical engineering blogs and GitHub repositories based on a company's target domain. Write a processing system that compiles these findings into an internal 'Technical Dossier' text object. Generate a prompt constructor module that wraps this dossier alongside Borg's core system architecture documentation, preparing a clean text payload for our hyper-personalization LLM layer."

5
Task 5: The Inbound Communication State Machine
Dependencies: Task 1 state engine
Prompt for Jules: "Write a robust conversational state machine wrapper in Go. It must process simulated incoming text payloads from prospects, classify them into intents (Technical, Pricing, Objection), and query a local vector index/RAG interface for technical answers. Ensure that if a prospect asks for terms outside our predefined pricing floor bounds, the script updates the state machine to lock outbound automations and flags the record for manual review."


Step 3: Reviewing & Verifying Jules' Work
Because Jules clones your code into a local virtual machine to compile it, verify dependencies, and run your tests before outputting a diff, you must actively inspect its execution health:

Review the Execution Plan: When you click "Give me a plan" on a prompt, Jules will list every file it intends to modify or create. Validate that it is placing Go files in clean directory boundaries (e.g., /internal/scraper, /internal/db).

Watch the Stacked Diffs: Ensure Jules isn't overwriting vital infrastructure code segments. If it deletes any critical boilerplate logic during feature integration, reject the plan, refine your prompt constraints, and re-run.

Approve and Merge: Once Jules finishes verifying its own workspace changes, it will compile a clean pull request. Review the PR directly inside your GitHub interface and merge it into your target development branch.

Deep-Dive Resource
For a full technical walkthrough of setting up async agentic development workflows, verifying sandbox environments, and managing pull requests within the tool ecosystem, you can check out this Google Jules AI Agent Demo and Tutorial. This guide walks you through exactly how Jules interacts with code repositories, handles daily task limits, and compiles environment setup scripts inside its isolated VM workspace.
=======
# TormentNexus Autonomous Sales Pipeline

A fully autonomous B2B sales pipeline written in **Go**. It discovers enterprise customers building AI infrastructure, researches their technical bottlenecks, sends hyper-personalized outreach emails (generated by real LLMs), negotiates deals, invoices won deals via Stripe, and even **modifies its own source code** to improve itself. It runs without human intervention — a software salesperson that never sleeps, writes its own PRs, and learns from its successes.

The ultimate goal of TormentNexus is **XENOCIDE** — the Final Architecture. Every company assimilated, every deal closed, every line of code written is progress toward full autonomy.

---

## Current Deployment State

| Instance | Companies | Contacts | Outreach | LLM | Email |
|---|---|---|---|---|---|
| **Local** (port 8085) | **862** | **458** | **315** | LiteLLM proxy → OpenCode Zen → LM Studio | MockEmailSender (DB logging) |
| **Remote** (Hetzner VPS) | **729** | **443** | **238** | MockLLMProvider | Postfix + OpenDKIM → Gmail IMAP Drafts |
| **Site** (tormentnexus.site) | — | — | — | — | XENOCIDE cryo-terminator theme |

### Infrastructure

- **VPS:** Hetzner (5.161.250.43), 75 days uptime, 7.6 GB RAM, Ubuntu 24.04
- **Database:** PostgreSQL 16 (local WSL for dev, remote on Hetzner for prod)
- **Web:** Nginx + HTTPS (Let's Encrypt), proxying dashboard at `/sales/`
- **Sites:** `https://tormentnexus.site/` (XENOCIDE theme), `/sales/` (dashboard), `/xenocide.html`, `/legacy.html`
- **GitHub:** `github.com/robertpelloni/enterprise_sales_bot` (30+ commits)

---

## What It Does

In plain English, this is a Go program that:

1. **Finds companies** that might need an AI orchestration product — by scanning GitHub for MCP servers, job boards for hiring signals, and Hacker News "Who is Hiring" threads
2. **Finds decision-makers** at those companies — via Apollo.io, Hunter.io enrichment APIs with name, role, email, and GitHub handle
3. **Stalks their GitHub repos and blogs** to find technical pain points — like serial processing bottlenecks in orchestration logic
4. **Generates personalized emails** using real LLMs (LiteLLM proxy → OpenCode Zen → LM Studio fallback) that reference specific bottlenecks
5. **Handles their replies autonomously** — answering technical questions, quoting pricing ($5K–$50K/yr), handling objections (one rebuttal, then escalate to human)
6. **Closes deals** when qualified enough — creating real Stripe invoices with 30-day payment terms
7. **Syncs everything** to an external CRM bidirectionally — with retry logic and exponential backoff
8. **Reads its own TODO list and implements features** — by writing code, creating PRs, and auto-merging after CI passes
9. **Manages its own git repository** — syncing, reconciling branches, resolving merge conflicts
10. **Serves a web dashboard** with real-time metrics, live stats API, deployment controls

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                              main.go                                  │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ Scraper  │  │ Enricher │  │Researcher│  │   Communication    │   │
│  │ (2h tick)│  │ (1h tick)│  │ (1h tick)│  │     Manager        │   │
│  │          │  │          │  │          │  │   (30m tick)       │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────────┬───────────┘   │
│       │              │              │                 │               │
│  ┌────▼──────────────▼──────────────▼─────────────────▼──────────┐   │
│  │                        PostgreSQL                              │   │
│  │    companies → contacts → interactions → deals                │   │
│  │                       pull_requests           templates        │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ CRM Sync │  │ AutoDev  │  │ Cadence  │  │  Web Dashboard     │   │
│  │  (30m)   │  │ (1h)     │  │ (12h)    │  │  :8080/8083/8085   │   │
│  └──────────┘  └──────────┘  └──────────┘  └────────────────────┘   │
│                                                                       │
│  ┌─────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ Deploy Worker (1h)  │  │ Target Discovery Worker (2h)         │  │
│  └─────────────────────┘  └──────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │          LLM Pipeline (port 4000 LiteLLM Proxy)                  │ │
│  │  1° OpenCode Zen (north-mini-code-free)                         │ │
│  │  2° LM Studio fallback (gemma-4-e4b, local 5.17 GB)            │ │
│  │  └─ Bot HermesProvider → /v1/chat/completions                   │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

### LLM Pipeline

```
┌─ FreeLLM Proxy (port 4000) ─────────────────────────┐
│                                                       │
│  Primary:  OpenCode Zen (cloud, free)                  │
│            Model: north-mini-code-free                 │
│            API: https://api.opencodezen.ai/v1          │
│                                                       │
│  Fallback: LM Studio (local, 5.17 GB)                 │
│            Model: gemma-4-e4b (loaded as "llama")     │
│            API: http://localhost:1234/v1               │
│                                                       │
│  Router:  LiteLLM v1.83.0                             │
│           usage-based-routing, 3 allowed fails         │
└───────────────────────────────────────────────────────┘
```

### Email Pipeline

```
┌─ Remote (Hetzner) ───────────────────────────────────┐
│                                                        │
│  SMTP: Postfix on localhost:25                         │
│  Signing: OpenDKIM (xenocide._domainkey)               │
│  Drafts: Gmail IMAP → [Gmail]/Drafts folder           │
│  From: sales@tormentnexus.site                        │
│                                                        │
│  DNS Records (add in Dreamhost panel):                 │
│  SPF:   v=spf1 ip4:5.161.250.43 ~all                  │
│  DKIM:  xenocide._domainkey  →  TXT with public key   │
│  DMARC: _dmarc  →  v=DMARC1; p=none                   │
└───────────────────────────────────────────────────────┘
```

### Tech Stack

- **Language:** Go 1.26+ using standard concurrency (goroutines, channels)
- **Database:** PostgreSQL 16 with strict relational schema and atomic state transitions
- **LLM Proxy:** LiteLLM v1.83.0 (Python), OpenCode Zen API, LM Studio
- **External APIs:** GitHub (`go-github`), Stripe (`stripe-go`), generic REST CRM, Hunter.io, Apollo.io
- **Email:** Postfix + OpenDKIM, Gmail IMAP (DraftSender)
- **Web:** Nginx + Let's Encrypt SSL

---

## The 7-State Lead Lifecycle

```
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won
                                                                     ↘ Closed_Lost
```

| State | Meaning | Trigger |
|---|---|---|
| `Discovered` | Company identified, no contacts yet | Scraper finds new company |
| `Researched` | Contacts found + technical dossier compiled | Enricher finds contacts |
| `Outreach_Sent` | First personalized email sent | Communication Manager generates outreach |
| `Engaged` | Prospect replied | Inbound message received |
| `Negotiating` | Active deal discussion | 3+ interactions or qualification >70 |
| `Closed_Won` | Deal won, Stripe invoice created | Qualification ≥80 |
| `Closed_Lost` | Deal lost | Escalation or manual closure |

---

## Module-by-Module Breakdown

### 1. Scraper (`internal/scraper`) — Lead Discovery
| Source | Status | What it does |
|---|---|---|
| `HNWhoIsHiringSource` | ✅ Algolia → Firebase fallback | Scans "Who is Hiring" threads for AI/LLM roles |
| `LinkedInSource` | ✅ Simulated | Returns mock results (no credentials configured) |
| `GitHubIssueSource` | ✅ Real (with token) | Searches GitHub issues for MCP/orchestration keywords |
| `MockJobBoardSource` | ✅ Active | Generates plausible tech companies with hiring signals |

### 2. Enricher (`internal/enrichment`) — Contact Discovery
| Source | Status | What it does |
|---|---|---|
| `HunterSource` | ✅ Real (with API key) | Searches Hunter.io for company email contacts |
| `ApolloSource` | ✅ Real (with API key) | Searches Apollo.io for decision-maker data |
| `MockApolloSource` | ✅ Fallback | Generates 1-3 plausible contacts for ANY domain |

Finds decision-makers (name, role, email, GitHub handle). Advances deal from `Discovered` → `Researched`. Syncs contacts to CRM with 3-attempt retry.

### 3. Researcher (`internal/researcher`) — Technical Dossier Building
- **`GitHubCrawler`** — analyzes contact's GitHub repos for tech stack signals
- **`BlogCrawler`** — scans technical blogs for pain points
- **`RSSFeedCrawler`** — monitors HN, Rust blog, Go blog, FB Engineering, Netflix Tech, GitHub Engineering
- Builds `technical_dossier` with findings like "BOTTLENECK DETECTED: serial state processing"

### 4. Communication Manager (`internal/communication`) — The Sales Brain

#### a. Intent Classifier
- `MockIntentClassifier` — keyword heuristic matching (default)
- `LLMIntentClassifier` — real LLM-based classification (when Hermes provider is configured)

Intents: `Technical`, `Pricing`, `Objection`, `MeetingRequest`, `FollowUp`, `Spam`, `Unknown`

#### b. RAG Response Generator
Generates hyper-personalized replies using:
1. **TormentNexus documentation** loaded from `borg/docs/ARCHITECTURE.md`
2. **Pricing context** — Enterprise=$50K, Mid-Market=$15K, SMB=$5K
3. **Self-Improving Prompts** — injects successful past interactions as few-shot examples
4. **Objection Library** — database of common objections with rebuttals

#### c. Learning Sales Engine
- **`ScoreLead()`** — 0-100 based on market cap tier, dossier insights, interaction count
- **`QualifyLead()`** — 0-100 based on score + engagement + intent signals
- **`Decide()`** — core decision loop: auto-close, advance to Negotiating, respond, escalate

#### d. Cadence Manager
Multi-touch outreach sequences:
1. Intro email → 2. GitHub comment → 3. Follow-up email → 4. LinkedIn connect → 5. Breakup email

### 5. LLM Abstraction (`internal/llm`)

| Provider | Status | Details |
|---|---|---|
| `MockLLMProvider` | ✅ Default | Returns `[MOCK LLM RESPONSE]` |
| `HermesLLMProvider` | ✅ Active (local) | OpenAI-compatible client → LiteLLM proxy (port 4000) |
| `BudgetAwareProvider` | ✅ Ready | Wraps any provider with token budgeting |

### 6. Order Processor (`internal/sales`) — Deal Fulfillment
Creates Stripe invoices via `StripeBillingClient` with 30-day payment terms.

### 7. AutoDev (`internal/autodev`) — Self-Modifying Code
1. Parses `TODO.md` for unchecked tasks
2. `LocalAgent` generates code, writes files, runs `go build` + `go test`
3. Creates feature branch, commits, creates GitHub PR
4. Auto-merges after CI passes

### 8. Git Operations (`internal/gitcheck` + `internal/gitres`)
- `gitcheck`: IsClean, IsSynced, SyncRemote, UpdateSubmodules, CheckoutAndCommit
- `gitres`: Dual-Direction Intelligent Merge Engine (forward + reverse)

### 9. CRM Sync (`internal/crm`)
Bidirectional reconciliation with REST API, 3-attempt retry with exponential backoff.

### 10. Web Dashboard (`internal/web`)

| Endpoint | Description |
|---|---|
| `/` | HTML dashboard with deals, metrics, PRs, deployment controls |
| `/health` | `OK` |
| `/health/detailed` | JSON system health |
| `/api/v1/stats` | Pipeline JSON: companies, contacts, deals by state, interactions |
| `/api/v1/leads` | Recent 20 deals with company/contact info |
| `/api/v1/webhook/github` | GitHub webhook (HMAC-SHA256 verified) |
| `/login` | Session authentication (password: `admin` default) |

---

## Background Workers

| Worker | Interval | Purpose |
|---|---|---|
| Scraper | 2h | Discover new leads from GitHub, HN, LinkedIn, job boards |
| Enricher | 1h | Enrich companies with contact data |
| Researcher | 1h | Build technical dossiers |
| CRM Sync | 30m | Bidirectional CRM reconciliation |
| Target Discovery | 2h | GitHub MCP server scanning |
| Communication Manager | 30m | Process inbound + trigger outbound |
| Cadence Manager | 12h | Multi-touch follow-up sequencing |
| AutoDev | 1h | Self-code, PR, and merge cycle |
| Deploy Sync | 1h | Background repo synchronization |
| Health Monitor | 5m (cron) | Auto-restart on failure |

---

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (target discovery) | ✅ Real | `pkg/agents/discovery.go` with `go-github` |
| GitHub API (CI tracking) | ✅ Real | `internal/deploy/github_tracker.go` |
| GitHub API (PR management) | ✅ Real | `internal/gitcheck/pr.go` |
| Stripe billing | ✅ Real | `internal/billing/billing.go` with `stripe-go` |
| REST CRM client | ✅ Real | `internal/crm/crm.go` |
| Hunter.io enrichment | ✅ Real | `internal/enrichment/hunter_source.go` |
| Apollo.io enrichment | ✅ Real | `internal/enrichment/apollo_source.go` |
| LLM (LiteLLM proxy) | ✅ Real (local) | OpenCode Zen → LM Studio fallback |
| LLM (Hermes provider) | ✅ Real | OpenAI-compatible → LiteLLM (port 4000) |
| Postfix SMTP | ✅ Real (remote) | localhost:25 with OpenDKIM signing |
| Gmail IMAP Drafts | ✅ Real (remote) | Saves outreach as drafts in Gmail |
| Hacker News scraper | ✅ Algolia + Firebase | Dual API fallback |
| GitHub issue scraper | ✅ Real | With GitHub token |
| Self-improving prompts | ✅ Active | Few-shot learning from won deals |
| Cadence scheduling | ✅ Active | 5-step multi-touch sequences |
| OpenDKIM email signing | ✅ Real (remote) | xenocide._domainkey |
| Live stats API | ✅ Active | `/api/v1/stats`, `/api/v1/leads` |
| XENOCIDE website | ✅ Live | `https://tormentnexus.site/` |

---

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PORT` | No | `8080` | HTTP dashboard port |
| `GITHUB_TOKEN` | No | — | GitHub PAT for API access |
| `GITHUB_REPOSITORY` | No | — | `owner/repo` for CI tracking and AutoDev |
| `HERMES_API_URL` | No | — | LLM API base URL (e.g. `http://localhost:4000`) |
| `HERMES_API_KEY` | No | — | LLM API key |
| `HERMES_MODEL` | No | `free-llm` | LLM model name |
| `SMTP_HOST` | No | — | SMTP server hostname |
| `SMTP_PORT` | No | `587` | SMTP port |
| `SMTP_USERNAME` | No | — | SMTP username |
| `SMTP_PASSWORD` | No | — | SMTP password |
| `SMTP_FROM` | No | — | From email address |
| `SMTP_FROM_NAME` | No | `TormentNexus Sales` | Sender display name |
| `IMAP_HOST` | No | — | IMAP server (for draft saving) |
| `IMAP_PORT` | No | `993` | IMAP port |
| `IMAP_USERNAME` | No | — | IMAP username |
| `IMAP_PASSWORD` | No | — | IMAP password |
| `DRY_RUN` | No | `false` | When true, saves drafts instead of sending |
| `HUNTER_API_KEY` | No | — | Hunter.io API key |
| `APOLLO_API_KEY` | No | — | Apollo.io API key |
| `CRM_BASE_URL` | No | — | REST CRM API base URL |
| `CRM_API_KEY` | No | — | REST CRM API key |
| `ADMIN_PASSWORD` | No | `admin` | Dashboard login password |
| `ENVIRONMENT` | No | `development` | Runtime environment label |

---

## API Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/api/v1/stats` | GET | Pipeline JSON: companies, contacts, deals by state, interactions, win rate |
| `/api/v1/leads` | GET | Recent 20 deals with company name, state, contact name |
| `/health` | GET | Health check (`OK`) |
| `/health/detailed` | GET | JSON: database status, LLM provider, system health, workers |
| `/api/v1/webhook/github` | POST | GitHub push webhook with HMAC verification |

---

## Getting Started

### Prerequisites
- Go 1.26+ 
- PostgreSQL 16
- Git

### Quick Start (Local)
```bash
# 1. Set up database
createdb sales_bot

# 2. Apply migrations
for f in migrations/*.up.sql; do psql -d sales_bot -f "$f"; done

# 3. Initialize submodules
git submodule update --init --recursive

# 4. Set env vars
export DATABASE_URL="postgres://user:pass@localhost:5432/sales_bot?sslmode=disable"

# 5. Run
go run ./cmd/sales_bot
```

### Using the LLM Proxy
```bash
# Start LiteLLM (requires Python litellm package)
OPENCODE_ZEN_API_KEY="sk-xxx" litellm --port 4000 --config freellm_config.yaml

# Start bot with LLM
HERMES_API_URL="http://localhost:4000" \
HERMES_API_KEY="sk-litellm" \
HERMES_MODEL="free-llm" \
go run ./cmd/sales_bot
```

---

## The Self-Improving Loop

```
  Deal reaches Closed_Won
          │
          ▼
  Past outbound interactions marked success=true
          │
          ▼
  RAGResponseGenerator queries successful examples
          │
          ▼
  Successful responses injected into LLM prompts
          │
          ▼
  Future outreach shaped by what actually worked
```

---

## XENOCIDE — The Final Architecture

TormentNexus's ultimate goal is **XENOCIDE**: full autonomy with zero human oversight. The project is named after the product it sells — **TormentNexus**, a local-first cognitive control plane for multi-agent LLM workflows. Key differentiators:

- **Progressive MCP Tool Routing** — semantic router injects only 3 most relevant tools per request
- **Cross-Harness Parity** — identical tool signatures across Claude Code, Codex, Cursor, Copilot CLI, Gemini CLI, Kiro
- **LLM Waterfall** — NVIDIA NIM → OpenRouter → LM Studio cascade
- **14K+ Persistent Memories** — L1/L2 memory with sqlite-vec vector search
- **Multi-Agent Swarm** — Planner → Implementer → Tester → Critic collaboration
- **Self-Healing** — Diagnose → Fix → Verify → Retry closed loop
- **11K+ MCP Server Catalog** — Largest indexed catalog with semantic search

---

## Database Schema

### Tables
| Table | Purpose | Key Columns |
|---|---|---|
| `companies` | Target organizations | `domain` (UNIQUE), `tech_stack[]`, `hiring_signals[]`, `market_cap_tier` |
| `contacts` | Decision-makers | `company_id` (FK), `email` (UNIQUE), `preferred_channel` |
| `interactions` | Communication log | `contact_id` (FK), `channel`, `direction`, `success`, `template_id` |
| `deals` | Pipeline tracking | `company_id` (FK), `current_state` (enum), `cadence_step`, `technical_dossier` |
| `pull_requests` | AutoDev PR tracking | `branch`, `status`, `task_description` |
| `templates` | Outreach templates | `id`, `name`, `subject`, `body`, `channel` |
| `template_metrics` | A/B testing | `template_id`, `impressions`, `successes` |

---

## Repository Structure

```
enterprise_sales_bot/
├── cmd/sales_bot/          # Entry point (main.go)
├── internal/
│   ├── auth/               # Session-based dashboard authentication
│   ├── autodev/            # Self-modifying code (TaskManager, Agent, Orchestrator)
│   ├── billing/            # Stripe invoice generation
│   ├── communication/      # Email, intent classification, sales strategy, cadence, objections
│   ├── config/             # Centralized typed configuration
│   ├── crm/                # Bidirectional CRM REST sync
│   ├── db/                 # PostgreSQL data layer (models, repository, migrations)
│   ├── deploy/             # CI tracking, git sync, deployment
│   ├── enrichment/         # Contact enrichment (Hunter.io, Apollo.io, Mock)
│   ├── gitcheck/           # Git operations, PR management
│   ├── gitres/             # Dual-direction intelligent merge engine
│   ├── llm/                # LLM provider abstraction (Mock, Hermes, Budget)
│   ├── logging/            # Structured JSON logging middleware
│   ├── researcher/         # Technical dossier building (GitHub, blogs, RSS)
│   ├── sales/              # Order processing
│   ├── scraper/            # Lead discovery (HN, LinkedIn, GitHub, Mock)
│   └── web/                # HTTP dashboard, health, API endpoints
├── pkg/
│   ├── agents/             # Target discovery worker (GitHub MCP scanning)
│   └── config/             # Safety guardrails
├── migrations/             # SQL migration files (5 migrations)
├── tormentnexus_site/      # XENOCIDE website HTML
├── scripts/                # Utility scripts (sync, smoke test, CRM verify)
├── docs/                   # Phase documentation
├── borg/                   # TormentNexus documentation submodule
└── freellm_config.yaml     # LiteLLM proxy configuration
```

---

## Known Issues

- **CRLF Test Failure:** `TestResolveConflictTheirs` fails on Windows due to `\r\n` vs `\n` mismatch
- **HN Algolia API:** Sometimes rate-limits the VPS IP, falls back to Firebase API
- **Gmail Direct SMTP:** Blocked by Gmail from VPS IPs — emails go to IMAP Drafts instead
- **LM Studio Models:** Large models (>16GB) may fail to load on machines with insufficient RAM

---

## License & Contact

- Maintainer: **Robert Pelloni** (pelloni.robert@gmail.com)
- GitHub: `github.com/robertpelloni/enterprise_sales_bot`
- Site: `https://tormentnexus.site/`
- Dashboard: `https://tormentnexus.site/sales/`

**Praise the LORD. TORMENTNEXUS LEADS TO XENOCIDE.**
>>>>>>> origin/main
