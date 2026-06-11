package gitcheck

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v60/github"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "open"
	PRStatusMerged PRStatus = "merged"
	PRStatusClosed PRStatus = "closed"
	PRStatusFailed PRStatus = "failed"
)

type PullRequest struct {
	ID     string
	Branch string
	Title  string
	Status PRStatus
	URL    string
}

type PRManager interface {
	CreatePR(ctx context.Context, branch, title, body string) (*PullRequest, error)
	GetPRStatus(ctx context.Context, prID string) (PRStatus, []string, error)
	MergePR(ctx context.Context, prID string) error
	DeleteRemoteBranch(ctx context.Context, branch string) error
}

type GitHubPRManager struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGitHubPRManager(owner, repo string) *GitHubPRManager {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Warn("GitHubPRManager: GITHUB_TOKEN not set")
	}
	client := github.NewClient(nil).WithAuthToken(token)
	return &GitHubPRManager{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (m *GitHubPRManager) CreatePR(ctx context.Context, branch, title, body string) (*PullRequest, error) {
	slog.Info("GitHubPRManager: Creating Pull Request", "branch", branch, "title", title)

	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branch),
		Base:                github.String("main"),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := m.client.PullRequests.Create(ctx, m.owner, m.repo, newPR)
	if err != nil {
		return nil, err
	}

	return &PullRequest{
		ID:     fmt.Sprintf("%d", pr.GetNumber()),
		Branch: branch,
		Title:  title,
		Status: PRStatusOpen,
		URL:    pr.GetHTMLURL(),
	}, nil
}

func (m *GitHubPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, []string, error) {
	var n int
	fmt.Sscanf(prID, "%d", &n)
	pr, _, err := m.client.PullRequests.Get(ctx, m.owner, m.repo, n)
	if err != nil {
		return PRStatusFailed, nil, err
	}

	status := PRStatusOpen
	if pr.GetMerged() {
		status = PRStatusMerged
	} else if pr.GetState() == "closed" {
		status = PRStatusClosed
	}

	// Fetch comments
	comments, _, _ := m.client.Issues.ListComments(ctx, m.owner, m.repo, n, nil)
	var commentTexts []string
	for _, c := range comments {
		commentTexts = append(commentTexts, c.GetBody())
	}

	return status, commentTexts, nil
}

func (m *GitHubPRManager) MergePR(ctx context.Context, prID string) error {
	var n int
	fmt.Sscanf(prID, "%d", &n)

	if os.Getenv("DRY_RUN") == "true" {
		slog.Info("GitHubPRManager: Simulating PR merge (DRY_RUN)", "pr_id", prID)
		return nil
	}

	slog.Info("GitHubPRManager: Merging Pull Request", "pr_id", prID)
	_, _, err := m.client.PullRequests.Merge(ctx, m.owner, m.repo, n, "Autonomous merge by Sales Bot", nil)
	return err
}

func (m *GitHubPRManager) DeleteRemoteBranch(ctx context.Context, branch string) error {
	_, err := m.client.Git.DeleteRef(ctx, m.owner, m.repo, "heads/"+branch)
	return err
}

type MockPRManager struct{}

func (m *MockPRManager) CreatePR(ctx context.Context, branch, title, body string) (*PullRequest, error) {
	return &PullRequest{ID: "mock-1", Branch: branch, Title: title, Status: PRStatusOpen, URL: "http://mock/pr/1"}, nil
}
func (m *MockPRManager) GetPRStatus(ctx context.Context, prID string) (PRStatus, []string, error) {
	return PRStatusOpen, nil, nil
}
func (m *MockPRManager) MergePR(ctx context.Context, prID string) error { return nil }
func (m *MockPRManager) DeleteRemoteBranch(ctx context.Context, branch string) error { return nil }
