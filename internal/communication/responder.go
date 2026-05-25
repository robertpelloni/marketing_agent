package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// MockResponseGenerator provides template-based replies for testing.
type MockResponseGenerator struct{}

func (m *MockResponseGenerator) Generate(ctx context.Context, contact db.Contact, interaction db.Interaction, intent Intent) (string, error) {
	log.Printf("MockResponseGenerator: Generating response for intent: %s", intent)

	switch intent {
	case IntentTechnical:
		return fmt.Sprintf("Hi %s, that's a great technical question. Borg uses a Go-based headless daemon for multi-agent LLM coordination. You can find more details in our technical dossier.", contact.Name), nil
	case IntentPricing:
		return fmt.Sprintf("Hello %s, regarding pricing, we offer tiered enterprise plans based on orchestration volume. I'll have a technical lead follow up with specifics.", contact.Name), nil
	case IntentObjection:
		return fmt.Sprintf("I understand your concerns, %s. Many engineering teams find that Borg significantly reduces state management complexity. Would a quick technical demo help?", contact.Name), nil
	default:
		return "Thank you for your message. How can I assist you further with Borg?", nil
	}
}
