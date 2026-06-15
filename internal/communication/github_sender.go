package communication

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type GitHubSender struct {
	client *github.Client
}

func NewGitHubSender() *GitHubSender {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("GitHubSender: GITHUB_TOKEN not set, outreach will be simulated.")
		return &GitHubSender{}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &GitHubSender{
		client: github.NewClient(tc),
	}
}

func (s *GitHubSender) SendComment(ctx context.Context, owner, repo string, issueNumber int, body string) error {
	if s.client == nil {
		log.Printf("GitHub simulation: comment on %s/%s#%d: %s", owner, repo, issueNumber, body)
		return nil
	}

	_, _, err := s.client.Issues.CreateComment(ctx, owner, repo, issueNumber, &github.IssueComment{
		Body: github.String(body),
	})
	if err != nil {
		return fmt.Errorf("github sender: failed to create comment: %w", err)
	}

	log.Printf("GitHub outreach sent to %s/%s#%d", owner, repo, issueNumber)
	return nil
}
