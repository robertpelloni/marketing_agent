package enrichment

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

type apollo401Transport struct{}

func (apollo401Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 401,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"error":"invalid api key"}`)),
		Request:    req,
	}, nil
}

// An invalid Apollo key must surface as an error (counted as a failure by
// the fallback chain), not as an empty result that looked like "no
// contacts" in telemetry while the credential lane was dead.
func TestApolloEnrich_Classifies401AsFailure(t *testing.T) {
	src := &ApolloSource{
		APIKey:     "invalid-key",
		HTTPClient: &http.Client{Transport: apollo401Transport{}},
	}

	_, err := src.Enrich(context.Background(), db.Company{Name: "Acme", Domain: "acme.com"})
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 in error, got %v", err)
	}

	// Through the fallback chain, the attempt must land in Failures.
	fb := NewFallbackSource([]EnrichmentSource{src}, []string{"Apollo401Stats"})
	_, _ = fb.Enrich(context.Background(), db.Company{Name: "Acme", Domain: "acme.com"})
	if got := SourceStats()["Apollo401Stats"]; got.Failures != 1 || got.Empty != 0 {
		t.Fatalf("expected 1 failure / 0 empty for Apollo 401, got %+v", got)
	}
}
