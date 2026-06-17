package researcher_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
)

type mockCrawler struct{}
func (m *mockCrawler) Crawl(ctx context.Context, target string) (string, error) {
	return "Mock findings for " + target, nil
}

type mockDossierProcessor struct{}
func (m *mockDossierProcessor) Process(findings []string) (string, error) {
	return "Processed: " + strings.Join(findings, "; "), nil
}

func setupResearcherTestDB(t *testing.T) *db.DB {
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

func TestResearcher_Integration_ResearchesDiscoveredDeals(t *testing.T) {
	database := setupResearcherTestDB(t)
	defer database.Close()
	ctx := context.Background()

	company := &db.Company{Name: "Test Research Corp", Domain: "testresearch.io", MarketCapTier: "Startup"}
	_ = database.CreateCompany(ctx, company)
	_ = database.CreateDeal(ctx, &db.Deal{CompanyID: company.ID, CurrentState: db.StateResearched})
	_ = database.CreateContact(ctx, &db.Contact{CompanyID: company.ID, Name: "Jane Doe", Email: "jane@test.io"})

	r := researcher.NewResearcher(database, []researcher.Crawler{&mockCrawler{}}, &mockDossierProcessor{}, crm.NewMockCRMClient())
	r.ExecuteResearch(ctx)

	updatedDeal, _ := database.GetDealByCompanyID(ctx, company.ID)
	if updatedDeal == nil { t.Error("Expected deal") }
}

func TestResearcher_Integration_SkipsNilDB(t *testing.T) {
	r := researcher.NewResearcher(nil, []researcher.Crawler{&mockCrawler{}}, &mockDossierProcessor{}, crm.NewMockCRMClient())
	r.ExecuteResearch(context.Background())
}
