package crm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Worker coordinates the synchronization between the local database and the external CRM.
type Worker struct {
	db     *db.DB
	client CRMClient
}

// NewWorker creates a new CRM synchronization worker.
func NewWorker(database *db.DB, client CRMClient) *Worker {
	return &Worker{
		db:     database,
		client: client,
	}
}

// Run starts the periodic CRM synchronization process.
func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("CRM Worker: Synchronization started (interval: %v)...", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("CRM Worker: Synchronization stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			w.sync(ctx)
		}
	}
}

func (w *Worker) sync(ctx context.Context) {
	log.Println("CRM Worker: Executing sync cycle...")

	// 0. Pull new interactions from CRM
	newInteractions, err := w.client.GetNewInteractions(ctx)
	if err != nil {
		log.Printf("CRM Worker: Error fetching new interactions: %v", err)
	} else {
		for _, interaction := range newInteractions {
			log.Printf("CRM Worker: Processing new interaction from CRM: %s", interaction.RawText)
			// In a real system, we'd lookup the contact by email and persist
			// For now, we log the activity.
		}
	}

	// 1. Reconcile updates from CRM
	updates, err := w.client.GetLeadUpdates(ctx)
	if err != nil {
		log.Printf("CRM Worker: Error fetching updates: %v", err)
	} else {
		for _, update := range updates {
			log.Printf("CRM Worker: Reconciling update for lead %s to %s", update.ID, update.NewState)
			// In a real system, we'd map external IDs. For now, we assume numeric mapping.
			var dealID int64
			if _, err := fmt.Sscanf(update.ID, "%d", &dealID); err == nil {
				if err := w.db.UpdateDealState(ctx, dealID, update.NewState); err != nil {
					log.Printf("CRM Worker: Failed to update deal %d: %v", dealID, err)
				}
			}
		}
	}

	// 2. Push local Negotiating/Closed deals to CRM
	// This is a simplified implementation
	deals, err := w.db.ListDealsByState(ctx, db.StateNegotiating)
	if err != nil {
		log.Printf("CRM Worker: Error listing negotiating deals: %v", err)
		return
	}

	for _, deal := range deals {
		company, _ := w.db.GetCompanyByID(ctx, deal.CompanyID)
		if company != nil {
			// 2a. Push local updates to CRM
			if err := w.client.PushDeal(ctx, deal, *company, "WorkerSync"); err != nil {
				log.Printf("CRM Worker: Error pushing deal %d: %v", deal.ID, err)
			}

			// 2b. Pull latest details from CRM to keep local state synchronized
			details, err := w.client.FetchDealDetails(ctx, deal.ID)
			if err != nil {
				log.Printf("CRM Worker: Error fetching deal details for %d: %v", deal.ID, err)
				continue
			}

			if details != nil {
				log.Printf("CRM Worker: Synchronizing details for deal %d from CRM", deal.ID)
				// Update local deal pricing and requirements if they differ
				if err := w.db.UpdateDealDetails(ctx, deal.ID, details.QuotedPricing, details.CustomRequirements); err != nil {
					log.Printf("CRM Worker: Error updating local deal %d: %v", deal.ID, err)
				}
			}
		}
	}
}
