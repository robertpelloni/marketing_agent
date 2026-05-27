package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	reconcile := flag.Bool("reconcile", false, "Run branch reconciliation and exit")
	flag.Parse()

	if *reconcile {
		log.Println("Running Intelligent Merge Engine...")
		if err := gitres.ReconcileBranches(); err != nil {
			log.Fatalf("Reconciliation failed: %v", err)
		}
		log.Println("Reconciliation complete.")
		return
	}

	log.Println("Starting Enterprise Sales Bot...")

	// 1. Initialize Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local development if not provided
		dbURL = "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// 2. Setup Scraper
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sources := []scraper.LeadSource{
		&scraper.MockJobBoardSource{},
	}
	s := scraper.NewScraper(database, sources)

	// Run scraper in background
	keywords := []string{"AI Engineer", "LLM Orchestration", "Agentic Workflows"}
	go s.Run(ctx, 1*time.Hour, keywords)

	// 2b. Setup Enricher
	enrichmentSources := []enrichment.EnrichmentSource{
		&enrichment.MockApolloSource{},
	}
	e := enrichment.NewEnricher(database, enrichmentSources)

	// Run enricher in background
	go e.Run(ctx, 1*time.Hour)

	// 2c. Setup Researcher
	crawlers := []researcher.Crawler{
		&researcher.GitHubCrawler{},
		&researcher.BlogCrawler{},
	}
	processor := &researcher.DefaultDossierProcessor{}
	r := researcher.NewResearcher(database, crawlers, processor)

	// Run researcher in background
	go r.Run(ctx, 1*time.Hour)

	// 2ca. Setup CRM Integration
	var crmClient crm.CRMClient
	crmBaseURL := os.Getenv("CRM_BASE_URL")
	crmAPIKey := os.Getenv("CRM_API_KEY")

	if crmBaseURL != "" && crmAPIKey != "" {
		log.Println("CRM: Initializing production REST CRM client.")
		crmClient = crm.NewRestCRMClient(crmBaseURL, crmAPIKey)
	} else {
		log.Println("CRM: Initializing mock CRM client (missing configuration).")
		crmClient = crm.NewMockCRMClient()
	}

	crmWorker := crm.NewWorker(database, crmClient)

	// Run CRM sync in background
	go crmWorker.Run(ctx, 30*time.Minute)

	// 2cb. Setup Borg Outreach System
	outreachWorker := agents.NewTargetDiscoveryWorker(database)

	// Run outreach discovery in background
	go outreachWorker.Run(ctx, 2*time.Hour)

	// 2d. Setup Deployer
	ciTracker := &deploy.MockCITracker{}
	deployer := deploy.NewDeployer(ciTracker)

	// 2da. Setup Deployer background sync and monitoring
	syncIntervalStr := os.Getenv("DEPLOY_SYNC_INTERVAL")
	if syncIntervalStr != "" {
		if interval, err := time.ParseDuration(syncIntervalStr); err == nil {
			go deployer.Run(ctx, interval)
			go deployer.MonitorDeployment(ctx, interval)
		} else {
			log.Printf("Warning: Invalid DEPLOY_SYNC_INTERVAL: %v", err)
		}
	}

	// 2e. Setup Communication Manager
	classifier := &communication.MockIntentClassifier{}
	responder := &communication.MockResponseGenerator{}
	strategy := communication.NewLearningSalesEngine(database)
	commManager := communication.NewManager(database, classifier, responder, strategy)

	// Run communication poller in background
	go commManager.Run(ctx, 30*time.Minute)

	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := &autodev.LocalAgent{}
	prManager := &gitcheck.GitHubPRManager{}
	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)

	// Run autodev worker in background (every 1 hour)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database, deployer, ciTracker, taskManager)
	go func() {
		if err := webServer.ListenAndServe(":8080"); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
