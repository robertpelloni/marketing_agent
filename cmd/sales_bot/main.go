package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/config"
	"github.com/robertpelloni/enterprise_sales_bot/internal/contentgen"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/sales"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	reconcile := flag.Bool("reconcile", false, "Run branch reconciliation and exit")
	inventory := flag.Bool("inventory", false, "Generate submodule inventory and exit")
	flag.Parse()

	if *inventory {
		log.Println("Generating Submodule Inventory...")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			log.Fatalf("Failed to generate inventory: %v", err)
		}
		fmt.Println(table)
		return
	}

	if *reconcile {
		log.Println("Running Intelligent Merge Engine...")
		if err := gitres.ReconcileBranches(); err != nil {
			log.Fatalf("Reconciliation failed: %v", err)
		}
		log.Println("Reconciliation complete.")
		return
	}

	log.Println("Starting TormentNexus Autonomous Sales Bot...")

	// 0. Load Configuration
	cfg := config.Load()

	// 1. Initialize Database
	database, err := db.NewDB(cfg.DatabaseURL)
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

	// 2ca. Setup CRM Integration
	var crmClient crm.CRMClient
	if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" {
		log.Printf("CRM: Initializing production REST CRM client at %s", cfg.CRMBaseURL)
		crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
	} else {
		log.Println("CRM: Initializing mock CRM client (missing configuration).")
		crmClient = crm.NewMockCRMClient()
	}

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
			log.Printf("CI: Initializing GitHub CI Tracker and Dispatcher for %s", cfg.GitHubRepository)
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}
	if ciTracker == nil {
		log.Println("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY).")
		ciTracker = &deploy.MockCITracker{}
	}
	deployer := deploy.NewDeployer(ciTracker, dispatcher)

	// 2da. Setup Deployer background sync and monitoring
	go deployer.Run(ctx, cfg.DeploySyncInterval)
	go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)

	// 2da. Setup LLM Provider
	llmProvider := &llm.MockLLMProvider{}

	// 2e. Setup Email Sender — SMTP, IMAP Drafts, or Mock
	var emailSender communication.EmailSender
	if cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" && !cfg.DryRun {
		log.Printf("Email: Initializing SMTP sender via %s:%d as %s", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername)
		emailSender = communication.NewSMTPSender(communication.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
	} else if cfg.DryRun && cfg.IMAPHost != "" && cfg.IMAPUsername != "" && cfg.IMAPPassword != "" {
		log.Printf("Email: DRY RUN mode — saving drafts to %s via IMAP.", cfg.IMAPFolder)
		emailSender = communication.NewDraftSender(cfg.IMAPHost, cfg.IMAPPort, cfg.IMAPUsername, cfg.IMAPPassword)
	} else {
		log.Println("Email: No email sender configured — using MockEmailSender.")
		emailSender = &communication.MockEmailSender{}
	}

	// 2f. Setup Communication Manager
	classifier := &communication.MockIntentClassifier{}
	responder := communication.NewRAGResponseGenerator(database, llmProvider)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)

	// 2fa. Setup Order Processing
	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, emailSender)

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
			log.Printf("Autodev: Initializing GitHub PR Manager for %s", cfg.GitHubRepository)
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
		log.Println("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY).")
		prManager = &gitcheck.MockPRManager{}
	}

	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)

	// Run autodev worker in background (every 1 hour)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 3a. Start Autonomous Blog Generator (daily)
	blogGen := contentgen.NewBlogGenerator(llmProvider, database)
	go blogGen.Run(ctx, 24*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database, deployer, ciTracker, taskManager, llmProvider)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           webServer,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		log.Printf("Web Dashboard: Listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Web server error: %v", err)
		}
	}()

	// 5. Graceful Shutdown Implementation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down: Signal received, initiating graceful drain...")

	// Cancel background workers via context
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Web server shutdown error: %v", err)
	}

	// Wait for workers to finish
	time.Sleep(2 * time.Second)
	log.Println("Shutting down: Done.")
}
