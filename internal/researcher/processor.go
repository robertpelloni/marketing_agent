package researcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

const maxResearchLeadsPerCycle = 100

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

	slog.Info("Researcher worker started...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Researcher worker stopping: Draining in-flight work...")
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
	if r.db == nil {
		slog.Info("Researcher: DB unavailable, skipping research cycle")
		return
	}

	// Find deals in Researched state (contacts found, now need deep technical context)
	deals, err := r.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info(fmt.Sprintf("Researcher: Error listing deals: %v", err))
		return
	}

	processed := 0
	for _, deal := range deals {
		if processed >= maxResearchLeadsPerCycle {
			break
		}
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Skip leads that already have a compiled technical dossier so hourly
		// cycles only research newly enriched leads instead of re-crawling all
		// ~5,000 Researched leads every hour.
		if strings.TrimSpace(deal.TechnicalDossier) != "" {
			continue
		}

		contacts, err := r.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		primaryContact, _ := db.SelectOutreachContact(contacts)
		if err := r.researchLead(ctx, deal, primaryContact); err != nil {
			slog.Info(fmt.Sprintf("Researcher: Error researching lead %d: %v", deal.ID, err))
			continue
		}
		processed++
	}
}

// buildResearchTargets derives clean, non-PII company/handle targets for
// crawlers, stripping raw email addresses and consumer webmail domains (gmail,
// icloud, etc.) that previously polluted GitHub/Blog/RSS queries.
func buildResearchTargets(company *db.Company, contact db.Contact) []string {
	var targets []string
	seen := make(map[string]struct{})
	add := func(raw string) {
		val := strings.TrimSpace(raw)
		if val == "" || strings.Contains(val, "@") {
			return
		}
		key := strings.ToLower(val)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		targets = append(targets, val)
	}

	if gh := strings.TrimSpace(contact.GitHubHandle); gh != "" {
		add(gh)
	}

	if company != nil {
		domain := strings.TrimSpace(company.Domain)
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimPrefix(domain, "www.")
		if strings.HasPrefix(strings.ToLower(domain), "github.com/") {
			rest := strings.Trim(domain[len("github.com/"):], "/")
			parts := strings.Split(rest, "/")
			if len(parts) >= 1 && parts[0] != "" {
				add(parts[0])
			}
			if len(parts) >= 2 && parts[1] != "" {
				add(parts[1])
			}
		} else if domain != "" && !db.IsPersonalEmailDomain(domain) {
			add(domain)
		}

		add(company.Name)
	}

	if email := strings.TrimSpace(contact.Email); email != "" && !db.IsAutomatedOrUnroutableEmail(email) {
		if at := strings.IndexByte(email, '@'); at > 0 && at+1 < len(email) {
			emailDomain := strings.ToLower(email[at+1:])
			if !db.IsPersonalEmailDomain(emailDomain) && emailDomain != "github.com" {
				add(emailDomain)
			}
		}
	}

	// Cap target count per lead so multi-field companies don't multiply crawl calls.
	if len(targets) > 2 {
		targets = targets[:2]
	}
	return targets
}

func (r *Researcher) researchLead(ctx context.Context, deal db.Deal, contact db.Contact) error {
	var findings []string

	company, _ := r.db.GetCompanyByID(ctx, deal.CompanyID)
	targets := buildResearchTargets(company, contact)

	for _, crawler := range r.crawlers {
		for _, target := range targets {
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

	// Calculate Unified Intent Score
	intentAggregator := &IntentAggregator{}
	var intentScore int
	if company != nil {
		score := intentAggregator.Aggregate(*company, dossier)
		intentScore = score.Score
		slog.Info(fmt.Sprintf("Researcher: Unified Intent Score for %s is %d", company.Name, intentScore))
	}

	err = r.db.UpdateTechnicalDossier(ctx, deal.ID, dossier)
	if err != nil {
		return err
	}

	// Synchronize the updated deal (including dossier) with the CRM (with retry logic)
	if r.crmClient != nil && company != nil {
		updatedDeal := deal
		updatedDeal.TechnicalDossier = dossier
		companyCopy := *company
		go func() {
			maxRetries := 3
			for i := 0; i < maxRetries; i++ {
				// We push the deal with a "Researcher" route to indicate provenance
				if err := r.crmClient.PushDeal(ctx, updatedDeal, companyCopy, "Researcher"); err != nil {
					slog.Info(fmt.Sprintf("Researcher Warning: Failed to push updated dossier to CRM (attempt %d/%d): %v", i+1, maxRetries, err))
					time.Sleep(time.Duration(i+1) * 2 * time.Second)
					continue
				}
				return
			}
			slog.Info(fmt.Sprintf("Researcher Error: CRM dossier push failed after %d attempts for deal %d", maxRetries, deal.ID))
		}()
	}

	slog.Info(fmt.Sprintf("Researcher: Successfully compiled technical dossier for deal %d", deal.ID))
	return nil
}
