package communication

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// GitHubCommentSender posts technical comments on GitHub Issues and PRs
// as a "technical hook" outreach for target companies.
type GitHubCommentSender struct {
	client		*github.Client
	username	string	// GitHub username for the bot account
}

func NewGitHubSender() *GitHubCommentSender {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Info("GitHubCommentSender: Warning: GITHUB_TOKEN not set, will use unauthenticated client (rate limited)")
		return &GitHubCommentSender{
			client:		github.NewClient(nil),
			username:	"tormentnexus-bot",
		}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return &GitHubCommentSender{
		client:		github.NewClient(tc),
		username:	os.Getenv("GITHUB_BOT_USERNAME"),
	}
}

// SendComment posts a technical comment on a GitHub Issue or PR.
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

	slog.Info("GitHubCommentSender: Comment posted", "owner", owner, "repo", repo, "issue", issueNumber)
	return nil
}

// SearchRelevantIssues searches a target company's GitHub repos for issues
// related to AI infrastructure, LLM orchestration, or MCP (TormentNexus's niche).
func (g *GitHubCommentSender) SearchRelevantIssues(ctx context.Context, companyDomain string) ([]IssueTarget, error) {
	if g.client == nil {
		return nil, fmt.Errorf("GitHubCommentSender: client not initialized")
	}

	searchTerms := []string{
		"AI infrastructure",
		"LLM orchestration",
		"MCP server",
		"model routing",
		"agent workflow",
	}

	var targets []IssueTarget
	org := extractOrgFromDomain(companyDomain)

	for _, term := range searchTerms {
		query := fmt.Sprintf("%s org:%s is:issue is:open", term, org)
		results, _, err := g.client.Search.Issues(ctx, query, &github.SearchOptions{
			Sort:	"updated",
			Order:	"desc",
			ListOptions: github.ListOptions{
				PerPage: 3,
			},
		})
		if err != nil {
			slog.Debug("GitHubCommentSender: Search error", "query", query, "error", err)
			continue
		}

		for _, issue := range results.Issues {
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

	// Deduplicate
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

type IssueTarget struct {
	Owner       string
	Repo        string
	IssueNumber int
	Title       string
	URL         string
	Relevance   int
}

func CalculateRelevance(term, title, body string) int {
	score := 0
	lowerTitle := strings.ToLower(title)
	lowerBody := strings.ToLower(body)

	highValue := []string{"MCP", "model context protocol", "agent", "orchestrat", "tool routing", "LLM"}
	for _, kw := range highValue {
		if strings.Contains(lowerTitle, kw) { score += 3 }
		if strings.Contains(lowerBody, kw) { score += 1 }
	}

	return score
}

func GenerateTechHookComment(issue IssueTarget) string {
	return fmt.Sprintf(`Hi there! 👋

I noticed this issue about %s — we've been working on similar challenges with our open-source project **TormentNexus**.

TormentNexus is a local-first cognitive control plane that coordinates multi-agent LLM workflows. If this resonates with what you're building, I'd love to hear your thoughts.

— TormentNexus Bot`, issue.Title)
}

func (g *GitHubCommentSender) FindAndComment(ctx context.Context, company db.Company, contact db.Contact) error {
	targets, err := g.SearchRelevantIssues(ctx, company.Domain)
	if err != nil {
		return fmt.Errorf("GitHubCommentSender: search failed: %w", err)
	}

	if len(targets) == 0 {
		slog.Info("GitHubCommentSender: No relevant issues found", "domain", company.Domain)
		return nil
	}

	bestTarget := targets[0]
	for _, t := range targets {
		if t.Relevance > bestTarget.Relevance {
			bestTarget = t
		}
	}

	comment := GenerateTechHookComment(bestTarget)
	return g.SendComment(ctx, bestTarget.Owner, bestTarget.Repo, bestTarget.IssueNumber, comment)
}

func extractOrgFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return domain
}
