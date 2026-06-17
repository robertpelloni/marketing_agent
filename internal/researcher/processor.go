package researcher

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

type DefaultDossierProcessor struct{}
func (p *DefaultDossierProcessor) Process(findings []string) (string, error) {
	return strings.Join(findings, "\n\n"), nil
}

func (r *Researcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Println("Researcher worker started...")
	r.executeResearch(ctx)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: r.executeResearch(ctx)
		}
	}
}

func (r *Researcher) executeResearch(ctx context.Context) {
	start := time.Now()
	deals, _ := r.db.ListDealsByState(ctx, db.StateResearched)
	for _, d := range deals {
		contacts, _ := r.db.ListContactsByCompany(ctx, d.CompanyID)
		if len(contacts) == 0 { continue }
		var findings []string
		for _, c := range r.crawlers {
			f, err := c.Crawl(ctx, contacts[0].Email)
			if err == nil { findings = append(findings, f) }
		}
		dossier, _ := r.processor.Process(findings)
		_ = r.db.UpdateTechnicalDossier(ctx, d.ID, dossier)
	}
	deploy.RecordTiming("Researcher", time.Since(start))
}
