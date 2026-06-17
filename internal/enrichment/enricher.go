package enrichment

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/webhook"
)

type EnrichmentSource interface {
	Enrich(ctx context.Context, company db.Company) ([]db.Contact, error)
	HealthCheck(ctx context.Context) error
}

type Enricher struct {
	db        *db.DB
	sources   []EnrichmentSource
	crmClient crm.CRMClient
	webhook   *webhook.Dispatcher
}

func NewEnricher(database *db.DB, sources []EnrichmentSource, crm crm.CRMClient) *Enricher {
	return &Enricher{
		db:        database,
		sources:   sources,
		crmClient: crm,
		webhook:   webhook.NewDispatcher(os.Getenv("WEBHOOK_URL")),
	}
}

func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	slog.Info("Enricher worker started...")
	e.ExecuteEnrichment(ctx)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: e.ExecuteEnrichment(ctx)
		}
	}
}

func (e *Enricher) ExecuteEnrichment(ctx context.Context) {
	if e.db == nil { return }
	start := time.Now()
	deals, _ := e.db.ListDealsByState(ctx, db.StateDiscovered)
	for _, d := range deals {
		comp, _ := e.db.GetCompanyByID(ctx, d.CompanyID)
		if comp == nil { continue }
		for _, s := range e.sources {
			contacts, _ := s.Enrich(ctx, *comp)
			for _, c := range contacts {
				c.CompanyID = d.CompanyID
				_ = e.db.CreateContact(ctx, &c)
			}
		}
		if err := e.db.UpdateDealState(ctx, d.ID, db.StateResearched); err == nil {
			if e.webhook != nil {
				_ = e.webhook.Dispatch(ctx, d.ID, db.StateResearched)
			}
		}
	}
	deploy.RecordTiming("Enricher", time.Since(start))
}
