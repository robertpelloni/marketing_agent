package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"os/signal"
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/config"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/sales"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"
	"github.com/robertpelloni/enterprise_sales_bot/internal/mail"
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	reconcile := flag.Bool("reconcile", false, "Run branch reconciliation and exit")
	inventory := flag.Bool("inventory", false, "Generate submodule inventory and exit")
	flag.Parse()

	if *inventory {
		slog.Info("Generating Submodule Inventory...")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			slog.Error("Failed to generate inventory", "error", err)
			os.Exit(1)
		}
		fmt.Println(table)
		return
	}

	if *reconcile {
		slog.Info("Running Intelligent Merge Engine...")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Reconciliation failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Reconciliation complete.")
		return
	}

	slog.Info("Starting TormentNexus Autonomous Sales Bot...")

	// 0. Load Configuration
	cfg := config.Load()

	// 1. Initialize Database
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Could not connect to database", "error", err)
		os.Exit(1)
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

	// 2ca. Setup CRM Integration
	var crmClient crm.CRMClient
	switch cfg.CRMProvider {
	case "hubspot":
		if cfg.CRMAPIKey != "" {
			slog.Info("CRM: Initializing HubSpot CRM client.")
			crmClient = crm.NewHubSpotCRMClient(cfg.CRMAPIKey)
		}
	case "salesforce":
		if cfg.CRMBaseURL != "" {
			slog.Info("CRM: Initializing Salesforce CRM client.")
			crmClient = crm.NewSalesforceCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey, cfg.SalesforceClientID, cfg.SalesforceClientSecret, cfg.SalesforceAuthURL)
		}
	default:
		if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" {
			slog.Info("CRM: Initializing production REST CRM client", "url", cfg.CRMBaseURL)
			crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
		}
	}

	if crmClient == nil {
		slog.Info("CRM: Initializing mock CRM client", "provider", cfg.CRMProvider)
		crmClient = crm.NewMockCRMClient()
	}

	// 2cb. Setup CRM Field Mappings
	crmClient.SetFieldMapping(crm.FieldMapping{
		DealNameProperty:     cfg.CRMDealNameProp,
		DealStageProperty:    cfg.CRMDealStageProp,
		DealAmountProperty:   cfg.CRMDealAmountProp,
		DealDossierProperty:  cfg.CRMDealDossierProp,
		ContactEmailProperty: cfg.CRMContactEmailProp,
	})

	// 2b. Setup Enricher
	enrichmentSources := []enrichment.EnrichmentSource{
		&enrichment.MockApolloSource{},
	}
	e := enrichment.NewEnricher(database, enrichmentSources, crmClient)

	// Run enricher in background
	go e.Run(ctx, 1*time.Hour)

	// 2c. Setup Researcher
	crawlers := []researcher.Crawler{
		&researcher.GitHubCrawler{Client: http.DefaultClient},
		&researcher.BlogCrawler{Client: http.DefaultClient},
	}
	processor := &researcher.DefaultDossierProcessor{}
	r := researcher.NewResearcher(database, crawlers, processor, crmClient)

	// Run researcher in background
	go r.Run(ctx, 1*time.Hour)

	crmWorker := crm.NewWorker(database, crmClient)

	// Run CRM sync in background
	go crmWorker.Run(ctx, 30*time.Minute)

	// 2cb. Setup TormentNexus Outreach System
	outreachWorker := agents.NewTargetDiscoveryWorker(database)

	// Run outreach discovery in background
	go outreachWorker.Run(ctx, 2*time.Hour)

	// 2d. Setup Deployer
	var ciTracker deploy.CITracker
	var dispatcher deploy.WorkflowDispatcher
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			// #nosec G706 -- Repository name is used for context in initialization logs
			slog.Info("CI: Initializing GitHub CI Tracker and Dispatcher", "repo", cfg.GitHubRepository)
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}
	if ciTracker == nil {
		slog.Info("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY).")
		ciTracker = &deploy.MockCITracker{}
	}
	deployer := deploy.NewDeployer(ciTracker, dispatcher)

	// 2da. Setup Deployer background sync and monitoring
	go deployer.Run(ctx, cfg.DeploySyncInterval)
	go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)

	// 2da. Setup LLM Provider
	llmProvider := &llm.MockLLMProvider{}

	// 2db. Setup Email direct sender
	var emailSender mail.EmailSender
	if cfg.SMTPHost != "" {
		slog.Info("SMTP: Initializing SMTP email sender", "host", cfg.SMTPHost)
		emailSender = mail.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	}

	// 2e. Setup Communication Manager
	classifier := &communication.MockIntentClassifier{}
	responder := communication.NewRAGResponseGenerator(database, llmProvider)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)

	// 2ea. Setup Order Processing
	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, crmClient, emailSender)

	// Run communication poller in background
	go commManager.Run(ctx, 30*time.Minute)

	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := &autodev.LocalAgent{}

	var prManager gitcheck.PRManager
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			// #nosec G706 -- Repository name is used for context in initialization logs
			slog.Info("Autodev: Initializing GitHub PR Manager", "repo", cfg.GitHubRepository)
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
		slog.Info("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY).")
		prManager = &gitcheck.MockPRManager{}
	}

	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)

	// Run autodev worker in background (every 1 hour)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database, deployer, ciTracker, taskManager, crmClient, commManager, cfg.CRMProvider)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: webServer,
	}

	go func() {
		slog.Info("Web Dashboard: Listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Web server error", "error", err)
		}
	}()

	// 5. Graceful Shutdown Implementation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down: Signal received, initiating graceful drain...")

	// Cancel background workers via context
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Web server shutdown error", "error", err)
	}

	// Wait for workers to finish
	time.Sleep(2 * time.Second)
	slog.Info("Shutting down: Done.")
}
