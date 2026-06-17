package communication

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// Intent represents the classified purpose of an inbound message.
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

// IntentClassifier defines the interface for categorizing inbound communication.
type IntentClassifier interface {
	Classify(ctx context.Context, text string) (Intent, error)
}

// ResponseGenerator defines the interface for creating tailored replies.
type ResponseGenerator interface {
	Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error)
}

// OrderProcessor defines the interface for fulfillment after a win.
type OrderProcessor interface {
	ProcessOrder(ctx context.Context, deal db.Deal) error
}

// Manager coordinates the inbound communication state machine.
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

// NewManager creates a new communication Manager.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, sender EmailSender, github *GitHubSender, linkedin *LinkedInSender, registry *llm.PromptRegistry) *Manager {
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
	}
}

// Run starts the periodic inbound communication processing loop.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Communication Manager: Background poller started (interval: %v)...", interval)

	// Run immediately on startup
	m.pollAndProcess(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Communication Manager: Background poller stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	// Check for 'Researched' deals that haven't had an outbound interaction yet
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		log.Printf("Comm Manager: Error polling deals: %v", err)
		return
	}

	for _, deal := range deals {
		contacts, err := m.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		// Check if we already sent outreach
		interactions, _ := m.db.ListInteractionsByContact(ctx, contacts[0].ID)
		hasOutbound := false
		for _, i := range interactions {
			if i.Direction == "Outbound" {
				hasOutbound = true
				break
			}
		}

		if !hasOutbound && !deal.ApprovalRequired {
			log.Printf("Comm Manager: Initiating autonomous outreach for deal %d to %s", deal.ID, contacts[0].Email)
			// Trigger outreach
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
				log.Printf("Comm Manager Error: Failed to initiate outreach for deal %d: %v", deal.ID, err)
			}
			if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil {
				log.Printf("Comm Manager Error: Failed to update deal state to OutreachSent for deal %d: %v", deal.ID, err)
			}
		}
	}
}

// DefaultChannelForContact returns the channel to use when communicating with a contact.
func DefaultChannelForContact(contact db.Contact) string {
	if contact.PreferredChannel != "" {
		return contact.PreferredChannel
	}
	return string(db.ChannelEmail)
}

// ProcessInbound handles a new inbound message from a contact.
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	channel := DefaultChannelForContact(contact)

	// 1. Persist inbound interaction
	inbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   channel,
		Direction: "Inbound",
		RawText:   text,
	}
	_ = m.db.CreateInteraction(ctx, &inbound)

	// 2. Classify intent
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil { return "", err }

	company, _ := m.db.GetCompanyByID(ctx, contact.CompanyID)
	interactions, _ := m.db.ListInteractionsByContact(ctx, contact.ID)
	deal, _ := m.db.GetDealByCompanyID(ctx, contact.CompanyID)

	salesCtx := SalesContext{
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: interactions,
		LatestIntent: intent,
	}

	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil { return "", err }

	if action == ActionEscalate {
		return "I've flagged this for our technical lead. We will get back to you shortly.", nil
	}

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil { return "", err }

	// 3a. Record outcome if deal was won or positive sentiment
	if action == ActionAdvanceState && m.registry != nil {
		updatedDeal, _ := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if updatedDeal != nil && updatedDeal.CurrentState == db.StateClosedWon {
			// Record success for the active prompt version
			// In a real scenario, we'd track which version was used in the interaction
			m.registry.RecordOutcome("outreach-reply", "current", true)
		}
	}

	// 4. Persist outbound interaction
	outbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   channel,
		Direction: "Outbound",
		RawText:   replyText,
		Summary:   fmt.Sprintf("Reply to intent: %s", intent),
	}
	_ = m.db.CreateInteraction(ctx, &outbound)

	// 5. Send communication
	if m.sender != nil && contact.Email != "" && channel == "email" {
		subject := fmt.Sprintf("Re: %s — TormentNexus", company.Name)
		if text == "START_OUTREACH" { subject = fmt.Sprintf("TormentNexus for %s — Quick Question", company.Name) }
		_ = m.sender.Send(ctx, EmailMessage{To: contact.Email, Subject: subject, Body: replyText})
	}

	return replyText, nil
}

func (m *Manager) GetDB() *db.DB { return m.db }
