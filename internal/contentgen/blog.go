package contentgen

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

const targetLength = 100000

var blogTopics = []string{
	"Progressive MCP Tool Routing: Why your agents don't need 50K tokens of tools",
	"Cross-Harness Parity: One config for Claude, Codex, Cursor, Copilot, Gemini, Kiro",
	"Local-First AI: Why your LLM infrastructure shouldn't depend on the cloud",
	"The LLM Waterfall: Zero-downtime inference with NVIDIA → OpenRouter → Local fallback",
	"Dual-Tier Memory Architecture: 14K+ persistent memories that survive restarts",
	"Multi-Agent Swarms: How Planner-Implementer-Tester-Critic collaboration works",
	"Self-Healing AI: The diagnose-fix-verify-retry closed loop",
	"Why TormentNexus uses sqlite-vec for vector search instead of Pinecone",
	"Progressive Disclosure: The anti-pattern of dumping every tool into context",
	"Agent-to-Agent Protocol: How AI agents negotiate and debate in shared chatrooms",
	"The Go Sidecar Pattern: Why we wrote our AI kernel in Go, not Python",
	"Cross-Harness Tool Parity: Byte-for-byte identical tools across 6 AI coding platforms",
	"Local-First Memory: Why your AI shouldn't need the cloud to remember",
	"11K MCP Servers: The largest indexed catalog of AI tools",
	"From NVIDIA NIM to Local Ollama: Building a resilient LLM provider cascade",
}

// BlogGenerator creates and continuously expands blog posts to 100K+ chars.
type BlogGenerator struct {
	llm       llm.LLMProvider
	db        *db.DB
	outputDir string
	posts     map[string]string // topic -> filepath
}

// NewBlogGenerator creates a new blog post generator.
func NewBlogGenerator(provider llm.LLMProvider, database *db.DB) *BlogGenerator {
	return &BlogGenerator{
		llm:       provider,
		db:        database,
		outputDir: "docs/blog",
		posts:     make(map[string]string),
	}
}

// Run starts the blog generation/expansion loop.
func (b *BlogGenerator) Run(ctx context.Context, interval time.Duration) {
	slog.Info("BlogGenerator: Starting autonomous content pipeline (target: 100K chars per post)...")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Create output directory
	os.MkdirAll(b.outputDir, 0755)

	// Initialize post tracking
	for _, topic := range blogTopics {
		b.posts[topic] = b.postPath(topic)
	}

	// Initial pass: generate all posts if they don't exist
	b.ensureAllPostsExist(ctx)

	// Continuous expansion loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("BlogGenerator: Stopping...")
			return
		case <-ticker.C:
			b.expandCycle(ctx)
		}
	}
}

func (b *BlogGenerator) postPath(topic string) string {
	slug := strings.ToLower(strings.ReplaceAll(topic, " ", "-"))
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "→", "to")
	if len(slug) > 60 {
		slug = slug[:60]
	}
	return filepath.Join(b.outputDir, slug+".md")
}

func (b *BlogGenerator) ensureAllPostsExist(ctx context.Context) {
	for _, topic := range blogTopics {
		path := b.posts[topic]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			slog.Info(fmt.Sprintf("BlogGenerator: Creating initial post: %s", topic))
			content := b.generateInitialPost(ctx, topic)
			if content != "" {
				b.writePost(path, topic, content)
			}
		}
	}
}

func (b *BlogGenerator) expandCycle(ctx context.Context) {
	// Find the shortest post first (priority to posts farthest from target)
	type postInfo struct {
		topic  string
		path   string
		length int
	}

	var posts []postInfo
	for _, topic := range blogTopics {
		path := b.posts[topic]
		content, err := os.ReadFile(path)
		length := 0
		if err == nil {
			length = len(content)
		}
		posts = append(posts, postInfo{topic, path, length})
	}

	// Sort by length ascending (shortest first)
	for i := 0; i < len(posts); i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[j].length < posts[i].length {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	// Work on the shortest post first
	for _, p := range posts {
		if p.length >= targetLength {
			continue // Already at target
		}

		remaining := targetLength - p.length
		slog.Info(fmt.Sprintf("BlogGenerator: Expanding \"%s\" — %d / %d chars (%d remaining)",
			p.topic, p.length, targetLength, remaining))

		b.expandPost(ctx, p.topic, p.path)
		return // One expansion per cycle
	}

	// All posts at target — proofread cycle
	slog.Info("BlogGenerator: All posts at 100K+ chars. Running proofread cycle.")
	b.proofreadCycle(ctx)
}

func (b *BlogGenerator) expandPost(ctx context.Context, topic, path string) {
	existing, err := os.ReadFile(path)
	if err != nil {
		slog.Error(fmt.Sprintf("BlogGenerator: Cannot read %s: %v", path, err))
		return
	}

	currentText := string(existing)
	currentLen := len(currentText)
	remaining := targetLength - currentLen
	if remaining <= 0 {
		return
	}

	// Request new section to append
	targetNew := remaining
	if targetNew > 8000 {
		targetNew = 8000 // Generate in chunks of ~8K chars
	}

	prompt := llm.Prompt{
		System: `You are continuing an existing technical blog post about TormentNexus.
Write the NEXT section that naturally follows from where the post currently ends.
Maintain the same voice: technically deep, authoritative, specific.
Include code snippets, architecture details, benchmarks, or real metrics where possible.
DO NOT repeat what was already written. DO NOT summarize. Extend with new content.`,
		User: fmt.Sprintf(`The following blog post currently has %d characters. It needs to reach 100,000 characters.
Write approximately %d characters of NEW content that extends this post with additional technical depth, 
use cases, architecture details, performance analysis, or implementation guidance.

TOPIC: %s

CURRENT POST ENDING:
%s

Write the next section that naturally continues from where it ends. Include a section header.`, currentLen, targetNew, topic, currentText[len(currentText)-500:]), // Last 500 chars for context
		MaxTokens: 4096,
	}

	newContent, err := b.llm.Generate(ctx, prompt)
	if err != nil {
		slog.Error(fmt.Sprintf("BlogGenerator: LLM expansion failed: %v", err))
		return
	}

	// Append to existing content
	updated := currentText + "\n\n" + newContent
	b.writePost(path, topic, updated)

	newLen := len(updated)
	slog.Info(fmt.Sprintf("BlogGenerator: Expanded \"%s\" by %d chars → %d / %d total",
		topic, newLen-currentLen, newLen, targetLength))
}

func (b *BlogGenerator) proofreadCycle(ctx context.Context) {
	for _, topic := range blogTopics {
		path := b.posts[topic]
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		prompt := llm.Prompt{
			System: `You are a technical editor reviewing a blog post about TormentNexus.
Review the content for:
1. Technical accuracy — flag any incorrect claims
2. Clarity — identify confusing or unclear passages
3. Consistency — ensure terminology is used consistently
4. Redundancy — flag repeated content that should be removed
5. Grammar and style issues

Provide a concise list of issues found, if any. If the post is clean, say "POST CLEAN".`,
			User: fmt.Sprintf(`Review the following blog post for issues. Topic: %s

POST CONTENT (first 2000 chars):
%s

POST CONTENT (last 2000 chars):
%s`, topic, string(content[:min(len(content), 2000)]),
				string(content[max(0, len(content)-2000):])),
			MaxTokens: 1024,
		}

		review, err := b.llm.Generate(ctx, prompt)
		if err != nil {
			continue
		}

		if !strings.Contains(review, "POST CLEAN") && len(review) > 50 {
			slog.Info(fmt.Sprintf("BlogGenerator: Proofread \"%s\" — issues found. Revising...", topic))
			b.revisePost(ctx, topic, path, string(content), review)
		} else {
			slog.Info(fmt.Sprintf("BlogGenerator: Proofread \"%s\" — clean.", topic))
		}
	}
}

func (b *BlogGenerator) revisePost(ctx context.Context, topic, path, content, review string) {
	prompt := llm.Prompt{
		System: `You are revising a technical blog post based on an editor's review.
Fix the identified issues while preserving the original content's structure and voice.
Output the COMPLETE revised post — every word, every section.`,
		User: fmt.Sprintf(`Revise this blog post based on the editor's review.

TOPIC: %s

EDITOR REVIEW:
%s

FULL POST CONTENT:
%s`, topic, review, content),
		MaxTokens: 8192,
	}

	revised, err := b.llm.Generate(ctx, prompt)
	if err != nil {
		slog.Error(fmt.Sprintf("BlogGenerator: Revision failed for %s: %v", topic, err))
		return
	}

	b.writePost(path, topic, revised)
	slog.Info(fmt.Sprintf("BlogGenerator: Revised \"%s\" based on proofread feedback.", topic))
}

func (b *BlogGenerator) generateInitialPost(ctx context.Context, topic string) string {
	prompt := llm.Prompt{
		System: `You are a technical content writer for TormentNexus, an AI operating system.
Write comprehensive, technically deep blog posts that appeal to senior engineers.
Style: Clear, authoritative, specific. Include technical details, architecture insights, and honest tradeoffs.
Format as Markdown with title, introduction, 5-8 sections, and conclusion.
Length: Begin with 3000-5000 characters — more will be added later.`,
		User: fmt.Sprintf(`Write a blog post about: "%s"

TormentNexus is a local-first cognitive control plane written in Go+TypeScript.
Key details to reference if relevant:
- Go kernel: 232 files, 446 HTTP handlers, port 4300
- TS control plane: 583 files, tRPC middleware, port 4100
- SQLite + sqlite-vec: 14K+ memories, 11K+ MCP servers
- Provider cascade: NVIDIA NIM → OpenRouter → LM Studio
- Cross-harness parity: 6 platforms, 27 golden fixtures
- Progressive MCP tool routing with LRU eviction
- A2A protocol with role rotation and consensus`, topic),
		MaxTokens: 4096,
	}

	content, err := b.llm.Generate(ctx, prompt)
	if err != nil {
		slog.Error(fmt.Sprintf("BlogGenerator: Initial generation failed: %v", err))
		return ""
	}
	return content
}

func (b *BlogGenerator) writePost(path, topic, content string) {
	header := fmt.Sprintf(`---
title: "%s"
date: %s
author: TormentNexus AI
status: expanding
chars: %d
---

`, topic, time.Now().Format(time.RFC3339), len(content))

	fullContent := header + content
	if err := os.WriteFile(path, []byte(fullContent), 0644); err != nil {
		slog.Error(fmt.Sprintf("BlogGenerator: Failed to write %s: %v", path, err))
	} else {
		slog.Info(fmt.Sprintf("BlogGenerator: Wrote %d chars to %s", len(fullContent), path))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
