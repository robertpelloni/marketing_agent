package communication

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// CadenceStep defines a single touch point in an outreach sequence.
type CadenceStep struct {
	StepNumber     int           // Order in the sequence (1-based)
	Channel        db.Channel    // email, linkedin, github
	DelayAfterPrev time.Duration // How long to wait after the previous step
	TemplateID     string        // Identifier for the message template
	Subject        string        // Subject line override
}

// CadenceSchedule defines a multi-touch outreach sequence.
type CadenceSchedule struct {
	Name        string        // e.g., "Standard B2B Tech Outreach"
	Description string        // Human-readable description
	Steps       []CadenceStep // Ordered list of touch points
	MaxAttempts int           // Maximum total touches before marking as exhausted
}

// CalculateAdaptiveDelay calculates a dynamic delay for a cadence step.
// For repeated touches on the same channel (e.g. Email follow-ups), it progressively
// increases delay (24h -> 48h -> 72h) to avoid fatigue and improve reply rates.
func CalculateAdaptiveDelay(step CadenceStep, priorAttempts int) time.Duration {
	baseDelay := step.DelayAfterPrev
	if baseDelay <= 0 {
		return 0
	}
	// If it's a follow-up touch (priorAttempts >= 2), add progressive spacing
	if step.Channel == db.ChannelEmail && priorAttempts >= 2 {
		multiplier := priorAttempts - 1
		if multiplier > 3 {
			multiplier = 3
		}
		return baseDelay * time.Duration(multiplier)
	}
	return baseDelay
}

// DefaultCadence returns the default outreach cadence for new leads.
func DefaultCadence() CadenceSchedule {
	return CadenceSchedule{
		Name:        "Standard B2B Tech Outreach",
		Description: "Default 5-touch sequence across email, GitHub, and LinkedIn",
		MaxAttempts: 10,
		Steps: []CadenceStep{
			{
				StepNumber:     1,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 0, // First touch immediately
				TemplateID:     "intro-email",
				Subject:        "HyperNexus for %s — Quick Question",
			},
			{
				StepNumber:     2,
				Channel:        db.ChannelGitHub,
				DelayAfterPrev: 24 * time.Hour, // 1 day after email
				TemplateID:     "github-hook",
				Subject:        "",
			},
			{
				StepNumber:     3,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 24 * time.Hour, // 1 day after GitHub
				TemplateID:     "followup-email",
				Subject:        "Re: HyperNexus for %s — Thoughts?",
			},
			{
				StepNumber:     4,
				Channel:        db.ChannelLinkedIn,
				DelayAfterPrev: 24 * time.Hour, // 1 day after email
				TemplateID:     "linkedin-connect",
				Subject:        "HyperNexus — AI Infrastructure",
			},
			{
				StepNumber:     5,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 48 * time.Hour, // 2 days after LinkedIn
				TemplateID:     "breakup-email",
				Subject:        "Should I close your file?",
			},
		},
	}
}

// CadenceTracker tracks where a lead is in their outreach cadence.
type CadenceTracker struct {
	db *db.DB
}

// NewCadenceTracker creates a new cadence tracker.
func NewCadenceTracker(database *db.DB) *CadenceTracker {
	return &CadenceTracker{db: database}
}

// CadenceProgress represents the current cadence state for a deal.
type CadenceProgress struct {
	DealID            int64
	ScheduleName      string
	NextStepNumber    int
	LastTouchTime     time.Time
	NextScheduledTime time.Time
	TotalAttempts     int
	IsExhausted       bool
}

// GetNextStep determines which cadence step to execute next for a deal.
func (ct *CadenceTracker) GetNextStep(ctx context.Context, dealID int64, schedule CadenceSchedule, interactions []db.Interaction) (*CadenceStep, *CadenceProgress, error) {
	stepByTemplate := make(map[string]int, len(schedule.Steps))
	for _, s := range schedule.Steps {
		if s.TemplateID != "" {
			stepByTemplate[s.TemplateID] = s.StepNumber
		}
	}

	totalAttempts := 0
	maxCompletedStep := 0
	lastOutboundTime := time.Time{}

	for _, i := range interactions {
		switch i.Direction {
		case "Outbound":
			totalAttempts++
			if i.CreatedAt.After(lastOutboundTime) {
				lastOutboundTime = i.CreatedAt
			}
			if stepNum, ok := stepByTemplate[i.TemplateID]; ok && stepNum > maxCompletedStep {
				maxCompletedStep = stepNum
			}
		case "Skipped":
			if stepNum, ok := stepByTemplate[i.TemplateID]; ok && stepNum > maxCompletedStep {
				maxCompletedStep = stepNum
			}
		}
	}

	// Check if max attempts exceeded
	if totalAttempts >= schedule.MaxAttempts {
		return nil, &CadenceProgress{
			DealID:         dealID,
			ScheduleName:   schedule.Name,
			NextStepNumber: len(schedule.Steps) + 1,
			TotalAttempts:  totalAttempts,
			IsExhausted:    true,
		}, nil
	}

	// Determine which step to execute next monotonically so multi-email sequences
	// (Step 1 email -> Step 2 GitHub -> Step 3 email -> Step 4 LinkedIn -> Step 5 email)
	// never loop back from Step 3 to Step 2.
	nextStepNumber := 1
	if maxCompletedStep > 0 {
		nextStepNumber = maxCompletedStep + 1
	}
	if totalAttempts+1 > nextStepNumber {
		nextStepNumber = totalAttempts + 1
	}

	// Find the actual step
	var nextStep *CadenceStep
	for i, step := range schedule.Steps {
		if step.StepNumber == nextStepNumber {
			nextStep = &schedule.Steps[i]
			break
		}
	}

	if nextStep == nil {
		return nil, &CadenceProgress{
			DealID:         dealID,
			ScheduleName:   schedule.Name,
			NextStepNumber: nextStepNumber,
			TotalAttempts:  totalAttempts,
			IsExhausted:    nextStepNumber > len(schedule.Steps),
		}, nil
	}

	// Calculate adaptive delay based on channel and attempts
	delay := CalculateAdaptiveDelay(*nextStep, totalAttempts)
	nextScheduled := lastOutboundTime.Add(delay)
	if lastOutboundTime.IsZero() {
		nextScheduled = time.Now() // First step is immediate
	}

	return nextStep, &CadenceProgress{
		DealID:            dealID,
		ScheduleName:      schedule.Name,
		NextStepNumber:    nextStepNumber,
		LastTouchTime:     lastOutboundTime,
		NextScheduledTime: nextScheduled,
		TotalAttempts:     totalAttempts,
		IsExhausted:       false,
	}, nil
}

// IsTimeForNextStep checks if it's time to execute the next cadence step.
func (ct *CadenceTracker) IsTimeForNextStep(progress CadenceProgress) bool {
	if progress.IsExhausted {
		return false
	}
	return time.Now().After(progress.NextScheduledTime)
}

// ShouldEngageContact is a convenience method that combines GetNextStep and IsTimeForNextStep.
// Returns the next step if it's time to engage, nil otherwise.
func (ct *CadenceTracker) ShouldEngageContact(ctx context.Context, dealID int64, interactions []db.Interaction) (*CadenceStep, error) {
	schedule := DefaultCadence()
	nextStep, progress, err := ct.GetNextStep(ctx, dealID, schedule, interactions)
	if err != nil {
		return nil, fmt.Errorf("cadence: get next step: %w", err)
	}

	if progress.IsExhausted {
		slog.Info(fmt.Sprintf("Cadence: Deal %d outreach exhausted after %d attempts", dealID, progress.TotalAttempts))
		return nil, nil
	}

	if !ct.IsTimeForNextStep(*progress) {
		slog.Info(fmt.Sprintf("Cadence: Deal %d not ready for next step (next at %s)", dealID, progress.NextScheduledTime.Format(time.RFC3339)))
		return nil, nil
	}

	return nextStep, nil
}

// CadenceAwareManager extends the Communication Manager with cadence scheduling.
// It integrates with the existing Manager to provide multi-touch outreach.
type CadenceAwareManager struct {
	*Manager
	tracker        *CadenceTracker
	linkedinSender *LinkedInSender
	githubSender   *GitHubCommentSender
}

// NewCadenceAwareManager wraps an existing Manager with cadence tracking.
func NewCadenceAwareManager(mgr *Manager, database *db.DB) *CadenceAwareManager {
	botRepo := os.Getenv("GITHUB_BOT_REPOSITORY")
	if botRepo == "" {
		botRepo = "HyperNexusSoft/HyperNexus" // Default to HyperNexus project repo
	}
	return &CadenceAwareManager{
		Manager:        mgr,
		tracker:        NewCadenceTracker(database),
		linkedinSender: NewLinkedInSender(),
		githubSender:   NewGitHubCommentSender(botRepo),
	}
}

// RunCadenceStarts the cadence-aware outreach loop.
func (cam *CadenceAwareManager) RunCadence(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("CadenceAwareManager: Multi-touch outreach scheduler started (interval: %v)", interval))

	// Run immediately on startup
	cam.checkCadence(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("CadenceAwareManager: Scheduler stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			cam.checkCadence(ctx)
		}
	}
}

// checkCadence checks all active deals for cadence-appropriate outreach.
func (cam *CadenceAwareManager) checkCadence(ctx context.Context) {
	if cam.db == nil {
		slog.Info("CadenceAwareManager: DB unavailable, skipping cadence check")
		return
	}

	deals, err := cam.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info(fmt.Sprintf("CadenceAwareManager: Error listing deals: %v", err))
		return
	}

	// Add OutreachSent and Engaged states to follow-up cadence
	for _, state := range []db.LeadState{db.StateOutreachSent, db.StateEngaged} {
		additional, err := cam.db.ListDealsByState(ctx, state)
		if err == nil {
			deals = append(deals, additional...)
		}
	}

	for _, deal := range deals {
		contacts, err := cam.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		interactions, err := cam.db.ListInteractionsByContact(ctx, contacts[0].ID)
		if err != nil {
			continue
		}

		nextStep, err := cam.tracker.ShouldEngageContact(ctx, deal.ID, interactions)
		if err != nil {
			slog.Info(fmt.Sprintf("CadenceAwareManager: Cadence check error for deal %d: %v", deal.ID, err))
			continue
		}

		if nextStep == nil {
			continue // Not time yet or exhausted
		}

		slog.Info(fmt.Sprintf("CadenceAwareManager: Executing cadence step %d (%s) for deal %d via %s",
			nextStep.StepNumber, nextStep.TemplateID, deal.ID, nextStep.Channel))

		// Trigger outreach based on channel
		switch nextStep.Channel {
		case db.ChannelEmail:
			recipientEmail := strings.TrimSpace(contacts[0].Email)
			if recipientEmail == "" {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Skipping email step %d for deal %d — contact has no email", nextStep.StepNumber, deal.ID))
				_ = cam.db.CreateInteraction(ctx, &db.Interaction{
					ContactID:  contacts[0].ID,
					Channel:    nextStep.Channel.String(),
					Direction:  "Skipped",
					Summary:    fmt.Sprintf("Skipped cadence step %d (%s): missing email", nextStep.StepNumber, nextStep.TemplateID),
					TemplateID: nextStep.TemplateID,
				})
				_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)
				continue
			}

			// Load template and render content
			tmpl, err := cam.db.GetTemplate(ctx, nextStep.TemplateID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Template %s not found for deal %d: %v", nextStep.TemplateID, deal.ID, err))
				// Fallback to ProcessInbound
				if _, err := cam.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
					slog.Info(fmt.Sprintf("CadenceAwareManager: Fallback email step failed for deal %d: %v", deal.ID, err))
				}
				continue
			}

			// Build SalesContext for template rendering
			company, err := cam.db.GetCompanyByID(ctx, deal.CompanyID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Could not get company %d: %v", deal.CompanyID, err))
				continue
			}

			contact := contacts[0]
			salesCtx := SalesContext{
				Company:      *company,
				Deal:         deal,
				Contact:      contact,
				Interactions: interactions,
				LatestIntent: IntentGeneral, // We're proactively reaching out
			}

			// Type-assert to *RAGResponseGenerator to access GenerateFromTemplate
			ragResponder, ok := cam.responder.(*RAGResponseGenerator)
			if !ok {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Responder does not support templates for deal %d", deal.ID))
				continue
			}

			// Generate subject and body from template
			subject, body, err := ragResponder.GenerateFromTemplate(ctx, tmpl, salesCtx)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Template rendering failed for deal %d: %v", deal.ID, err))
				continue
			}

			// Send email FIRST before persisting Outbound interaction or advancing cadence step
			if cam.sender != nil {
				emailMsg := EmailMessage{
					To:      recipientEmail,
					Subject: subject,
					Body:    body,
				}
				if err := cam.sender.Send(ctx, emailMsg); err != nil {
					slog.Info(fmt.Sprintf("CadenceAwareManager: Email send failed for deal %d: %v", deal.ID, err))
					cam.recordOutreachLog(ctx, company.Name, recipientEmail, tmpl.ID, "failed")
					continue
				}
				slog.Info(fmt.Sprintf("CadenceAwareManager: Email sent successfully for deal %d (step %d, template %s)", deal.ID, nextStep.StepNumber, tmpl.ID))
				cam.recordOutreachLog(ctx, company.Name, recipientEmail, tmpl.ID, "sent")
			} else {
				slog.Info(fmt.Sprintf("CadenceAwareManager: No email sender configured — email logged for deal %d", deal.ID))
			}

			// Record template impression for A/B testing metrics
			if impErr := cam.db.RecordTemplateImpression(ctx, tmpl.ID); impErr != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to record template impression for %s: %v", tmpl.ID, impErr))
			}

			// Create outbound interaction after successful delivery
			outbound := db.Interaction{
				ContactID:  contacts[0].ID,
				Channel:    nextStep.Channel.String(),
				Direction:  "Outbound",
				RawText:    body,
				Summary:    fmt.Sprintf("Cadence step %d: %s via %s", nextStep.StepNumber, tmpl.ID, nextStep.Channel),
				TemplateID: tmpl.ID,
			}
			if err := cam.db.CreateInteraction(ctx, &outbound); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to create interaction for deal %d: %v", deal.ID, err))
				continue
			}

			if deal.CurrentState == db.StateResearched {
				_ = cam.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent)
			}

			// Update cadence step in DB
			if err := cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to update cadence step for deal %d: %v", deal.ID, err))
			}

		case db.ChannelLinkedIn:
			if strings.TrimSpace(contacts[0].LinkedInURL) == "" {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Skipping LinkedIn step %d for deal %d — no LinkedIn profile URL populated", nextStep.StepNumber, deal.ID))
				_ = cam.db.CreateInteraction(ctx, &db.Interaction{
					ContactID:  contacts[0].ID,
					Channel:    nextStep.Channel.String(),
					Direction:  "Skipped",
					Summary:    fmt.Sprintf("Skipped cadence step %d (%s): missing LinkedIn URL", nextStep.StepNumber, nextStep.TemplateID),
					TemplateID: nextStep.TemplateID,
				})
				_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)
				continue
			}

			// Load template
			tmpl, err := cam.db.GetTemplate(ctx, nextStep.TemplateID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Template %s not found for LinkedIn step: %v", nextStep.TemplateID, err))
				continue
			}

			company, err := cam.db.GetCompanyByID(ctx, deal.CompanyID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Could not get company details for LinkedIn step: %v", err))
				continue
			}

			salesCtx := SalesContext{
				Company:      *company,
				Deal:         deal,
				Contact:      contacts[0],
				Interactions: interactions,
				LatestIntent: IntentGeneral,
			}

			ragResponder, ok := cam.responder.(*RAGResponseGenerator)
			if !ok {
				slog.Info("CadenceAwareManager: Responder does not support templates for LinkedIn")
				continue
			}

			subject, body, err := ragResponder.GenerateFromTemplate(ctx, tmpl, salesCtx)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: LinkedIn template rendering failed: %v", err))
				continue
			}

			// Trigger send via active LinkedInSender BEFORE persisting Outbound interaction
			if cam.linkedinSender != nil {
				lkMsg := LinkedInMessage{
					ProfileURL: contacts[0].LinkedInURL,
					Subject:    subject,
					Body:       body,
				}
				if err := cam.linkedinSender.Send(ctx, lkMsg); err != nil {
					slog.Info(fmt.Sprintf("CadenceAwareManager: LinkedIn send failed for deal %d: %v", deal.ID, err))
					continue
				}
				slog.Info(fmt.Sprintf("CadenceAwareManager: LinkedIn message sent successfully to %s (step %d, template %s)", contacts[0].LinkedInURL, nextStep.StepNumber, tmpl.ID))
			}

			outbound := db.Interaction{
				ContactID:  contacts[0].ID,
				Channel:    nextStep.Channel.String(),
				Direction:  "Outbound",
				RawText:    body,
				Summary:    fmt.Sprintf("Cadence step %d: %s via %s", nextStep.StepNumber, tmpl.ID, nextStep.Channel),
				TemplateID: tmpl.ID,
			}
			if err := cam.db.CreateInteraction(ctx, &outbound); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to log LinkedIn interaction: %v", err))
				continue
			}

			// Update cadence step in DB
			if err := cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to update cadence step for deal %d: %v", deal.ID, err))
			}

		case db.ChannelGitHub:
			tmpl, err := cam.db.GetTemplate(ctx, nextStep.TemplateID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Template %s not found for GitHub step: %v", nextStep.TemplateID, err))
				continue
			}

			company, err := cam.db.GetCompanyByID(ctx, deal.CompanyID)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Could not get company details for GitHub step: %v", err))
				continue
			}

			salesCtx := SalesContext{
				Company:      *company,
				Deal:         deal,
				Contact:      contacts[0],
				Interactions: interactions,
				LatestIntent: IntentGeneral,
			}

			ragResponder, ok := cam.responder.(*RAGResponseGenerator)
			if !ok {
				slog.Info("CadenceAwareManager: Responder does not support templates for GitHub")
				continue
			}

			_, body, err := ragResponder.GenerateFromTemplate(ctx, tmpl, salesCtx)
			if err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: GitHub template rendering failed: %v", err))
				continue
			}

			// Search and post comment BEFORE logging Outbound interaction
			if cam.githubSender != nil {
				orgName := extractOrgFromDomain(company.Domain)
				if orgName == "" {
					_ = cam.db.CreateInteraction(ctx, &db.Interaction{
						ContactID:  contacts[0].ID,
						Channel:    nextStep.Channel.String(),
						Direction:  "Skipped",
						Summary:    fmt.Sprintf("Skipped cadence step %d (%s): empty GitHub org", nextStep.StepNumber, tmpl.ID),
						TemplateID: tmpl.ID,
					})
					_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)
					continue
				}
				targets, searchErr := cam.githubSender.SearchRelevantIssues(ctx, orgName)
				if searchErr != nil {
					slog.Info(fmt.Sprintf("CadenceAwareManager: GitHub issue search failed for org %s (deal %d): %v", orgName, deal.ID, searchErr))
					continue
				}
				if len(targets) == 0 {
					slog.Info(fmt.Sprintf("CadenceAwareManager: No relevant GitHub issues found for org %s — skipping step %d", orgName, nextStep.StepNumber))
					_ = cam.db.CreateInteraction(ctx, &db.Interaction{
						ContactID:  contacts[0].ID,
						Channel:    nextStep.Channel.String(),
						Direction:  "Skipped",
						Summary:    fmt.Sprintf("Skipped cadence step %d (%s): no matching open GitHub issues for %s", nextStep.StepNumber, tmpl.ID, orgName),
						TemplateID: tmpl.ID,
					})
					_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)
					continue
				}
				target := targets[0]
				if err := cam.githubSender.SendComment(ctx, target.Owner, target.Repo, target.IssueNumber, body); err != nil {
					slog.Info(fmt.Sprintf("CadenceAwareManager: GitHub comment failed for deal %d: %v", deal.ID, err))
					continue
				}
				slog.Info(fmt.Sprintf("CadenceAwareManager: GitHub comment posted successfully on %s/%s#%d (step %d)", target.Owner, target.Repo, target.IssueNumber, nextStep.StepNumber))
			}

			outbound := db.Interaction{
				ContactID:  contacts[0].ID,
				Channel:    nextStep.Channel.String(),
				Direction:  "Outbound",
				RawText:    body,
				Summary:    fmt.Sprintf("Cadence step %d: %s via %s", nextStep.StepNumber, tmpl.ID, nextStep.Channel),
				TemplateID: tmpl.ID,
			}
			if err := cam.db.CreateInteraction(ctx, &outbound); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to log GitHub interaction: %v", err))
				continue
			}

			// Update cadence step in DB
			if err := cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber); err != nil {
				slog.Info(fmt.Sprintf("CadenceAwareManager: Failed to update cadence step for deal %d: %v", deal.ID, err))
			}
		}
	}
}

// recordOutreachLog best-effort records email outreach attempts to outreach_log
// so /api/v1/outreach/stats and the consolidated dashboard stay in sync.
func (cam *CadenceAwareManager) recordOutreachLog(ctx context.Context, companyName, email, category, status string) {
	if cam.db == nil || cam.db.Conn == nil {
		return
	}
	_, _ = cam.db.Conn.ExecContext(ctx,
		`INSERT INTO outreach_log (company_name, contact_email, category, status, sent_at) VALUES ($1, $2, $3, $4, NOW())`,
		companyName, email, category, status,
	)
}
