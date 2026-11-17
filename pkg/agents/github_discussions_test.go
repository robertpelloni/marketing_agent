package agents

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const searchFixture = `{
  "data": {
    "search": {
      "edges": [
        {"node": {"id": "D_1", "title": "MCP tool routing help", "url": "https://github.com/o/r/discussions/1", "body": "how?", "repository": {"nameWithOwner": "acme/widgets"}}},
        {"node": {"id": "D_2", "title": "Agent memory persistence", "url": "https://github.com/o2/r2/discussions/2", "body": "local?", "repository": {"nameWithOwner": "beta/tools"}}},
        {"node": {}}
      ]
    }
  }
}`

func TestSearchDiscussionsParsesGraphQL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchFixture))
	}))
	defer srv.Close()

	g := NewGitHubDiscussionsEngager("tok")
	g.apiURL = srv.URL

	got, err := g.searchDiscussions(context.Background(), "MCP in:title")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 discussions (empty node skipped), got %d: %+v", len(got), got)
	}
	if got[0].DiscussionNodeID != "D_1" || got[0].Title != "MCP tool routing help" || got[0].RepoOwner != "acme" || got[0].RepoName != "widgets" {
		t.Errorf("unexpected first result: %+v", got[0])
	}
	if got[1].RepoOwner != "beta" || got[1].RepoName != "tools" {
		t.Errorf("unexpected second result: %+v", got[1])
	}
}

func TestSearchDiscussionsSurfacesGraphQLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Bad credentials"}]}`))
	}))
	defer srv.Close()

	g := NewGitHubDiscussionsEngager("tok")
	g.apiURL = srv.URL

	if _, err := g.searchDiscussions(context.Background(), "x"); err == nil || !strings.Contains(err.Error(), "Bad credentials") {
		t.Fatalf("expected GraphQL error to surface, got: %v", err)
	}
}

func TestPostReplySendsMutation(t *testing.T) {
	var sawMutation bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		sawMutation = strings.Contains(string(buf[:n]), "addDiscussionComment") && strings.Contains(string(buf[:n]), "D_1")
		_, _ = w.Write([]byte(`{"data":{"addDiscussionComment":{"comment":{"id":"DC_1"}}}}`))
	}))
	defer srv.Close()

	g := NewGitHubDiscussionsEngager("tok")
	g.apiURL = srv.URL

	if err := g.postReply(context.Background(), "D_1", "hello"); err != nil {
		t.Fatalf("postReply failed: %v", err)
	}
	if !sawMutation {
		t.Error("expected addDiscussionComment mutation with discussion id in request body")
	}
}

func TestEngageSkipsWhenPostingDisabled(t *testing.T) {
	t.Setenv("GITHUB_POSTING_DISABLED", "true")
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte(searchFixture))
	}))
	defer srv.Close()

	g := NewGitHubDiscussionsEngager("tok")
	g.apiURL = srv.URL
	g.engage(context.Background())

	if called {
		t.Error("engage must not call the API when GITHUB_POSTING_DISABLED=true")
	}
}
