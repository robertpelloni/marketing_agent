package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// GitHubCITracker implements the CITracker interface using the GitHub Actions API.
type GitHubCITracker struct {
	owner string
	repo  string
	token string
	// apiBase is overridable for tests; production uses api.github.com.
	apiBase string
}

// NewGitHubCITracker creates a new tracker instance.
func NewGitHubCITracker(owner, repo string) *GitHubCITracker {
	return &GitHubCITracker{
		owner:   owner,
		repo:    repo,
		token:   os.Getenv("GITHUB_TOKEN"),
		apiBase: "https://api.github.com",
	}
}

// doStatusRequest performs one Actions API request, optionally with auth.
func (g *GitHubCITracker) doStatusRequest(ctx context.Context, url string, withAuth bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if withAuth && g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return http.DefaultClient.Do(req)
}

type workflowRunsResponse struct {
	WorkflowRuns []struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"workflow_runs"`
}

func (g *GitHubCITracker) GetLatestStatus(ctx context.Context, branch string) (CIStatus, error) {
	base := g.apiBase
	if base == "" {
		base = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/actions/runs?branch=%s&per_page=1", base, g.owner, g.repo, branch)

	resp, err := g.doStatusRequest(ctx, url, true)
	if err != nil {
		return CIStatusUnknown, err
	}

	// A revoked/expired token must not blind health reporting for a public
	// repo: workflow runs are public data, so retry once without auth.
	// (Prod lost system_health this way when its GITHUB_TOKEN was revoked
	// while CI itself was green.)
	if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && g.token != "" {
		_ = resp.Body.Close()
		resp, err = g.doStatusRequest(ctx, url, false)
		if err != nil {
			return CIStatusUnknown, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return CIStatusUnknown, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var data workflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return CIStatusUnknown, err
	}

	if len(data.WorkflowRuns) == 0 {
		return CIStatusPending, nil
	}

	run := data.WorkflowRuns[0]
	if run.Status != "completed" {
		return CIStatusPending, nil
	}

	switch run.Conclusion {
	case "success":
		return CIStatusSuccess, nil
	case "failure", "cancelled", "timed_out", "action_required":
		return CIStatusFailure, nil
	case "skipped":
		return CIStatusSuccess, nil // Treat skipped as non-failure for gating
	default:
		return CIStatusUnknown, nil
	}
}

func (g *GitHubCITracker) GetSystemHealth(ctx context.Context) (string, error) {
	// For health, we can check the status of the main branch
	status, err := g.GetLatestStatus(ctx, "main")
	if err != nil {
		return "Health check failed", err
	}
	return fmt.Sprintf("Main branch status: %s", status), nil
}
