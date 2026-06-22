package enrichment

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"log/slog"
>>>>>>> origin/main
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	log.Println("Enricher worker started...")
=======
	slog.Info("Enricher worker started...")

	// Run immediately on startup
	e.executeEnrichment(ctx)
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			log.Println("Enricher worker stopping...")
=======
			slog.Info("Enricher worker stopping: Draining in-flight work...")
>>>>>>> origin/main
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
<<<<<<< HEAD
	// 1. Find deals in Discovered state
	deals, err := e.db.ListDealsByState(ctx, db.StateDiscovered)
	if err != nil {
		log.Printf("Enricher: Error listing discovered deals: %v", err)
=======
	if e.db == nil {
		// DB not available – skip enrichment to avoid panic in dev/test environments
		slog.Info("Enricher: DB unavailable, skipping enrichment cycle")
		return
	}
	// 1. Find deals in Discovered state
	deals, err := e.db.ListDealsByState(ctx, db.StateDiscovered)
	if err != nil {
		slog.Info(fmt.Sprintf("Enricher: Error listing discovered deals: %v", err))
>>>>>>> origin/main
		return
	}

	for _, deal := range deals {
		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
<<<<<<< HEAD
			log.Printf("Enricher: Error getting company %d: %v", deal.CompanyID, err)
=======
			slog.Info(fmt.Sprintf("Enricher: Error getting company %d: %v", deal.CompanyID, err))
>>>>>>> origin/main
			continue
		}

		err = e.enrichCompany(ctx, deal, *company)
		if err != nil {
<<<<<<< HEAD
			log.Printf("Enricher: Error enriching company %s: %v", company.Name, err)
=======
			slog.Info(fmt.Sprintf("Enricher: Error enriching company %s: %v", company.Name, err))
>>>>>>> origin/main
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) error {
	for _, source := range e.sources {
		contacts, err := source.Enrich(ctx, company)
		if err != nil {
<<<<<<< HEAD
			log.Printf("Enricher: Error from source: %v", err)
=======
			slog.Info(fmt.Sprintf("Enricher: Error from source: %v", err))
>>>>>>> origin/main
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
<<<<<<< HEAD
				log.Printf("Enricher: Error persisting contact %s: %v", contact.Name, err)
=======
				slog.Info(fmt.Sprintf("Enricher: Error persisting contact %s: %v", contact.Name, err))
>>>>>>> origin/main
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return fmt.Errorf("failed to update deal state: %w", err)
			}
<<<<<<< HEAD
			log.Printf("Enricher: Successfully enriched %s with %d contacts", company.Name, len(contacts))
=======

			// Synchronize newly found contacts with the CRM (with retry logic)
			if e.crmClient != nil {
				go func() {
					maxRetries := 3
					for i := 0; i < maxRetries; i++ {
						if err := e.crmClient.SyncContacts(ctx, company.ID, contacts); err != nil {
							slog.Info(fmt.Sprintf("Enricher Warning: Failed to sync contacts to CRM (attempt %d/%d): %v", i+1, maxRetries, err))
							time.Sleep(time.Duration(i+1) * 2 * time.Second)
							continue
						}
						return
					}
					slog.Info(fmt.Sprintf("Enricher Error: CRM contact synchronization failed after %d attempts for company %d", maxRetries, company.ID))
				}()
			}

			slog.Info(fmt.Sprintf("Enricher: Successfully enriched %s with %d contacts", company.Name, len(contacts)))
>>>>>>> origin/main
			return nil
		}
	}
	return nil
}

<<<<<<< HEAD
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

=======
// MockApolloSource is a legacy stub that no longer generates mock contacts.
// Only real enrichment sources (Hunter.io, Apollo.io) are used.
type MockApolloSource struct{}

func (m *MockApolloSource) Enrich(_ context.Context, _ db.Company) ([]db.Contact, error) {
>>>>>>> origin/main
	return nil, nil
}
