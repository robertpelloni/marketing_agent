package enrichment

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	// Circuit breaker: after circuitThreshold failed calls inside the
	// cooldown window the source stops calling GitHub until the window
	// slides past. Prod showed ~160 failed retries in 30 minutes when
	// GitHub search broke persistently from the Hetzner IP. A sliding
	// window (not a consecutive counter) is required: successful empty
	// org searches interleave the failures and would reset a consecutive
	// counter forever, which is exactly what prod logs showed.
	circuitThreshold int
	cooldown         time.Duration
	failureTimes     []time.Time
	openUntil        time.Time
	tokenInvalid     atomic.Bool
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

// recordSuccess notes a successful GitHub call. It intentionally does not
// clear the failure window: interleaved empty successes are part of the
// failure pattern this breaker exists to catch.
func (g *GitHubEnrichSource) recordSuccess() {}

// recordFailure records a transient failure in the sliding window and opens
// the breaker at the threshold (default 10 failures, default 5 minute
// window/cooldown).
func (g *GitHubEnrichSource) recordFailure() {
	g.mu.Lock()
	defer g.mu.Unlock()
	cd := g.cooldown
	if cd <= 0 {
		cd = 5 * time.Minute
	}
	now := time.Now()
	cutoff := now.Add(-cd)
	kept := g.failureTimes[:0]
	for _, t := range g.failureTimes {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	g.failureTimes = append(kept, now)
	threshold := g.circuitThreshold
	if threshold <= 0 {
		threshold = 10
	}
	if len(g.failureTimes) >= threshold && now.After(g.openUntil) {
		g.openUntil = now.Add(cd)
		slog.Info(fmt.Sprintf("GitHubEnrichSource: circuit open after %d failures in %v, cooling down %v", len(g.failureTimes), cd, cd))
	}
}

// tripRateLimitCircuit immediately opens the circuit breaker for the given
// cooldown window when GitHub reports exhausted rate limits (HTTP 403/429 with
// X-RateLimit-Remaining: 0 or secondary rate limit message).
func (g *GitHubEnrichSource) tripRateLimitCircuit(wait time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if wait <= 0 {
		if g.cooldown > 0 {
			wait = g.cooldown
		} else {
			wait = 15 * time.Minute
		}
	}
	until := time.Now().Add(wait)
	if until.After(g.openUntil) {
		g.openUntil = until
		slog.Warn(fmt.Sprintf("GitHubEnrichSource: rate limit exhausted; circuit open for %v", wait.Round(time.Second)))
	}
}

// rateLimitCooldown inspects GitHub rate-limit headers (Retry-After and
// X-RateLimit-Reset) to compute a bounded cooldown duration.
func (g *GitHubEnrichSource) rateLimitCooldown(h http.Header) time.Duration {
	if ra := strings.TrimSpace(h.Get("Retry-After")); ra != "" {
		if secs, err := strconv.ParseInt(ra, 10, 64); err == nil && secs > 0 {
			d := time.Duration(secs) * time.Second
			if d > time.Hour {
				return time.Hour
			}
			return d
		}
	}
	if resetStr := strings.TrimSpace(h.Get("X-RateLimit-Reset")); resetStr != "" {
		if epoch, err := strconv.ParseInt(resetStr, 10, 64); err == nil && epoch > 0 {
			wait := time.Until(time.Unix(epoch, 0)) + 5*time.Second
			if wait < time.Minute {
				wait = time.Minute
			}
			if wait > time.Hour {
				wait = time.Hour
			}
			return wait
		}
	}
	if g.cooldown > 0 {
		return g.cooldown
	}
	return 15 * time.Minute
}

// isRetryableGitHubStatus reports whether a response status is worth
// retrying: rate limiting and server-side failures are transient, other
// 4xx answers will not change on a retry.
func isRetryableGitHubStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// applyAuth attaches the configured GitHub token unless an earlier 401 proved
// the token was revoked/expired.
func (g *GitHubEnrichSource) applyAuth(req *http.Request) {
	if g.Token != "" && !g.tokenInvalid.Load() {
		req.Header.Set("Authorization", "token "+g.Token)
	}
}

// doWithRetry executes req with bounded exponential backoff. The request is
// cloned per attempt. Retry-After (seconds) is honored when GitHub sends it.
// If an authenticated request receives 401 Unauthorized (revoked/expired PAT),
// the token is disabled in-memory and the request is immediately retried
// unauthenticated against GitHub's public API.
func (g *GitHubEnrichSource) doWithRetry(req *http.Request) (*http.Response, error) {
	attempts := g.maxAttempts()
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		cloned := req.Clone(req.Context())
		if g.tokenInvalid.Load() {
			cloned.Header.Del("Authorization")
		}
		resp, err := g.HTTPClient.Do(cloned)
		if err == nil && resp.StatusCode == http.StatusUnauthorized && cloned.Header.Get("Authorization") != "" {
			_ = resp.Body.Close()
			if g.tokenInvalid.CompareAndSwap(false, true) {
				slog.Warn("GitHubEnrichSource: GITHUB_TOKEN returned 401 Unauthorized; falling back to unauthenticated public GitHub API")
			}
			unauthReq := req.Clone(req.Context())
			unauthReq.Header.Del("Authorization")
			resp, err = g.HTTPClient.Do(unauthReq)
		}
		if err == nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) {
			remainingZero := strings.TrimSpace(resp.Header.Get("X-RateLimit-Remaining")) == "0"
			rateLimited := remainingZero
			if !rateLimited && resp.StatusCode == http.StatusForbidden {
				snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				lowerSnippet := strings.ToLower(string(snippet))
				if strings.Contains(lowerSnippet, "rate limit") || strings.Contains(lowerSnippet, "secondary rate") {
					rateLimited = true
				} else {
					_ = resp.Body.Close()
					resp.Body = io.NopCloser(strings.NewReader(string(snippet)))
				}
			}
			if rateLimited {
				wait := g.rateLimitCooldown(resp.Header)
				_ = resp.Body.Close()
				g.tripRateLimitCircuit(wait)
				return nil, fmt.Errorf("github: rate limit exceeded (HTTP %d, cooldown %v)", resp.StatusCode, wait.Round(time.Second))
			}
		}
		if err == nil && !isRetryableGitHubStatus(resp.StatusCode) {
			g.recordSuccess()
			if resp.StatusCode == http.StatusOK && strings.TrimSpace(resp.Header.Get("X-RateLimit-Remaining")) == "0" {
				g.tripRateLimitCircuit(g.rateLimitCooldown(resp.Header))
			}
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
			// Prod diagnosis (2026-11-13): authenticated GitHub search works
			// from the Hetzner host over HTTP/1.1 and unauthenticated HTTP/2,
			// but the Go client's HTTP/2 search calls reset persistently
			// (PROTOCOL_ERROR) under load. Force HTTP/1.1 for this client;
			// the retry/backoff/circuit breaker above remains as the safety
			// net if GitHub throttles anyway. Note: ForceAttemptHTTP2=false
			// alone does NOT disable HTTP/2 on a default Transport — an
			// empty, non-nil TLSNextProto map is the documented kill switch.
			Transport: &http.Transport{
				ForceAttemptHTTP2: false,
				TLSNextProto:      map[string]func(string, *tls.Conn) http.RoundTripper{},
			},
		},
	}
}

// Enrich implements EnrichmentSource by searching GitHub for the company's
// repos and extracting committer emails from recent commits.
func (g *GitHubEnrichSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if g.circuitOpen() {
		return nil, fmt.Errorf("github: circuit open (cooling down after repeated failures)")
	}
	// Fast path: if the company domain already names a specific GitHub repo
	// (e.g. "github.com/owner/repo" from TargetDiscoveryWorker), inspect its
	// commits directly first without spending Search API quota.
	var repos []struct{ FullName string }
	if directRepo := directRepoFullName(company.Domain); directRepo != "" {
		repos = append(repos, struct{ FullName string }{FullName: directRepo})
	}

	orgName := g.companyToOrgName(company)
	if len(repos) == 0 {
		if orgName == "" {
			return nil, fmt.Errorf("github: could not determine org name for %s", company.Name)
		}

		slog.Info(fmt.Sprintf("GitHubEnrichSource: Searching for commits at org %s (%s)", orgName, company.Name))

		// Try to find the org's repos
		var err error
		repos, err = g.searchOrgRepos(ctx, orgName)
		if err != nil {
			repos, err = g.searchByCompanyName(ctx, company.Name)
			if err != nil {
				return nil, fmt.Errorf("github: repo search failed for %s: %w", orgName, err)
			}
		}
	}

	if len(repos) == 0 {
		slog.Info(fmt.Sprintf("GitHubEnrichSource: No repos found for %s", orgName))
		return nil, nil
	}

	seen := make(map[string]bool)
	var contacts []db.Contact
	var lastCommitErr error
	hadCleanRepoCheck := false

	for _, repo := range repos {
		if len(contacts) >= 5 {
			break
		}

		commits, err := g.fetchRecentCommits(ctx, repo.FullName, 20)
		if err != nil {
			lastCommitErr = err
			if g.circuitOpen() {
				break
			}
			continue
		}
		hadCleanRepoCheck = true

		for _, commit := range commits {
			email := strings.TrimSpace(commit.Commit.Author.Email)
			name := strings.TrimSpace(commit.Commit.Author.Name)
			lowerEmail := strings.ToLower(email)

			if lowerEmail == "" || seen[lowerEmail] {
				continue
			}

			if db.IsAutomatedCommitAuthor(name, email) {
				continue
			}

			seen[lowerEmail] = true
			if name == "" {
				if at := strings.IndexByte(email, '@'); at > 0 {
					name = email[:at]
				} else {
					name = email
				}
			}

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
		if !hadCleanRepoCheck && lastCommitErr != nil {
			return nil, lastCommitErr
		}
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
	g.applyAuth(req)

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
	g.applyAuth(req)

	resp, err := g.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusConflict {
		// 404 (deleted/private repo) or 409 ("Git Repository is empty.") is a
		// permanent empty state rather than a transient API failure.
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for %s commits", resp.StatusCode, fullName)
	}

	var commits []githubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	return commits, nil
}

// directRepoFullName extracts "owner/repo" when domain is formatted as
// "github.com/owner/repo" or "https://github.com/owner/repo".
func directRepoFullName(domain string) string {
	d := strings.TrimSpace(domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "www.")
	if !strings.HasPrefix(strings.ToLower(d), "github.com/") {
		return ""
	}
	rest := strings.Trim(d[len("github.com/"):], "/")
	parts := strings.Split(rest, "/")
	if len(parts) >= 2 && isValidOrgName(parts[0]) && isValidRepoSegment(parts[1]) {
		return parts[0] + "/" + parts[1]
	}
	return ""
}

func isValidRepoSegment(s string) bool {
	if len(s) < 1 || len(s) > 100 {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// companyToOrgName derives a GitHub org name from a company record.
// Includes sanity checks to avoid passing garbage HN comment text as GitHub queries.
func (g *GitHubEnrichSource) companyToOrgName(company db.Company) string {
	// Prefer explicit github.com/<owner>[/<repo>] domain when present, since
	// TargetDiscoveryWorker stores repo name in company.Name and owner/repo in Domain.
	d := strings.TrimSpace(company.Domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "www.")
	if strings.HasPrefix(strings.ToLower(d), "github.com/") {
		rest := strings.Trim(d[len("github.com/"):], "/")
		parts := strings.Split(rest, "/")
		if len(parts) >= 1 && isValidOrgName(parts[0]) {
			return parts[0]
		}
	}

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
