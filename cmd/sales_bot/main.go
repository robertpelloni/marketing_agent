package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
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
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// 0a. Recovery & Heartbeat (System Robustness)
	log.Println("HEARTBEAT: TormentNexus Autonomous Sales Bot starting initialization...")
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "CRITICAL PANIC RECOVERED: %v\n", r)
			fmt.Fprintf(os.Stderr, "STACK TRACE:\n%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

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
		log.Printf("WARNING: Could not connect to database: %v", err)
	} else {
		defer database.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Setup Scraper — HN "Who is Hiring" + LinkedIn + GitHub Issues + Mock fallback
	sources := []scraper.LeadSource{
		&scraper.HNWhoIsHiringSource{Client: http.DefaultClient},
		&scraper.LinkedInSource{Client: http.DefaultClient},
		&scraper.GitHubIssueSource{Client: http.DefaultClient, Token: cfg.GitHubToken},
		&scraper.MockJobBoardSource{},
	}
	s := scraper.NewScraper(database, sources)

	keywords := []string{"AI Engineer", "LLM Orchestration", "Agentic Workflows", "AI Platform", "ML Infrastructure"}
	go s.Run(ctx, 2*time.Hour, keywords)

	// 2a. Setup CRM Integration
	var crmClient crm.CRMClient
	if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" {
		log.Printf("CRM: Initializing production REST CRM client at %s", cfg.CRMBaseURL)
		crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
	} else {
		log.Println("CRM: Initializing mock CRM client (missing configuration).")
		crmClient = crm.NewMockCRMClient()
	}

	// 2b. Setup Enricher — Hunter.io + Apollo.io + Mock with FallbackSource
	var enrichmentSources []enrichment.EnrichmentSource
	var sourceNames []string

	if cfg.HunterAPIKey != "" {
		log.Println("Enrichment: Initializing Hunter.io source.")
		enrichmentSources = append(enrichmentSources, enrichment.NewHunterSource(cfg.HunterAPIKey))
		sourceNames = append(sourceNames, "Hunter.io")
	}

	if cfg.ApolloAPIKey != "" {
		log.Println("Enrichment: Initializing Apollo.io source.")
		enrichmentSources = append(enrichmentSources, enrichment.NewApolloSource(cfg.ApolloAPIKey))
		sourceNames = append(sourceNames, "Apollo.io")
	}

	if len(enrichmentSources) == 0 {
		enrichmentSources = append(enrichmentSources, &enrichment.MockApolloSource{})
		sourceNames = append(sourceNames, "Mock")
	} else {
		enrichmentSources = append(enrichmentSources, &enrichment.MockApolloSource{})
		sourceNames = append(sourceNames, "Mock (fallback)")
	}

	fallbackSource := enrichment.NewFallbackSource(enrichmentSources, sourceNames)
	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{fallbackSource}, crmClient)
	go e.Run(ctx, 1*time.Hour)

	// 2c. Setup Researcher
	crawlers := []researcher.Crawler{
		&researcher.GitHubCrawler{Client: http.DefaultClient},
		&researcher.BlogCrawler{Client: http.DefaultClient},
	}
	processor := &researcher.DefaultDossierProcessor{}
	r := researcher.NewResearcher(database, crawlers, processor, crmClient)
	go r.Run(ctx, 1*time.Hour)

	crmWorker := crm.NewWorker(database, crmClient)
	go crmWorker.Run(ctx, 30*time.Minute)

	// 2d. Setup Target Discovery
	outreachWorker := agents.NewTargetDiscoveryWorker(database)
	go outreachWorker.Run(ctx, 2*time.Hour)

	// 2k. Setup Blog Intelligence
	blogWorker := scraper.NewBlogWorker(database)
	go blogWorker.Run(ctx, 4*time.Hour)

	// 2e. Setup Deployer
	var ciTracker deploy.CITracker
	var dispatcher deploy.WorkflowDispatcher

	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			log.Printf("CI: Initializing GitHub CI Tracker and Dispatcher for %s", cfg.GitHubRepository)
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}

	if ciTracker == nil {
		ciTracker = &deploy.MockCITracker{}
	}

	deployer := deploy.NewDeployer(ciTracker, dispatcher)
	go deployer.Run(ctx, cfg.DeploySyncInterval)
	go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)

	// 2f. Setup LLM Provider — Hermes or Mock
	var llmProvider llm.LLMProvider
	var promptRegistry *llm.PromptRegistry

	if cfg.HermesAPIURL != "" && cfg.HermesAPIKey != "" {
		llmProvider = llm.NewHermesLLMProvider(llm.HermesConfig{
			BaseURL: cfg.HermesAPIURL,
			APIKey:  cfg.HermesAPIKey,
			Model:   cfg.HermesModel,
		})

		promptRegistry = llm.NewPromptRegistry("data/prompt_registry.json")
		promptRegistry.RegisterVersion("outreach-reply", "Intent: ${intent}. Dossier: ${dossier}. Company: ${company}. ${negative} Generate a professional reply.")

		if err := llmProvider.(*llm.HermesLLMProvider).HealthCheck(ctx); err != nil {
			log.Printf("LLM: WARNING — Hermes health check failed: %v", err)
		}
	} else {
		llmProvider = &llm.MockLLMProvider{}
	}

	// 2g. Setup Intent Classifier
	var classifier communication.IntentClassifier
	if cfg.HermesAPIURL != "" && cfg.HermesAPIKey != "" {
		classifier = communication.NewLLMIntentClassifier(llmProvider)
	} else {
		classifier = &communication.MockIntentClassifier{}
	}

	responder := communication.NewRAGResponseGenerator(database, llmProvider, promptRegistry)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)

	// 2h. Setup Email Sender
	var emailSender communication.EmailSender
	if cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" && !cfg.DryRun {
		emailSender = communication.NewSMTPSender(communication.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
	} else {
		emailSender = &communication.MockEmailSender{}
	}

	ghSender := communication.NewGitHubSender()
	liSender := communication.NewLinkedInSender()

	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, emailSender, ghSender, liSender, promptRegistry)
	go commManager.Run(ctx, 30*time.Minute)

	// 2j. Setup Cadence-aware outreach scheduling
	cadenceManager := communication.NewCadenceAwareManager(commManager, database)
	go cadenceManager.RunCadence(ctx, 12*time.Hour)

	// 2i. Setup IMAP Email Receiver
	if cfg.IMAPHost != "" && cfg.IMAPUsername != "" && cfg.IMAPPassword != "" {
		imapReceiver := communication.NewEmailReceiver(communication.IMAPConfig{
			Host:     cfg.IMAPHost,
			Port:     cfg.IMAPPort,
			Username: cfg.IMAPUsername,
			Password: cfg.IMAPPassword,
			Folder:   cfg.IMAPFolder,
		}, commManager)
		go imapReceiver.Run(ctx, cfg.IMAPPollInterval)
	}

	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := autodev.NewLocalAgent(llmProvider)

	var prManager gitcheck.PRManager
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
		prManager = &gitcheck.MockPRManager{}
	}

	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database, deployer, ciTracker, taskManager, llmProvider, promptRegistry)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           webServer,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Web Dashboard: Listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Web server error: %v", err)
		}
	}()

	// 5. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down: Signal received, initiating graceful drain...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Web server shutdown error: %v", err)
	}
	time.Sleep(2 * time.Second)
	log.Println("Shutting down: Done.")
}
