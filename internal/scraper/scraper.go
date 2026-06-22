package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LeadSource defines an interface for discovering potential leads.
type LeadSource interface {
	Discover(ctx context.Context, keywords []string) ([]db.Company, error)
}

// Scraper coordinates the discovery and persistence of leads.
type Scraper struct {
	db      *db.DB
	sources []LeadSource
}

// NewScraper creates a new Scraper instance.
func NewScraper(database *db.DB, sources []LeadSource) *Scraper {
	return &Scraper{
		db:      database,
		sources: sources,
	}
}

// Run starts the background discovery process.
func (s *Scraper) Run(ctx context.Context, interval time.Duration, keywords []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Scraper worker started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scraper worker stopping: Draining in-flight work")
			return
		case <-ticker.C:
			s.ExecuteDiscovery(ctx, keywords)
		}
	}
}

// ExecuteDiscovery manually triggers a discovery cycle (exported for testing).
func (s *Scraper) ExecuteDiscovery(ctx context.Context, keywords []string) {
	for _, source := range s.sources {
		companies, err := source.Discover(ctx, keywords)
		if err != nil {
			slog.Error("Error discovering leads from source", "error", err)
			continue
		}

		for _, company := range companies {
			err := s.processDiscoveredCompany(ctx, company)
			if err != nil {
				slog.Error("Error processing company", "company_name", company.Name, "error", err)
			}
		}
	}
}

func (s *Scraper) processDiscoveredCompany(ctx context.Context, company db.Company) error {
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

	slog.Info("Successfully discovered and persisted new lead", "company_name", company.Name, "domain", company.Domain)
	return nil
}

<<<<<<< HEAD
// MockJobBoardSource is a simulated lead source for testing and initial development.
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
// MockJobBoardSource is a legacy stub that no longer generates mock data.
// All lead sources are now real API integrations.
type MockJobBoardSource struct{}

func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	return nil, nil
>>>>>>> origin/main
}
