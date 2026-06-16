package communication

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
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

// DefaultCadence returns the default outreach cadence for new leads.
func DefaultCadence() CadenceSchedule {
	return CadenceSchedule{
		Name:        "Standard B2B Tech Outreach",
		Description: "Default 5-touch sequence across email, GitHub, and LinkedIn",
		MaxAttempts: 5,
		Steps: []CadenceStep{
			{
				StepNumber:     1,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 0, // First touch immediately
				TemplateID:     "intro-email",
				Subject:        "TormentNexus for %s — Quick Question",
			},
			{
				StepNumber:     2,
				Channel:        db.ChannelGitHub,
				DelayAfterPrev: 48 * time.Hour, // 2 days after email
				TemplateID:     "github-hook",
				Subject:        "",
			},
			{
				StepNumber:     3,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 72 * time.Hour, // 3 days after GitHub
				TemplateID:     "followup-email",
				Subject:        "Re: TormentNexus for %s — Thoughts?",
			},
			{
				StepNumber:     4,
				Channel:        db.ChannelLinkedIn,
				DelayAfterPrev: 96 * time.Hour, // 4 days after email
				TemplateID:     "linkedin-connect",
				Subject:        "TormentNexus — AI Infrastructure",
			},
			{
				StepNumber:     5,
				Channel:        db.ChannelEmail,
				DelayAfterPrev: 168 * time.Hour, // 7 days after LinkedIn
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
	totalAttempts := 0
	lastOutboundTime := time.Time{}
	var lastChannel string

	for _, i := range interactions {
		if i.Direction == "Outbound" {
			totalAttempts++
			if i.CreatedAt.After(lastOutboundTime) {
				lastOutboundTime = i.CreatedAt
				lastChannel = i.Channel
			}
		}
	}

	// Check if max attempts exceeded
	if totalAttempts >= schedule.MaxAttempts {
		return nil, &CadenceProgress{
			DealID:        dealID,
			ScheduleName:  schedule.Name,
			NextStepNumber: len(schedule.Steps) + 1,
			TotalAttempts: totalAttempts,
			IsExhausted:   true,
		}, nil
	}

	// Determine which step to execute next
	nextStepNumber := 1
	if lastOutboundTime.IsZero() {
		// No outbound yet — first step
		nextStepNumber = 1
	} else {
		// Find the step matching the last channel
		for i, step := range schedule.Steps {
			if string(step.Channel) == lastChannel {
				nextStepNumber = step.StepNumber + 1
				break
			}
			// If we've passed all steps up to this point without match,
			// the next step is the current one
			if step.StepNumber > totalAttempts {
				nextStepNumber = step.StepNumber
				break
			}
			_ = i
		}
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
			DealID:        dealID,
			ScheduleName:  schedule.Name,
			NextStepNumber: nextStepNumber,
			TotalAttempts: totalAttempts,
			IsExhausted:   nextStepNumber > len(schedule.Steps),
		}, nil
	}

	// Calculate next scheduled time
	nextScheduled := lastOutboundTime.Add(nextStep.DelayAfterPrev)
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
		log.Printf("Cadence: Deal %d outreach exhausted after %d attempts", dealID, progress.TotalAttempts)
		return nil, nil
	}

	if !ct.IsTimeForNextStep(*progress) {
		log.Printf("Cadence: Deal %d not ready for next step (next at %s)", dealID, progress.NextScheduledTime.Format(time.RFC3339))
		return nil, nil
	}

	return nextStep, nil
}

// CadenceAwareManager extends the Communication Manager with cadence scheduling.
// It integrates with the existing Manager to provide multi-touch outreach.
type CadenceAwareManager struct {
	*Manager
	tracker *CadenceTracker
}

// NewCadenceAwareManager wraps an existing Manager with cadence tracking.
func NewCadenceAwareManager(mgr *Manager, database *db.DB) *CadenceAwareManager {
	return &CadenceAwareManager{
		Manager: mgr,
		tracker: NewCadenceTracker(database),
	}
}

// RunCadenceStarts the cadence-aware outreach loop.
func (cam *CadenceAwareManager) RunCadence(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("CadenceAwareManager: Multi-touch outreach scheduler started (interval: %v)", interval)

	// Run immediately on startup
	cam.checkCadence(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("CadenceAwareManager: Scheduler stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			cam.checkCadence(ctx)
		}
	}
}

// checkCadence checks all active deals for cadence-appropriate outreach.
func (cam *CadenceAwareManager) checkCadence(ctx context.Context) {
	deals, err := cam.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		log.Printf("CadenceAwareManager: Error listing deals: %v", err)
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
			log.Printf("CadenceAwareManager: Cadence check error for deal %d: %v", deal.ID, err)
			continue
		}

		if nextStep == nil {
			continue // Not time yet or exhausted
		}

		log.Printf("CadenceAwareManager: Executing cadence step %d (%s) for deal %d via %s",
			nextStep.StepNumber, nextStep.TemplateID, deal.ID, nextStep.Channel)

		// Trigger outreach based on channel
		switch nextStep.Channel {
		case db.ChannelEmail:
			// Use the existing ProcessInbound mechanism for email
			if _, err := cam.Manager.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
				log.Printf("CadenceAwareManager: Email step failed for deal %d: %v", deal.ID, err)
			}
		case db.ChannelLinkedIn:
			// LinkedIn step — log for now (needs LinkedInSender)
			if cam.linkedin != nil { cam.linkedin.SendMessage(ctx, contacts[0].LinkedInURL, "TormentNexus AI - follow up") }
		case db.ChannelGitHub:
			// GitHub step — log for now (needs GitHubCommentSender)
			if cam.github != nil && contacts[0].GitHubHandle != "" { cam.github.SendComment(ctx, "robertpelloni", "enterprise_sales_bot", 1, "Technical inquiry regarding TormentNexus") }
		}
	}
}
