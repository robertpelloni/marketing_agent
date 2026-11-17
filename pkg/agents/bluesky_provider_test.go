package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Fake AT Protocol server: createSession echoes the identifier as the
// handle; createRecord requires a Bearer token.
func fakeBsky(t *testing.T, sessionCalls *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "createSession"):
			*sessionCalls++
			var body struct {
				Identifier string `json:"identifier"`
				Password   string `json:"password"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Password == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{
				"accessJwt": "tok-" + body.Identifier,
				"handle":    body.Identifier,
				"did":       "did:plc:" + body.Identifier,
			})
		case strings.HasSuffix(r.URL.Path, "createRecord"):
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer tok-") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"uri": "at://post/1"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestBlueskyPostsAsRequestedAccount(t *testing.T) {
	calls := 0
	srv := fakeBsky(t, &calls)
	defer srv.Close()

	p := NewBlueskyProvider(map[string]string{
		"tormentnexus.bsky.social": "pw1",
		"hypernexus.bsky.social":   "pw2",
	})
	p.baseURL = srv.URL

	for _, acct := range []string{"tormentnexus.bsky.social", "hypernexus.bsky.social"} {
		if err := p.Post(context.Background(), PostRequest{Brand: "x", Platform: "bluesky", Content: "hi", AccountID: acct}); err != nil {
			t.Errorf("post as %s failed: %v", acct, err)
		}
	}
	if calls != 2 {
		t.Errorf("expected 2 sessions (one per account), got %d", calls)
	}
}

func TestBlueskyRetriesSessionOnceOnFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "createSession") {
			calls++
			if calls == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"accessJwt": "tok-a", "handle": "mdmatk.bsky.social", "did": "did:plc:a"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"uri": "at://post/1"})
	}))
	defer srv.Close()

	p := NewBlueskyProvider(map[string]string{"mdmatk.bsky.social": "pw"})
	p.baseURL = srv.URL
	p.retryBackoff = 0

	if err := p.Post(context.Background(), PostRequest{Brand: "tormentnexus", Platform: "bluesky", Content: "hi", AccountID: "mdmatk.bsky.social"}); err != nil {
		t.Fatalf("post must succeed after one session retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 createSession calls (1 fail + 1 retry), got %d", calls)
	}
}

func TestBlueskyRefusesUnknownAccount(t *testing.T) {
	calls := 0
	srv := fakeBsky(t, &calls)
	defer srv.Close()

	p := NewBlueskyProvider(map[string]string{"tormentnexus.bsky.social": "pw1"})
	p.baseURL = srv.URL

	err := p.Post(context.Background(), PostRequest{Brand: "hypernexus", Platform: "bluesky", Content: "hi", AccountID: "hypernexus.bsky.social"})
	if err == nil || !strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("expected no-credentials refusal, got: %v", err)
	}
	if calls != 0 {
		t.Errorf("must not even create a session for an unknown account, got %d calls", calls)
	}
}
