package sales

import (
	"context"
	"errors"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/billing"
	"github.com/robertpelloni/marketing_agent/internal/crm"
	"github.com/robertpelloni/marketing_agent/internal/db"
)

// Mock Billing Client for detailed testing
type errorBillingClient struct {}

func (e *errorBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "", errors.New("billing service down")
}

func (e *errorBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (billing.InvoiceStatus, error) {
	return billing.InvoicePending, nil
}

type successBillingClient struct {}

func (s *successBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "INV-TEST", nil
}

func (s *successBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (billing.InvoiceStatus, error) {
	return billing.InvoicePending, nil
}

// Mock CRM Client for detailed testing
type errorCRMClient struct {
	crm.MockCRMClient
}

func (e *errorCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	return errors.New("crm service down")
}

// Mock objects for testing
type mockDB struct {
	company *db.Company
	err     error
}

func (m *mockDB) GetCompanyByID(ctx context.Context, id int64) (*db.Company, error) {
	return m.company, m.err
}

func TestProcessOrder_BillingFailure(t *testing.T) {
	billingClient := &errorBillingClient{}
	crmClient := &crm.MockCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err == nil {
		t.Fatal("Expected billing failure error, got nil")
	}
	if err.Error() != "failed to create invoice: billing service down" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestProcessOrder_Success(t *testing.T) {
	billingClient := &successBillingClient{}
	crmClient := &crm.MockCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err != nil {
		t.Fatalf("ProcessOrder failed: %v", err)
	}
}

func TestProcessOrder_CRMWarning(t *testing.T) {
	// CRM sync failure should only log a warning and not return an error in the current implementation.
	billingClient := &successBillingClient{}
	crmClient := &errorCRMClient{}
	dbMock := &mockDB{company: &db.Company{ID: 1, Name: "TestCorp"}}

	p := NewOrderProcessor(dbMock, billingClient, crmClient)

	err := p.ProcessOrder(context.Background(), db.Deal{ID: 1, CompanyID: 1})
	if err != nil {
		t.Fatalf("ProcessOrder should not fail on CRM error, but got: %v", err)
	}
}

func (m *successBillingClient) CancelSubscription(ctx context.Context, subID string, prorate bool) error { return nil }
func (m *errorBillingClient) CancelSubscription(ctx context.Context, subID string, prorate bool) error { return errors.New("cancel failed") }
func (m *successBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier string, successURL string, cancelURL string) (string, error) { return "http://checkout.url", nil }
func (m *errorBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier string, successURL string, cancelURL string) (string, error) { return "", errors.New("failed") }
