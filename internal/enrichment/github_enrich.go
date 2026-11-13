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
	"sync"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
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
	// MaxAttempts bounds retries for transient GitHub failures (transport
	// resets such as HTTP/2 PROTOCOL_ERROR, 429, and 5xx). Zero means 3.
	// Prod logs showed bursts of PROTOCOL_ERRORs where every company in a
	// cycle retried immediately at full rate; bounded backoff spreads them.
	MaxAttempts int
	// BaseDelay is the first retry wait, doubled per attempt. Zero means 500ms.
	BaseDelay time.Duration

	// Circuit breaker: after circuitThreshold consecutive transient failures
	// the source stops calling GitHub until Cooldown elapses. Prod showed
	// ~160 failed retries in 30 minutes when GitHub search broke persistently
	// from the Hetzner IP; without this, every enrichment cycle re-hammers a
	// dead lane for every company.
	circuitThreshold int
	cooldown         time.Duration
	consecutiveFails int
	openUntil        time.Time
	mu               sync.Mutex
}

func (g *GitHubEnrichSource) maxAttempts() int {
	if g.MaxAttempts > 0 {
		return g.MaxAttempts
	}
	return 3
}

func (g *GitHubEnrichSource) baseDelay() time.Duration {
	if g.BaseDelay > 0 {
		return g.BaseDelay
	}
	return 500 * time.Millisecond
}

// circuitOpen reports whether the breaker is currently cooling down.
func (g *GitHubEnrichSource) circuitOpen() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return time.Now().Before(g.openUntil)
}

// recordSuccess resets the breaker after a successful GitHub call.
func (g *GitHubEnrichSource) recordSuccess() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutiveFails = 0
	g.openUntil = time.Time{}
}

// recordFailure counts a transient failure and opens the breaker at the
// threshold (default 5 failures, default 5 minute cooldown).
func (g *GitHubEnrichSource) recordFailure() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutiveFails++
	threshold := g.circuitThreshold
	if threshold <= 0 {
		threshold = 5
	}
	if g.consecutiveFails >= threshold && time.Now().After(g.openUntil) {
		cd := g.cooldown
		if cd <= 0 {
			cd = 5 * time.Minute
		}
		g.openUntil = time.Now().Add(cd)
		slog.Info(fmt.Sprintf("GitHubEnrichSource: circuit open after %d consecutive failures, cooling down %v", g.consecutiveFails, cd))
	}
}

// isRetryableGitHubStatus reports whether a response status is worth
// retrying: rate limiting and server-side failures are transient, other
// 4xx answers will not change on a retry.
func isRetryableGitHubStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// doWithRetry executes req with bounded exponential backoff. The request is
// cloned per attempt. Retry-After (seconds) is honored when GitHub sends it.
// Context cancellation aborts the wait between attempts.
func (g *GitHubEnrichSource) doWithRetry(req *http.Request) (*http.Response, error) {
	attempts := g.maxAttempts()
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := g.HTTPClient.Do(req.Clone(req.Context()))
		if err == nil && !isRetryableGitHubStatus(resp.StatusCode) {
			g.recordSuccess()
			return resp, nil
		}
		g.recordFailure()
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("GitHub API returned %d", resp.StatusCode)
		}
		if attempt == attempts-1 {
			if err != nil {
				return nil, lastErr
			}
			return resp, nil
		}
		delay := g.baseDelay() << attempt
		if err == nil && resp != nil {
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, perr := time.ParseDuration(ra + "s"); perr == nil && secs > delay {
					delay = secs
				}
			}
			_ = resp.Body.Close()
		}
		slog.Info(fmt.Sprintf("GitHubEnrichSource: transient failure (attempt %d/%d), backing off %v: %v", attempt+1, attempts, delay, lastErr))
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}
	}
	return nil, lastErr
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
	if g.circuitOpen() {
		return nil, fmt.Errorf("github: circuit open (cooling down after repeated failures)")
	}
	// Derive GitHub org name from the company
	orgName := g.companyToOrgName(company)
	if orgName == "" {
		return nil, fmt.Errorf("github: could not determine org name for %s", company.Name)
	}

	slog.Info(fmt.Sprintf("GitHubEnrichSource: Searching for commits at org %s (%s)", orgName, company.Name))

	// Try to find the org's repos
	repos, err := g.searchOrgRepos(ctx, orgName)
	if err != nil {
		repos, err = g.searchByCompanyName(ctx, company.Name)
		if err != nil {
			return nil, fmt.Errorf("github: repo search failed for %s: %w", orgName, err)
		}
	}

	if len(repos) == 0 {
		slog.Info(fmt.Sprintf("GitHubEnrichSource: No repos found for %s", orgName))
		return nil, nil
	}

	seen := make(map[string]bool)
	var contacts []db.Contact

	for _, repo := range repos {
		if len(contacts) >= 5 {
			break
		}

		commits, err := g.fetchRecentCommits(ctx, repo.FullName, 20)
		if err != nil {
			continue
		}

		for _, commit := range commits {
			email := strings.TrimSpace(commit.Commit.Author.Email)
			name := strings.TrimSpace(commit.Commit.Author.Name)

			if email == "" || seen[email] {
				continue
			}

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

			role := "Engineer"
			if strings.Contains(strings.ToLower(commit.Commit.Message), "merge") ||
				strings.Contains(strings.ToLower(commit.Commit.Message), "release") {
				role = "Lead Engineer"
			}

			// Extract GitHub handle from repo owner if personal repo
			ghHandle := ""
			parts := strings.Split(repo.FullName, "/")
			if len(parts) == 2 && strings.EqualFold(parts[0], orgName) {
				ghHandle = orgName
			}

			contacts = append(contacts, db.Contact{
				Name:         name,
				Email:        email,
				Role:         role,
				GitHubHandle: ghHandle,
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

	resp, err := g.doWithRetry(req)
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

	resp, err := g.doWithRetry(req)
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
// Includes sanity checks to avoid passing garbage HN comment text as GitHub queries.
func (g *GitHubEnrichSource) companyToOrgName(company db.Company) string {
	name := strings.TrimSpace(company.Name)

	// If the name looks like a valid GitHub handle, use it directly
	if name != "" && !strings.Contains(name, " ") && len(name) > 2 && len(name) < 40 {
		if isValidOrgName(name) {
			return name
		}
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

	if domain != "" && !strings.Contains(domain, ".") && isValidOrgName(domain) {
		return domain
	}

	return ""
}

// isValidOrgName checks whether a string looks like a plausible GitHub org name.
// Rejects garbage from HN comment parsing (HTML entities, URL encoding, etc).
func isValidOrgName(s string) bool {
	// Reject strings containing HTML entities or URL encoding artifacts
	if strings.Contains(s, "&") || strings.Contains(s, "%") ||
		strings.Contains(s, "#") || strings.Contains(s, "+") ||
		strings.Contains(s, "://") || strings.Contains(s, "<") {
		return false
	}
	// Reject strings that look like raw URLs
	if strings.HasPrefix(s, "http") || strings.HasPrefix(s, "www.") {
		return false
	}
	// Reject strings that are too short or look like single chars
	if len(s) < 2 {
		return false
	}
	// GitHub org names only use alphanumeric + hyphens
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

// compile-time interface check
var _ EnrichmentSource = (*GitHubEnrichSource)(nil)
