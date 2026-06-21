package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// TwitterSource implements LeadSource by searching Twitter/X for
// companies discussing AI infrastructure pain points.
type TwitterSource struct {
	Client   *http.Client
	Username string
	Password string
}

// Discover searches Twitter for companies hiring or discussing AI orchestration.
func (t *TwitterSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("TwitterSource: Searching for AI/LLM signals...")

	if t.Username == "" || t.Password == "" {
		slog.Info("TwitterSource: No credentials configured, returning simulated results")
		return t.simulate(ctx, keywords)
	}

	slog.Info("TwitterSource: Credentials configured, attempting headless scrape")
	return t.scrapeTwitter(ctx, keywords)
}

// scrapeTwitter uses headless browser to search Twitter/X
func (t *TwitterSource) scrapeTwitter(ctx context.Context, keywords []string) ([]db.Company, error) {
	// Twitter/X scraping via browser automation would go here
	// For now, fall back to simulation
	slog.Info("TwitterSource: Browser-based scraping not yet implemented, using simulation")
	return t.simulate(ctx, keywords)
}

// simulate returns mock companies from Twitter-like signals
func (t *TwitterSource) simulate(ctx context.Context, keywords []string) ([]db.Company, error) {
	simulatedCompanies := []db.Company{
		{
			Name:          "NeuroCore AI",
			Domain:        "neurocore.ai",
			TechStack:     []string{"Python", "PyTorch", "Kubernetes", "LLMs"},
			HiringSignals: []string{"Hiring: ML Infrastructure Engineer", "Building multi-agent orchestration platform"},
			MarketCapTier: "Startup",
		},
		{
			Name:          "Cortex Dynamics",
			Domain:        "cortexdynamics.io",
			TechStack:     []string{"Go", "Rust", "gRPC", "Redis"},
			HiringSignals: []string{"Hiring: Distributed Systems Engineer", "Series B funded, scaling AI infrastructure"},
			MarketCapTier: "Mid-Market",
		},
		{
			Name:          "Synthwave Labs",
			Domain:        "synthwavelabs.com",
			TechStack:     []string{"TypeScript", "Node.js", "AWS", "LangChain"},
			HiringSignals: []string{"Building AI agent framework", "YC W25"},
			MarketCapTier: "Startup",
		},
		{
			Name:          "OmniInfer",
			Domain:        "omniinfer.tech",
			TechStack:     []string{"Go", "Python", "TensorFlow", "Kafka"},
			HiringSignals: []string{"Building LLM inference platform", "Hiring: AI Platform Engineer"},
			MarketCapTier: "Mid-Market",
		},
		{
			Name:          "Apex Orchestration",
			Domain:        "apexorchestra.com",
			TechStack:     []string{"Rust", "Go", "gRPC", "Docker"},
			HiringSignals: []string{"Multi-agent coordination platform", "Hiring: Senior Backend Engineer"},
			MarketCapTier: "Startup",
		},
	}

	// Filter by keyword relevance
	var matches []db.Company
	for _, c := range simulatedCompanies {
		for _, kw := range keywords {
			lower := strings.ToLower(c.Name + " " + strings.Join(c.TechStack, " ") + " " + strings.Join(c.HiringSignals, " "))
			if strings.Contains(lower, strings.ToLower(kw)) {
				matches = append(matches, c)
				break
			}
		}
	}

	if len(matches) == 0 {
		// Return first 3 if no keyword matches
		if len(simulatedCompanies) > 3 {
			matches = simulatedCompanies[:3]
		} else {
			matches = simulatedCompanies
		}
	}

	slog.Info(fmt.Sprintf("TwitterSource: Simulated discovery of %d companies from Twitter signals", len(matches)))
	return matches, nil
}

func (t *TwitterSource) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}
