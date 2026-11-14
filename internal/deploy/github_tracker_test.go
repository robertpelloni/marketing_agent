package deploy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGitHubCITracker_GetLatestStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"workflow_runs": [
				{
					"status": "completed",
					"conclusion": "success"
				}
			]
		}`)
	}))
	defer server.Close()

	tracker := &GitHubCITracker{owner: "robertpelloni", repo: "marketing_agent", apiBase: server.URL}
	status, err := tracker.GetLatestStatus(context.Background(), "main")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != CIStatusSuccess {
		t.Fatalf("expected CIStatusSuccess, got %v", status)
	}
}

// A revoked token (401 with auth) must fall back to an unauthenticated
// request: workflow runs for a public repo are public data, and health
// reporting should not go blind because a credential died.
func TestGitHubCITracker_RetriesWithoutAuthOn401(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"message":"Bad credentials"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"workflow_runs":[{"status":"completed","conclusion":"success"}]}`)
	}))
	defer server.Close()

	tracker := &GitHubCITracker{owner: "robertpelloni", repo: "marketing_agent", token: "revoked-token", apiBase: server.URL}
	status, err := tracker.GetLatestStatus(context.Background(), "main")
	if err != nil {
		t.Fatalf("expected unauthenticated fallback to succeed, got %v", err)
	}
	if status != CIStatusSuccess {
		t.Fatalf("expected CIStatusSuccess, got %v", status)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("expected exactly 2 requests (auth + unauth), got %d", got)
	}
}
