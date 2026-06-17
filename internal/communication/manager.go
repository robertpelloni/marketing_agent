package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

type Intent string
const (
	IntentTechnical      Intent = "Technical"
	IntentPricing        Intent = "Pricing"
	IntentObjection      Intent = "Objection"
	IntentMeetingRequest Intent = "MeetingRequest"
	IntentFollowUp       Intent = "FollowUp"
	IntentSpam           Intent = "Spam"
	IntentUnknown        Intent = "Unknown"
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
	github     *GitHubSender
	linkedin   *LinkedInSender
	registry   *llm.PromptRegistry
}

func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, sender EmailSender, github *GitHubSender, linkedin *LinkedInSender, registry *llm.PromptRegistry) *Manager {
	return &Manager{db: database, classifier: classifier, responder:  responder, strategy:   strategy, processor:  processor, sender:     sender, github:     github, linkedin:   linkedin, registry:   registry}
}

func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	m.pollAndProcess(ctx)
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C: m.pollAndProcess(ctx)
		}
	}
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	deals, _ := m.db.ListDealsByState(ctx, db.StateResearched)
	for _, deal := range deals {
		contacts, _ := m.db.ListContactsByCompany(ctx, deal.CompanyID)
		if len(contacts) == 0 { continue }
		interactions, _ := m.db.ListInteractionsByContact(ctx, contacts[0].ID)
		hasOutbound := false
		for _, i := range interactions {
			if i.Direction == "Outbound" { hasOutbound = true; break }
		}
		if !hasOutbound && !deal.ApprovalRequired {
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
	salesCtx := SalesContext{Company: *company, Deal: *deal, Contact: contact, Interactions: ints, LatestIntent: intent}
	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil { return "", err }
	if action == ActionEscalate { return "Escalated to human lead.", nil }
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil { return "", err }
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
	_ = m.db.CreateInteraction(ctx, &db.Interaction{ContactID: contact.ID, Channel: channel, Direction: "Outbound", RawText: replyText})
	if m.sender != nil && channel == "email" {
		_ = m.sender.Send(ctx, EmailMessage{To: contact.Email, Subject: "TormentNexus follow up", Body: replyText})
	}
	return replyText, nil
}

func (m *Manager) triggerWebhook(ctx context.Context, dealID int64, newState db.LeadState) {
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
