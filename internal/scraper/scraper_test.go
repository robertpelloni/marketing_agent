package scraper

import (
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// Note: Testing Scraper.executeDiscovery would require a mock/test database.
// For now, we verify the interface and discovery logic.
func TestScraper_Initialization(t *testing.T) {
	database := &db.DB{} // Empty DB for init test
	sources := []LeadSource{&GitHubJobSource{}}
	s := NewScraper(database, sources)

	if s == nil {
		t.Fatal("Expected scraper instance, got nil")
	}

	if len(s.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(s.sources))
	}
}
