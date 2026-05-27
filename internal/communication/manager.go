package communication

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
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
}

// NewManager creates a new communication Manager.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor) *Manager {
	return &Manager{
		db:         database,
		classifier: classifier,
		responder:  responder,
		strategy:   strategy,
		processor:  processor,
	}
}

// Run starts the periodic inbound communication processing loop.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Communication Manager: Background poller started (interval: %v)...", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Communication Manager: Background poller stopping...")
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	// In a real scenario, this would poll an IMAP server or Webhook queue.
	// For this architecture, we simulate by checking for 'Researched' deals
	// that haven't had an outbound interaction yet (triggering initial outreach).
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

		if !hasOutbound {
			log.Printf("Comm Manager: Initiating autonomous outreach for deal %d to %s", deal.ID, contacts[0].Email)
			// Trigger outreach
			m.ProcessInbound(ctx, contacts[0], "START_OUTREACH") // Internal trigger
			m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent)
		}
	}
}

// ProcessInbound handles a new inbound message from a contact.
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	// 1. Persist inbound interaction
	inbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   "Email", // Default for now
		Direction: "Inbound",
		RawText:   text,
	}
	err := m.db.CreateInteraction(ctx, &inbound)
	if err != nil {
		return "", err
	}

	// 2. Classify intent
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil {
		return "", err
	}

	// 2a. Decide next action using strategy engine
	company, err := m.db.GetCompanyByID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company for strategy: %w", err)
	}

	interactions, err := m.db.ListInteractionsByContact(ctx, contact.ID)
	if err != nil {
		log.Printf("Warning: failed to list interactions for strategy: %v", err)
	}

	deal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get deal for strategy: %w", err)
	}

	salesCtx := SalesContext{
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: interactions,
		LatestIntent: intent,
	}

	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil {
		return "", err
	}

	if action == ActionEscalate {
		log.Printf("UI: Deal %d escalated for human review.", salesCtx.Deal.ID)
		return "I've flagged this for our technical lead to review. We will get back to you shortly.", nil
	}

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil {
		return "", err
	}

	// 3a. Handle order processing if deal was won
	if action == ActionAdvanceState {
		updatedDeal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err == nil && updatedDeal.CurrentState == db.StateClosedWon && m.processor != nil {
			log.Printf("Comm Manager: Triggering order processor for won deal %d", updatedDeal.ID)
			if err := m.processor.ProcessOrder(ctx, *updatedDeal); err != nil {
				log.Printf("Comm Manager Error: Order processing failed: %v", err)
			}
		}
	}

	// 4. Persist outbound interaction
	outbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   "Email",
		Direction: "Outbound",
		RawText:   replyText,
		Summary:   fmt.Sprintf("Reply to intent: %s", intent),
	}
	err = m.db.CreateInteraction(ctx, &outbound)
	if err != nil {
		return "", err
	}

	return replyText, nil
}
