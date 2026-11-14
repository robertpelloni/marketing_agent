package enrichment

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func dbCompanyForCircuit() db.Company {
	return db.Company{Name: "acme", Domain: "acme.com"}
}

// flakyTransport fails the first failTimes requests with a transport error
// (simulating the HTTP/2 PROTOCOL_ERROR resets seen in prod logs), then
// serves a canned GitHub search response.
type flakyTransport struct {
	calls     atomic.Int32
	failTimes int32
	status    int
	body      string
}

func (f *flakyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	n := f.calls.Add(1)
	if n <= f.failTimes {
		return nil, errors.New("stream error: stream ID 3; PROTOCOL_ERROR; received from peer")
	}
	return &http.Response{
		StatusCode: f.status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(f.body)),
		Request:    req,
	}, nil
}

func TestGitHubSearchRetriesTransientFailures(t *testing.T) {
	rt := &flakyTransport{failTimes: 2, status: 200, body: `{"items":[{"full_name":"acme/widgets"}]}`}
	src := &GitHubEnrichSource{
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
	}

	repos, err := src.searchRepos(context.Background(), "acme in:name")
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	if len(repos) != 1 || repos[0].FullName != "acme/widgets" {
		t.Fatalf("unexpected repos: %+v", repos)
	}
	if got := rt.calls.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestGitHubSearchGivesUpAfterMaxAttempts(t *testing.T) {
	rt := &flakyTransport{failTimes: 100, status: 200, body: `{}`}
	src := &GitHubEnrichSource{
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
	}

	_, err := src.searchRepos(context.Background(), "acme in:name")
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if got := rt.calls.Load(); got != 3 {
		t.Fatalf("expected exactly 3 attempts, got %d", got)
	}
}

func TestGitHubCircuitOpensAfterRepeatedFailures(t *testing.T) {
	rt := &flakyTransport{failTimes: 100, status: 200, body: `{}`}
	src := &GitHubEnrichSource{
		HTTPClient:       &http.Client{Transport: rt},
		MaxAttempts:      1,
		BaseDelay:        time.Millisecond,
		circuitThreshold: 3,
		cooldown:         time.Minute,
	}

	// Three failing Enrich calls trip the breaker (each does 2 searches).
	for i := 0; i < 3; i++ {
		_, _ = src.searchRepos(context.Background(), "acme in:name")
	}
	if !src.circuitOpen() {
		t.Fatal("expected circuit to open after threshold failures")
	}

	callsBefore := rt.calls.Load()
	_, err := src.Enrich(context.Background(), dbCompanyForCircuit())
	if err == nil || !strings.Contains(err.Error(), "circuit open") {
		t.Fatalf("expected circuit-open error, got %v", err)
	}
	if rt.calls.Load() != callsBefore {
		t.Fatal("circuit-open Enrich must not call GitHub")
	}
}

func TestGitHubCircuitIgnoresInterleavedSuccesses(t *testing.T) {
	// Prod pattern: a successful empty org search lands between every
	// burst of name-search failures. The window must still trip.
	src := &GitHubEnrichSource{circuitThreshold: 3, cooldown: time.Minute}
	for i := 0; i < 3; i++ {
		src.recordFailure()
		src.recordSuccess()
	}
	if !src.circuitOpen() {
		t.Fatal("expected circuit to open despite interleaved successes")
	}
}

func TestGitHubSearchDoesNotRetryClientErrors(t *testing.T) {
	rt := &flakyTransport{failTimes: 0, status: 404, body: `{"message":"Not Found"}`}
	src := &GitHubEnrichSource{
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
	}

	_, err := src.searchRepos(context.Background(), "acme in:name")
	if err == nil {
		t.Fatal("expected 404 error")
	}
	if got := rt.calls.Load(); got != 1 {
		t.Fatalf("expected no retries on 404, got %d attempts", got)
	}
}

type authFallbackTransport struct {
	authCalls   atomic.Int32
	unauthCalls atomic.Int32
	body        string
}

func (a *authFallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") != "" {
		a.authCalls.Add(1)
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"message":"Bad credentials"}`)),
			Request:    req,
		}, nil
	}
	a.unauthCalls.Add(1)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(a.body)),
		Request:    req,
	}, nil
}

func TestGitHubEnrichFallsBackUnauthenticatedOn401AndUsesDirectRepoDomain(t *testing.T) {
	rt := &authFallbackTransport{
		body: `[{"sha":"abc123","commit":{"author":{"name":"Jane Founder","email":"jane@acme.dev"},"message":"release v1.0"}}]`,
	}
	src := &GitHubEnrichSource{
		Token:       "revoked-token",
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 2,
		BaseDelay:   time.Millisecond,
	}

	contacts, err := src.Enrich(context.Background(), db.Company{
		Name:   "mcp-router",
		Domain: "github.com/acme-org/mcp-router",
	})
	if err != nil {
		t.Fatalf("expected Enrich to succeed via unauthenticated fallback, got %v", err)
	}
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].Email != "jane@acme.dev" || contacts[0].GitHubHandle != "acme-org" {
		t.Fatalf("unexpected contact: %+v", contacts[0])
	}
	if rt.authCalls.Load() != 1 {
		t.Fatalf("expected revoked token to be tried once, got %d", rt.authCalls.Load())
	}
	if rt.unauthCalls.Load() != 1 {
		t.Fatalf("expected 1 unauthenticated call (direct commits fast path), got %d", rt.unauthCalls.Load())
	}

	// Second call must skip the revoked token completely.
	_, err = src.Enrich(context.Background(), db.Company{
		Name:   "mcp-router",
		Domain: "github.com/acme-org/mcp-router",
	})
	if err != nil {
		t.Fatalf("second Enrich failed: %v", err)
	}
	if rt.authCalls.Load() != 1 {
		t.Fatalf("expected revoked token not to be retried on subsequent calls, got %d", rt.authCalls.Load())
	}
}
