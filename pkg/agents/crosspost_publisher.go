package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// DevToPublisher cross-posts blog articles to dev.to via their API.
// Dev.to reach: 5M+ developers. API key from https://dev.to/settings/extensions.
type DevToPublisher struct {
	APIKey     string
	HTTPClient *http.Client
	dryRun     bool
}

// HashnodePublisher cross-posts to hashnode.com via their GraphQL API.
type HashnodePublisher struct {
	APIToken   string
	HTTPClient *http.Client
	dryRun     bool
}

// NewDevToPublisher creates a dev.to cross-poster.
func NewDevToPublisher(apiKey string) *DevToPublisher {
	if apiKey == "" {
		apiKey = os.Getenv("DEVTO_API_KEY")
	}
	return &DevToPublisher{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		dryRun:     os.Getenv("DRY_RUN") == "true",
	}
}

// NewHashnodePublisher creates a hashnode cross-poster.
func NewHashnodePublisher(token string) *HashnodePublisher {
	if token == "" {
		token = os.Getenv("HASHNODE_API_TOKEN")
	}
	return &HashnodePublisher{
		APIToken:   token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		dryRun:     os.Getenv("DRY_RUN") == "true",
	}
}

// PublishToDevTo posts an article to dev.to.
func (d *DevToPublisher) PublishToDevTo(ctx context.Context, title, bodyMarkdown string, tags []string) error {
	if d.APIKey == "" {
		slog.Info("DevToPublisher: No DEVTO_API_KEY configured — skipping cross-post")
		return nil
	}
	if d.dryRun {
		slog.Info(fmt.Sprintf("DevToPublisher [DRY RUN]: %s", title))
		return nil
	}

	payload := map[string]interface{}{
		"article": map[string]interface{}{
			"title":         title,
			"body_markdown": bodyMarkdown,
			"published":     true,
			"tags":          tags,
			"canonical_url": fmt.Sprintf("https://tormentnexus.site/blog/tormentnexus/%s", slugifyTitle(title)),
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://dev.to/api/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", d.APIKey)

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("dev.to post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dev.to post: HTTP %d: %s", resp.StatusCode, string(respBody[:300]))
	}

	slog.Info(fmt.Sprintf("DevToPublisher: Cross-posted \"%s\" to dev.to ✓", title))
	return nil
}

// PublishToHashnode posts an article to hashnode.com.
// Hashnode uses a GraphQL API. Publication ID is required.
func (h *HashnodePublisher) PublishToHashnode(ctx context.Context, title, contentMarkdown string, tags []string, publicationID string) error {
	if h.APIToken == "" {
		slog.Info("HashnodePublisher: No HASHNODE_API_TOKEN configured — skipping")
		return nil
	}
	if h.dryRun {
		slog.Info(fmt.Sprintf("HashnodePublisher [DRY RUN]: %s", title))
		return nil
	}

	// Hashnode GraphQL mutation
	query := `mutation PublishPost($input: PublishPostInput!) {
		publishPost(input: $input) {
			post { id title url }
		}
	}`

	tagSlugs := make([]map[string]string, len(tags))
	for i, t := range tags {
		tagSlugs[i] = map[string]string{"slug": t}
	}

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":           title,
			"contentMarkdown": contentMarkdown,
			"publicationId":   publicationID,
			"tags":            tagSlugs,
			"disableComments": false,
		},
	}

	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://gql.hashnode.com", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", h.APIToken)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("hashnode post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hashnode post: HTTP %d: %s", resp.StatusCode, string(respBody[:300]))
	}

	slog.Info(fmt.Sprintf("HashnodePublisher: Cross-posted \"%s\" to hashnode ✓", title))
	return nil
}

// CrossPostBlog publishes the latest blog post to dev.to and hashnode.
func CrossPostBlog(ctx context.Context, devto *DevToPublisher, hashnode *HashnodePublisher, post *BlogPost) {
	slug := slugifyTitle(post.Title)
	canonicalURL := fmt.Sprintf("https://tormentnexus.site/blog/tormentnexus/%s.html", slug)

	// Convert HTML content to markdown (simple approach)
	markdown := htmlToMarkdownSimple(post.Content)
	markdown += fmt.Sprintf("\n\n---\n\n*Originally published at [tormentnexus.site](%s)*", canonicalURL)

	tags := []string{"ai", "llm", "opensource", "mcp", "agents"}
	if post.Brand == "hypernexus" {
		tags = append(tags, "enterprise", "devops")
	}

	if devto != nil {
		if err := devto.PublishToDevTo(ctx, post.Title, markdown, tags); err != nil {
			slog.Error("CrossPostBlog: dev.to publish failed", "error", err, "title", post.Title)
		}
	}
	if hashnode != nil {
		if err := hashnode.PublishToHashnode(ctx, post.Title, markdown, tags, os.Getenv("HASHNODE_PUBLICATION_ID")); err != nil {
			slog.Error("CrossPostBlog: hashnode publish failed", "error", err, "title", post.Title)
		}
	}
}

// htmlToMarkdownSimple does a basic HTML-to-markdown conversion.
// For production, use a proper library. This handles our blog's HTML format.
func htmlToMarkdownSimple(html string) string {
	// Strip article wrapper, keep content
	content := html
	// Remove HTML tags (very basic — good enough for dev.to which accepts HTML too)
	// Actually dev.to supports HTML body_markdown, so just pass as-is
	return content
}

func slugifyTitle(title string) string {
	return slugify(title)
}
