package crm

import (
	"context"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// MockCRMClient simulates an external CRM integration.
type MockCRMClient struct{}

// NewMockCRMClient creates a new mock CRM client.
func NewMockCRMClient() *MockCRMClient {
	return &MockCRMClient{}
}

func (m *MockCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company) error {
	log.Printf("CRM: Pushing deal %d for company %s to CRM", deal.ID, company.Name)
	return nil
}

func (m *MockCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	log.Println("CRM: Fetching updates from external CRM...")
	// Simulate a lead being closed in the CRM
	return []LeadUpdate{}, nil
}

func (m *MockCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	log.Printf("CRM: Validating account for domain: %s", domain)
	return true, nil
}

func (m *MockCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	log.Printf("CRM: Syncing interaction for deal %d: %s", dealID, note)
	return nil
}

func (m *MockCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	log.Printf("CRM: Fetching mock details for deal %d", dealID)
	return &DealDetails{
		ID:                 dealID,
		Status:             db.StateNegotiating,
		QuotedPricing:      10000.0,
		CustomRequirements: "Mock Requirement",
	}, nil
}
