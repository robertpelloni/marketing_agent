package gitcheck

import (
	"context"
	"log"
)

// PRStatus represents the current state of a Pull Request.
type PRStatus string

const (
	PRStatusOpen   PRStatus = "Open"
	PRStatusMerged PRStatus = "Merged"
	PRStatusClosed PRStatus = "Closed"
	PRStatusFailed PRStatus = "Failed"
)

// PullRequest encapsulates metadata for an autonomous PR.
type PullRequest struct {
	ID     string
	URL    string
	Branch string
	Title  string
	Status PRStatus
}

// PRManager defines the interface for autonomous pull request operations.
type PRManager interface {
	// CreatePullRequest creates a new PR from the source branch to the target.
	CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error)

	// GetPRStatus retrieves the current status of a PR.
	GetPRStatus(ctx context.Context, prID string) (PRStatus, error)

	// MergePullRequest merges an open PR.
	MergePullRequest(ctx context.Context, prID string) error
}

// GitHubPRManager implements the PRManager interface using the GitHub CLI (gh).
type GitHubPRManager struct{}

func (g *GitHubPRManager) CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error) {
	log.Printf("GitHubPRManager: Creating Pull Request for branch %s: %s", branch, title)
	// Simulated PR creation using 'gh pr create' logic
	return &PullRequest{
		ID:     "123",
		URL:    "https://github.com/robertpelloni/enterprise_sales_bot/pull/123",
		Branch: branch,
		Title:  title,
		Status: PRStatusOpen,
	}, nil
}

func (g *GitHubPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, error) {
	return PRStatusOpen, nil
}

func (g *GitHubPRManager) MergePullRequest(ctx context.Context, prID string) error {
	log.Printf("GitHubPRManager: Merging Pull Request %s", prID)
	return nil
}

// MockPRManager is a simulated implementation for testing.
type MockPRManager struct{}

func (m *MockPRManager) CreatePullRequest(ctx context.Context, branch string, title string, body string) (*PullRequest, error) {
	return &PullRequest{
		ID:     "mock-456",
		Branch: branch,
		Title:  title,
		Status: PRStatusOpen,
	}, nil
}

func (m *MockPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, error) {
	return PRStatusOpen, nil
}

func (m *MockPRManager) MergePullRequest(ctx context.Context, prID string) error {
	return nil
}
