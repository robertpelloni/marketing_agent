package researcher

import (
	"context"
	"log/slog"
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

	slog.Info("Researcher worker started...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Researcher worker stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			r.ExecuteResearch(ctx)
		}
	}
}

func (r *Researcher) ExecuteResearch(ctx context.Context) {
	if r.db == nil { return }
	start := time.Now()
	// Find deals in Researched state (contacts found, now need deep technical context)
	deals, err := r.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info("Researcher: Error listing deals", "error", err)
		return
	}

	for _, deal := range deals {
		contacts, err := r.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		err = r.researchLead(ctx, deal, contacts[0])
		if err != nil {
			slog.Info("Researcher: Error researching lead", "deal", deal.ID, "error", err)
		}
	}
	deploy.RecordTiming("Researcher", time.Since(start))
}

func (r *Researcher) researchLead(ctx context.Context, deal db.Deal, contact db.Contact) error {
	var findings []string

	// Crawl for each source
	targets := []string{contact.GitHubHandle, contact.Email}
	for _, crawler := range r.crawlers {
		for _, target := range targets {
			if target == "" { continue }
			finding, err := crawler.Crawl(ctx, target)
			if err == nil && finding != "" {
				findings = append(findings, finding)
			}
		}
	}

	dossier, err := r.processor.Process(findings)
	if err != nil {
		return err
	}

	// Update deal with dossier
	err = r.db.UpdateTechnicalDossier(ctx, deal.ID, dossier)
	if err != nil {
		return err
	}

	// Synchronize the updated deal (including dossier) with the CRM (with retry logic)
	if r.crmClient != nil {
		company, _ := r.db.GetCompanyByID(ctx, deal.CompanyID)
		if company != nil {
			updatedDeal := deal
			updatedDeal.TechnicalDossier = dossier
			go func() {
				maxRetries := 3
				for i := 0; i < maxRetries; i++ {
					if err := r.crmClient.PushDeal(ctx, updatedDeal, *company, "Researcher"); err != nil {
						slog.Info("Researcher Warning: Failed to push updated dossier to CRM", "attempt", i+1, "error", err)
						time.Sleep(time.Duration(i+1) * 2 * time.Second)
						continue
					}
					return
				}
				slog.Info("Researcher Error: CRM dossier push failed after max retries", "deal", deal.ID)
			}()
		}
	}

	slog.Info("Researcher: Successfully compiled technical dossier", "deal", deal.ID)
	return nil
}
