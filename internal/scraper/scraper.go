package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

type LeadSource interface {
	Discover(ctx context.Context, keywords []string) ([]db.Company, error)
}

type Scraper struct {
	db      *db.DB
	sources []LeadSource
}

func NewScraper(database *db.DB, sources []LeadSource) *Scraper {
	return &Scraper{
		db:      database,
		sources: sources,
	}
}

func (s *Scraper) Run(ctx context.Context, interval time.Duration, keywords []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Scraper worker started...")
	s.ExecuteDiscovery(ctx, keywords)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scraper worker stopping...")
			return
		case <-ticker.C:
			s.ExecuteDiscovery(ctx, keywords)
		}
	}
}

func (s *Scraper) ExecuteDiscovery(ctx context.Context, keywords []string) {
	start := time.Now()
	slog.Info("Scraper: Polling for leads...")
	for _, source := range s.sources {
		companies, err := source.Discover(ctx, keywords)
		if err != nil {
			slog.Info("Error discovering leads from source", "error", err)
			continue
		}

		for _, company := range companies {
			err := s.processDiscoveredCompany(ctx, company)
			if err != nil {
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

	existing, err := s.db.GetCompanyByDomain(ctx, company.Domain)
	if err == nil && existing != nil {
		return nil
	}

	err = s.db.CreateCompany(ctx, &company)
	if err != nil {
		return fmt.Errorf("failed to persist company: %w", err)
	}

	deal := db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateDiscovered,
	}
	err = s.db.CreateDeal(ctx, &deal)
	if err != nil {
		return fmt.Errorf("failed to create initial deal: %w", err)
	}

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
}
