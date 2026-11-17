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
	"strings"
	"time"
)

// GitHubDiscussionsEngager searches for relevant GitHub Discussions in the
// MCP/AI ecosystem and posts helpful responses mentioning TormentNexus.
// GitHub Discussions are high-signal: developers actively asking for solutions.
type GitHubDiscussionsEngager struct {
	Token      string
	HTTPClient *http.Client
	dryRun     bool
	apiURL     string
}

// NewGitHubDiscussionsEngager creates a GitHub Discussions engagement bot.
func NewGitHubDiscussionsEngager(token string) *GitHubDiscussionsEngager {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &GitHubDiscussionsEngager{
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		dryRun:     os.Getenv("DRY_RUN") == "true",
		apiURL:     "https://api.github.com/graphql",
	}
}

// Run starts periodic discussion search and engagement.
func (g *GitHubDiscussionsEngager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("GitHubDiscussions: Engagement bot started (interval: %v)", interval))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.engage(ctx)
		}
	}
}

func (g *GitHubDiscussionsEngager) engage(ctx context.Context) {
	if g.Token == "" {
		slog.Info("GitHubDiscussions: No GITHUB_TOKEN — skipping")
		return
	}

	// Posting is operator-gated: with GITHUB_POSTING_DISABLED=true this
	// worker must not search or reply at all (prod stance since the token
	// refresh; replies would be real public comments).
	if os.Getenv("GITHUB_POSTING_DISABLED") == "true" {
		slog.Info("GitHubDiscussions: Posting disabled via GITHUB_POSTING_DISABLED — skipping cycle")
		return
	}

	// Search for MCP/agent-related discussions
	queries := []string{
		"MCP tool routing OR model context protocol in:title",
		"multi-agent orchestration LLM in:title",
		"agent memory persistence local in:title",
		"LLM waterfall failover in:title",
		"Claude Code tool parity in:title",
	}

	for _, query := range queries {
		discussions, err := g.searchDiscussions(ctx, query)
		if err != nil {
			slog.Warn("GitHubDiscussions: search failed", "query", query, "error", err)
			continue
		}

		for _, d := range discussions {
			if g.dryRun {
				slog.Info(fmt.Sprintf("GitHubDiscussions [DRY RUN]: %s — %s", d.Title, d.URL))
				continue
			}

			comment := g.buildComment(d)
			if err := g.postReply(ctx, d.DiscussionNodeID, comment); err != nil {
				slog.Warn("GitHubDiscussions: reply failed", "title", d.Title, "error", err)
				continue
			}

			slog.Info(fmt.Sprintf("GitHubDiscussions: Replied to \"%s\"", d.Title))
			time.Sleep(30 * time.Second) // rate limit
		}
	}
}

type discussionResult struct {
	Title            string
	URL              string
	Body             string
	DiscussionNodeID string
	RepoOwner        string
	RepoName         string
}

func (g *GitHubDiscussionsEngager) searchDiscussions(ctx context.Context, query string) ([]discussionResult, error) {
	// GitHub Discussions search via GraphQL
	gql := fmt.Sprintf(`{
		search(query: "%s", type: DISCUSSION, first: 5) {
			edges {
				node {
					... on Discussion {
						id
						title
						url
						body
						repository { nameWithOwner }
					}
				}
			}
		}
	}`, query)

	payload, _ := json.Marshal(map[string]string{"query": gql})
	req, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL, bytes.NewReader(payload))
	req.Header.Set("Authorization", "bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discussions search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discussions search: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	// Proper JSON decoding of the GraphQL envelope (replaces the old
	// regex/skeleton stub that discarded every response and returned 0).
	var envelope struct {
		Data struct {
			Search struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						Title      string `json:"title"`
						URL        string `json:"url"`
						Body       string `json:"body"`
						Repository struct {
							NameWithOwner string `json:"nameWithOwner"`
						} `json:"repository"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"search"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("discussions search: decode response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("discussions search: GraphQL error: %s", envelope.Errors[0].Message)
	}

	var results []discussionResult
	for _, e := range envelope.Data.Search.Edges {
		n := e.Node
		if n.ID == "" || n.Title == "" {
			continue // non-Discussion or empty node
		}
		r := discussionResult{Title: n.Title, URL: n.URL, Body: n.Body, DiscussionNodeID: n.ID}
		if owner, name, ok := strings.Cut(n.Repository.NameWithOwner, "/"); ok {
			r.RepoOwner, r.RepoName = owner, name
		}
		results = append(results, r)
	}
	return results, nil
}

func (g *GitHubDiscussionsEngager) buildComment(d discussionResult) string {
	return fmt.Sprintf(`Hey! I noticed this discussion about %s — this is exactly the kind of problem **TormentNexus** solves.

TormentNexus is a local-first cognitive control plane (open source: github.com/HyperNexusSoft/HyperNexus) that handles:
- **Progressive MCP tool routing** — semantic vector search ranks and injects only the most relevant tools
- **Cross-harness parity** — identical tool signatures across Claude Code, Cursor, Codex, Gemini CLI, Copilot, Windsurf
- **L1/L2 memory** — 14K+ persisted memories with sqlite-vec semantic search
- **LLM waterfall** — transparent cascade through providers on 429/5xx errors

Happy to answer any questions!`, d.Title)
}

func (g *GitHubDiscussionsEngager) postReply(ctx context.Context, discussionNodeID, comment string) error {
	// Real addDiscussionComment mutation (was a no-op stub that reported
	// success without posting anything).
	mutation := `mutation($discussionId: ID!, $body: String!) {
		addDiscussionComment(input: {discussionId: $discussionId, body: $body}) {
			comment { id }
		}
	}`
	payload, _ := json.Marshal(map[string]any{
		"query": mutation,
		"variables": map[string]string{
			"discussionId": discussionNodeID,
			"body":         comment,
		},
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL, bytes.NewReader(payload))
	req.Header.Set("Authorization", "bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("discussions reply: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discussions reply: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return fmt.Errorf("discussions reply: decode response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("discussions reply: GraphQL error: %s", envelope.Errors[0].Message)
	}
	return nil
}
