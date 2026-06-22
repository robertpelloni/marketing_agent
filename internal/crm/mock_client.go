package crm

import (
	"context"
<<<<<<< HEAD
	"log"

=======
	"log/slog"

	"fmt"
>>>>>>> origin/main
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// MockCRMClient simulates an external CRM integration.
<<<<<<< HEAD
type MockCRMClient struct{}
=======
type MockCRMClient struct {
	PushDealCalled        bool
	SyncInteractionCalled bool
	SyncContactsCalled    bool
	GetLeadUpdatesCalled  bool
	LatestNote            string
	UpdatesToReturn       []LeadUpdate
}
>>>>>>> origin/main

// NewMockCRMClient creates a new mock CRM client.
func NewMockCRMClient() *MockCRMClient {
	return &MockCRMClient{}
}

func (m *MockCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
<<<<<<< HEAD
	log.Printf("CRM: Pushing deal %d for company %s (Route: %s) to CRM", deal.ID, company.Name, route)
=======
	slog.Info(fmt.Sprintf("CRM: Pushing deal %d for company %s (Route: %s) to CRM", deal.ID, company.Name, route))
	m.PushDealCalled = true
>>>>>>> origin/main
	return nil
}

func (m *MockCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
<<<<<<< HEAD
	log.Println("CRM: Fetching updates from external CRM...")
	// Simulate a lead being closed in the CRM
=======
	slog.Info("CRM: Fetching updates from external CRM...")
	m.GetLeadUpdatesCalled = true
	if m.UpdatesToReturn != nil {
		return m.UpdatesToReturn, nil
	}
>>>>>>> origin/main
	return []LeadUpdate{}, nil
}

func (m *MockCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
<<<<<<< HEAD
	log.Printf("CRM: Validating account for domain: %s", domain)
=======
	slog.Info(fmt.Sprintf("CRM: Validating account for domain: %s", domain))
>>>>>>> origin/main
	return true, nil
}

func (m *MockCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
<<<<<<< HEAD
	log.Printf("CRM: Syncing interaction for deal %d: %s", dealID, note)
=======
	slog.Info(fmt.Sprintf("CRM: Syncing interaction for deal %d: %s", dealID, note))
	m.SyncInteractionCalled = true
	m.LatestNote = note
	return nil
}

func (m *MockCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	slog.Info(fmt.Sprintf("CRM: Syncing %d contacts for company %d", len(contacts), companyID))
	m.SyncContactsCalled = true
>>>>>>> origin/main
	return nil
}

func (m *MockCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
<<<<<<< HEAD
	log.Printf("CRM: Fetching mock details for deal %d", dealID)
=======
	slog.Info(fmt.Sprintf("CRM: Fetching mock details for deal %d", dealID))
>>>>>>> origin/main
	return &DealDetails{
		ID:                 dealID,
		Status:             db.StateNegotiating,
		QuotedPricing:      10000.0,
		CustomRequirements: "Mock Requirement",
	}, nil
}
