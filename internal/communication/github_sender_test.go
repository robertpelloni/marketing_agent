package communication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v60/github"
)

func TestGitHubCommentSender_SendComment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		_ , _ = w.Write([]byte(`{"id": 123, "body": "test comment"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := github.NewClient(nil)
	u, _ := url.Parse(server.URL + "/")
	client.BaseURL = u

	sender := &GitHubCommentSender{
		client: client,
		repo:   "test/repo",
	}

	err := sender.SendComment(context.Background(), "owner", "repo", 1, "test comment")
	if err != nil {
		t.Fatalf("SendComment failed: %v", err)
	}
}

func TestCalculateRelevance(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		title    string
		body     string
		minScore int
	}{
		{
			name:     "High relevance MCP",
			term:     "MCP server",
			title:    "New MCP server for PostgreSQL",
			body:     "I am building a Model Context Protocol server",
			minScore: 3,
		},
		{
			name:     "Low relevance random",
			term:     "AI",
			title:    "Fix typo",
			body:     "Just a small fix",
			minScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateRelevance(tt.term, tt.title, tt.body)
			if score < tt.minScore {
				t.Errorf("Score %d < expected %d", score, tt.minScore)
			}
		})
	}
}
