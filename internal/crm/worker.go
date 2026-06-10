package crm

import (
	"context"
	"fmt"
	"log/slog"
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

	slog.Info("CRM Worker Synchronization started (interval )...", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("CRM Worker: Synchronization stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			w.sync(ctx)
		}
	}
}

func (w *Worker) sync(ctx context.Context) {
	slog.Info("CRM Worker: Executing sync cycle...")

	// 1. Reconcile updates from CRM
	updates, err := w.client.GetLeadUpdates(ctx)
	if err != nil {
		slog.Error("CRM Worker Error fetching updates", "error", err)
	} else {
		for _, update := range updates {
			slog.Info("CRM Worker Reconciling update for lead  to", "update_ID", update.ID, "update_NewState", update.NewState)
			// In a real system, we'd map external IDs. For now, we assume numeric mapping.
			var dealID int64
			if _, err := fmt.Sscanf(update.ID, "%d", &dealID); err == nil {
				if err := w.db.UpdateDealState(ctx, dealID, update.NewState); err != nil {
					slog.Error("CRM Worker Failed to update deal", "dealID", dealID, "error", err)
				}
			}
		}
	}

	// 2. Push local Negotiating/Closed deals to CRM
	// This is a simplified implementation
	deals, err := w.db.ListDealsByState(ctx, db.StateNegotiating)
	if err != nil {
		slog.Error("CRM Worker Error listing negotiating deals", "error", err)
		return
	}

	for _, deal := range deals {
		company, _ := w.db.GetCompanyByID(ctx, deal.CompanyID)
		if company != nil {
			// 2a. Push local updates to CRM
			if err := w.client.PushDeal(ctx, deal, *company, "WorkerSync"); err != nil {
				slog.Error("CRM Worker Error pushing deal", "deal_ID", deal.ID, "error", err)
			}

			// 2b. Pull latest details from CRM to keep local state synchronized
			details, err := w.client.FetchDealDetails(ctx, deal.ID)
			if err != nil {
				slog.Error("CRM Worker Error fetching deal details for", "deal_ID", deal.ID, "error", err)
				continue
			}

			if details != nil {
				slog.Info("CRM Worker Synchronizing details for deal  from CRM", "deal_ID", deal.ID)
				// Update local deal pricing and requirements if they differ
				if err := w.db.UpdateDealDetails(ctx, deal.ID, details.QuotedPricing, details.CustomRequirements); err != nil {
					slog.Error("CRM Worker Error updating local deal", "deal_ID", deal.ID, "error", err)
				}
			}
		}
	}
}
