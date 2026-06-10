package researcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// DefaultDossierProcessor implements the DossierProcessor interface.
type DefaultDossierProcessor struct{}

func (p *DefaultDossierProcessor) Process(findings []string) (string, error) {
	if len(findings) == 0 {
		return "", nil
	}
	return strings.Join(findings, "\n\n"), nil
}

// PromptFormatter constructs a hyper-personalized outreach prompt.
type PromptFormatter struct {
	TormentNexusContext string
}

func (f *PromptFormatter) Format(dossier string) string {
	return fmt.Sprintf(`### TormentNexus OUTREACH CONTEXT ###
%s

### TARGET TECHNICAL FINDINGS ###
%s

### HYPER-PERSONALIZED HOOK ###
[Drafting specialized technical outreach based on the discovered bottleneck...]`, f.TormentNexusContext, dossier)
}

// Run starts the background research process.
func (r *Researcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Researcher worker started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Researcher worker stopping: Draining in-flight work")
			return
		case <-ticker.C:
			r.executeResearch(ctx)
		}
	}
}

// ExecuteResearch manually triggers a research cycle (exported for testing).
func (r *Researcher) ExecuteResearch(ctx context.Context) {
	r.executeResearch(ctx)
}

func (r *Researcher) executeResearch(ctx context.Context) {
	// Find deals in Researched state (contacts found, now need deep technical context)
	deals, err := r.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Error("Researcher: Error listing deals", "error", err)
		return
	}

	for _, deal := range deals {
		contacts, err := r.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		err = r.researchLead(ctx, deal, contacts[0])
		if err != nil {
			slog.Error("Researcher: Error researching lead", "deal_id", deal.ID, "error", err)
		}
	}
}

func (r *Researcher) researchLead(ctx context.Context, deal db.Deal, contact db.Contact) error {
	var findings []string

	// Crawl for each source
	targets := []string{contact.GitHubHandle, contact.Email} // Simplified targets
	for _, crawler := range r.crawlers {
		for _, target := range targets {
			if target == "" {
				continue
			}
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
					// We push the deal with a "Researcher" route to indicate provenance
					if err := r.crmClient.PushDeal(ctx, updatedDeal, *company, "Researcher"); err != nil {
						slog.Warn("Researcher: Failed to push updated dossier to CRM", "attempt", i+1, "max_retries", maxRetries, "deal_id", deal.ID, "error", err)
						time.Sleep(time.Duration(i+1) * 2 * time.Second)
						continue
					}
					return
				}
				slog.Error("Researcher: CRM dossier push failed after all attempts", "deal_id", deal.ID)
			}()
		}
	}

	slog.Info("Researcher: Successfully compiled technical dossier", "deal_id", deal.ID)
	return nil
}
