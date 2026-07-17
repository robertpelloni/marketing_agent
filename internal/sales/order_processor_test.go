package sales

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/billing"
	"gitlab.com/robertpelloni/marketing_agent/internal/crm"
	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// Mock Billing Client for detailed testing
type errorBillingClient struct {}

func (e *errorBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "", errors.New("billing service down")
}

func (e *errorBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (billing.InvoiceStatus, error) {
	return billing.InvoicePending, nil
}

func (e *errorBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier billing.Tier, successURL, cancelURL string) (string, error) {
	return "", errors.New("billing service down")
}

func (e *errorBillingClient) GetSubscription(ctx context.Context, subID string) (*billing.SubscriptionInfo, error) {
	return nil, errors.New("billing service down")
}

func (e *errorBillingClient) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error {
	return errors.New("billing service down")
}

func (e *errorBillingClient) UpdateSubscriptionSeats(ctx context.Context, subID string, seats int) error {
	return errors.New("billing service down")
}

func (e *errorBillingClient) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) (string, error) {
	return "", errors.New("billing service down")
}

type successBillingClient struct {}

func (s *successBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "INV-TEST", nil
}

func (s *successBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (billing.InvoiceStatus, error) {
	return billing.InvoicePending, nil
}

func (s *successBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier billing.Tier, successURL, cancelURL string) (string, error) {
	return "https://checkout.stripe.com/test", nil
}

func (s *successBillingClient) GetSubscription(ctx context.Context, subID string) (*billing.SubscriptionInfo, error) {
	return &billing.SubscriptionInfo{StripeSubID: subID, State: "active"}, nil
}

func (s *successBillingClient) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error {
	return nil
}

func (s *successBillingClient) UpdateSubscriptionSeats(ctx context.Context, subID string, seats int) error {
	return nil
}

func (s *successBillingClient) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) (string, error) {
	return "EVENT-OK", nil
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
