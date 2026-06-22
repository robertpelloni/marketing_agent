package main

import (
	"context"
<<<<<<< HEAD
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"os/signal"
=======
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
>>>>>>> origin/main
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
<<<<<<< HEAD
	"github.com/robertpelloni/enterprise_sales_bot/internal/config"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
=======
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
>>>>>>> origin/main
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/sales"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"
<<<<<<< HEAD
	"github.com/robertpelloni/enterprise_sales_bot/internal/mail"
=======
>>>>>>> origin/main
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	reconcile := flag.Bool("reconcile", false, "Run branch reconciliation and exit")
	inventory := flag.Bool("inventory", false, "Generate submodule inventory and exit")
	flag.Parse()

	if *inventory {
<<<<<<< HEAD
		slog.Info("Generating Submodule Inventory...")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			slog.Error("Failed to generate inventory", "error", err)
			os.Exit(1)
=======
		log.Println("Generating Submodule Inventory...")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			log.Fatalf("Failed to generate inventory: %v", err)
>>>>>>> origin/main
		}
		fmt.Println(table)
		return
	}

	if *reconcile {
<<<<<<< HEAD
		slog.Info("Running Intelligent Merge Engine...")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Reconciliation failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Reconciliation complete.")
		return
	}

	slog.Info("Starting TormentNexus Autonomous Sales Bot...")
=======
		log.Println("Running Intelligent Merge Engine...")
		if err := gitres.ReconcileBranches(); err != nil {
			log.Fatalf("Reconciliation failed: %v", err)
		}
		log.Println("Reconciliation complete.")
		return
	}

	log.Println("Starting TormentNexus Autonomous Sales Bot...")
>>>>>>> origin/main

	// 0. Load Configuration
	cfg := config.Load()

	// 1. Initialize Database
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
<<<<<<< HEAD
		slog.Error("Could not connect to database", "error", err)
		os.Exit(1)
=======
		log.Fatalf("Could not connect to database: %v", err)
>>>>>>> origin/main
	}
	defer database.Close()

	// 2. Setup Scraper
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sources := []scraper.LeadSource{
<<<<<<< HEAD
		&scraper.MockJobBoardSource{},
=======
		&scraper.TwitterSource{
			Client:            http.DefaultClient,
			BearerToken:       cfg.TwitterBearerToken,
			APIKey:            cfg.TwitterAPIKey,
			APIKeySecret:      cfg.TwitterAPIKeySecret,
			AccessToken:       cfg.TwitterAccessToken,
			AccessTokenSecret: cfg.TwitterAccessTokenSecret,
		},
		&scraper.LinkedInSource{
			Client: http.DefaultClient,
		},
>>>>>>> origin/main
	}
	s := scraper.NewScraper(database, sources)

	// Run scraper in background
	keywords := []string{"AI Engineer", "LLM Orchestration", "Agentic Workflows"}
	go s.Run(ctx, 1*time.Hour, keywords)

	// 2ca. Setup CRM Integration
	var crmClient crm.CRMClient
<<<<<<< HEAD
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

=======
	if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" {
		log.Printf("CRM: Initializing production REST CRM client at %s", cfg.CRMBaseURL)
		crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
	} else {
		log.Println("CRM: Initializing mock CRM client (missing configuration).")
		crmClient = crm.NewMockCRMClient()
	}

>>>>>>> origin/main
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

<<<<<<< HEAD
=======
	crmWorker := crm.NewWorker(database, crmClient)

	// Run CRM sync in background
	go crmWorker.Run(ctx, 30*time.Minute)

>>>>>>> origin/main
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
<<<<<<< HEAD
			slog.Info("CI: Initializing GitHub CI Tracker and Dispatcher", "repo", cfg.GitHubRepository)
=======
			log.Printf("CI: Initializing GitHub CI Tracker and Dispatcher for %s", cfg.GitHubRepository)
>>>>>>> origin/main
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}
	if ciTracker == nil {
<<<<<<< HEAD
		slog.Info("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY).")
=======
		log.Println("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY).")
>>>>>>> origin/main
		ciTracker = &deploy.MockCITracker{}
	}
	deployer := deploy.NewDeployer(ciTracker, dispatcher)

	// 2da. Setup Deployer background sync and monitoring
	go deployer.Run(ctx, cfg.DeploySyncInterval)
	go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)

	// 2da. Setup LLM Provider
	llmProvider := &llm.MockLLMProvider{}

<<<<<<< HEAD
	// 2db. Setup Email direct sender
	var emailSender mail.EmailSender
	if cfg.SMTPHost != "" {
		slog.Info("SMTP: Initializing SMTP email sender", "host", cfg.SMTPHost)
		emailSender = mail.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	}

	// 2e. Setup Communication Manager
=======
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
>>>>>>> origin/main
	classifier := &communication.MockIntentClassifier{}
	responder := communication.NewRAGResponseGenerator(database, llmProvider)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)

<<<<<<< HEAD
	// 2ea. Setup Order Processing
	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, crmClient, emailSender)
=======
	// 2fa. Setup Order Processing
	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, emailSender)
>>>>>>> origin/main

	// Run communication poller in background
	go commManager.Run(ctx, 30*time.Minute)

<<<<<<< HEAD
	// 2ea. Setup CRM and Email Pollers (Inbound Ingestion)
	crmWorker := crm.NewWorker(database, crmClient, commManager)
	go crmWorker.Run(ctx, 30*time.Minute)

	imapPoller := mail.NewIMAPPoller(database, commManager, cfg.IMAPHost, cfg.IMAPUser, cfg.IMAPPass)
	go imapPoller.Run(ctx, 30*time.Minute)

=======
>>>>>>> origin/main
	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := &autodev.LocalAgent{}

	var prManager gitcheck.PRManager
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			// #nosec G706 -- Repository name is used for context in initialization logs
<<<<<<< HEAD
			slog.Info("Autodev: Initializing GitHub PR Manager", "repo", cfg.GitHubRepository)
=======
			log.Printf("Autodev: Initializing GitHub PR Manager for %s", cfg.GitHubRepository)
>>>>>>> origin/main
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
<<<<<<< HEAD
		slog.Info("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY).")
=======
		log.Println("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY).")
>>>>>>> origin/main
		prManager = &gitcheck.MockPRManager{}
	}

	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)

	// Run autodev worker in background (every 1 hour)
	go orchestrator.Run(ctx, 1*time.Hour)

<<<<<<< HEAD
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
=======
	// 3a. Start Autonomous Blog Generator (daily)
	blogGen := contentgen.NewBlogGenerator(llmProvider, database)
	go blogGen.Run(ctx, 3*time.Minute)

	// 3b. Start Stats API Server (port 8086, no auth)
	go func() {
		statsMux := http.NewServeMux()
		statsMux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "https://tormentnexus.site")
			ctx := r.Context()
			companies, _ := database.CountCompanies(ctx)
			contacts, _ := database.CountContacts(ctx)
			interactions, _ := database.CountInteractions(ctx)
			stateCounts := make(map[string]int)
			states, _ := database.CountDealsByState(ctx)
			for _, st := range states {
				stateCounts[string(st.State)] = st.Count
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"companies": companies, "contacts": contacts,
				"interactions": interactions, "deals": stateCounts,
				"status": "operational",
			})
		})
		statsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "OK") })
		log.Printf("Stats API: Listening on :8086")
		if err := http.ListenAndServe(":8086", statsMux); err != nil {
			log.Printf("Stats API error: %v", err)
		}
	}()

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
>>>>>>> origin/main
		}
	}()

	// 5. Graceful Shutdown Implementation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

<<<<<<< HEAD
	slog.Info("Shutting down: Signal received, initiating graceful drain...")
=======
	log.Println("Shutting down: Signal received, initiating graceful drain...")
>>>>>>> origin/main

	// Cancel background workers via context
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
<<<<<<< HEAD
		slog.Error("Web server shutdown error", "error", err)
=======
		log.Printf("Web server shutdown error: %v", err)
>>>>>>> origin/main
	}

	// Wait for workers to finish
	time.Sleep(2 * time.Second)
<<<<<<< HEAD
	slog.Info("Shutting down: Done.")
=======
	log.Println("Shutting down: Done.")
>>>>>>> origin/main
}
