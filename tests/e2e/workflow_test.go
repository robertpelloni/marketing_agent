package e2e

import (
	"context"
<<<<<<< HEAD
	"os"
=======
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
>>>>>>> origin/main
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
<<<<<<< HEAD
=======
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
>>>>>>> origin/main
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
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
<<<<<<< HEAD
	enricher := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}})
=======
	// For production verification, we use a mock CRM server and the real RestCRMClient
	// to test the HTTP integration layer.
	mux := http.NewServeMux()
	var mu sync.Mutex
	calls := make(map[string]int)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls[r.URL.Path]++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/updates" {
			w.Write([]byte("[]"))
		}
	})
	crmServer := httptest.NewServer(mux)
	defer crmServer.Close()

	realCRM := crm.NewRestCRMClient(crmServer.URL, "e2e-token")

	enricher := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, realCRM)
>>>>>>> origin/main
	enricher.ExecuteEnrichment(ctx)

	// Verify contact was created
	contacts, err := database.ListContactsByCompany(ctx, deal.CompanyID)
	if err != nil || len(contacts) == 0 {
		t.Fatal("Expected a contact to be created during enrichment")
	}

	// 2c. Research Phase
<<<<<<< HEAD
	res := researcher.NewResearcher(database, []researcher.Crawler{&researcher.GitHubCrawler{}}, &researcher.DefaultDossierProcessor{})
=======
	res := researcher.NewResearcher(database, []researcher.Crawler{&researcher.GitHubCrawler{}}, &researcher.DefaultDossierProcessor{}, realCRM)
>>>>>>> origin/main
	res.ExecuteResearch(ctx)

	// Verify dossier was compiled
	updatedDeal, _ := database.GetDealByCompanyID(ctx, deal.CompanyID)
	if updatedDeal.TechnicalDossier == "" {
		t.Error("Expected technical dossier to be compiled")
	}

	// 2d. Outreach Phase
	classifier := &communication.MockIntentClassifier{}
<<<<<<< HEAD
	responder := communication.NewRAGResponseGenerator(&llm.MockLLMProvider{})
	strategy := communication.NewLearningSalesEngine(database, nil, nil)
	comm := communication.NewManager(database, classifier, responder, strategy, nil)

	// Simulate inbound pricing inquiry
	reply, err := comm.ProcessInbound(ctx, contacts[0], "How much does Borg cost?")
=======
	responder := communication.NewRAGResponseGenerator(database, &llm.MockLLMProvider{})
	strategy := communication.NewLearningSalesEngine(database, realCRM, nil)
	comm := communication.NewManager(database, classifier, responder, strategy, nil, realCRM, nil)

	// Simulate inbound pricing inquiry
	reply, err := comm.ProcessInbound(ctx, contacts[0], "How much does TormentNexus cost?")
>>>>>>> origin/main
	if err != nil {
		t.Fatalf("Failed to process inbound: %v", err)
	}
	if reply == "" {
		t.Error("Expected outreach reply")
	}

	// 2e. Negotiation & Closing Phase
	// Simulate positive intent after outreach
<<<<<<< HEAD
	reply, err = comm.ProcessInbound(ctx, contacts[0], "This looks interesting, let's proceed with a proposal.")
=======
	_, err = comm.ProcessInbound(ctx, contacts[0], "This looks interesting, let's proceed with a proposal.")
>>>>>>> origin/main
	if err != nil {
		t.Fatalf("Failed to process follow-up: %v", err)
	}

	// Verify deal advanced (QualifyLead should be high)
	wonDeal, _ := database.GetDealByCompanyID(ctx, deal.CompanyID)
	if wonDeal.CurrentState != db.StateClosedWon {
		t.Errorf("Expected deal to be Closed_Won, got %s", wonDeal.CurrentState)
	}

<<<<<<< HEAD
	// 3. Autonomous Task Generation Phase
	tmpTodo, _ := os.CreateTemp("", "TODO_E2E.md")
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte("- [ ] E2E Task"), 0644)
=======
	// Verify CRM synchronization occurred via HTTP
	time.Sleep(100 * time.Millisecond) // Wait for async retries/pushes
	mu.Lock()
	if calls["/deals"] == 0 {
		t.Error("Expected CRM PushDeal HTTP call")
	}
	if calls[fmt.Sprintf("/companies/%d/contacts", deal.CompanyID)] == 0 {
		t.Error("Expected CRM SyncContacts HTTP call")
	}
	mu.Unlock()

	// 3. Autonomous Task Generation Phase
	tmpTodo, err := os.CreateTemp("", "TODO_E2E.md")
	if err != nil {
		t.Fatalf("Failed to create E2E TODO: %v", err)
	}
	defer os.Remove(tmpTodo.Name())
	if err := os.WriteFile(tmpTodo.Name(), []byte("- [ ] E2E Task"), 0644); err != nil {
		t.Fatalf("Failed to write E2E TODO: %v", err)
	}
>>>>>>> origin/main

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
<<<<<<< HEAD
	tmpTodo, _ := os.CreateTemp("", "TODO_PILOT.md")
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte("- [ ] Implement autonomous sales-feature"), 0644)
=======
	tmpTodo, err := os.CreateTemp("", "TODO_PILOT.md")
	if err != nil {
		t.Fatalf("Failed to create pilot TODO: %v", err)
	}
	defer os.Remove(tmpTodo.Name())
	if err := os.WriteFile(tmpTodo.Name(), []byte("- [ ] Implement autonomous sales-feature"), 0644); err != nil {
		t.Fatalf("Failed to write pilot TODO: %v", err)
	}
>>>>>>> origin/main

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
<<<<<<< HEAD
=======

func TestCRMReconciliationWorkflow(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping CRM reconciliation test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// 1. Setup a lead in Negotiating state
	company := &db.Company{Name: "Reconciliation Corp", Domain: "recon.io"}
	database.CreateCompany(ctx, company)
	deal := &db.Deal{CompanyID: company.ID, CurrentState: db.StateNegotiating}
	database.CreateDeal(ctx, deal)

	// 2. Setup mock CRM server to provide updates
	mux := http.NewServeMux()
	mux.HandleFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		// Mock a state change from Negotiating to Won
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"ID": "%d", "NewState": "Closed_Won", "Notes": "Closed via external CRM portal"}]`, deal.ID)
	})
	mux.HandleFunc(fmt.Sprintf("/deals/%d", deal.ID), func(w http.ResponseWriter, r *http.Request) {
		// Mock updated deal details
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": %d, "status": "Closed_Won", "quoted_pricing": 15000.0, "custom_requirements": "SLA upgrade confirmed" }`, deal.ID)
	})
	// Allow PushDeal to succeed
	mux.HandleFunc("/deals", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	crmServer := httptest.NewServer(mux)
	defer crmServer.Close()

	realCRM := crm.NewRestCRMClient(crmServer.URL, "recon-token")
	worker := crm.NewWorker(database, realCRM, nil)

	// 3. Trigger reconciliation
	// crm.Worker.sync is internal, but we can call Run once or simulate it.
	// For this test, we verify that the worker logic is functional.
	// We'll use a wrapper or just call the sync logic if it was exported.
	// Since sync is private, we simulate the logic or use the background loop.

	// Execute a single sync cycle (we'll need to export it or use a test helper)
	// For now, let's just trigger the background worker briefly.
	go worker.Run(ctx, 100*time.Millisecond)

	// Wait for reconciliation to occur
	time.Sleep(500 * time.Millisecond)

	// 4. Verify local state matches CRM updates
	updatedDeal, _ := database.GetDealByCompanyID(ctx, company.ID)
	if updatedDeal.CurrentState != db.StateClosedWon {
		t.Errorf("Reconciliation failed: expected state Closed_Won, got %s", updatedDeal.CurrentState)
	}
	if updatedDeal.QuotedPricing != 15000.0 {
		t.Errorf("Reconciliation failed: expected pricing 15000.0, got %f", updatedDeal.QuotedPricing)
	}
}
>>>>>>> origin/main
