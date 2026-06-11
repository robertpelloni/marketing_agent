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
	"sync"
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
		slog.Info("Generating Submodule Inventory")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			slog.Error("Failed to generate inventory", "error", err)
			os.Exit(1)
		}
		fmt.Println(table)
		return
	}

	if *reconcile {
		slog.Info("Running Intelligent Merge Engine")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Reconciliation failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Reconciliation complete")
		return
	}

	slog.Info("Starting TormentNexus Autonomous Sales Bot")

	// 0. Load Configuration
	cfg := config.Load()

	// 1. Initialize Database
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Could not connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// 2. Setup Context and WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 3. Setup Providers and Clients
	var crmClient crm.CRMClient
	switch cfg.CRMProvider {
	case "hubspot":
		if cfg.CRMAPIKey != "" {
			slog.Info("CRM: Initializing HubSpot CRM client")
			crmClient = crm.NewHubSpotCRMClient(cfg.CRMAPIKey)
		}
	case "salesforce":
		if cfg.CRMBaseURL != "" {
			slog.Info("CRM: Initializing Salesforce CRM client")
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

	llmProvider := &llm.MockLLMProvider{}
	billingClient := &billing.MockBillingClient{}

	var emailSender mail.EmailSender
	if cfg.SMTPHost != "" {
		slog.Info("SMTP: Initializing SMTP email sender", "host", cfg.SMTPHost)
		emailSender = mail.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	}

	// 4. Setup Core Modules
	classifier := &communication.MockIntentClassifier{}
	responder := communication.NewRAGResponseGenerator(database, llmProvider)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)
	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, crmClient, emailSender)

	// 5. Setup Workers
	s := scraper.NewScraper(database, []scraper.LeadSource{&scraper.MockJobBoardSource{}})
	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, crmClient)
	r := researcher.NewResearcher(database, []researcher.Crawler{
		&researcher.GitHubCrawler{Client: http.DefaultClient},
		&researcher.BlogCrawler{Client: http.DefaultClient},
	}, &researcher.DefaultDossierProcessor{}, crmClient)
	crmWorker := crm.NewWorker(database, crmClient, commManager)
	outreachWorker := agents.NewTargetDiscoveryWorker(database)

	var ciTracker deploy.CITracker
	var dispatcher deploy.WorkflowDispatcher
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			slog.Info("CI: Initializing GitHub CI Tracker and Dispatcher", "repository", cfg.GitHubRepository)
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}
	if ciTracker == nil {
		slog.Info("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY)")
		ciTracker = &deploy.MockCITracker{}
	}
	deployer := deploy.NewDeployer(ciTracker, dispatcher)

	taskManager := autodev.NewTaskManager("TODO.md")
	agent := &autodev.LocalAgent{}
	var prManager gitcheck.PRManager
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			slog.Info("Autodev: Initializing GitHub PR Manager", "repository", cfg.GitHubRepository)
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
		slog.Info("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY)")
		prManager = &gitcheck.MockPRManager{}
	}
	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)

	// 6. Run Workers in Background
	wg.Add(7)
	go func() { defer wg.Done(); s.Run(ctx, 1*time.Hour, []string{"AI Engineer", "LLM Orchestration"}) }()
	go func() { defer wg.Done(); e.Run(ctx, 1*time.Hour) }()
	go func() { defer wg.Done(); r.Run(ctx, 1*time.Hour) }()
	go func() { defer wg.Done(); crmWorker.Run(ctx, 30*time.Minute) }()
	go func() { defer wg.Done(); commManager.Run(ctx, 30*time.Minute) }()
	go func() { defer wg.Done(); outreachWorker.Run(ctx, 2*time.Hour) }()
	go func() { defer wg.Done(); orchestrator.Run(ctx, 1*time.Hour) }()

	wg.Add(2)
	go func() { defer wg.Done(); deployer.Run(ctx, cfg.DeploySyncInterval) }()
	go func() { defer wg.Done(); deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval) }()

	// 7. Start Web Server
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

	// 8. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down: Signal received, initiating graceful drain")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Web server shutdown error", "error", err)
	}

	wg.Wait()
	slog.Info("Shutting down: Done")
}
