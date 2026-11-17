package scraper

import (
	"context"
	"testing"
)

func TestLinkedInSourceDisabledSkipsScrape(t *testing.T) {
	t.Setenv("LINKEDIN_DISABLED", "true")
	t.Setenv("LINKEDIN_USERNAME", "someone")
	t.Setenv("LINKEDIN_PASSWORD", "secret")

	src := &LinkedInSource{}
	companies, err := src.Discover(context.Background(), []string{"ai"})
	if err != nil {
		t.Fatalf("disabled lane must not error, got: %v", err)
	}
	if len(companies) != 0 {
		t.Fatalf("disabled lane must return no companies (no scrape, no fabricated leads), got %d", len(companies))
	}
}
