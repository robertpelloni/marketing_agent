package communication

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

<<<<<<< HEAD
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
=======
	"github.com/google/go-github/v60/github"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"golang.org/x/oauth2"
>>>>>>> origin/main
)

// GitHubCommentSender posts technical comments on GitHub Issues and PRs
// as a "technical hook" outreach for target companies.
type GitHubCommentSender struct {
	client		*github.Client
<<<<<<< HEAD
	username	string	// GitHub username for the bot account
}

func NewGitHubSender() *GitHubCommentSender {
=======
	repo		string	// e.g. "robertpelloni/enterprise_sales_bot"
	username	string	// GitHub username for the bot account
}

// NewGitHubCommentSender creates a GitHubCommentSender.
// If GITHUB_TOKEN is set, it creates an authenticated client.
// repo is the bot's repository for generating relevant comments.
func NewGitHubCommentSender(repo string) *GitHubCommentSender {
>>>>>>> origin/main
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Info("GitHubCommentSender: Warning: GITHUB_TOKEN not set, will use unauthenticated client (rate limited)")
		return &GitHubCommentSender{
			client:		github.NewClient(nil),
<<<<<<< HEAD
=======
			repo:		repo,
>>>>>>> origin/main
			username:	"tormentnexus-bot",
		}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return &GitHubCommentSender{
		client:		github.NewClient(tc),
<<<<<<< HEAD
=======
		repo:		repo,
>>>>>>> origin/main
		username:	os.Getenv("GITHUB_BOT_USERNAME"),
	}
}

// SendComment posts a technical comment on a GitHub Issue or PR.
<<<<<<< HEAD
=======
// owner/repo: the target repository
// issueNumber: the issue or PR number to comment on
// commentBody: the message body to post
>>>>>>> origin/main
func (g *GitHubCommentSender) SendComment(ctx context.Context, owner, repo string, issueNumber int, commentBody string) error {
	if g.client == nil {
		return fmt.Errorf("GitHubCommentSender: client not initialized")
	}

	comment := &github.IssueComment{
		Body: github.String(commentBody),
	}

	_, _, err := g.client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)
	if err != nil {
		return fmt.Errorf("GitHubCommentSender: failed to create comment on %s/%s#%d: %w", owner, repo, issueNumber, err)
	}

<<<<<<< HEAD
	slog.Info("GitHubCommentSender: Comment posted", "owner", owner, "repo", repo, "issue", issueNumber)
=======
	slog.Info(fmt.Sprintf("GitHubCommentSender: Comment posted on %s/%s#%d", owner, repo, issueNumber))
>>>>>>> origin/main
	return nil
}

// SearchRelevantIssues searches a target company's GitHub repos for issues
// related to AI infrastructure, LLM orchestration, or MCP (TormentNexus's niche).
<<<<<<< HEAD
=======
// Returns issue URLs and titles.
>>>>>>> origin/main
func (g *GitHubCommentSender) SearchRelevantIssues(ctx context.Context, companyDomain string) ([]IssueTarget, error) {
	if g.client == nil {
		return nil, fmt.Errorf("GitHubCommentSender: client not initialized")
	}

<<<<<<< HEAD
=======
	// Extract owner from domain or use a search approach
	// We search for issues mentioning keywords in the company's GitHub repos
>>>>>>> origin/main
	searchTerms := []string{
		"AI infrastructure",
		"LLM orchestration",
		"MCP server",
		"model routing",
		"agent workflow",
<<<<<<< HEAD
	}

	var targets []IssueTarget
	org := extractOrgFromDomain(companyDomain)

	for _, term := range searchTerms {
		query := fmt.Sprintf("%s org:%s is:issue is:open", term, org)
=======
		"multi-agent",
		"tool orchestration",
		"prompt management",
		"AI observability",
	}

	var targets []IssueTarget

	for _, term := range searchTerms {
		query := fmt.Sprintf("%s org:%s is:issue is:open", term, companyDomain)
>>>>>>> origin/main
		results, _, err := g.client.Search.Issues(ctx, query, &github.SearchOptions{
			Sort:	"updated",
			Order:	"desc",
			ListOptions: github.ListOptions{
				PerPage: 3,
			},
		})
		if err != nil {
<<<<<<< HEAD
			slog.Debug("GitHubCommentSender: Search error", "query", query, "error", err)
=======
			slog.Info(fmt.Sprintf("GitHubCommentSender: Search error for %q: %v", query, err))
>>>>>>> origin/main
			continue
		}

		for _, issue := range results.Issues {
<<<<<<< HEAD
=======
			// Split owner/repo from issue URL
>>>>>>> origin/main
			parts := strings.Split(issue.GetRepositoryURL(), "/")
			repoOwner := ""
			repoName := ""
			if len(parts) >= 2 {
				repoOwner = parts[len(parts)-2]
				repoName = parts[len(parts)-1]
			}

			targets = append(targets, IssueTarget{
				Owner:		repoOwner,
				Repo:		repoName,
				IssueNumber:	issue.GetNumber(),
				Title:		issue.GetTitle(),
				URL:		issue.GetHTMLURL(),
				Relevance:	CalculateRelevance(term, issue.GetTitle(), issue.GetBody()),
			})
		}
	}

<<<<<<< HEAD
	// Deduplicate
=======
	// Deduplicate by URL
>>>>>>> origin/main
	seen := make(map[string]bool)
	var unique []IssueTarget
	for _, t := range targets {
		if !seen[t.URL] {
			seen[t.URL] = true
			unique = append(unique, t)
		}
	}

	return unique, nil
}

<<<<<<< HEAD
type IssueTarget struct {
	Owner       string
	Repo        string
	IssueNumber int
	Title       string
	URL         string
	Relevance   int
}

=======
// IssueTarget represents a GitHub issue or PR that is relevant for outreach.
type IssueTarget struct {
	Owner		string
	Repo		string
	IssueNumber	int
	Title		string
	URL		string
	Relevance	int	// higher = more relevant
}

// CalculateRelevance scores how relevant an issue is to TormentNexus.
>>>>>>> origin/main
func CalculateRelevance(term, title, body string) int {
	score := 0
	lowerTitle := strings.ToLower(title)
	lowerBody := strings.ToLower(body)

<<<<<<< HEAD
	highValue := []string{"MCP", "model context protocol", "agent", "orchestrat", "tool routing", "LLM"}
	for _, kw := range highValue {
		if strings.Contains(lowerTitle, kw) { score += 3 }
		if strings.Contains(lowerBody, kw) { score += 1 }
=======
	// High-value keywords
	highValue := []string{"MCP", "model context protocol", "agent", "orchestrat", "tool routing", "LLM"}
	for _, kw := range highValue {
		if strings.Contains(lowerTitle, kw) {
			score += 3
		}
		if strings.Contains(lowerBody, kw) {
			score += 1
		}
	}

	// Medium-value keywords
	mediumValue := []string{"infrastructure", "pipeline", "workflow", "automation", "integration", "deploy"}
	for _, kw := range mediumValue {
		if strings.Contains(lowerTitle, kw) {
			score += 2
		}
		if strings.Contains(lowerBody, kw) {
			score += 1
		}
	}

	// Bonus for matching search term
	if strings.Contains(strings.ToLower(term), lowerTitle) || strings.Contains(lowerTitle, strings.ToLower(term)) {
		score += 2
>>>>>>> origin/main
	}

	return score
}

<<<<<<< HEAD
func GenerateTechHookComment(issue IssueTarget) string {
	return fmt.Sprintf(`Hi there! 👋

I noticed this issue about %s — we've been working on similar challenges with our open-source project **TormentNexus**.

TormentNexus is a local-first cognitive control plane that coordinates multi-agent LLM workflows. If this resonates with what you're building, I'd love to hear your thoughts.

— TormentNexus Bot`, issue.Title)
}

func (g *GitHubCommentSender) FindAndComment(ctx context.Context, company db.Company, contact db.Contact) error {
	targets, err := g.SearchRelevantIssues(ctx, company.Domain)
=======
// GenerateTechHookComment generates a helpful, value-first comment related to
// the issue that positions TormentNexus as a solution.
func GenerateTechHookComment(issue IssueTarget) string {
	comment := fmt.Sprintf(`Hi there! 👋

I noticed this issue about %s — we've been working on similar challenges with our open-source project **TormentNexus**.

TormentNexus is a local-first cognitive control plane that coordinates multi-agent LLM workflows with:
- **Progressive MCP Tool Routing** — semantic router that injects only the 3 most relevant tools per request (no 50K-token tool dumps)
- **Cross-Harness Tool Parity** — byte-for-byte identical tool signatures for Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf
- **LLM Waterfall** — seamless cascading through cloud providers → aggregators → local models on 429/5xx errors
- **Local-First Memory** — 14K+ persisted memories with sqlite-vec semantic search, surviving restarts

If this resonates with what you're building, I'd love to hear your thoughts. We're always looking for feedback from teams pushing the boundaries of AI infrastructure.

— TormentNexus Bot`, issue.Title)

	return comment
}

// FindAndComment is a high-level operation that searches for relevant issues
// in a target organization's GitHub repos and posts a technical hook comment.
func (g *GitHubCommentSender) FindAndComment(ctx context.Context, company db.Company, contact db.Contact) error {
	// Search for relevant issues
	targets, err := g.SearchRelevantIssues(ctx, extractOrgFromDomain(company.Domain))
>>>>>>> origin/main
	if err != nil {
		return fmt.Errorf("GitHubCommentSender: search failed: %w", err)
	}

	if len(targets) == 0 {
<<<<<<< HEAD
		slog.Info("GitHubCommentSender: No relevant issues found", "domain", company.Domain)
		return nil
	}

	bestTarget := targets[0]
	for _, t := range targets {
=======
		slog.Info(fmt.Sprintf("GitHubCommentSender: No relevant issues found for %s", company.Domain))
		return nil
	}

	// Pick the most relevant issue
	bestTarget := targets[0]
	for _, t := range targets[1:] {
>>>>>>> origin/main
		if t.Relevance > bestTarget.Relevance {
			bestTarget = t
		}
	}

<<<<<<< HEAD
	comment := GenerateTechHookComment(bestTarget)
	return g.SendComment(ctx, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber, comment)
}

func extractOrgFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 1 {
=======
	// Generate and post the comment
	comment := GenerateTechHookComment(bestTarget)
	if err := g.SendComment(ctx, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber, comment); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("GitHubCommentSender: Technical hook posted for %s on %s/%s#%d",
		company.Name, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber))

	return nil
}

// extractOrgFromDomain extracts the likely GitHub org name from a company domain.
func extractOrgFromDomain(domain string) string {
	// Handle URLs
	if strings.HasPrefix(domain, "http") {
		parsed, err := url.Parse(domain)
		if err == nil {
			domain = parsed.Host
		}
	}

	// Remove TLD and subdomain
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
>>>>>>> origin/main
		return parts[0]
	}
	return domain
}
