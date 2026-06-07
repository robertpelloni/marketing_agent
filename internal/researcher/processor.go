package researcher

import (
	"context"
	"fmt"
	"log"
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
	BorgContext string
}

func (f *PromptFormatter) Format(dossier string) string {
	return fmt.Sprintf(`### BORG OUTREACH CONTEXT ###
%s

### TARGET TECHNICAL FINDINGS ###
%s

### HYPER-PERSONALIZED HOOK ###
[Drafting specialized technical outreach based on the discovered bottleneck...]`, f.BorgContext, dossier)
}

// Run starts the background research process.
func (r *Researcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("Researcher worker started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Researcher worker stopping: Draining in-flight work...")
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
		log.Printf("Researcher: Error listing deals: %v", err)
		return
	}

	for _, deal := range deals {
		contacts, err := r.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		err = r.researchLead(ctx, deal, contacts[0])
		if err != nil {
			log.Printf("Researcher: Error researching lead %d: %v", deal.ID, err)
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

	log.Printf("Researcher: Successfully compiled technical dossier for deal %d", deal.ID)
	return nil
}
