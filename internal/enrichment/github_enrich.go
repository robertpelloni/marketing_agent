package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/marketing_agent/internal/db"
)

// GitHubEnrichSource implements EnrichmentSource by extracting email addresses
// from GitHub commit history. When Hunter and Apollo fail (no contacts found
// or API errors), this fallback extracts committer emails from the org's
// public repositories to enable basic outreach.
//
// GitHub commits always include the author's email address, so this is a
// reliable last-resort enrichment source for any company with public repos.
type GitHubEnrichSource struct {
	Token      string
	HTTPClient *http.Client
}

// githubCommit represents a single commit from the GitHub API.
type githubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// githubSearchRepoResult represents a repo search result.
type githubSearchRepoResult struct {
	Items []struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
		Owner    *struct {
			Login string `json:"login"`
			URL   string `json:"url"`
		} `json:"owner"`
	} `json:"items"`
}

// NewGitHubEnrichSource creates a new GitHub-based enrichment source.
// Falls back to GITHUB_TOKEN env var if token is empty.
func NewGitHubEnrichSource(token string) *GitHubEnrichSource {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		slog.Info("GitHubEnrichSource: No GITHUB_TOKEN set — will only work for public repos without auth")
	}
	return &GitHubEnrichSource{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enrich implements EnrichmentSource by searching GitHub for the company's
// repos and extracting committer emails from recent commits.
func (g *GitHubEnrichSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	// Derive GitHub org name from the company
	orgName := g.companyToOrgName(company)
	if orgName == "" {
		return nil, fmt.Errorf("github: could not determine org name for %s", company.Name)
	}

	slog.Info(fmt.Sprintf("GitHubEnrichSource: Searching for commits at org %s (%s)", orgName, company.Name))

	// Try to find the org's repos
	repos, err := g.searchOrgRepos(ctx, orgName)
	if err != nil {
		// Try searching by company name as a fallback
		repos, err = g.searchByCompanyName(ctx, company.Name)
		if err != nil {
			return nil, fmt.Errorf("github: repo search failed for %s: %w", orgName, err)
		}
	}

	if len(repos) == 0 {
		slog.Info(fmt.Sprintf("GitHubEnrichSource: No repos found for %s", orgName))
		return nil, nil
	}

	// Extract unique committer emails from all repos
	seen := make(map[string]bool)
	var contacts []db.Contact

	for _, repo := range repos {
		if len(contacts) >= 5 {
			break // cap at 5 contacts
		}

		commits, err := g.fetchRecentCommits(ctx, repo.FullName, 20)
		if err != nil {
			slog.Info(fmt.Sprintf("GitHubEnrichSource: Failed to fetch commits for %s: %v", repo.FullName, err))
			continue
		}

		for _, commit := range commits {
			email := strings.TrimSpace(commit.Commit.Author.Email)
			name := strings.TrimSpace(commit.Commit.Author.Name)

			if email == "" || seen[email] {
				continue
			}

			// Skip no-reply, action bots, and system emails to avoid bounces
			lowerEmail := strings.ToLower(email)
			if strings.Contains(lowerEmail, "noreply") ||
				strings.Contains(lowerEmail, "github-actions") ||
				strings.Contains(lowerEmail, "web-flow") ||
				strings.HasPrefix(lowerEmail, "support@") ||
				strings.HasPrefix(lowerEmail, "admin@") ||
				strings.HasPrefix(lowerEmail, "info@") {
				continue
			}

			seen[email] = true

			// Derive role from commit count heuristic
			role := "Engineer"
			if strings.Contains(strings.ToLower(commit.Commit.Message), "merge") ||
				strings.Contains(strings.ToLower(commit.Commit.Message), "release") {
				role = "Lead Engineer"
			}

			contacts = append(contacts, db.Contact{
				Name:  name,
				Email: email,
				Role:  role,
			})

			slog.Info(fmt.Sprintf("GitHubEnrichSource: Found committer %s <%s> in %s", name, email, repo.FullName))
		}
	}

	if len(contacts) == 0 {
		slog.Info(fmt.Sprintf("GitHubEnrichSource: No usable commit emails found for %s", orgName))
		return nil, nil
	}

	return contacts, nil
}

// searchOrgRepos searches for repos belonging to a GitHub org or user.
func (g *GitHubEnrichSource) searchOrgRepos(ctx context.Context, orgName string) ([]struct {
	FullName string
}, error) {
	query := fmt.Sprintf("org:%s", url.QueryEscape(orgName))
	return g.searchRepos(ctx, query)
}

// searchByCompanyName searches for repos by company name as a fallback.
func (g *GitHubEnrichSource) searchByCompanyName(ctx context.Context, companyName string) ([]struct {
	FullName string
}, error) {
	query := fmt.Sprintf("%s in:name", url.QueryEscape(companyName))
	return g.searchRepos(ctx, query)
}

// searchRepos performs a GitHub repo search.
func (g *GitHubEnrichSource) searchRepos(ctx context.Context, query string) ([]struct {
	FullName string
}, error) {
	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=updated&per_page=5", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var result githubSearchRepoResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var repos []struct {
		FullName string
	}
	for _, item := range result.Items {
		repos = append(repos, struct{ FullName string }{FullName: item.FullName})
	}

	return repos, nil
}

// fetchRecentCommits fetches recent commits from a GitHub repo.
func (g *GitHubEnrichSource) fetchRecentCommits(ctx context.Context, fullName string, count int) ([]githubCommit, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/commits?per_page=%d", fullName, count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for %s commits", resp.StatusCode, fullName)
	}

	var commits []githubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	return commits, nil
}

// companyToOrgName derives a GitHub org name from a company record.
func (g *GitHubEnrichSource) companyToOrgName(company db.Company) string {
	// If the company name is a GitHub-style handle, use it directly
	name := strings.TrimSpace(company.Name)
	if name != "" && !strings.Contains(name, " ") && len(name) > 2 {
		return name
	}

	// Try extracting from domain
	domain := strings.TrimSpace(company.Domain)
	domain = strings.TrimSuffix(domain, ".com")
	domain = strings.TrimSuffix(domain, ".io")
	domain = strings.TrimSuffix(domain, ".tech")
	domain = strings.TrimSuffix(domain, ".ai")
	domain = strings.TrimSuffix(domain, ".dev")
	domain = strings.TrimSuffix(domain, ".app")
	domain = strings.TrimSuffix(domain, ".org")
	domain = strings.TrimSuffix(domain, ".net")

	if domain != "" && !strings.Contains(domain, ".") {
		return domain
	}

	return ""
}

// compile-time interface check
var _ EnrichmentSource = (*GitHubEnrichSource)(nil)
