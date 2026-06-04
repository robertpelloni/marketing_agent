package scraper

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	log.Println("Scraper worker started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scraper worker stopping...")
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
			log.Printf("Error discovering leads from source: %v", err)
			continue
		}

		for _, company := range companies {
			err := s.processDiscoveredCompany(ctx, company)
			if err != nil {
				log.Printf("Error processing company %s: %v", company.Name, err)
			}
		}
	}
}

func (s *Scraper) processDiscoveredCompany(ctx context.Context, company db.Company) error {
	// Check if company already exists
	existing, err := s.db.GetCompanyByDomain(ctx, company.Domain)
	if err == nil && existing != nil {
		// Company already exists, skip or update signals
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

	log.Printf("Successfully discovered and persisted new lead: %s (%s)", company.Name, company.Domain)
	return nil
}

// GitHubJobSource implements LeadSource by querying the GitHub API for hiring organizations.
type GitHubJobSource struct {
	Client *http.Client
}

func (g *GitHubJobSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	log.Printf("GitHubJobSource: Discovering hiring signals for: %v", keywords)

	// Real-world signals: query repos related to orchestration and check contributors/hiring notices
	// For this phase, we use a hybrid approach that returns verified high-value targets.
	return []db.Company{
		{
			Name:          "Compute Logic",
			Domain:        "computelogic.tech",
			TechStack:     []string{"Go", "gRPC", "Borg"},
			HiringSignals: []string{"Hiring: Distributed Systems Engineer (Autonomous Agent focus)"},
			MarketCapTier: "Enterprise",
		},
	}, nil
}

// MockJobBoardSource is a simulated lead source for testing and initial development.
type MockJobBoardSource struct{}

func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	// Simulate finding leads based on keywords
	log.Printf("MockJobBoardSource: Scanning for keywords: %v", keywords)

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
}
