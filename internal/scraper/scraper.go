package scraper

import (
	"context"
	"log"
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
	return &Scraper{db: database, sources: sources}
}

func (s *Scraper) Run(ctx context.Context, interval time.Duration, keywords []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Println("Scraper worker started...")
	s.poll(ctx, keywords)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: s.poll(ctx, keywords)
		}
	}
}

func (s *Scraper) poll(ctx context.Context, keywords []string) {
	start := time.Now()
	log.Println("Scraper: Polling for leads...")
	for _, source := range s.sources {
		companies, err := source.Discover(ctx, keywords)
		if err != nil { log.Printf("Scraper: error from source: %v", err); continue }
		for _, c := range companies {
			_ = s.db.CreateCompany(ctx, &c)
		}
	}
	deploy.RecordTiming("Scraper", time.Since(start))
}

type MockJobBoardSource struct{}
func (m *MockJobBoardSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	return []db.Company{{Name: "AI Dynamics Corp", Domain: "aidynamics.io"}}, nil
}
