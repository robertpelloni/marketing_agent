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

type staticCommitTransport struct {
	calls  atomic.Int32
	status int
	header http.Header
	body   string
}

func (s *staticCommitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls.Add(1)
	hdr := make(http.Header)
	for k, vals := range s.header {
		for _, v := range vals {
			hdr.Add(k, v)
		}
	}
	return &http.Response{
		StatusCode: s.status,
		Header:     hdr,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Request:    req,
	}, nil
}

func TestGitHubEnrichSkipsBotAndAgentCommitsAndFindsHumanMaintainer(t *testing.T) {
	rt := &staticCommitTransport{
		status: http.StatusOK,
		body: `[
			{"sha":"1","commit":{"author":{"name":"Codex","email":"codex@openai.com"},"message":"chore: automated fix"}},
			{"sha":"2","commit":{"author":{"name":"renovate[bot]","email":"bot@renovateapp.com"},"message":"chore(deps): bump"}},
			{"sha":"3","commit":{"author":{"name":"Scanner","email":"bot@mcp-scanner.github.io"},"message":"update report"}},
			{"sha":"4","commit":{"author":{"name":"Claude Code","email":"claude@anthropic.com"},"message":"refactor"}},
			{"sha":"5","commit":{"author":{"name":"Alex Talbot","email":"talbot@acme.dev"},"message":"release v2.0"}}
		]`,
	}
	src := &GitHubEnrichSource{
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 2,
		BaseDelay:   time.Millisecond,
	}

	contacts, err := src.Enrich(context.Background(), db.Company{
		Name:   "mcp-gateway",
		Domain: "github.com/acme-dev/mcp-gateway",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contacts) != 1 {
		t.Fatalf("expected only 1 human contact, got %d: %+v", len(contacts), contacts)
	}
	if contacts[0].Email != "talbot@acme.dev" || contacts[0].Name != "Alex Talbot" {
		t.Fatalf("unexpected contact selected: %+v", contacts[0])
	}
}

func TestGitHubEnrich403RateLimitOpensCircuitAndReturnsError(t *testing.T) {
	hdr := make(http.Header)
	hdr.Set("X-RateLimit-Remaining", "0")
	hdr.Set("Retry-After", "120")

	rt := &staticCommitTransport{
		status: http.StatusForbidden,
		header: hdr,
		body:   `{"message":"API rate limit exceeded for 5.161.250.43."}`,
	}
	src := &GitHubEnrichSource{
		HTTPClient:  &http.Client{Transport: rt},
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
	}

	fb := NewFallbackSource(
		[]EnrichmentSource{NewApolloSource("unused"), NewHunterSource("unused"), src},
		[]string{"ApolloSkipRL", "HunterSkipRL", "GitHubRLStats"},
	)

	contacts, report, err := fb.EnrichWithReport(context.Background(), db.Company{
		Name:   "agent-kit",
		Domain: "github.com/acme-dev/agent-kit",
	})
	if err != nil {
		t.Fatalf("FallbackSource.EnrichWithReport should not return top-level error, got %v", err)
	}
	if len(contacts) != 0 {
		t.Fatalf("expected 0 contacts when rate-limited, got %d", len(contacts))
	}
	if rt.calls.Load() != 1 {
		t.Fatalf("expected immediate circuit trip without retries on X-RateLimit-Remaining=0, got %d calls", rt.calls.Load())
	}
	if !src.circuitOpen() {
		t.Fatal("expected GitHubEnrichSource circuit to be open after 403 rate limit")
	}
	if report == nil || !report.HadSourceError || !report.AllCircuitOpen {
		t.Fatalf("expected HadSourceError=true and AllCircuitOpen=true, got %+v", report)
	}

	stats := SourceStats()
	if stats["GitHubRLStats"].Failures != 1 || stats["GitHubRLStats"].Empty != 0 {
		t.Fatalf("expected 1 failure / 0 empty for rate-limited GitHub commits, got %+v", stats["GitHubRLStats"])
	}
	if stats["ApolloSkipRL"].Attempts != 0 || stats["HunterSkipRL"].Attempts != 0 {
		t.Fatalf("expected domain-only sources to be skipped uncounted on github.com repo leads, got apollo=%+v hunter=%+v",
			stats["ApolloSkipRL"], stats["HunterSkipRL"])
	}
}
