package enrichment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("Enricher worker started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Enricher worker stopping: Draining in-flight work...")
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
		log.Printf("Enricher: Error listing discovered deals: %v", err)
		return
	}

	for _, deal := range deals {
		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
			log.Printf("Enricher: Error getting company %d: %v", deal.CompanyID, err)
			continue
		}

		err = e.enrichCompany(ctx, deal, *company)
		if err != nil {
			log.Printf("Enricher: Error enriching company %s: %v", company.Name, err)
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) error {
	for _, source := range e.sources {
		contacts, err := source.Enrich(ctx, company)
		if err != nil {
			log.Printf("Enricher: Error from source: %v", err)
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
				log.Printf("Enricher: Error persisting contact %s: %v", contact.Name, err)
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return fmt.Errorf("failed to update deal state: %w", err)
			}

			// Synchronize newly found contacts with the CRM
			if e.crmClient != nil {
				if err := e.crmClient.SyncContacts(ctx, company.ID, contacts); err != nil {
					log.Printf("Enricher Warning: Failed to sync contacts to CRM: %v", err)
				}
			}

			log.Printf("Enricher: Successfully enriched %s with %d contacts", company.Name, len(contacts))
			return nil
		}
	}
	return nil
}

// MockApolloSource is a simulated enrichment source.
type MockApolloSource struct{}

func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	log.Printf("MockApolloSource: Searching for contacts at %s", company.Domain)

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
