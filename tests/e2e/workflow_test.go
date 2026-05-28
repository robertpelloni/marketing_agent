package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
)

func TestEndToEndSalesWorkflow(t *testing.T) {
	// 1. Initial Setup (DB)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping E2E test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 2. Lead Discovery Phase
	sources := []scraper.LeadSource{&scraper.MockJobBoardSource{}}
	s := scraper.NewScraper(database, sources)

	// Discover a lead
	s.ExecuteDiscovery(ctx, []string{"E2E-TEST"})

	// Verify lead exists in DB
	deals, err := database.ListRecentDeals(ctx, 1)
	if err != nil || len(deals) == 0 {
		t.Fatal("Expected a lead to be created in DB")
	}
	deal := deals[0]

	// 2b. Enrichment Phase
	enricher := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}})
	enricher.ExecuteEnrichment(ctx)

	// Verify contact was created
	contacts, err := database.ListContactsByCompany(ctx, deal.CompanyID)
	if err != nil || len(contacts) == 0 {
		t.Fatal("Expected a contact to be created during enrichment")
	}

	// 2c. Research Phase
	res := researcher.NewResearcher(database, []researcher.Crawler{&researcher.GitHubCrawler{}}, &researcher.DefaultDossierProcessor{})
	res.ExecuteResearch(ctx)

	// Verify dossier was compiled
	updatedDeal, _ := database.GetDealByCompanyID(ctx, deal.CompanyID)
	if updatedDeal.TechnicalDossier == "" {
		t.Error("Expected technical dossier to be compiled")
	}

	// 2d. Outreach Phase
	classifier := &communication.MockIntentClassifier{}
	responder := &communication.MockResponseGenerator{}
	strategy := communication.NewLearningSalesEngine(database, nil, nil)
	comm := communication.NewManager(database, classifier, responder, strategy, nil)

	// Simulate inbound pricing inquiry
	reply, err := comm.ProcessInbound(ctx, contacts[0], "How much does Borg cost?")
	if err != nil {
		t.Fatalf("Failed to process inbound: %v", err)
	}
	if reply == "" {
		t.Error("Expected outreach reply")
	}

	// 3. Autonomous Task Generation Phase
	tmpTodo, _ := os.CreateTemp("", "TODO_E2E.md")
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte("- [ ] E2E Task"), 0644)

	manager := autodev.NewTaskManager(tmpTodo.Name())
	agent := &autodev.MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}
	orchestrator := autodev.NewOrchestrator(database, manager, agent, prManager, tracker)

	// Skip sync for E2E
	os.Setenv("SKIP_AUTODEV_SYNC", "true")
	defer os.Unsetenv("SKIP_AUTODEV_SYNC")

	// Execute task lifecycle
	orchestrator.ExecuteStep(ctx)

	// Verify PR was "created" in DB
	prs, err := database.ListActivePullRequests(ctx)
	if err != nil {
		t.Fatalf("Failed to list PRs: %v", err)
	}

	found := false
	for _, pr := range prs {
		if pr.Title == "Autonomous Update: E2E Task" {
			found = true
			break
		}
	}

	if !found {
		t.Error("E2E Task PR not found in database")
	}
}

func TestAutonomousCodeGeneration_Pilot(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping Pilot test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to initialize database for pilot: %v", err)
	}

	// 1. Prepare TODO
	tmpTodo, _ := os.CreateTemp("", "TODO_PILOT.md")
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte("- [ ] Implement autonomous sales-feature"), 0644)

	manager := autodev.NewTaskManager(tmpTodo.Name())
	agent := &autodev.LocalAgent{} // Real LocalAgent for code gen
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}
	orchestrator := autodev.NewOrchestrator(database, manager, agent, prManager, tracker)

	os.Setenv("SKIP_AUTODEV_SYNC", "true")
	os.Setenv("SKIP_AUTODEV_TESTS", "true")
	defer os.Unsetenv("SKIP_AUTODEV_SYNC")
	defer os.Unsetenv("SKIP_AUTODEV_TESTS")

	// 2. Trigger Loop
	orchestrator.ExecuteStep(ctx)

	// 3. Verify file creation
	if _, err := os.Stat("internal/sales/feature.go"); os.IsNotExist(err) {
		t.Errorf("Autonomous code generation failed: internal/sales/feature.go not found")
	}
}
