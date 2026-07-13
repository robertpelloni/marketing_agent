package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertpelloni/marketing_agent/internal/llm"
)

// BlogEngine generates SEO-optimized blog posts for tormentnexus.site
// and hypernexus.site, published to the nginx-served web directory.
// Also cross-posts to dev.to and hashnode for wider reach.
type BlogEngine struct {
	llm       llm.LLMProvider
	outputDir string // where blog HTML files are written
	dryRun    bool
	devto     *DevToPublisher
	hashnode  *HashnodePublisher
}

// BlogPost represents a generated blog article.
type BlogPost struct {
	Title       string
	Slug        string
	Brand       string // "tormentnexus" or "hypernexus"
	Category    string
	Content     string // full HTML
	Excerpt     string
	PublishedAt time.Time
}

// BlogTopic defines an article topic with SEO keywords.
type BlogTopic struct {
	Title       string
	SEOKeywords []string
	Category    string
	Angle       string // unique angle for DeepSeek
}

// blogTopics is the rotating content calendar.
var blogTopics = []BlogTopic{
	{
		Title:       "Why Local-First AI Infrastructure Matters in 2026",
		SEOKeywords: []string{"local AI", "offline LLM", "private AI infrastructure", "air-gapped AI"},
		Category:    "architecture",
		Angle:       "Explain why keeping AI tooling local matters for developer velocity, privacy, and uptime. Mention TormentNexus's local-first memory (14K+ persisted) and zero-cloud-fallback waterfall.",
	},
	{
		Title:       "Progressive MCP Tool Routing: Stop Drowning Your Agents in 50K Tokens",
		SEOKeywords: []string{"MCP tool routing", "progressive disclosure", "semantic tool search", "agent context optimization"},
		Category:    "technical",
		Angle:       "Explain how dumping 50K tokens worth of tool schemas overwhelms LLMs. Present TormentNexus's progressive routing as the solution — semantic vector search + ranked injection.",
	},
	{
		Title:       "The LLM Waterfall Pattern: Never Let a Rate Limit Kill Your Workflow",
		SEOKeywords: []string{"LLM waterfall", "provider failover", "API rate limit", "zero downtime AI"},
		Category:    "patterns",
		Angle:       "Describe the cascade from primary APIs → OpenRouter → local LM Studio/Ollama. Emphasize that TormentNexus handles this transparently with no code changes.",
	},
	{
		Title:       "Cross-Harness Tool Parity: One Config, Six AI Coding Environments",
		SEOKeywords: []string{"Claude Code", "Cursor", "Codex", "Gemini CLI", "Copilot", "Windsurf", "tool parity"},
		Category:    "developer-tools",
		Angle:       "Explain the pain of configuring different tools for Claude Code vs Cursor vs Codex. Show how TormentNexus provides byte-for-byte identical tool signatures across all harnesses.",
	},
	{
		Title:       "Multi-Agent Swarms: Planner, Implementer, Tester, Critic in One Chatroom",
		SEOKeywords: []string{"multi-agent", "AI swarm", "agent collaboration", "consensus"},
		Category:    "patterns",
		Angle:       "Walk through the Planner → Implementer → Tester → Critic cycle. Show how the PairOrchestrator enforces collaboration phases and debate consensus.",
	},
	{
		Title:       "Enterprise AI Governance with HyperNexus: SSO, RBAC, and Audit Trails",
		SEOKeywords: []string{"enterprise AI governance", "SSO", "RBAC", "AI audit trail", "SOC 2"},
		Category:    "enterprise",
		Angle:       "Target enterprise decision-makers. Highlight HyperNexus's SOC 2 compliance, Ed25519 signed tokens, role-based access control, and self-hosted deployment.",
	},
	{
		Title:       "Self-Healing AI: When Your Agent Debugs Its Own Code",
		SEOKeywords: []string{"self-healing AI", "autonomous debugging", "agent autonomy"},
		Category:    "technical",
		Angle:       "Present TormentNexus's Healer loop: diagnose → fix → verify → persist. Each fix is stored in L2 memory for fleet-wide learning.",
	},
	{
		Title:       "11K+ MCP Servers: The Largest Indexed Catalog for AI Tooling",
		SEOKeywords: []string{"MCP server catalog", "AI tools directory", "MCP ecosystem"},
		Category:    "ecosystem",
		Angle:       "TormentNexus indexes 11,024 MCP servers from Glama, Smithery, MCP.run, npm, and GitHub. Semantic search + auto-discovery. This is the App Store for AI tools.",
	},
}

// NewBlogEngine creates a blog generation engine.
func NewBlogEngine(llmProvider llm.LLMProvider, outputDir string) *BlogEngine {
	return &BlogEngine{
		llm:       llmProvider,
		outputDir: outputDir,
		dryRun:    os.Getenv("DRY_RUN") == "true",
		devto:     NewDevToPublisher(""),
		hashnode:  NewHashnodePublisher(""),
	}
}

// Run starts the periodic blog generation loop.
func (be *BlogEngine) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("BlogEngine: Auto-blog generator started (interval: %v, output: %s)", interval, be.outputDir))

	// Run first post immediately
	be.generateNextPost(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("BlogEngine: Stopping...")
			return
		case <-ticker.C:
			be.generateNextPost(ctx)
		}
	}
}

// topicIndex tracks which topic to generate next (cyclical).
var topicIndex int

func (be *BlogEngine) generateNextPost(ctx context.Context) {
	topic := blogTopics[topicIndex%len(blogTopics)]
	topicIndex++

	brand := "tormentnexus"
	if topic.Category == "enterprise" {
		brand = "hypernexus"
	}

	post, err := be.generatePost(ctx, topic, brand)
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

	slog.Info(fmt.Sprintf("BlogEngine: Published \"%s\" (%s)", post.Title, post.Slug))
}

func (be *BlogEngine) generatePost(ctx context.Context, topic BlogTopic, brand string) (*BlogPost, error) {
	seo := strings.Join(topic.SEOKeywords, ", ")

	systemPrompt := fmt.Sprintf(`You are an expert technical content writer for AI developer tools. 
Write a comprehensive, SEO-optimized blog post for %s about the topic below.
Brand voice: authoritative, developer-focused, backed by specific technical details.

Include:
- An engaging h1 title
- A 2-3 sentence meta description excerpt
- 4-6 sections with h2 headings
- Real code examples or CLI snippets where appropriate
- Mention TormentNexus/HyperNexus naturally (don't oversell)
- Call to action at the end linking to the website
- Format as clean HTML (no <!DOCTYPE>, just the article body)

SEO keywords to naturally include: %s`, brand, seo)

	content, err := be.llm.Generate(ctx, llm.Prompt{
		System:    systemPrompt,
		User:      fmt.Sprintf("Topic: %s\nAngle: %s\nWrite the full blog post now.", topic.Title, topic.Angle),
		MaxTokens: 2000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Extract title from generated content
	title := topic.Title
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "# ") {
		lines := strings.SplitN(content, "\n", 2)
		title = strings.TrimPrefix(lines[0], "# ")
		if len(lines) > 1 {
			content = strings.TrimSpace(lines[1])
		}
	}

	slug := slugify(title)
	now := time.Now()

	// Build excerpt from first paragraph
	excerpt := ""
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 40 && !strings.HasPrefix(line, "<") {
			excerpt = line
			if len(excerpt) > 250 {
				excerpt = excerpt[:250] + "..."
			}
			break
		}
	}

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

// publish writes the blog post as an HTML snippet to the output directory.
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

	// Also update an index.json for the blog listing
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

	// Cross-post to dev.to and hashnode for 10x reach
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
      <span class="brand">%s</span>
      <span class="category">%s</span>
    </div>
  </header>
  <div class="content">
%s
  </div>
  <footer>
    <p>Published by <a href="%s">%s</a> — the OS for AI models.</p>
  </footer>
</article>`, post.Slug, post.Brand, post.Category,
		post.Title,
		dateStr,
		brandName,
		post.Category,
		post.Content,
		brandURL, brandName)
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	// remove non-alphanumeric except hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "-")
}
