package agents

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"log/slog"
>>>>>>> origin/main
	"os"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
<<<<<<< HEAD
=======
	"github.com/robertpelloni/enterprise_sales_bot/internal/metrics"
>>>>>>> origin/main
)

// TargetDiscoveryWorker scans for new opportunities (e.g., GitHub, MCP servers).
type TargetDiscoveryWorker struct {
	db *db.DB
}

// NewTargetDiscoveryWorker creates a new discovery worker.
func NewTargetDiscoveryWorker(database *db.DB) *TargetDiscoveryWorker {
	return &TargetDiscoveryWorker{db: database}
}

// Run starts the target discovery background loop.
func (w *TargetDiscoveryWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	log.Printf("TormentNexus Outreach: Target discovery worker started (interval: %v)...", interval)
=======
	slog.Info("TormentNexus Outreach: Target discovery worker started", "interval", interval)
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			log.Println("TormentNexus Outreach: Target discovery worker stopping...")
=======
			slog.Info("TormentNexus Outreach: Target discovery worker stopping")
>>>>>>> origin/main
			return
		case <-ticker.C:
			w.discover(ctx)
		}
	}
}

func (w *TargetDiscoveryWorker) discover(ctx context.Context) {
<<<<<<< HEAD
	log.Println("TormentNexus Outreach: Scanning for new MCP server repositories on GitHub...")
=======
	slog.Info("TormentNexus Outreach: Scanning for new MCP server repositories on GitHub")
>>>>>>> origin/main

	client := github.NewClient(nil)
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		client = github.NewClient(nil).WithAuthToken(token)
	}

	query := "model-context-protocol OR mcp-server language:Go language:TypeScript"
	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	result, _, err := client.Search.Repositories(ctx, query, opts)
	if err != nil {
<<<<<<< HEAD
		log.Printf("TormentNexus Outreach Error: GitHub search failed: %v", err)
=======
		slog.Error("TormentNexus Outreach: GitHub search failed", "error", err)
>>>>>>> origin/main
		return
	}

	for _, repo := range result.Repositories {
		domain := fmt.Sprintf("github.com/%s", repo.GetFullName())
<<<<<<< HEAD
			// #nosec G706 -- Domain name is used for context in informational logs
		log.Printf("TormentNexus Outreach: Evaluating repository: %s", domain)
=======
		slog.Info("TormentNexus Outreach: Evaluating repository", "domain", domain)
>>>>>>> origin/main

		// Check if company already exists
		existing, _ := w.db.GetCompanyByDomain(ctx, domain)
		if existing != nil {
			continue
		}

		// Create new lead
		company := &db.Company{
			Name:           repo.GetName(),
			Domain:         domain,
			TechStack:      []string{repo.GetLanguage()},
			HiringSignals:  []string{"Active Open Source contributor"},
			MarketCapTier:  "SMB", // Default for discovered repos
		}

		if err := w.db.CreateCompany(ctx, company); err != nil {
<<<<<<< HEAD
				// #nosec G706 -- Domain name is used for context in error logs
			log.Printf("TormentNexus Outreach Warning: Failed to create company %s: %v", domain, err)
=======
			slog.Warn("TormentNexus Outreach: Failed to create company", "domain", domain, "error", err)
>>>>>>> origin/main
			continue
		}

		deal := &db.Deal{
			CompanyID:    company.ID,
			CurrentState: db.StateDiscovered,
		}

		if err := w.db.CreateDeal(ctx, deal); err != nil {
<<<<<<< HEAD
				// #nosec G706 -- Domain name is used for context in error logs
			log.Printf("TormentNexus Outreach Warning: Failed to create deal for %s: %v", domain, err)
		} else {
				// #nosec G706 -- Domain name is used for context in success logs
			log.Printf("TormentNexus Outreach Success: New lead discovered: %s", domain)
=======
			slog.Warn("TormentNexus Outreach: Failed to create deal", "domain", domain, "error", err)
		} else {
			metrics.LeadsDiscovered.Inc()
			slog.Info("TormentNexus Outreach: New lead discovered", "domain", domain)
>>>>>>> origin/main
		}
	}
}
