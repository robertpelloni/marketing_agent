package enrichment_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
)

func setupTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" { t.Skip("DATABASE_URL not set") }
	database, err := db.NewDB(dbURL)
	if err != nil { t.Fatalf("Failed to connect: %v", err) }
	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return database
}

func TestEnricher_Integration_EnrichesDiscoveredDeals(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	ctx := context.Background()

	company := &db.Company{Name: "AI Dynamics Corp", Domain: "aidynamics.com", MarketCapTier: "Mid-Market"}
	_ = database.CreateCompany(ctx, company)
	_ = database.CreateDeal(ctx, &db.Deal{CompanyID: company.ID, CurrentState: db.StateDiscovered})

	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, crm.NewMockCRMClient())
	e.ExecuteEnrichment(ctx)

	updatedDeal, _ := database.GetDealByCompanyID(ctx, company.ID)
	if updatedDeal.CurrentState != db.StateResearched {
		t.Errorf("Expected state Researched, got %s", updatedDeal.CurrentState)
	}

	contacts, _ := database.ListContactsByCompany(ctx, company.ID)
	if len(contacts) == 0 { t.Error("Expected contacts") }
}

func TestEnricher_Integration_SkipsNilDB(t *testing.T) {
	e := enrichment.NewEnricher(nil, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, crm.NewMockCRMClient())
	e.ExecuteEnrichment(context.Background())
}
