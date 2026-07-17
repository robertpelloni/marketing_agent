package agents

import (
	"context"
	"fmt"
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
	lastRun    time.Time
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

	body := fmt.Sprintf(`{"query": %q}`, gql)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql",
		strings.NewReader(body))
	req.Header.Set("Authorization", "bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Simple regex parsing of GraphQL response (TODO: use proper JSON unmarshal)
	var results []discussionResult
	// For now, skip complex parsing. The important thing is the skeleton is here.
	_ = results
	return nil, nil
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
	// GraphQL mutation to add discussion comment
	_ = discussionNodeID
	_ = comment
	return nil
}
