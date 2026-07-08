package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"os"

	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/stripe/stripe-go/v81"
)

func TestHandleStripeWebhook_CheckoutSessionCompleted(t *testing.T) {
	ctx := context.Background()

	// Initialize test DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" { t.Skip("no database url") }
	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test, DB not available: %v", err)
	}

	err = database.RunMigrations(ctx)
	if err != nil {
		t.Skipf("Skipping integration test, migrations failed: %v", err)
	}
	defer database.Close()

	// Set up test data
	company := &db.Company{
		Name:   "Stripe Webhook Test Corp",
		Domain: "stripe-webhook-test.io",
	}
	if err := database.CreateCompany(ctx, company); err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateNegotiating, // Set to negotiating initially
		QuotedPricing: 0,
	}
	if err := database.CreateDeal(ctx, deal); err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Create test server

	server := NewServer(database, nil, nil, nil, nil)

	// Create mock webhook payload
	session := stripe.CheckoutSession{
		ID:          "cs_test_12345",
		AmountTotal: 9900, // $99.00
		Metadata: map[string]string{
			"deal_id": strconv.FormatInt(deal.ID, 10),
		},
	}

	sessionBytes, _ := json.Marshal(session)

	event := stripe.Event{
		Type: "checkout.session.completed",
		Data: &stripe.EventData{
			Raw: json.RawMessage(sessionBytes),
		},
	}

	payloadBytes, _ := json.Marshal(event)

	// Execute request
	req, _ := http.NewRequest("POST", "/api/v1/webhook/stripe", bytes.NewBuffer(payloadBytes))
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	// Verify database changes
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get updated deal: %v", err)
	}

	if updatedDeal.CurrentState != db.StateClosedWon {
		t.Errorf("Expected deal state to be Closed_Won, got %s", updatedDeal.CurrentState)
	}

	if updatedDeal.QuotedPricing != 99.0 {
		t.Errorf("Expected deal pricing to be 99.0, got %f", updatedDeal.QuotedPricing)
	}

    // Check audit logs
    logs, err := database.ListRecentAuditLogs(ctx, 10)
    if err != nil {
        t.Fatalf("Failed to get audit logs: %v", err)
    }

    found := false
    for _, log := range logs {
        if log.EntityID == deal.ID && log.Type == "deal_transition" && log.Action == string(db.StateClosedWon) {
            found = true
            break
        }
    }
    if !found {
        t.Errorf("Expected audit log for deal transition not found")
    }
}
