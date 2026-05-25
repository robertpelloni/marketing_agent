package crm

import (
	"context"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LeadUpdate represents a change in lead status from the CRM.
type LeadUpdate struct {
	ID        string
	NewState  db.LeadState
	Notes     string
}

// CRMClient defines the interface for interacting with external CRM systems.
type CRMClient interface {
	// PushDeal synchronizes a local deal to the CRM.
	PushDeal(ctx context.Context, deal db.Deal, company db.Company) error

	// GetLeadUpdates retrieves status changes from the CRM for local reconciliation.
	GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error)

	// ValidateAccount checks if a company account is valid and active in the CRM.
	ValidateAccount(ctx context.Context, domain string) (bool, error)
}
