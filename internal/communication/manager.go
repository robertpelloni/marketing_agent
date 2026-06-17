package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/webhook"
)

type Intent string
const (
	IntentTechnical		Intent	= "Technical"
	IntentPricing		Intent	= "Pricing"
	IntentObjection		Intent	= "Objection"
	IntentMeetingRequest	Intent	= "MeetingRequest"
	IntentFollowUp		Intent	= "FollowUp"
	IntentSpam		Intent	= "Spam"
	IntentUnknown		Intent	= "Unknown"
	IntentGeneral		Intent	= "General"
)

type IntentClassifier interface { Classify(ctx context.Context, text string) (Intent, error) }
type ResponseGenerator interface { Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) }
type OrderProcessor interface { ProcessOrder(ctx context.Context, deal db.Deal) error }

type Manager struct {
	db         *db.DB
	classifier IntentClassifier
	responder  ResponseGenerator
	strategy   SalesStrategy
	processor  OrderProcessor
	sender     EmailSender
	github     *GitHubCommentSender
	linkedin   *LinkedInSender
	registry   *llm.PromptRegistry
	objections *ObjectionLibrary
	webhook    *webhook.Dispatcher
}

func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, sender EmailSender, github *GitHubCommentSender, linkedin *LinkedInSender, registry *llm.PromptRegistry) *Manager {
	return &Manager{
		db:         database,
		classifier: classifier,
		responder:  responder,
		strategy:   strategy,
		processor:  processor,
		sender:     sender,
		github:     github,
		linkedin:   linkedin,
		registry:   registry,
		webhook:    webhook.NewDispatcher(os.Getenv("WEBHOOK_URL")),
	}
}

func (m *Manager) SetObjectionLibrary(lib *ObjectionLibrary) {
	m.objections = lib
}

func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Communication Manager: Background poller started", "interval", interval)

	m.pollAndProcess(ctx)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Communication Manager stopping...")
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

func (m *Manager) ExecutePoll(ctx context.Context) {
	m.pollAndProcess(ctx)
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	if m.db == nil { return }
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Error("Comm Manager: Error polling deals", "error", err)
		return
	}

	for _, deal := range deals {
		contacts, _ := m.db.ListContactsByCompany(ctx, deal.CompanyID)
		if len(contacts) == 0 { continue }
		interactions, _ := m.db.ListInteractionsByContact(ctx, contacts[0].ID)
		hasOutbound := false
		for _, i := range interactions {
			if i.Direction == "Outbound" { hasOutbound = true; break }
		}

		if !hasOutbound && !deal.ApprovalRequired {
			slog.Info("Comm Manager: Initiating autonomous outreach", "deal", deal.ID, "to", contacts[0].Email)
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err == nil {
				_ = m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent)
				m.triggerWebhook(ctx, deal.ID, db.StateOutreachSent)
			}
		}
	}
}

func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	channel := "email"
	if contact.PreferredChannel != "" { channel = contact.PreferredChannel }

	_ = m.db.CreateInteraction(ctx, &db.Interaction{ContactID: contact.ID, Channel: channel, Direction: "Inbound", RawText: text})
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil { return "", err }

	company, _ := m.db.GetCompanyByID(ctx, contact.CompanyID)
	ints, _ := m.db.ListInteractionsByContact(ctx, contact.ID)
	deal, _ := m.db.GetDealByCompanyID(ctx, contact.CompanyID)

	salesCtx := SalesContext{
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: ints,
		LatestIntent: intent,
	}

	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil { return "", err }
	if action == ActionEscalate { return "Escalated to human lead.", nil }

	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil { return "", err }

	var responseID string
	if m.objections != nil {
		sentiment := AnalyzeSentiment(text)
		if sentiment.Sentiment == SentimentNegative || sentiment.Sentiment == SentimentMixed {
			match := m.objections.MatchObjection(ctx, text, sentiment, deal.CurrentState)
			if match != nil {
				replyText = match.Response.Text + "\n\n" + replyText
				responseID = match.Response.ID
				m.objections.RecordOutcome(responseID, false)
			}
		}
	}

	if action == ActionAdvanceState {
		updatedDeal, _ := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if updatedDeal != nil {
			m.triggerWebhook(ctx, updatedDeal.ID, updatedDeal.CurrentState)
			if updatedDeal.CurrentState == db.StateClosedWon {
				if m.registry != nil { m.registry.RecordOutcome("outreach-reply", "current", true) }
				if m.processor != nil { _ = m.processor.ProcessOrder(ctx, *updatedDeal) }
			}
		}
	}

	_ = m.db.CreateInteraction(ctx, &db.Interaction{
		ContactID:  contact.ID,
		Channel:    channel,
		Direction:  "Outbound",
		RawText:    replyText,
		Summary:    fmt.Sprintf("Reply to intent: %s", intent),
		ResponseID: responseID,
	})

	if m.sender != nil && channel == "email" {
		subject := "TormentNexus follow up"
		if text == "START_OUTREACH" { subject = "Quick question" }
		_ = m.sender.Send(ctx, EmailMessage{To: contact.Email, Subject: subject, Body: replyText})
	} else if channel == "github" && m.github != nil {
		_ = m.github.FindAndComment(ctx, *company, contact)
	}

	return replyText, nil
}

func (m *Manager) triggerWebhook(ctx context.Context, dealID int64, newState db.LeadState) {
	if m.webhook != nil {
		go func() {
			_ = m.webhook.Dispatch(context.Background(), dealID, newState)
		}()
		return
	}
	// Old trigger logic (legacy fallback if dispatcher not initialized)
	url := os.Getenv("WEBHOOK_URL")
	if url == "" { return }
	payload, _ := json.Marshal(map[string]interface{}{"deal_id": dealID, "event": "state_change", "new_state": newState, "ts": time.Now()})
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil { _ = resp.Body.Close() }
}

func (m *Manager) GetDB() *db.DB { return m.db }

func (m *Manager) ApproveDeal(ctx context.Context, dealID int64) error {
	err := m.db.UpdateDealState(ctx, dealID, db.StateNegotiating)
	if err == nil {
		m.triggerWebhook(ctx, dealID, db.StateNegotiating)
	}
	return err
}
