package communication

import (
	"context"
	"fmt"

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
	Generate(ctx context.Context, contact db.Contact, interaction db.Interaction, intent Intent) (string, error)
}

// Manager coordinates the inbound communication state machine.
type Manager struct {
	db         *db.DB
	classifier IntentClassifier
	responder  ResponseGenerator
}

// NewManager creates a new communication Manager.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator) *Manager {
	return &Manager{
		db:         database,
		classifier: classifier,
		responder:  responder,
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

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, contact, inbound, intent)
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
