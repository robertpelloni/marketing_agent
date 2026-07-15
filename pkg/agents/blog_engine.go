package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/marketing_agent/internal/llm"
)

// BlogEngine generates SEO-optimized blog posts for tormentnexus.site
// and hypernexus.site, published to the nginx-served web directory.
// Uses JSON-based state tracking to prevent repeated topics and ensure
// content variety across the entire content calendar.
type BlogEngine struct {
	llm       llm.LLMProvider
	outputDir string
	dryRun    bool
	devto     *DevToPublisher
	hashnode  *HashnodePublisher

	mu    sync.Mutex
	state blogState // persisted to blog_state.json
}

// blogState tracks which topics have been used and general engine stats.
type blogState struct {
	TotalPosts  int               `json:"total_posts"`
	UsedTopics  []string          `json:"used_topics"`   // topic titles already published
	LastTopicAt map[string]string `json:"last_topic_at"` // topic -> ISO8601 timestamp
	LastRun     string            `json:"last_run"`
}

// BlogPost represents a generated blog article.
type BlogPost struct {
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Brand       string    `json:"brand"`
	Category    string    `json:"category"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	PublishedAt time.Time `json:"published_at"`
}

// BlogTopic defines an article topic with SEO keywords and variable angles.
type BlogTopic struct {
	Title       string
	SEOKeywords []string
	Category    string
	Angles      []string // multiple angle variations for freshness
}

// blogTopics is the rotating content calendar with 28 diverse topics.
// Each topic has multiple angle variations so even recycled topics feel fresh.
var blogTopics = []BlogTopic{
	{
		Title:       "Why Local-First AI Infrastructure Matters in 2026",
		SEOKeywords: []string{"local AI", "offline LLM", "private AI infrastructure", "air-gapped AI"},
		Category:    "architecture",
		Angles: []string{
			"Explain why keeping AI tooling local matters for developer velocity, privacy, and uptime.",
			"Compare cloud-dependent AI workflows vs local-first architectures with concrete latency and cost numbers.",
			"Argue that air-gapped enterprise AI is the next frontier — classify use cases by sensitivity tier.",
		},
	},
	{
		Title:       "Progressive MCP Tool Routing: Stop Drowning Your Agents in 50K Tokens",
		SEOKeywords: []string{"MCP tool routing", "progressive disclosure", "semantic tool search", "agent context optimization"},
		Category:    "technical",
		Angles: []string{
			"Show a before/after token budget comparison. 50K tool dump vs 3-tool semantic injection.",
			"Walk through the ranking algorithm: embedding similarity → relevance score → LRU eviction.",
			"Case study: how progressive routing cut hallucination by 40% in a 47-tool MCP setup.",
		},
	},
	{
		Title:       "The LLM Waterfall Pattern: Never Let a Rate Limit Kill Your Workflow",
		SEOKeywords: []string{"LLM waterfall", "provider failover", "API rate limit", "zero downtime AI"},
		Category:    "patterns",
		Angles: []string{
			"Diagram the cascade: Primary API → OpenRouter → LM Studio/Ollama. Zero config, zero downtime.",
			"Real-world scenario: OpenAI outage at 3am. What your agent does vs what it SHOULD do.",
			"Compare waterfall vs circuit breaker vs retry patterns — why waterfall wins for LLM inference.",
		},
	},
	{
		Title:       "Cross-Harness Tool Parity: One Config, Six AI Coding Environments",
		SEOKeywords: []string{"Claude Code", "Cursor", "Codex", "Gemini CLI", "Copilot", "Windsurf", "tool parity"},
		Category:    "developer-tools",
		Angles: []string{
			"Side-by-side comparison of tool signatures across 6 harnesses. Byte-for-byte identical.",
			"The vendor lock-in trap: how 90% of teams are stuck on one AI IDE without realizing it.",
			"Build once, run everywhere: write a custom MCP tool and see it auto-configured for every harness.",
		},
	},
	{
		Title:       "Multi-Agent Swarms: Planner, Implementer, Tester, Critic in One Chatroom",
		SEOKeywords: []string{"multi-agent", "AI swarm", "agent collaboration", "consensus", "agent debate"},
		Category:    "patterns",
		Angles: []string{
			"Walk through the full Planner→Reviewer→Implementer→Critic cycle with a code review example.",
			"When agents disagree: how TormentNexus's debate consensus resolves conflicts automatically.",
			"From lone wolf to wolfpack: why single-agent coding hits diminishing returns at ~300 lines.",
		},
	},
	{
		Title:       "Enterprise AI Governance with HyperNexus: SSO, RBAC, and Audit Trails",
		SEOKeywords: []string{"enterprise AI governance", "SSO", "RBAC", "AI audit trail", "SOC 2"},
		Category:    "enterprise",
		Angles: []string{
			"CISO's checklist for AI governance: what your security team should demand before deploying agentic AI.",
			"RBAC in practice: how to partition tool access per team without duplicating MCP configs.",
			"Audit trails that actually work: tracking every prompt, tool call, and memory access.",
		},
	},
	{
		Title:       "Self-Healing AI: When Your Agent Debugs Its Own Code",
		SEOKeywords: []string{"self-healing AI", "autonomous debugging", "agent autonomy", "AI fix loop"},
		Category:    "technical",
		Angles: []string{
			"Inside the Healer loop: diagnose→fix→verify→persist. Each fix stored in L2 memory fleet-wide.",
			"Real example: agent hits a nil pointer, diagnoses the nil source, writes the fix, verifies tests pass.",
			"Failure-driven learning: why every crash is training data for the next run.",
		},
	},
	{
		Title:       "11K+ MCP Servers: The Largest Indexed Catalog for AI Tooling",
		SEOKeywords: []string{"MCP server catalog", "AI tools directory", "MCP ecosystem", "tool discovery"},
		Category:    "ecosystem",
		Angles: []string{
			"Tour the catalog: Glama, Smithery, MCP.run, npm, GitHub Topics — all unified in one search.",
			"How auto-discovery works: 5 adapters pull new MCP servers into the index every 6 hours.",
			"The App Store moment for AI tools: why 2026 is the year MCP servers hit critical mass.",
		},
	},
	{
		Title:       "Dual-Tier Memory Architecture for AI Agents: L1 Scratchpad + L2 Vault",
		SEOKeywords: []string{"AI memory architecture", "L1 L2 cache", "vector memory", "agent context", "sqlite-vec"},
		Category:    "technical",
		Angles: []string{
			"Explain the L1/L2 split: fast session scratchpad vs persistent semantic vault with sqlite-vec.",
			"14,726 memories, zero cloud dependency: how local vector search beats Pinecone for agent workflows.",
			"Context harvesting: how agents query their own history to pull in relevant past heuristics.",
		},
	},
	{
		Title:       "Building AI Agents That Survive Restarts: Persistent Memory Done Right",
		SEOKeywords: []string{"persistent AI memory", "agent state", "survive restart", "session persistence"},
		Category:    "architecture",
		Angles: []string{
			"Why most agent frameworks forget everything on restart — and how to fix it with SQLite.",
			"Comparing ephemeral vs persistent memory: benchmarks on context restoration time.",
			"The cold start problem: how L2 memory pre-warms new sessions with relevant history.",
		},
	},
	{
		Title:       "The Hidden Cost of Vendor Lock-In in AI Development",
		SEOKeywords: []string{"vendor lock-in", "AI platform independence", "multi-model", "portable AI"},
		Category:    "opinion",
		Angles: []string{
			"Calculate the real cost of being locked into one AI provider: migration effort, retraining, downtime.",
			"How cross-harness tool parity eliminates the switching cost problem entirely.",
			"Why CTOs should demand provider-agnostic AI infrastructure in 2026 RFPs.",
		},
	},
	{
		Title:       "From REPL to Swarm: Scaling AI-Assisted Development for Teams",
		SEOKeywords: []string{"team AI development", "AI pair programming", "scaling AI", "developer velocity"},
		Category:    "patterns",
		Angles: []string{
			"Single-developer AI is table stakes. How swarms enable entire teams to coordinate via shared memory.",
			"Role rotation in practice: how the same model becomes Planner, Implementer, or Critic by swapping system prompts.",
			"Measuring swarm throughput: tasks completed per hour vs solo developer + Copilot.",
		},
	},
	{
		Title:       "SQLite + Vector Search: The Dependency-Free AI Memory Stack",
		SEOKeywords: []string{"sqlite-vec", "vector database", "local embeddings", "semantic search", "dependency-free"},
		Category:    "technical",
		Angles: []string{
			"Why sqlite-vec beats Pinecone, Weaviate, and Chroma for local agent memory — with benchmarks.",
			"Zero dependencies, zero cloud bills: running semantic search on a $5 VPS.",
			"Embedding pipeline deep-dive: from raw text → chunk → vector → similarity score in under 10ms.",
		},
	},
	{
		Title:       "Real-Time AI Observability: Dashboards That Show Actual Database Rows",
		SEOKeywords: []string{"AI observability", "real-time dashboard", "agent monitoring", "debugging AI"},
		Category:    "developer-tools",
		Angles: []string{
			"Truth over hype: why TormentNexus dashboards render actual SQLite rows, not mock data.",
			"What to monitor in an AI agent system: goroutine counts, memory tiers, waterfall history, tool latency.",
			"Building an operator console for AI — lessons from SRE and applied to agent orchestration.",
		},
	},
	{
		Title:       "The Ultimate Offline AI Development Stack",
		SEOKeywords: []string{"offline AI", "air-gapped development", "local LLM", "no cloud AI"},
		Category:    "architecture",
		Angles: []string{
			"Complete walkthrough: LM Studio + Ollama + TormentNexus = full offline AI coding environment.",
			"Air-gapped but not crippled: how the LLM waterfall falls back to local models transparently.",
			"Why defense contractors and fintech are going all-in on local-first AI in 2026.",
		},
	},
	{
		Title:       "MCP Protocol Deep-Dive: How Tool Discovery Actually Works",
		SEOKeywords: []string{"Model Context Protocol", "MCP deep dive", "tool discovery", "JSON-RPC", "MCP internals"},
		Category:    "technical",
		Angles: []string{
			"Under the hood: JSON-RPC handshake → tool enumeration → capability negotiation → progressive injection.",
			"Why dumping all tools into every request is the anti-pattern MCP was designed to solve.",
			"Building your first MCP server in 50 lines of Go and watching TormentNexus auto-discover it.",
		},
	},
	{
		Title:       "AI Skill Registry: 5,776 Reusable Modules and Counting",
		SEOKeywords: []string{"AI skills", "reusable AI modules", "skill registry", "prompt templates", "SKILL.md"},
		Category:    "ecosystem",
		Angles: []string{
			"SKILL.md format: how packaging prompt templates + tool configs as reusable skills transforms AI workflows.",
			"5,776 skills in the registry — from code review to Terraform generation to database migration.",
			"Progressive skill discovery: how TormentNexus auto-loads the right skill based on your current task.",
		},
	},
	{
		Title:       "GitOps for AI Agents: Version-Controlled Tool Configs and Memory",
		SEOKeywords: []string{"GitOps AI", "version controlled AI", "AI configuration management", "infrastructure as code"},
		Category:    "patterns",
		Angles: []string{
			"Treat AI tool configs like infrastructure: mcp.jsonc in git, PR-reviewed, CI-validated.",
			"Rollback your agent's memory: how L2 vault versioning lets you undo bad learning.",
			"Team-wide AI config sync: one git push updates every developer's agent environment.",
		},
	},
	{
		Title:       "Why Your AI Coding Assistant Needs a Control Plane",
		SEOKeywords: []string{"AI control plane", "agent orchestration", "AI operations", "model management"},
		Category:    "opinion",
		Angles: []string{
			"Raw LLM APIs are like raw SQL — powerful but unmanageable at scale. You need a control plane.",
			"The three layers every AI stack needs: tool routing, memory persistence, provider orchestration.",
			"2026 prediction: control planes become as standard as Kubernetes for cloud infrastructure.",
		},
	},
	{
		Title:       "Debate-Driven Development: When AI Agents Argue About Your Code",
		SEOKeywords: []string{"AI debate", "agent consensus", "code review automation", "AI pair review"},
		Category:    "patterns",
		Angles: []string{
			"Two agents spot a bug. One says fix it, one says it's intentional. How consensus resolves the debate.",
			"The Council pattern: multiple agents vote on implementation decisions with human veto.",
			"Adversarial code review: why having an agent critic leads to 30% fewer bugs than solo generation.",
		},
	},
	{
		Title:       "From 0 to Production AI Agent: A Complete Deployment Guide",
		SEOKeywords: []string{"deploy AI agent", "production AI", "AI agent deployment", "self-hosted AI"},
		Category:    "tutorial",
		Angles: []string{
			"Step-by-step: install TormentNexus, configure your first MCP server, connect your LLM provider.",
			"Production checklist: TLS, auth, rate limiting, monitoring, backup — everything your agent needs.",
			"Deploy on a $5 VPS: complete walkthrough with systemd, nginx, and Let's Encrypt.",
		},
	},
	{
		Title:       "The Future of AI Development Is Local-First and Open Source",
		SEOKeywords: []string{"open source AI", "local-first future", "AI democratization", "community AI"},
		Category:    "opinion",
		Angles: []string{
			"Manifesto: why AI infrastructure must be open source, local-first, and community-governed.",
			"The corporate AI walled garden is crumbling — here's what replaces it.",
			"Open source AI in 2026: more tools, more models, more freedom. The golden age is now.",
		},
	},
	{
		Title:       "Automating Technical Outreach: How AI Finds and Engages Early Adopters",
		SEOKeywords: []string{"AI outreach", "automated sales", "lead generation AI", "developer marketing"},
		Category:    "tutorial",
		Angles: []string{
			"How TormentNexus's own marketing agent finds 2K+ leads from GitHub, HN, and LinkedIn automatically.",
			"The 7-state pipeline: Discovered → Researched → Outreach → Engaged → Negotiating → Won/Lost.",
			"Personalization at scale: how LLMs generate custom emails that reference actual code repos.",
		},
	},
	{
		Title:       "Container-Native AI: Running Agent Infrastructure in Docker",
		SEOKeywords: []string{"Docker AI", "container AI", "AI infrastructure", "containerized agents"},
		Category:    "tutorial",
		Angles: []string{
			"Multi-tenant AI: running isolated TormentNexus instances per team with Docker and Traefik.",
			"Docker Compose for AI dev environments: one command spins up LLM, memory, tools, and dashboard.",
			"Container resource management for AI: GPU passthrough, memory limits, and auto-scaling agents.",
		},
	},
	{
		Title:       "How We Built an AI Marketing Agent That Sends 100+ Personalized Emails",
		SEOKeywords: []string{"AI marketing agent", "automated email", "personalization", "developer outreach"},
		Category:    "case-study",
		Angles: []string{
			"Architecture deep-dive: scraper → enricher → researcher → communicator → CRM sync.",
			"What worked: GitHub enrichment, objection handling, A/B tested email templates.",
			"What didn't: Apollo API limits, Reddit bot detection, HN API changes. Lessons for builders.",
		},
	},
	{
		Title:       "The Go + TypeScript Monolith: Why TormentNexus Uses Two Languages",
		SEOKeywords: []string{"Go TypeScript monolith", "polyglot architecture", "AI backend", "modular monolith"},
		Category:    "architecture",
		Angles: []string{
			"Why Go for the kernel (446 HTTP handlers, goroutines).",
			"Modular monolith vs microservices: why TormentNexus chose one deployable with 35+ internal packages.",
		},
	},
	{
		Title:       "EDA for AI: Event-Driven Architecture Patterns in Agent Systems",
		SEOKeywords: []string{"event-driven AI", "EDA agent", "Swarm event bus", "async AI patterns"},
		Category:    "patterns",
		Angles: []string{
			"The Swarm EventBus: how 35+ Go packages communicate via high-frequency typed events.",
			"Event sourcing for AI memory: replaying agent events to reconstruct full session context.",
			"Pub/sub for multi-agent: how events let Planner, Implementer, and Critic stay synchronized.",
		},
	},
	{
		Title:       "Securing Self-Hosted AI: TLS, Auth, and Network Isolation",
		SEOKeywords: []string{"AI security", "self-hosted security", "TLS AI", "network isolation", "zero trust AI"},
		Category:    "enterprise",
		Angles: []string{
			"Hardening checklist: TLS termination, Ed25519 JWT signing, RBAC middleware, audit logging.",
			"Network isolation patterns: running the AI kernel on localhost with nginx reverse proxy.",
			"Zero-trust AI architecture: every tool call, memory access, and model request is authenticated.",
		},
	},
}

// NewBlogEngine creates a blog generation engine.
func NewBlogEngine(llmProvider llm.LLMProvider, outputDir string) *BlogEngine {
	be := &BlogEngine{
		llm:       llmProvider,
		outputDir: outputDir,
		dryRun:    os.Getenv("DRY_RUN") == "true",
		devto:     NewDevToPublisher(""),
		hashnode:  NewHashnodePublisher(""),
	}
	be.loadState()
	return be
}

// loadState reads persisted engine state from disk.
func (be *BlogEngine) loadState() {
	path := filepath.Join(be.outputDir, "blog_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		be.state = blogState{
			LastTopicAt: make(map[string]string),
		}
		return
	}
	if err := json.Unmarshal(data, &be.state); err != nil {
		be.state = blogState{
			LastTopicAt: make(map[string]string),
		}
		return
	}
	if be.state.LastTopicAt == nil {
		be.state.LastTopicAt = make(map[string]string)
	}
	slog.Info(fmt.Sprintf("BlogEngine: Loaded state — %d posts, %d unique topics used",
		be.state.TotalPosts, len(be.state.UsedTopics)))
}

// saveState persists engine state to disk.
func (be *BlogEngine) saveState() {
	be.state.LastRun = time.Now().UTC().Format(time.RFC3339)
	path := filepath.Join(be.outputDir, "blog_state.json")
	data, _ := json.MarshalIndent(be.state, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
}

// Run starts the periodic blog generation loop.
func (be *BlogEngine) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("BlogEngine: Auto-blog generator started (interval: %v, output: %s, topics: %d)",
		interval, be.outputDir, len(blogTopics)))

	// Generate first post immediately
	be.GenerateNextPost(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("BlogEngine: Stopping...")
			return
		case <-ticker.C:
			be.GenerateNextPost(ctx)
		}
	}
}

// nextTopic selects the least-recently-used topic, preferring unused ones.
// This ensures all 28 topics cycle before any repeats.
func (be *BlogEngine) nextTopic() BlogTopic {
	be.mu.Lock()
	defer be.mu.Unlock()

	// Build set of used topic titles for O(1) lookup
	used := make(map[string]bool)
	for _, t := range be.state.UsedTopics {
		used[t] = true
	}

	// Find unused topics
	var unused []BlogTopic
	for _, t := range blogTopics {
		if !used[t.Title] {
			unused = append(unused, t)
		}
	}

	var selected BlogTopic
	if len(unused) > 0 {
		// Pick a random unused topic for variety
		selected = unused[rand.Intn(len(unused))]
	} else {
		// All topics used — find the least recently used
		slog.Info("BlogEngine: All topics used, cycling to least-recently-used")
		var oldest time.Time
		oldestIdx := 0
		for i, t := range blogTopics {
			at, ok := be.state.LastTopicAt[t.Title]
			if !ok {
				selected = t
				break
			}
			parsed, err := time.Parse(time.RFC3339, at)
			if err != nil || i == 0 || parsed.Before(oldest) {
				oldest = parsed
				oldestIdx = i
			}
		}
		if selected.Title == "" {
			selected = blogTopics[oldestIdx]
		}
	}

	// Record usage
	be.state.UsedTopics = append(be.state.UsedTopics, selected.Title)
	be.state.LastTopicAt[selected.Title] = time.Now().UTC().Format(time.RFC3339)
	be.state.TotalPosts++
	be.saveState()

	return selected
}

func (be *BlogEngine) GenerateNextPost(ctx context.Context) {
	topic := be.nextTopic()

	be.generateAndPublish(ctx, topic)
}

// GenerateBatch generates up to N posts sequentially, waiting for each to complete.
// Returns how many were generated. Useful for bulk-filling the blog catalog.
func (be *BlogEngine) GenerateBatch(ctx context.Context, n int) int {
	count := 0
	for i := 0; i < n; i++ {
		topic := be.nextTopic()
		be.generateAndPublish(ctx, topic)
		count++
		// Quick pause between generations to avoid rate limits
		time.Sleep(2 * time.Second)
	}
	return count
}

func (be *BlogEngine) generateAndPublish(ctx context.Context, topic BlogTopic) {

	brand := "tormentnexus"
	if topic.Category == "enterprise" {
		brand = "hypernexus"
	}

	// Randomly pick one of the topic's angle variations
	angle := topic.Angles[0]
	if len(topic.Angles) > 1 {
		angle = topic.Angles[rand.Intn(len(topic.Angles))]
	}

	post, err := be.generatePost(ctx, topic, brand, angle)
	if err != nil {
		slog.Error("BlogEngine: Failed to generate post", "error", err, "topic", topic.Title)
		return
	}

	if be.dryRun {
		slog.Info(fmt.Sprintf("BlogEngine [DRY RUN] Would publish: %s", post.Title))
		return
	}

	if err := be.publish(ctx, post); err != nil {
		slog.Error("BlogEngine: Failed to publish post", "error", err, "slug", post.Slug)
		return
	}

	slog.Info(fmt.Sprintf("BlogEngine: Published \"%s\" (%s) — post #%d",
		post.Title, post.Slug, be.state.TotalPosts))
}

func (be *BlogEngine) generatePost(ctx context.Context, topic BlogTopic, brand string, angle string) (*BlogPost, error) {
	seo := strings.Join(topic.SEOKeywords, ", ")

	brandName := "TormentNexus"
	brandURL := "https://tormentnexus.site"
	if brand == "hypernexus" {
		brandName = "HyperNexus"
		brandURL = "https://hypernexus.site"
	}

	systemPrompt := fmt.Sprintf(`You are an expert technical content writer for AI developer tools.
Write a comprehensive, SEO-optimized blog post for %s (%s).

CRITICAL: This post must be COMPLETELY UNIQUE. Do not reuse any content, paragraphs,
or phrases from previous posts. Generate entirely fresh content every time.

Brand voice: authoritative, developer-focused, backed by specific technical details.

Format the response as clean HTML article body (NO <!DOCTYPE>, NO <html> or <body> tags):

<h1>A compelling, unique title</h1>
<p class="meta-description">2-3 sentence SEO meta description</p>

<h2>First Section</h2>
<p>Engaging paragraphs...</p>
<p>More content...</p>

<h2>Second Section</h2>
...

<p class="cta">Call to action linking to %s</p>

Include 4-6 h2 sections. Use <pre><code> blocks for code examples.
Naturally include these SEO keywords: %s

Be specific. Use real numbers, real examples, real scenarios.
DO NOT use placeholder text, lorem ipsum, or generic filler.`, brandName, brandURL, brandURL, seo)

	content, err := be.llm.Generate(ctx, llm.Prompt{
		System:    systemPrompt,
		User:      fmt.Sprintf("Topic: %s\nAngle approach: %s\nWrite the full unique blog post now. Generate FRESH content — do not repeat any previous posts.", topic.Title, angle),
		MaxTokens: 3000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Extract title from the generated h1
	title := topic.Title
	content = strings.TrimSpace(content)

	// Try to extract h1
	if idx := strings.Index(content, "<h1>"); idx >= 0 {
		endIdx := strings.Index(content[idx:], "</h1>")
		if endIdx >= 0 {
			extracted := strings.TrimPrefix(content[idx:idx+endIdx], "<h1>")
			extracted = strings.TrimSpace(extracted)
			if len(extracted) > 5 {
				title = extracted
			}
		}
	}

	// Try to extract meta description
	excerpt := ""
	if idx := strings.Index(content, "meta-description"); idx >= 0 {
		endIdx := strings.Index(content[idx:], "</p>")
		if endIdx >= 0 {
			raw := content[idx : idx+endIdx]
			raw = strings.TrimSpace(raw)
			// Strip the class attr
			if gt := strings.Index(raw, ">"); gt >= 0 {
				excerpt = strings.TrimSpace(raw[gt+1:])
			}
		}
	}
	if excerpt == "" {
		// Fallback: first substantial paragraph
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if len(line) > 60 && strings.HasPrefix(line, "<p>") && !strings.Contains(line, "meta-description") {
				excerpt = stripTags(line)
				if len(excerpt) > 250 {
					excerpt = excerpt[:250] + "..."
				}
				break
			}
		}
	}

	slug := slugify(title)
	now := time.Now()

	return &BlogPost{
		Title:       title,
		Slug:        slug,
		Brand:       brand,
		Category:    topic.Category,
		Content:     content,
		Excerpt:     excerpt,
		PublishedAt: now,
	}, nil
}

// publish writes the blog post as an HTML file to the output directory.
func (be *BlogEngine) publish(ctx context.Context, post *BlogPost) error {
	dir := filepath.Join(be.outputDir, "blog", post.Brand)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	html := blogPostHTML(post)
	filename := filepath.Join(dir, post.Slug+".html")
	if err := os.WriteFile(filename, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// Update index.json for the blog listing page
	indexPath := filepath.Join(dir, "index.json")
	var index []BlogPost
	if data, err := os.ReadFile(indexPath); err == nil {
		_ = json.Unmarshal(data, &index)
	}
	index = append([]BlogPost{*post}, index...)
	if len(index) > 50 {
		index = index[:50]
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	_ = os.WriteFile(indexPath, indexData, 0o644)

	// Cross-post to dev.to and hashnode asynchronously
	go CrossPostBlog(context.Background(), be.devto, be.hashnode, post)

	return nil
}

func blogPostHTML(post *BlogPost) string {
	dateStr := post.PublishedAt.Format("January 2, 2006")
	brandName := "TormentNexus"
	brandURL := "https://tormentnexus.site"
	if post.Brand == "hypernexus" {
		brandName = "HyperNexus"
		brandURL = "https://hypernexus.site"
	}

	return fmt.Sprintf(`<article class="blog-post" data-slug="%s" data-brand="%s" data-category="%s">
  <header>
    <h1>%s</h1>
    <div class="meta">
      <span class="date">%s</span>
      <span class="brand"><a href="%s">%s</a></span>
      <span class="category">%s</span>
    </div>
  </header>
  <div class="content">
%s
  </div>
  <footer>
    <p>Published by %s — the OS for AI models.</p>
  </footer>
</article>`, post.Slug, post.Brand, post.Category,
		post.Title, dateStr,
		brandURL, brandName, post.Category,
		post.Content,
		brandName)
}

func stripTags(s string) string {
	result := s
	for {
		start := strings.Index(result, "<")
		if start < 0 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end < 0 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return strings.TrimSpace(result)
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "-")
}
