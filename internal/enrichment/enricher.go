package enrichment

import (
	"context"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// EnrichmentSource defines an interface for finding contact details for a company.
type EnrichmentSource interface {
	Enrich(ctx context.Context, company db.Company) ([]db.Contact, error)
}

// Enricher coordinates the enrichment of company leads with contact data.
type Enricher struct {
	db      *db.DB
	sources []EnrichmentSource
}

// NewEnricher creates a new Enricher instance.
func NewEnricher(database *db.DB, sources []EnrichmentSource) *Enricher {
	return &Enricher{
		db:      database,
		sources: sources,
	}
}
