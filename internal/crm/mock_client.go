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
	"fmt"
=======
	"fmt"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
>>>>>>> origin/main
)

// MockCRMClient simulates an external CRM integration.
<<<<<<< HEAD
type MockCRMClient struct{}
=======
type MockCRMClient struct {
<<<<<<< HEAD
	PushDealCalled		bool
	SyncInteractionCalled	bool
	SyncContactsCalled	bool
	GetLeadUpdatesCalled	bool
	LatestNote		string
	UpdatesToReturn		[]LeadUpdate
=======
	PushDealCalled        bool
	SyncInteractionCalled bool
	SyncContactsCalled    bool
	GetLeadUpdatesCalled  bool
	LatestNote            string
	UpdatesToReturn       []LeadUpdate
>>>>>>> origin/main
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
>>>>>>> origin/main
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
>>>>>>> origin/main
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
>>>>>>> origin/main
	m.SyncInteractionCalled = true
	m.LatestNote = note
	return nil
}

func (m *MockCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
<<<<<<< HEAD
	log.Printf("CRM: Syncing %d contacts for company %d", len(contacts), companyID)
=======
	slog.Info(fmt.Sprintf("CRM: Syncing %d contacts for company %d", len(contacts), companyID))
>>>>>>> origin/main
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
<<<<<<< HEAD
		ID:			dealID,
		Status:			db.StateNegotiating,
		QuotedPricing:		10000.0,
		CustomRequirements:	"Mock Requirement",
=======
		ID:                 dealID,
		Status:             db.StateNegotiating,
		QuotedPricing:      10000.0,
		CustomRequirements: "Mock Requirement",
>>>>>>> origin/main
	}, nil
}
<<<<<<< HEAD

func (m *MockCRMClient) SendEmail(ctx context.Context, contact db.Contact, subject, body string) error {
	log.Printf("CRM: Simulating email send to %s (Subject: %s)", contact.Email, subject)
	return nil
}

func (m *MockCRMClient) GetNewInteractions(ctx context.Context) ([]db.Interaction, error) {
	log.Println("CRM: Fetching mock interactions from CRM...")
	return []db.Interaction{}, nil
}

func (m *MockCRMClient) SetFieldMapping(mapping FieldMapping) {
	log.Println("CRM: Mock setting field mapping")
}
=======
>>>>>>> origin/main
