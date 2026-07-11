package billing_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/robertpelloni/marketing_agent/internal/billing"
	"github.com/robertpelloni/marketing_agent/internal/db"
)

func TestBilling_Integration_SubscriptionCreationFlow(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping test due to db err: %v", err)
	}

	if err := database.RunMigrations(ctx); err != nil {
		_ = database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}
	defer func() { _ = database.Close() }()

	store := billing.NewDBAdapter(database.Conn)

	// Create test company
	company := &db.Company{
		Name:          "Billing Test Corp",
		Domain:        "billingtest.io",
		MarketCapTier: "Mid-Market",
	}
	if err := database.CreateCompany(ctx, company); err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	// Test mapping subscription via Store DB Adapter directly
	// Simulating the internal save executed by handleCheckoutCompleted
	stripeSubID := "sub_12345_test"
	stripeCustID := "cus_12345_test"
	rate := 99.0
	seats := 1

	subInfo, err := store.CreateSubscription(ctx, company.ID, billing.TierProfessional, stripeSubID, stripeCustID, rate, seats, nil)
	if err != nil {
		t.Fatalf("Failed to create subscription via store: %v", err)
	}

	if subInfo == nil {
		t.Fatal("Expected subInfo to not be nil")
	}

	// Verify retrieval by Stripe ID
	fetchedSub, err := store.GetSubscriptionByStripeID(ctx, stripeSubID)
	if err != nil {
		t.Fatalf("Failed to get subscription by stripe ID: %v", err)
	}
	if fetchedSub.CompanyID != company.ID {
		t.Errorf("Expected CompanyID %d, got %d", company.ID, fetchedSub.CompanyID)
	}

	// Verify retrieval by Company ID
	fetchedByCompany, err := store.GetSubscriptionByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get subscription by company ID: %v", err)
	}
	if fetchedByCompany.StripeSubID != stripeSubID {
		t.Errorf("Expected StripeSubID %s, got %s", stripeSubID, fetchedByCompany.StripeSubID)
	}

	// Verify State Update
	err = store.UpdateSubscriptionState(ctx, stripeSubID, "active")
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	fetchedActive, _ := store.GetSubscriptionByStripeID(ctx, stripeSubID)
	if fetchedActive.State != "active" {
		t.Errorf("Expected state 'active', got %s", fetchedActive.State)
	}

	// Verify Period Update
	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)
	err = store.UpdateSubscriptionPeriod(ctx, stripeSubID, now, nextMonth)
	if err != nil {
		t.Fatalf("Failed to update period: %v", err)
	}

	// Verify Cancellation
	err = store.CancelSubscription(ctx, stripeSubID, nextMonth)
	if err != nil {
		t.Fatalf("Failed to cancel subscription: %v", err)
	}
}

func TestBilling_Integration_GrandfatheredRates(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping test due to db err: %v", err)
	}
	store := billing.NewDBAdapter(database.Conn)
	defer func() { _ = database.Close() }()

	company := &db.Company{
		Name:          "Rate Test Corp",
		Domain:        "ratetest.io",
		MarketCapTier: "Startup",
	}
	_ = database.CreateCompany(ctx, company)

	stripeSubID := "sub_rates_test"
	store.CreateSubscription(ctx, company.ID, billing.TierProfessional, stripeSubID, "cus_rate_test", 50.0, 1, nil)

	// Set grandfathered rate
	err = store.SetGrandfatheredRate(ctx, stripeSubID, 50.0)
	if err != nil {
		t.Fatalf("Failed to set grandfathered rate: %v", err)
	}

	// Verify grandfathered rate
	sub, _ := store.GetSubscriptionByStripeID(ctx, stripeSubID)
	if sub.GrandfatheredRate == nil {
		t.Fatal("Expected grandfathered rate to be set, got nil")
	}
	if *sub.GrandfatheredRate != 50.0 {
		t.Errorf("Expected grandfathered rate 50.0, got %f", *sub.GrandfatheredRate)
	}

	// Record price change
	err = store.RecordPriceChange(ctx, sub.ID, 50.0, 99.0)
	if err != nil {
		t.Fatalf("Failed to record price change: %v", err)
	}
}
