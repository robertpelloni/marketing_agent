package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// GitHubIssueSource implements LeadSource by searching GitHub issues for companies with relevant technical challenges.
type GitHubIssueSource struct {
	Client   *http.Client
	Token    string
	Keywords []string
	BaseURL  string // Optional API base URL override for testing (defaults to https://api.github.com)
}

func (g *GitHubIssueSource) apiBaseURL() string {
	if g.BaseURL != "" {
		return strings.TrimRight(g.BaseURL, "/")
	}
	return "https://api.github.com"
}

// githubIssueSearchResponse represents the GitHub Search API response for issues.
type githubIssueSearchResponse struct {
	TotalCount int `json:"total_count"`
	Issues     []struct {
		HTMLURL   string `json:"html_url"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		RepoURL   string `json:"repository_url"`
		CreatedAt string `json:"created_at"`
	} `json:"items"`
}

// githubRepo represents a GitHub repository.
type githubRepo struct {
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// Discover searches GitHub for open issues matching TormentNexus-relevant keywords,
// extracts organization information, and returns them as potential leads.
func (g *GitHubIssueSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	if g.Token == "" {
		g.Token = os.Getenv("GITHUB_TOKEN")
	}

	if g.Token == "" {
		slog.Info("GitHubIssueSource: No GITHUB_TOKEN set, returning empty results")
		return []db.Company{}, nil
	}

	if g.Client == nil {
		g.Client = http.DefaultClient
	}

	if len(g.Keywords) == 0 {
		g.Keywords = []string{
			"MCP",
			"model context protocol",
			"tool routing",
			"multi-agent",
			"agent orchestration",
			"LLM orchestration",
			"LLM management",
			"agent workflow",
			"tool calling",
			"function calling",
		}
	}

	slog.Info(fmt.Sprintf("GitHubIssueSource: Searching for issues with keywords: %v", g.Keywords))

	companiesMap := make(map[string]*db.Company)

	// Search for each keyword
	for _, keyword := range g.Keywords {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		issues, err := g.searchIssues(ctx, keyword)
		if err != nil {
			slog.Info(fmt.Sprintf("GitHubIssueSource: Error searching for '%s': %v", keyword, err))
			if strings.Contains(err.Error(), "status 429") || strings.Contains(err.Error(), "status 403") {
				slog.Info("GitHubIssueSource: GitHub Search API rate limit reached; stopping issue keyword loop early")
				break
			}
			continue
		}

		// Process each issue to extract company information
		for _, issue := range issues {
			orgName, repoName, err := g.extractOwnerRepoFromURL(issue.RepoURL)
			if err != nil {
				continue
			}

			// Skip well-known large orgs that aren't target customers
			if g.isExcludedOrg(orgName) {
				continue
			}

			domain := g.resolveIssueDomain(ctx, orgName, repoName)
			if domain == "" {
				continue
			}

			// Check if we already have this company
			if existing, ok := companiesMap[domain]; ok {
				// Append this issue as an additional hiring signal
				signal := fmt.Sprintf("GitHub Issue: %s - %s", issue.Title, issue.HTMLURL)
				existing.HiringSignals = append(existing.HiringSignals, signal)
			} else {
				// Fetch repo details via Core API (/repos/{owner}/{repo}) rather than
				// burning Search API quota (/search/repositories).
				repoInfo, err := g.getRepoDetails(ctx, orgName, repoName)
				if err != nil {
					repoInfo = &githubRepo{
						FullName: orgName + "/" + repoName,
					}
					repoInfo.Owner.Login = orgName
				}

				var techStack []string
				if repoInfo.Language != "" {
					techStack = []string{repoInfo.Language}
				}
				companyName := repoInfo.Owner.Login
				if companyName == "" {
					companyName = orgName
				}

				company := &db.Company{
					Name:          companyName,
					Domain:        domain,
					TechStack:     techStack,
					HiringSignals: []string{fmt.Sprintf("GitHub Issue: %s - %s", issue.Title, issue.HTMLURL)},
					MarketCapTier: g.inferMarketCap(repoInfo),
				}
				companiesMap[domain] = company
			}
		}
	}

	// Convert map to slice
	var companies []db.Company
	for _, company := range companiesMap {
		companies = append(companies, *company)
	}

	slog.Info(fmt.Sprintf("GitHubIssueSource: Discovered %d companies from GitHub issues", len(companies)))
	return companies, nil
}

// searchIssues queries the GitHub Search API for issues matching the given keyword.
func (g *GitHubIssueSource) searchIssues(ctx context.Context, keyword string) ([]struct {
	HTMLURL   string `json:"html_url"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	RepoURL   string `json:"repository_url"`
	CreatedAt string `json:"created_at"`
}, error) {
	// Build search query: keyword in title/body, open issues only, exclude very large orgs
	query := fmt.Sprintf("%s type:issue state:open", keyword)
	queryURL := fmt.Sprintf("%s/search/issues?q=%s&per_page=10", g.apiBaseURL(), url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && g.Token != "" {
		_ = resp.Body.Close()
		slog.Warn("GitHubIssueSource: GITHUB_TOKEN returned 401; clearing token and retrying unauthenticated")
		g.Token = ""
		req2, rerr := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		if rerr != nil {
			return nil, rerr
		}
		req2.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err = g.Client.Do(req2)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result githubIssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Issues, nil
}

// extractOrgFromRepoURL extracts the organization name from a GitHub repository URL.
// Example: "https://api.github.com/repos/acme-corp/my-repo" -> "acme-corp"
func (g *GitHubIssueSource) extractOrgFromRepoURL(repoURL string) (string, error) {
	owner, _, err := g.extractOwnerRepoFromURL(repoURL)
	return owner, err
}

// extractOwnerRepoFromURL extracts both owner and repository name from a GitHub API repo URL.
func (g *GitHubIssueSource) extractOwnerRepoFromURL(repoURL string) (string, string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(repoURL), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo URL: %s", repoURL)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	if owner == "" || repo == "" || owner == "repos" {
		return "", "", fmt.Errorf("invalid repo URL: %s", repoURL)
	}
	return owner, repo, nil
}

// getRepoDetails fetches repository metadata directly from the GitHub Core API
// (/repos/{owner}/{repo}) instead of burning Search API quota. Falls back to
// getOrgInfo when a test server only mocks /search/repositories.
func (g *GitHubIssueSource) getRepoDetails(ctx context.Context, owner, repo string) (*githubRepo, error) {
	repoURL := fmt.Sprintf("%s/repos/%s/%s", g.apiBaseURL(), url.PathEscape(owner), url.PathEscape(repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoURL, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" && g.Token != "GITHUB_TOKEN_PLACEHOLDER" {
		req.Header.Set("Authorization", "token "+g.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && g.Token != "" {
		_ = resp.Body.Close()
		slog.Warn("GitHubIssueSource: GITHUB_TOKEN returned 401 on repo details; clearing token and retrying unauthenticated")
		g.Token = ""
		req2, rerr := http.NewRequestWithContext(ctx, http.MethodGet, repoURL, nil)
		if rerr != nil {
			return nil, rerr
		}
		req2.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err = g.Client.Do(req2)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return g.getOrgInfo(ctx, owner)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for repo %s/%s", resp.StatusCode, owner, repo)
	}

	var details githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

// resolveIssueDomain prefers a verified website listed on the GitHub user/org
// profile and otherwise falls back to the canonical "github.com/owner/repo"
// identifier so downstream enrichment uses direct commit history without
// fabricating nonexistent .com/.io domains.
func (g *GitHubIssueSource) resolveIssueDomain(ctx context.Context, owner, repo string) string {
	if website := g.fetchOwnerWebsite(ctx, owner); website != "" {
		return website
	}
	if owner != "" && repo != "" {
		return fmt.Sprintf("github.com/%s/%s", owner, repo)
	}
	return ""
}

func (g *GitHubIssueSource) fetchOwnerWebsite(ctx context.Context, owner string) string {
	if g.Token == "GITHUB_TOKEN_PLACEHOLDER" && g.BaseURL == "" {
		return ""
	}
	apiURL := fmt.Sprintf("%s/users/%s", g.apiBaseURL(), url.PathEscape(owner))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" && g.Token != "GITHUB_TOKEN_PLACEHOLDER" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var profile struct {
		Blog string `json:"blog"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil || profile.Blog == "" {
		return ""
	}

	domain := normalizeProfileWebsiteDomain(profile.Blog)
	if domain != "" {
		slog.Info(fmt.Sprintf("GitHubIssueSource: Found verified profile website %s for %s", domain, owner))
	}
	return domain
}

func normalizeProfileWebsiteDomain(raw string) string {
	blog := strings.ToLower(strings.TrimSpace(raw))
	blog = strings.TrimPrefix(blog, "https://")
	blog = strings.TrimPrefix(blog, "http://")
	blog = strings.TrimPrefix(blog, "www.")
	blog = strings.Split(blog, "/")[0]
	blog = strings.Split(blog, "?")[0]
	blog = strings.Trim(blog, ".")
	if blog == "" || !strings.Contains(blog, ".") {
		return ""
	}
	// Social profile links and GitHub Pages are not corporate email domains.
	for _, nonCorp := range []string{"github.com", "github.io", "twitter.com", "x.com", "linkedin.com", "medium.com", "substack.com", "youtube.com"} {
		if blog == nonCorp || strings.HasSuffix(blog, "."+nonCorp) {
			return ""
		}
	}
	return blog
}

// getOrgInfo fetches repository information to extract organization details.
func (g *GitHubIssueSource) getOrgInfo(ctx context.Context, orgName string) (*githubRepo, error) {
	// Search for recent repos by this org to get org metadata
	query := fmt.Sprintf("user:%s", orgName)
	queryURL := fmt.Sprintf("%s/search/repositories?q=%s&sort=updated&per_page=1", g.apiBaseURL(), url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && g.Token != "" {
		_ = resp.Body.Close()
		slog.Warn("GitHubIssueSource: GITHUB_TOKEN returned 401; clearing token and retrying unauthenticated")
		g.Token = ""
		req2, rerr := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		if rerr != nil {
			return nil, rerr
		}
		req2.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err = g.Client.Do(req2)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for org %s", resp.StatusCode, orgName)
	}

	var result struct {
		Items []githubRepo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no repos found for org %s", orgName)
	}

	return &result.Items[0], nil
}

// isExcludedOrg checks if an organization should be excluded from targeting.
func (g *GitHubIssueSource) isExcludedOrg(orgName string) bool {
	// Exclude very large tech companies and GitHub itself
	excluded := []string{
		"microsoft", "google", "amazon", "facebook", "meta", "apple",
		"netflix", "uber", "airbnb", "twitter", "linkedin", "github",
		"gitlab", "atlassian", "jetbrains", "docker", "kubernetes",
		"canonical", "redhat", "ibm", "oracle", "salesforce", "slack",
		"stripe", "twilio", "mongodb", "elastic", "confluent",
		"anthropic", "anthropics", "openai", "modelcontextprotocol",
		"cloudflare", "vercel", "huggingface",
	}

	for _, ex := range excluded {
		if strings.EqualFold(orgName, ex) {
			return true
		}
	}
	return false
}

// orgToDomain maps a GitHub organization name to a realistic domain.
// Uses the GitHub API to fetch the org's listed website/blog URL first.
// Falls back to heuristics: AI/ML → .io, everything else → .com
func (g *GitHubIssueSource) orgToDomain(orgName string) string {
	name := strings.ToLower(orgName)

	// Remove common suffixes
	name = strings.TrimSuffix(name, "-inc")
	name = strings.TrimSuffix(name, "-corp")
	name = strings.TrimSuffix(name, "-hq")
	name = strings.TrimSuffix(name, "-dev")
	name = strings.TrimSuffix(name, "-labs")
	name = strings.TrimSuffix(name, "-tech")

	// Fetch org's actual website from GitHub API
	if domain := g.fetchOrgWebsite(orgName); domain != "" {
		return domain
	}

	// Heuristic: AI/ML companies → .io, everything else → .com
	if strings.Contains(name, "ai") || strings.Contains(name, "ml") ||
		strings.Contains(name, "neural") || strings.Contains(name, "logic") ||
		strings.Contains(name, "deep") || strings.Contains(name, "cortex") ||
		strings.Contains(name, "bot") || strings.Contains(name, "agent") ||
		strings.Contains(name, "intelli") || strings.Contains(name, "data") {
		return name + ".io"
	}

	return name + ".com"
}

// fetchOrgWebsite attempts to retrieve the organization's listed website
// from the GitHub API. Returns empty string on failure.
func (g *GitHubIssueSource) fetchOrgWebsite(orgName string) string {
	if g.Token == "GITHUB_TOKEN_PLACEHOLDER" || g.Token == "" {
		return "" // Skip real network lookup for test fixtures
	}
	apiURL := fmt.Sprintf("https://api.github.com/orgs/%s", url.PathEscape(orgName))

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	resp, err := g.Client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var orgData struct {
		Blog string `json:"blog"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgData); err != nil {
		return ""
	}

	if orgData.Blog == "" {
		return ""
	}

	// Extract domain from the blog URL
	blog := strings.TrimSpace(orgData.Blog)
	blog = strings.TrimPrefix(blog, "https://")
	blog = strings.TrimPrefix(blog, "http://")
	blog = strings.TrimPrefix(blog, "www.")
	blog = strings.Split(blog, "/")[0]
	blog = strings.Split(blog, "?")[0]

	if blog != "" && strings.Contains(blog, ".") {
		slog.Info(fmt.Sprintf("GitHubIssueSource: Found org website %s for %s", blog, orgName))
		return blog
	}

	return ""
}

// inferMarketCap infers a company's market cap tier from repository metadata.
func (g *GitHubIssueSource) inferMarketCap(repo *githubRepo) string {
	// Heuristics based on repo metadata
	if repo.Language == "" {
		return "Small Business"
	}

	// Star count and activity heuristics
	// (In a real implementation, we'd fetch star count from the repo)
	if strings.Contains(repo.Description, "enterprise") ||
		strings.Contains(repo.Description, "platform") {
		return "Mid-Market"
	}

	return "Small Business"
}
