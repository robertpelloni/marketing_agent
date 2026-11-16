package communication

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/go-github/v60/github"
	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func testGitHubClient(t *testing.T, base string, hc *http.Client) *github.Client {
	t.Helper()
	c := github.NewClient(hc)
	u, err := url.Parse(base + "/")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	c.BaseURL = u
	c.UploadURL = u
	return c
}

// A rejected token must switch search to the anonymous client instead of
// halting, and FindAndComment must skip posting gracefully (no failed send)
// while the token is known-rejected.
func TestGitHubCommentSender_AnonFallbackOn401(t *testing.T) {
	var authHits, anonHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/issues") {
			if r.Header.Get("Authorization") != "" {
				authHits.Add(1)
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintln(w, `{"message":"Bad credentials"}`)
				return
			}
			anonHits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"total_count":1,"items":[{"number":7,"title":"MCP server routing help","html_url":"https://github.com/acme/widgets/issues/7","repository_url":"https://api.github.com/repos/acme/widgets"}]}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	authClient := testGitHubClient(t, server.URL, &http.Client{
		Transport: &authTransport{token: "revoked"},
	})
	anonClient := testGitHubClient(t, server.URL, nil)
	sender := &GitHubCommentSender{client: authClient, anon: anonClient, repo: "robertpelloni/marketing_agent"}

	targets, err := sender.SearchRelevantIssues(context.Background(), "acme")
	if err != nil {
		t.Fatalf("expected anonymous fallback search to succeed, got %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("expected at least one issue target from anonymous search")
	}
	if !sender.tokenRejected {
		t.Fatal("expected tokenRejected to be set after 401")
	}
	if authHits.Load() == 0 || anonHits.Load() == 0 {
		t.Fatalf("expected auth + anon attempts, got auth=%d anon=%d", authHits.Load(), anonHits.Load())
	}

	// Posting must be skipped (nil error), not attempted and failed.
	err = sender.FindAndComment(context.Background(),
		db.Company{Name: "Acme", Domain: "acme.com"},
		db.Contact{Email: "dev@acme.com"})
	if err != nil {
		t.Fatalf("expected graceful skip while token rejected, got %v", err)
	}
}

type authTransport struct {
	token string
	base  http.RoundTripper
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "token "+a.token)
	base := a.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req2)
}
