package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Intent represents the classified purpose of an inbound message.
type Intent string

const (
	IntentTechnical Intent = "Technical"
	IntentPricing   Intent = "Pricing"
	IntentObjection Intent = "Objection"
	IntentSpam      Intent = "Spam"
	IntentUnknown   Intent = "Unknown"
)

// IntentClassifier defines the interface for categorizing inbound communication.
type IntentClassifier interface {
	Classify(ctx context.Context, text string) (Intent, error)
}

// ResponseGenerator defines the interface for creating tailored replies.
type ResponseGenerator interface {
	Generate(ctx context.Context, contact db.Contact, interaction db.Interaction, intent Intent, action Action) (string, error)
}

// Manager coordinates the inbound communication state machine.
type Manager struct {
	db         *db.DB
	classifier IntentClassifier
	responder  ResponseGenerator
	strategy   SalesStrategy
}

// NewManager creates a new communication Manager.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy) *Manager {
	return &Manager{
		db:         database,
		classifier: classifier,
		responder:  responder,
		strategy:   strategy,
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
	replyText, err := m.responder.Generate(ctx, contact, inbound, intent, action)
	if err != nil {
		return "", err
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
