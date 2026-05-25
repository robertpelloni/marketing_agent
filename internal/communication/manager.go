package communication

import (
	"context"
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
