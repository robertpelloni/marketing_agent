package agents

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHubSkipsTwitterWhenDisabled(t *testing.T) {
	t.Setenv("TWITTER_API_KEY", "test-key")
	t.Setenv("TWITTER_BEARER_TOKEN", "test-bearer")
	t.Setenv("TWITTER_DISABLED", "true")

	hub := NewSocialPosterHub()
	if hub.Provider("twitter") != nil {
		t.Fatal("expected twitter provider to be skipped when TWITTER_DISABLED=true")
	}
}

func TestHubRegistersTwitterWhenEnabled(t *testing.T) {
	t.Setenv("TWITTER_API_KEY", "test-key")
	t.Setenv("TWITTER_DISABLED", "")

	hub := NewSocialPosterHub()
	if hub.Provider("twitter") == nil {
		t.Fatal("expected twitter provider to be registered when not disabled")
	}
}

func TestTwitterPostCreditsDepletedIsSentinel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"detail":"credits depleted","status":402,"title":"Payment Required"}`))
	}))
	defer srv.Close()

	p := NewTwitterProvider("k", "s", "tok", "toksecret", "")
	p.baseURL = srv.URL

	err := p.Post(context.Background(), PostRequest{Brand: "hypernexus", Platform: "twitter", Content: "hello", AccountID: "@hypernexus_site"})
	if !errors.Is(err, ErrTwitterCreditsDepleted) {
		t.Fatalf("expected ErrTwitterCreditsDepleted, got: %v", err)
	}
}
