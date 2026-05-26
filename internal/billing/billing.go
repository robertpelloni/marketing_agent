package billing

import (
	"context"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// InvoiceStatus represents the current state of a deal's billing.
type InvoiceStatus string

const (
	InvoicePending InvoiceStatus = "Pending"
	InvoiceSent    InvoiceStatus = "Sent"
	InvoicePaid    InvoiceStatus = "Paid"
	InvoiceFailed  InvoiceStatus = "Failed"
)

// BillingClient defines the interface for interacting with financial systems (e.g., Stripe).
type BillingClient interface {
	// CreateInvoice generates a new billing record for a won deal.
	CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error)

	// GetInvoiceStatus retrieves the payment status of an invoice.
	GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error)
}

// MockBillingClient provides a simulated implementation for billing tasks.
type MockBillingClient struct{}

func (m *MockBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "INV-MOCK-123", nil
}

func (m *MockBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	return InvoicePending, nil
}
