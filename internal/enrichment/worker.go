package enrichment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Enricher worker started...")

	// Run immediately on startup
	e.executeEnrichment(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Enricher worker stopping: Draining in-flight work...")
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
	if e.db == nil {
		// DB not available – skip enrichment to avoid panic in dev/test environments
		slog.Info("Enricher: DB unavailable, skipping enrichment cycle")
		return
	}
	// 1. Find deals in Discovered state. This must use the enrichment query,
	// not ListDealsByState: that one inner-joins contacts for outreach, and a
	// deal with no contacts yet is exactly what enrichment exists to fix.
	deals, err := e.db.ListDealsForEnrichment(ctx, 100)
	if err != nil {
		slog.Info(fmt.Sprintf("Enricher: Error listing discovered deals: %v", err))
		return
	}

	for _, deal := range deals {
		select {
		case <-ctx.Done():
			return
		default:
		}

		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error getting company %d: %v", deal.CompanyID, err))
			continue
		}

		stopBatch, err := e.enrichCompany(ctx, deal, *company)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error enriching company %s: %v", company.Name, err))
		}
		if stopBatch {
			slog.Info("Enricher: All applicable enrichment sources are cooling down; pausing remaining batch until next cycle")
			return
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) (bool, error) {
	hadSourceError := false
	allCircuitOpen := false

	for _, source := range e.sources {
		if source == nil {
			continue
		}

		var contacts []db.Contact
		var err error
		if fb, ok := source.(*FallbackSource); ok {
			var report *FallbackReport
			contacts, report, err = fb.EnrichWithReport(ctx, company)
			if report != nil {
				if report.HadSourceError {
					hadSourceError = true
				}
				if report.AllCircuitOpen {
					allCircuitOpen = true
				}
			}
		} else {
			contacts, err = source.Enrich(ctx, company)
		}

		if err != nil {
			hadSourceError = true
			slog.Info(fmt.Sprintf("Enricher: Error from source: %v", err))
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
				slog.Info(fmt.Sprintf("Enricher: Error persisting contact %s: %v", contact.Name, err))
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return false, fmt.Errorf("failed to update deal state: %w", err)
			}

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
			return false, nil
		}
	}

	// Rotate FIFO enrichment queue (ORDER BY updated_at ASC) only when sources
	// completed cleanly with zero contacts (e.g. noreply-only repos or garbage
	// leads). Keep updated_at untouched on rate-limit/transient source errors so
	// those leads stay first in line for the next cycle.
	if !hadSourceError && e.db != nil {
		if err := e.db.TouchDealUpdatedAt(ctx, deal.ID); err != nil {
			slog.Info(fmt.Sprintf("Enricher: Failed to rotate deal %d updated_at: %v", deal.ID, err))
		}
	}

	return allCircuitOpen, nil
}
