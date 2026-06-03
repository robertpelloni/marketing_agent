package db

import (
	"context"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestDatabase_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// 1. Test Company Creation
	company := &Company{
		Name:          "Integration Test Corp",
		Domain:        "itest.io",
		TechStack:     []string{"Go", "Postgres"},
		HiringSignals: []string{"Hiring AI Engineers"},
		MarketCapTier: "Mid-Market",
	}

	err = database.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	if company.ID == 0 {
		t.Error("Expected company ID to be set")
	}

	// 2. Test Deal Creation
	deal := &Deal{
		CompanyID:    company.ID,
		CurrentState: StateDiscovered,
	}

	err = database.CreateDeal(ctx, deal)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// 3. Test Contact Creation
	contact := &Contact{
		CompanyID: company.ID,
		Name:      "Jane Doe",
		Role:      "CTO",
		Email:     "jane@itest.io",
	}

	err = database.CreateContact(ctx, contact)
	if err != nil {
		t.Fatalf("Failed to create contact: %v", err)
	}

	// 4. Test Interaction Creation
	interaction := &Interaction{
		ContactID: contact.ID,
		Channel:   "Email",
		Direction: "Inbound",
		RawText:   "Interested in Borg.",
	}

	err = database.CreateInteraction(ctx, interaction)
	if err != nil {
		t.Fatalf("Failed to create interaction: %v", err)
	}

	// 5. Test Listing and State Transitions
	deals, err := database.ListDealsByState(ctx, StateDiscovered)
	if err != nil {
		t.Fatalf("Failed to list deals: %v", err)
	}

	found := false
	for _, d := range deals {
		if d.ID == deal.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find the created deal in Discovered state")
	}

	err = database.UpdateDealState(ctx, deal.ID, StateResearched)
	if err != nil {
		t.Fatalf("Failed to update deal state: %v", err)
	}
}
