package enrichment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Enricher worker started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Enricher worker stopping: Draining in-flight work")
			return
		case <-ticker.C:
			e.executeEnrichment(ctx)
		}
	}
}

// ExecuteEnrichment manually triggers an enrichment cycle (exported for testing).
func (e *Enricher) ExecuteEnrichment(ctx context.Context) {
	e.executeEnrichment(ctx)
}

func (e *Enricher) executeEnrichment(ctx context.Context) {
	// 1. Find deals in Discovered state
	deals, err := e.db.ListDealsByState(ctx, db.StateDiscovered)
	if err != nil {
		slog.Error("Enricher: Error listing discovered deals", "error", err)
		return
	}

	for _, deal := range deals {
		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
			slog.Error("Enricher: Error getting company", "company_id", deal.CompanyID, "error", err)
			continue
		}

		err = e.enrichCompany(ctx, deal, *company)
		if err != nil {
			slog.Error("Enricher: Error enriching company", "company_name", company.Name, "error", err)
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) error {
	for _, source := range e.sources {
		contacts, err := source.Enrich(ctx, company)
		if err != nil {
			slog.Error("Enricher: Error from source", "error", err)
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
				slog.Error("Enricher: Error persisting contact", "contact_name", contact.Name, "error", err)
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return fmt.Errorf("failed to update deal state: %w", err)
			}

			// Synchronize newly found contacts with the CRM (with retry logic)
			if e.crmClient != nil {
				go func() {
					maxRetries := 3
					for i := 0; i < maxRetries; i++ {
						if err := e.crmClient.SyncContacts(ctx, company.ID, contacts); err != nil {
							slog.Warn("Enricher: Failed to sync contacts to CRM", "attempt", i+1, "max_retries", maxRetries, "company_id", company.ID, "error", err)
							time.Sleep(time.Duration(i+1) * 2 * time.Second)
							continue
						}
						return
					}
					slog.Error("Enricher: CRM contact synchronization failed after all attempts", "company_id", company.ID)
				}()
			}

			slog.Info("Enricher: Successfully enriched company", "company_name", company.Name, "contacts_count", len(contacts))
			return nil
		}
	}
	return nil
}

// MockApolloSource is a simulated enrichment source.
type MockApolloSource struct{}

func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	slog.Info("MockApolloSource: Searching for contacts", "domain", company.Domain)

	// Simulate finding contacts based on domain
	if company.Domain == "aidynamics.com" {
		return []db.Contact{
			{
				Name:         "Sarah Chen",
				Role:         "Director of AI",
				Email:        "sarah.chen@aidynamics.com",
				GitHubHandle: "schen-ai",
			},
		}, nil
	} else if company.Domain == "neuralsystems.io" {
		return []db.Contact{
			{
				Name:         "James Wilson",
				Role:         "Principal Systems Architect",
				Email:        "j.wilson@neuralsystems.io",
				GitHubHandle: "jwilson-sys",
			},
		}, nil
	}

	return nil, nil
}
