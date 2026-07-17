package communication

import (
	"context"
	"os"
	"testing"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/crm"
	"gitlab.com/robertpelloni/marketing_agent/internal/db"
	"gitlab.com/robertpelloni/marketing_agent/internal/llm"
)

func TestFeedbackLoop_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() { _ = database.Close() }()

	mockCRM := crm.NewMockCRMClient()
	mockLLM := &llm.MockLLMProvider{}
	strategy := NewLearningSalesEngine(database, mockCRM, mockLLM)

	// Setup testing data
	company := &db.Company{
		Name:          "FeedbackLoop Corp",
		Domain:        "feedbackloop.com",
		MarketCapTier: "Mid-Market",
	}
	if err := database.CreateCompany(ctx, company); err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	deal := &db.Deal{
		CompanyID:        company.ID,
		CurrentState:     db.StateEngaged,
		TechnicalDossier: "BOTTLENECK INFRASTRUCTURE",
	}
	if err := database.CreateDeal(ctx, deal); err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	contact := &db.Contact{
		CompanyID: company.ID,
		Name:      "Test Contact",
		Email:     "contact@feedbackloop.com",
	}
	if err := database.CreateContact(ctx, contact); err != nil {
		t.Fatalf("CreateContact failed: %v", err)
	}

	outbound := &db.Interaction{
		ContactID: contact.ID,
		Direction: "Outbound",
		RawText:   "Initial pitch.",
		Success:   false,
	}
	if err := database.CreateInteraction(ctx, outbound); err != nil {
		t.Fatalf("CreateInteraction failed: %v", err)
	}

	inbound := &db.Interaction{
		ContactID: contact.ID,
		Direction: "Inbound",
		RawText:   "Interested.",
		Success:   false,
	}
	if err := database.CreateInteraction(ctx, inbound); err != nil {
		t.Fatalf("CreateInteraction failed: %v", err)
	}

	// 1. Simulate StateClosedWon
	salesCtx := SalesContext{
		Deal:         *deal,
		Company:      *company,
		Contact:      *contact,
		Interactions: []db.Interaction{*outbound, *outbound, *outbound, *outbound, *inbound, *inbound, *inbound}, // To trigger shouldAdvanceState logic
		LatestIntent: IntentFollowUp, // FollowUp + high interactions = advancing to ClosedWon for mid-market
	}

	action, err := strategy.Decide(ctx, salesCtx)
	if err != nil {
		t.Fatalf("Decide failed: %v", err)
	}

	if action != ActionAdvanceState {
		t.Fatalf("Expected ActionAdvanceState, got %v", action)
	}

	time.Sleep(100 * time.Millisecond) // Give CRM go-routine time to finish or let DB apply

	// Verify deal state updated to StateClosedWon
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to fetch updated deal: %v", err)
	}
	if updatedDeal.CurrentState != db.StateClosedWon {
		t.Errorf("Expected StateClosedWon, got %v", updatedDeal.CurrentState)
	}

	// 2. Validate via GDPR export
	gdprData, err := database.ExportGDPRData(ctx, contact.Email)
	if err != nil {
		t.Fatalf("GDPR export failed: %v", err)
	}

	interactions, ok := gdprData["interactions"].([]db.Interaction)
	if !ok || len(interactions) == 0 {
		t.Fatal("Failed to extract interactions from GDPR export")
	}

	hasSuccess := false
	for _, i := range interactions {
		if i.Success {
			hasSuccess = true
			break
		}
	}

	if !hasSuccess {
		t.Error("Expected at least one interaction to be marked success=true after ClosedWon transition")
	}
}
