package crm

import (
	"context"
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
			log.Println("CRM Worker: Synchronization stopping...")
			return
		case <-ticker.C:
			w.sync(ctx)
		}
	}
}

func (w *Worker) sync(ctx context.Context) {
	log.Println("CRM Worker: Executing sync cycle...")

	// 1. Reconcile updates from CRM
	updates, err := w.client.GetLeadUpdates(ctx)
	if err != nil {
		log.Printf("CRM Worker: Error fetching updates: %v", err)
	} else {
		for _, update := range updates {
			log.Printf("CRM Worker: Reconciling update for lead %s to %s", update.ID, update.NewState)
			// Logic to find local lead by CRM ID and update state would go here
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
			if err := w.client.PushDeal(ctx, deal, *company); err != nil {
				log.Printf("CRM Worker: Error pushing deal %d: %v", deal.ID, err)
			}
		}
	}
}
