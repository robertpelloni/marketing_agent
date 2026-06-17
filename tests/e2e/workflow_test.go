package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
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

func TestFullWorkflow(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	ctx := context.Background()

	// 1. Discovery
	s := scraper.NewScraper(database, []scraper.LeadSource{&scraper.MockJobBoardSource{}})
	s.ExecuteDiscovery(ctx, []string{"AI"})
	deals, _ := database.ListRecentDeals(ctx, 10, 0)
	if len(deals) == 0 { t.Fatal("No deals discovered") }

	// 2. Enrichment
	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, crm.NewMockCRMClient())
	e.ExecuteEnrichment(ctx)

	// 3. Research
	r := researcher.NewResearcher(database, nil, &researcher.DefaultDossierProcessor{}, crm.NewMockCRMClient())
	r.ExecuteResearch(ctx)

	// 4. Communication
	llmProvider := &llm.MockLLMProvider{}
	registry := llm.NewPromptRegistry("data/test_registry.json")
	responder := communication.NewRAGResponseGenerator(database, llmProvider, registry)
	strategy := communication.NewLearningSalesEngine(database, crm.NewMockCRMClient(), llmProvider)
	mgr := communication.NewManager(database, &communication.MockIntentClassifier{}, responder, strategy, nil, nil, nil, nil, registry)
	mgr.ExecutePoll(ctx)

	updatedDeal, _ := database.GetDealByCompanyID(ctx, deals[0].CompanyID)
	if updatedDeal == nil { t.Fatal("Deal missing") }
}
