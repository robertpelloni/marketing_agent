package scraper

import (
	"context"
	"fmt"
	"log/slog"
<<<<<<< HEAD
	"net/http"
=======
>>>>>>> origin/main
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LeadSource defines an interface for discovering potential leads.
type LeadSource interface {
	Discover(ctx context.Context, keywords []string) ([]db.Company, error)
}

// Scraper coordinates the discovery and persistence of leads.
type Scraper struct {
<<<<<<< HEAD
	db	*db.DB
	sources	[]LeadSource
=======
	db      *db.DB
	sources []LeadSource
>>>>>>> origin/main
}

// NewScraper creates a new Scraper instance.
func NewScraper(database *db.DB, sources []LeadSource) *Scraper {
	return &Scraper{
<<<<<<< HEAD
		db:		database,
		sources:	sources,
=======
		db:      database,
		sources: sources,
>>>>>>> origin/main
	}
}

// Run starts the background discovery process.
func (s *Scraper) Run(ctx context.Context, interval time.Duration, keywords []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	slog.Info("Scraper worker started...")

	// Run immediately on startup
	s.ExecuteDiscovery(ctx, keywords)
=======
	slog.Info("Scraper worker started")
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			slog.Info("Scraper worker stopping: Draining in-flight work...")
=======
			slog.Info("Scraper worker stopping: Draining in-flight work")
>>>>>>> origin/main
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
<<<<<<< HEAD
			slog.Info(fmt.Sprintf("Error discovering leads from source: %v", err))
=======
			slog.Error("Error discovering leads from source", "error", err)
>>>>>>> origin/main
			continue
		}

		for _, company := range companies {
			err := s.processDiscoveredCompany(ctx, company)
			if err != nil {
<<<<<<< HEAD
				slog.Info(fmt.Sprintf("Error processing company %s: %v", company.Name, err))
=======
				slog.Error("Error processing company", "company_name", company.Name, "error", err)
>>>>>>> origin/main
			}
		}
	}
}

func (s *Scraper) processDiscoveredCompany(ctx context.Context, company db.Company) error {
<<<<<<< HEAD
	if s.db == nil {
		slog.Info("Scraper: DB unavailable, skipping company processing")
		return nil
	}

	// Check if company already exists
	existing, err := s.db.GetCompanyByDomain(ctx, company.Domain)
	if err == nil && existing != nil {
		// Company already exists, skip or update signals
=======
	// Check if company already exists
	existing, err := s.db.GetCompanyByDomain(ctx, company.Domain)
	if err == nil && existing != nil {
>>>>>>> origin/main
		return nil
	}

	// Create new company
	err = s.db.CreateCompany(ctx, &company)
	if err != nil {
		return fmt.Errorf("failed to persist company: %w", err)
	}

	// Initialize a deal in Discovered state
	deal := db.Deal{
<<<<<<< HEAD
		CompanyID:	company.ID,
		CurrentState:	db.StateDiscovered,
=======
		CompanyID:    company.ID,
		CurrentState: db.StateDiscovered,
>>>>>>> origin/main
	}
	err = s.db.CreateDeal(ctx, &deal)
	if err != nil {
		return fmt.Errorf("failed to create initial deal: %w", err)
	}

<<<<<<< HEAD
	slog.Info(fmt.Sprintf("Successfully discovered and persisted new lead: %s (%s)", company.Name, company.Domain))
	return nil
}

<<<<<<< HEAD
// MockJobBoardSource is a simulated lead source for testing and initial development.
type MockJobBoardSource struct{}

func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	// Simulate finding leads based on keywords
	slog.Info(fmt.Sprintf("MockJobBoardSource: Scanning for keywords: %v", keywords))

	return []db.Company{
		{
			Name:		"AI Dynamics Corp",
			Domain:		"aidynamics.com",
			TechStack:	[]string{"Python", "PyTorch", "Kubernetes"},
			HiringSignals:	[]string{"Hiring: Senior AI Platform Engineer"},
			MarketCapTier:	"Mid-Market",
		},
		{
			Name:		"Neural Systems Inc",
			Domain:		"neuralsystems.io",
			TechStack:	[]string{"Go", "Rust", "LLMs"},
			HiringSignals:	[]string{"Hiring: LLM Orchestration Architect"},
			MarketCapTier:	"Enterprise",
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
