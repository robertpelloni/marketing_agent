package communication

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type CadenceStep struct {
	StepNumber     int
	Channel        db.Channel
	DelayAfterPrev time.Duration
	TemplateID     string
	Subject        string
}

type CadenceSchedule struct {
	Name        string
	Description string
	Steps       []CadenceStep
	MaxAttempts int
}

func DefaultCadence() CadenceSchedule {
	return CadenceSchedule{
		Name:        "Standard B2B Tech Outreach",
		Description: "Default 5-touch sequence across email, GitHub, and LinkedIn",
		MaxAttempts: 5,
		Steps: []CadenceStep{
			{StepNumber: 1, Channel: db.ChannelEmail, DelayAfterPrev: 0, TemplateID: "intro-email", Subject: "TormentNexus for %s"},
			{StepNumber: 2, Channel: db.ChannelGitHub, DelayAfterPrev: 48 * time.Hour, TemplateID: "github-hook"},
			{StepNumber: 3, Channel: db.ChannelEmail, DelayAfterPrev: 72 * time.Hour, TemplateID: "followup-email", Subject: "Re: TormentNexus"},
			{StepNumber: 4, Channel: db.ChannelLinkedIn, DelayAfterPrev: 96 * time.Hour, TemplateID: "linkedin-connect"},
			{StepNumber: 5, Channel: db.ChannelEmail, DelayAfterPrev: 168 * time.Hour, TemplateID: "breakup-email", Subject: "Closing file"},
		},
	}
}

type CadenceTracker struct {
	db *db.DB
}

func NewCadenceTracker(database *db.DB) *CadenceTracker {
	return &CadenceTracker{db: database}
}

type CadenceProgress struct {
	DealID            int64
	ScheduleName      string
	NextStepNumber    int
	LastTouchTime     time.Time
	NextScheduledTime time.Time
	TotalAttempts     int
	IsExhausted       bool
}

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

	if totalAttempts >= schedule.MaxAttempts {
		return nil, &CadenceProgress{DealID: dealID, TotalAttempts: totalAttempts, IsExhausted: true}, nil
	}

	nextStepNumber := totalAttempts + 1
	var nextStep *CadenceStep
	for i, step := range schedule.Steps {
		if step.StepNumber == nextStepNumber {
			nextStep = &schedule.Steps[i]
			break
		}
	}

	if nextStep == nil {
		return nil, &CadenceProgress{DealID: dealID, TotalAttempts: totalAttempts, IsExhausted: true}, nil
	}

	nextScheduled := lastOutboundTime.Add(nextStep.DelayAfterPrev)
	if lastOutboundTime.IsZero() { nextScheduled = time.Now() }

	return nextStep, &CadenceProgress{
		DealID:            dealID,
		NextStepNumber:    nextStepNumber,
		LastTouchTime:     lastOutboundTime,
		NextScheduledTime: nextScheduled,
		TotalAttempts:     totalAttempts,
		IsExhausted:       false,
	}, nil
}

func (ct *CadenceTracker) IsTimeForNextStep(progress CadenceProgress) bool {
	if progress.IsExhausted { return false }
	return time.Now().After(progress.NextScheduledTime)
}

func (ct *CadenceTracker) ShouldEngageContact(ctx context.Context, dealID int64, interactions []db.Interaction) (*CadenceStep, error) {
	schedule := DefaultCadence()
	nextStep, progress, err := ct.GetNextStep(ctx, dealID, schedule, interactions)
	if err != nil { return nil, err }
	if progress.IsExhausted || !ct.IsTimeForNextStep(*progress) { return nil, nil }
	return nextStep, nil
}

type CadenceAwareManager struct {
	*Manager
	tracker *CadenceTracker
}

func NewCadenceAwareManager(mgr *Manager, database *db.DB) *CadenceAwareManager {
	return &CadenceAwareManager{
		Manager: mgr,
		tracker: NewCadenceTracker(database),
	}
}

func (cam *CadenceAwareManager) RunCadence(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	slog.Info("CadenceAwareManager: Outreach scheduler started", "interval", interval)
	cam.checkCadence(ctx)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: cam.checkCadence(ctx)
		}
	}
}

func (cam *CadenceAwareManager) checkCadence(ctx context.Context) {
	if cam.db == nil { return }
	deals, _ := cam.db.ListDealsByState(ctx, db.StateResearched)
	for _, state := range []db.LeadState{db.StateOutreachSent, db.StateEngaged} {
		additional, _ := cam.db.ListDealsByState(ctx, state)
		deals = append(deals, additional...)
	}

	for _, deal := range deals {
		contacts, _ := cam.db.ListContactsByCompany(ctx, deal.CompanyID)
		if len(contacts) == 0 { continue }
		interactions, _ := cam.db.ListInteractionsByContact(ctx, contacts[0].ID)
		nextStep, _ := cam.tracker.ShouldEngageContact(ctx, deal.ID, interactions)
		if nextStep == nil { continue }

		slog.Info("CadenceAwareManager: Executing step", "step", nextStep.StepNumber, "deal", deal.ID, "channel", nextStep.Channel)

		switch nextStep.Channel {
		case db.ChannelEmail:
			tmpl, err := cam.db.GetTemplate(ctx, nextStep.TemplateID)
			if err != nil {
				_, _ = cam.Manager.ProcessInbound(ctx, contacts[0], "START_OUTREACH")
				continue
			}
			company, _ := cam.db.GetCompanyByID(ctx, deal.CompanyID)
			salesCtx := SalesContext{Company: *company, Deal: deal, Contact: contacts[0], Interactions: interactions, LatestIntent: IntentGeneral}
			ragResponder, _ := cam.responder.(*RAGResponseGenerator)
			subject, body, _ := ragResponder.GenerateFromTemplate(ctx, tmpl, salesCtx)
			_ = cam.db.RecordTemplateImpression(ctx, tmpl.ID)
			_ = cam.db.CreateInteraction(ctx, &db.Interaction{ContactID: contacts[0].ID, Channel: "email", Direction: "Outbound", RawText: body, TemplateID: tmpl.ID})
			if cam.Manager.sender != nil {
				_ = cam.Manager.sender.Send(ctx, EmailMessage{To: contacts[0].Email, Subject: subject, Body: body})
			}
			_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)

		case db.ChannelLinkedIn:
			if cam.linkedin != nil { _ = cam.linkedin.SendMessage(ctx, contacts[0].LinkedInURL, "Follow up regarding TormentNexus") }
			_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)

		case db.ChannelGitHub:
			if cam.github != nil && contacts[0].GitHubHandle != "" {
				company, _ := cam.db.GetCompanyByID(ctx, deal.CompanyID)
				_ = cam.github.FindAndComment(ctx, *company, contacts[0])
			}
			_ = cam.db.SetCadenceStep(ctx, deal.ID, nextStep.StepNumber)
		}
	}
}
