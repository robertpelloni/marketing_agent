package main
import (
	"context"; "flag"; "fmt"; "log/slog"; "net/http"; "os"; "os/signal"; "syscall"; "time"
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
	_ "github.com/lib/pq"
)
func main() {
	reconcile := flag.Bool("reconcile", false, ""); inventory := flag.Bool("inventory", false, ""); flag.Parse()
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}); slog.SetDefault(slog.New(handler))
	if *inventory { slog.Info("Generating Submodule Inventory"); table, _ := gitcheck.GenerateSubmoduleInventory(); fmt.Println(table); return }
	if *reconcile { slog.Info("Running Intelligent Merge Engine"); if err := gitres.ReconcileBranches(); err != nil { slog.Error("Reconciliation failed", "error", err); os.Exit(1) }; return }
	slog.Info("Starting TormentNexus Autonomous Sales Bot")
	cfg := config.Load(); database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil { slog.Error("Could not connect to database", "error", err); os.Exit(1) }; defer database.Close()
	ctx, cancel := context.WithCancel(context.Background()); defer cancel()
	s := scraper.NewScraper(database, []scraper.LeadSource{&scraper.MockJobBoardSource{}}); go s.Run(ctx, 1*time.Hour, []string{"AI Engineer"})
	var crmClient crm.CRMClient
	switch cfg.CRMProvider {
	case "hubspot": if cfg.CRMAPIKey != "" { crmClient = crm.NewHubSpotCRMClient(cfg.CRMAPIKey) }
	case "salesforce": if cfg.CRMBaseURL != "" { crmClient = crm.NewSalesforceCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey, cfg.SalesforceClientID, cfg.SalesforceClientSecret, cfg.SalesforceAuthURL) }
	default: if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" { crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey) }
	}
	if crmClient == nil { crmClient = crm.NewMockCRMClient() }
	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{&enrichment.MockApolloSource{}}, crmClient); go e.Run(ctx, 1*time.Hour)
	r := researcher.NewResearcher(database, []researcher.Crawler{&researcher.GitHubCrawler{}}, &researcher.DefaultDossierProcessor{}, crmClient); go r.Run(ctx, 1*time.Hour)
	crmWorker := crm.NewWorker(database, crmClient); go crmWorker.Run(ctx, 30*time.Minute)
	outreachWorker := agents.NewTargetDiscoveryWorker(database); go outreachWorker.Run(ctx, 2*time.Hour)
	var ciTracker deploy.CITracker; var dispatcher deploy.WorkflowDispatcher
	if cfg.GitHubRepository != "" { ciTracker = deploy.NewGitHubCITracker("", ""); dispatcher = deploy.NewGitHubDispatcher("", "") }
	if ciTracker == nil { ciTracker = &deploy.MockCITracker{} }
	deployer := deploy.NewDeployer(ciTracker, dispatcher); go deployer.Run(ctx, cfg.DeploySyncInterval); go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)
	commManager := communication.NewManager(database, &communication.MockIntentClassifier{}, communication.NewRAGResponseGenerator(database, &llm.MockLLMProvider{}), communication.NewLearningSalesEngine(database, crmClient, &llm.MockLLMProvider{}), sales.NewOrderProcessor(database, &billing.MockBillingClient{}, crmClient), crmClient); go commManager.Run(ctx, 30*time.Minute)
	orchestrator := autodev.NewOrchestrator(database, autodev.NewTaskManager("TODO.md"), &autodev.LocalAgent{}, &gitcheck.MockPRManager{}, ciTracker); go orchestrator.Run(ctx, 1*time.Hour)
	webServer := web.NewServer(database, deployer, ciTracker, autodev.NewTaskManager("TODO.md"), crmClient, commManager, cfg.CRMProvider); srv := &http.Server{Addr: ":" + cfg.Port, Handler: webServer}
	go func() { if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { slog.Error("Web server error", "error", err) } }()
	sigChan := make(chan os.Signal, 1); signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM); <-sigChan
	slog.Info("Shutting down"); cancel(); shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second); defer shutdownCancel(); srv.Shutdown(shutdownCtx); time.Sleep(2 * time.Second); slog.Info("Done")
}
