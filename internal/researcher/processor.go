package researcher

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"log/slog"
>>>>>>> origin/main
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
<<<<<<< HEAD
	BorgContext string
}

func (f *PromptFormatter) Format(dossier string) string {
	return fmt.Sprintf(`### BORG OUTREACH CONTEXT ###
=======
	TormentNexusContext string
}

func (f *PromptFormatter) Format(dossier string) string {
	return fmt.Sprintf(`### TormentNexus OUTREACH CONTEXT ###
>>>>>>> origin/main
%s

### TARGET TECHNICAL FINDINGS ###
%s

### HYPER-PERSONALIZED HOOK ###
<<<<<<< HEAD
[Drafting specialized technical outreach based on the discovered bottleneck...]`, f.BorgContext, dossier)
=======
[Drafting specialized technical outreach based on the discovered bottleneck...]`, f.TormentNexusContext, dossier)
>>>>>>> origin/main
}

// Run starts the background research process.
func (r *Researcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	log.Println("Researcher worker started...")
=======
	slog.Info("Researcher worker started...")
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			log.Println("Researcher worker stopping...")
=======
			slog.Info("Researcher worker stopping: Draining in-flight work...")
>>>>>>> origin/main
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
<<<<<<< HEAD
	// Find deals in Researched state (contacts found, now need deep technical context)
	deals, err := r.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		log.Printf("Researcher: Error listing deals: %v", err)
=======
	if r.db == nil {
		slog.Info("Researcher: DB unavailable, skipping research cycle")
		return
	}

	// Find deals in Researched state (contacts found, now need deep technical context)
	deals, err := r.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info(fmt.Sprintf("Researcher: Error listing deals: %v", err))
>>>>>>> origin/main
		return
	}

	for _, deal := range deals {
		contacts, err := r.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		err = r.researchLead(ctx, deal, contacts[0])
		if err != nil {
<<<<<<< HEAD
			log.Printf("Researcher: Error researching lead %d: %v", deal.ID, err)
=======
			slog.Info(fmt.Sprintf("Researcher: Error researching lead %d: %v", deal.ID, err))
>>>>>>> origin/main
		}
	}
}

func (r *Researcher) researchLead(ctx context.Context, deal db.Deal, contact db.Contact) error {
	var findings []string

	// Crawl for each source
<<<<<<< HEAD
	targets := []string{contact.GitHubHandle, contact.Email}	// Simplified targets
=======
	targets := []string{contact.GitHubHandle, contact.Email} // Simplified targets
>>>>>>> origin/main
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
<<<<<<< HEAD
	company, _ := r.db.GetCompanyByID(ctx, deal.CompanyID)

	// Calculate Unified Intent Score
	intentAggregator := &IntentAggregator{}
	var intentScore int
	if company != nil {
		score := intentAggregator.Aggregate(*company, dossier)
		intentScore = score.Score
		slog.Info(fmt.Sprintf("Researcher: Unified Intent Score for %s is %d", company.Name, intentScore))
	}

=======
>>>>>>> origin/main
	err = r.db.UpdateTechnicalDossier(ctx, deal.ID, dossier)
	if err != nil {
		return err
	}

<<<<<<< HEAD
	log.Printf("Researcher: Successfully compiled technical dossier for deal %d", deal.ID)
=======
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
<<<<<<< HEAD
						log.Printf("Researcher Warning: Failed to push updated dossier to CRM (attempt %d/%d): %v", i+1, maxRetries, err)
=======
						slog.Info(fmt.Sprintf("Researcher Warning: Failed to push updated dossier to CRM (attempt %d/%d): %v", i+1, maxRetries, err))
>>>>>>> origin/main
						time.Sleep(time.Duration(i+1) * 2 * time.Second)
						continue
					}
					return
				}
<<<<<<< HEAD
				log.Printf("Researcher Error: CRM dossier push failed after %d attempts for deal %d", maxRetries, deal.ID)
=======
				slog.Info(fmt.Sprintf("Researcher Error: CRM dossier push failed after %d attempts for deal %d", maxRetries, deal.ID))
>>>>>>> origin/main
			}()
		}
	}

<<<<<<< HEAD
	log.Printf("Researcher: Successfully compiled technical dossier for deal %d", deal.ID)
=======
	slog.Info(fmt.Sprintf("Researcher: Successfully compiled technical dossier for deal %d", deal.ID))
>>>>>>> origin/main
	return nil
}
