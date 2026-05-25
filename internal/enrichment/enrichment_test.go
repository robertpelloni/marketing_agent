package enrichment

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestMockApolloSource_Enrich(t *testing.T) {
	source := &MockApolloSource{}
	company := db.Company{
		Name:   "AI Dynamics Corp",
		Domain: "aidynamics.com",
	}

	contacts, err := source.Enrich(context.Background(), company)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	if contacts[0].Name != "Sarah Chen" {
		t.Errorf("Expected Sarah Chen, got %s", contacts[0].Name)
	}
}

func TestEnricher_Initialization(t *testing.T) {
	database := &db.DB{}
	sources := []EnrichmentSource{&MockApolloSource{}}
	e := NewEnricher(database, sources)

	if e == nil {
		t.Fatal("Expected enricher instance, got nil")
	}

	if len(e.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(e.sources))
	}
}
