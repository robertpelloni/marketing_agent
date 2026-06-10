package sales

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// OrderDB defines the database interface needed by the OrderProcessor.
type OrderDB interface {
	GetCompanyByID(ctx context.Context, id int64) (*db.Company, error)
}

// Processor coordinates the transition from a won deal to an active order.
type Processor struct {
	db      OrderDB
	billing billing.BillingClient
	crm     crm.CRMClient
}

// NewOrderProcessor creates a new Processor instance.
func NewOrderProcessor(database OrderDB, billingClient billing.BillingClient, crmClient crm.CRMClient) *Processor {
	return &Processor{
		db:      database,
		billing: billingClient,
		crm:     crmClient,
	}
}

// ProcessOrder handles fulfillment for a deal that has been closed won.
func (p *Processor) ProcessOrder(ctx context.Context, deal db.Deal) error {
	slog.Info("OrderProcessor Processing fulfillment for deal", "deal_ID", deal.ID)

	company, err := p.db.GetCompanyByID(ctx, deal.CompanyID)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	// 1. Generate Invoice
	invoiceID, err := p.billing.CreateInvoice(ctx, deal, *company)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	slog.Info("OrderProcessor Invoice  created for deal", "invoiceID", invoiceID, "deal_ID", deal.ID)

	// 2. Synchronize with CRM
	err = p.crm.SyncInteraction(ctx, deal.ID, fmt.Sprintf("Order processed. Invoice: %s", invoiceID))
	if err != nil {
		slog.Error("OrderProcessor Warning CRM sync failed", "error", err)
	}

	// 3. (Optional) Trigger Provisioning
	// In a full implementation, this might call a provisioning service.
	slog.Info("OrderProcessor Fulfillment complete for deal", "deal_ID", deal.ID)

	return nil
}
