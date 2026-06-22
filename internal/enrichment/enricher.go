package enrichment

import (
	"context"
<<<<<<< HEAD
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

=======

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
>>>>>>> origin/main
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// EnrichmentSource defines an interface for finding contact details for a company.
>>>>>>> origin/main
type EnrichmentSource interface {
	Enrich(ctx context.Context, company db.Company) ([]db.Contact, error)
}

<<<<<<< HEAD
=======
// Enricher coordinates the enrichment of company leads with contact data.
>>>>>>> origin/main
type Enricher struct {
<<<<<<< HEAD
	db      *db.DB
	sources []EnrichmentSource
}

// NewEnricher creates a new Enricher instance.
func NewEnricher(database *db.DB, sources []EnrichmentSource) *Enricher {
	return &Enricher{
		db:      database,
		sources: sources,
=======
	db        *db.DB
	sources   []EnrichmentSource
	crmClient crm.CRMClient
}

<<<<<<< HEAD
func NewEnricher(database *db.DB, sources []EnrichmentSource, crm crm.CRMClient) *Enricher {
	return &Enricher{db: database, sources: sources, crmClient: crm}
}

func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Println("Enricher worker started...")
	e.executeEnrichment(ctx)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: e.executeEnrichment(ctx)
		}
	}
}

func (e *Enricher) executeEnrichment(ctx context.Context) {
	start := time.Now()
	deals, _ := e.db.ListDealsByState(ctx, db.StateDiscovered)
	for _, d := range deals {
		comp, _ := e.db.GetCompanyByID(ctx, d.CompanyID)
		if comp == nil { continue }
		for _, s := range e.sources {
			contacts, _ := s.Enrich(ctx, *comp)
			for _, c := range contacts {
				_ = e.db.CreateContact(ctx, &c)
			}
		}
		_ = e.db.UpdateDealState(ctx, d.ID, db.StateResearched)
	}
	deploy.RecordTiming("Enricher", time.Since(start))
}
=======
// NewEnricher creates a new Enricher instance.
func NewEnricher(database *db.DB, sources []EnrichmentSource, crmClient crm.CRMClient) *Enricher {
	return &Enricher{
		db:        database,
		sources:   sources,
		crmClient: crmClient,
>>>>>>> origin/main
	}
}
>>>>>>> origin/main
