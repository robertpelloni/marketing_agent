package scraper

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/robertpelloni/marketing_agent/internal/db"
)

// RedditSource implements LeadSource by searching Reddit for companies
// discussing MCP, LLM orchestration, and AI infrastructure.
type RedditSource struct {
	Client     *http.Client
	Username   string
	Password   string
	Subreddits []string
}

// NewRedditSource creates a new RedditSource.
func NewRedditSource() *RedditSource {
	return &RedditSource{
		Client: http.DefaultClient,
		Subreddits: []string{
			"LocalLLaMA",
			"MachineLearning",
			"selfhosted",
			"artificial",
		},
	}
}

// Discover implements LeadSource by searching Reddit for relevant discussions.
func (r *RedditSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	if r.Username == "" {
		r.Username = os.Getenv("REDDIT_USERNAME")
	}
	if r.Password == "" {
		r.Password = os.Getenv("REDDIT_PASSWORD")
	}

	if r.Client == nil {
		r.Client = http.DefaultClient
	}

	slog.Info("RedditSource: Searching for AI/LLM-related subreddit discussions...")

	// Reddit's API requires OAuth. For now, this returns placeholder results
	// since we rely on the CDP poster for actual Reddit outreach.
	return []db.Company{}, nil
}
