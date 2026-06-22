package scraper

import (
	"context"
	"fmt"
	"log/slog"
<<<<<<< HEAD
	"net/http"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

=======
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LeadSource defines an interface for discovering potential leads.
>>>>>>> origin/main
type LeadSource interface {
	Discover(ctx context.Context, keywords []string) ([]db.Company, error)
}

<<<<<<< HEAD
=======
// Scraper coordinates the discovery and persistence of leads.
>>>>>>> origin/main
type Scraper struct {
	db      *db.DB
	sources []LeadSource
}

<<<<<<< HEAD
=======
// NewScraper creates a new Scraper instance.
>>>>>>> origin/main
func NewScraper(database *db.DB, sources []LeadSource) *Scraper {
	return &Scraper{
		db:      database,
		sources: sources,
	}
}

<<<<<<< HEAD
=======
// Run starts the background discovery process.
>>>>>>> origin/main
func (s *Scraper) Run(ctx context.Context, interval time.Duration, keywords []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	slog.Info("Scraper worker started...")

	// Run immediately on startup
	s.poll(ctx, keywords)
=======
	slog.Info("Scraper worker started")
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			slog.Info("Scraper worker stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			s.poll(ctx, keywords)
=======
			slog.Info("Scraper worker stopping: Draining in-flight work")
			return
		case <-ticker.C:
			s.ExecuteDiscovery(ctx, keywords)
>>>>>>> origin/main
		}
	}
}

<<<<<<< HEAD
func (s *Scraper) poll(ctx context.Context, keywords []string) {
	start := time.Now()
	slog.Info("Scraper: Polling for leads...")
	for _, source := range s.sources {
		companies, err := source.Discover(ctx, keywords)
		if err != nil {
			slog.Info("Error discovering leads from source", "error", err)
=======
// ExecuteDiscovery manually triggers a discovery cycle (exported for testing).
func (s *Scraper) ExecuteDiscovery(ctx context.Context, keywords []string) {
	for _, source := range s.sources {
		companies, err := source.Discover(ctx, keywords)
		if err != nil {
			slog.Error("Error discovering leads from source", "error", err)
>>>>>>> origin/main
			continue
		}

		for _, company := range companies {
			err := s.processDiscoveredCompany(ctx, company)
			if err != nil {
<<<<<<< HEAD
				slog.Info("Error processing company", "name", company.Name, "error", err)
			}
		}
	}
	deploy.RecordTiming("Scraper", time.Since(start))
}

func (s *Scraper) processDiscoveredCompany(ctx context.Context, company db.Company) error {
	if s.db == nil {
		slog.Info("Scraper: DB unavailable, skipping company processing")
		return nil
	}

=======
				slog.Error("Error processing company", "company_name", company.Name, "error", err)
			}
		}
	}
}

func (s *Scraper) processDiscoveredCompany(ctx context.Context, company db.Company) error {
>>>>>>> origin/main
	// Check if company already exists
	existing, err := s.db.GetCompanyByDomain(ctx, company.Domain)
	if err == nil && existing != nil {
		return nil
	}

	// Create new company
	err = s.db.CreateCompany(ctx, &company)
	if err != nil {
		return fmt.Errorf("failed to persist company: %w", err)
	}

	// Initialize a deal in Discovered state
	deal := db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateDiscovered,
	}
	err = s.db.CreateDeal(ctx, &deal)
	if err != nil {
		return fmt.Errorf("failed to create initial deal: %w", err)
	}

<<<<<<< HEAD
	slog.Info("Successfully discovered and persisted new lead", "name", company.Name, "domain", company.Domain)
	return nil
}

type GitHubJobSource struct {
	Client *http.Client
}

func (g *GitHubJobSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("GitHubJobSource: Discovering hiring signals", "keywords", keywords)
	return []db.Company{
		{
			Name:          "Compute Logic",
			Domain:        "computelogic.tech",
			TechStack:     []string{"Go", "gRPC", "TormentNexus"},
			HiringSignals: []string{"Hiring: Distributed Systems Engineer (Autonomous Agent focus)"},
			MarketCapTier: "Enterprise",
		},
	}, nil
}

type MockJobBoardSource struct{}
func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("MockJobBoardSource: Scanning for keywords", "keywords", keywords)
	return []db.Company{
		{
			Name:          "AI Dynamics Corp",
			Domain:        "aidynamics.com",
			TechStack:     []string{"Python", "PyTorch", "Kubernetes"},
			HiringSignals: []string{"Hiring: Senior AI Platform Engineer"},
			MarketCapTier: "Mid-Market",
		},
		{
			Name:          "Neural Systems Inc",
			Domain:        "neuralsystems.io",
			TechStack:     []string{"Go", "Rust", "LLMs"},
			HiringSignals: []string{"Hiring: LLM Orchestration Architect"},
			MarketCapTier: "Enterprise",
		},
	}, nil
=======
	slog.Info("Successfully discovered and persisted new lead", "company_name", company.Name, "domain", company.Domain)
	return nil
}

// MockJobBoardSource is a legacy stub that no longer generates mock data.
// All lead sources are now real API integrations.
type MockJobBoardSource struct{}

func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	return nil, nil
>>>>>>> origin/main
}
