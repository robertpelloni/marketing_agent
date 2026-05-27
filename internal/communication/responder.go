package communication

import (
	"context"
	"fmt"
	"log"
)

// MockResponseGenerator provides template-based replies for testing.
type MockResponseGenerator struct{}

func (m *MockResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	log.Printf("MockResponseGenerator: Generating response for intent: %s, action: %s", salesCtx.LatestIntent, action)

	if action == ActionAdvanceState {
		return fmt.Sprintf("Hi %s, based on our conversation, I'm preparing a formal proposal for you.", salesCtx.Contact.Name), nil
	}

	switch salesCtx.LatestIntent {
	case IntentTechnical:
		return fmt.Sprintf("Hi %s, that's a great technical question. Borg uses a Go-based headless daemon for multi-agent LLM coordination. You can find more details in our technical dossier.", salesCtx.Contact.Name), nil
	case IntentPricing:
		return fmt.Sprintf("Hello %s, regarding pricing, we offer tiered enterprise plans based on orchestration volume. I'll have a technical lead follow up with specifics.", salesCtx.Contact.Name), nil
	case IntentObjection:
		return fmt.Sprintf("I understand your concerns, %s. Many engineering teams find that Borg significantly reduces state management complexity. Would a quick technical demo help?", salesCtx.Contact.Name), nil
	case IntentMeetingRequest:
		return fmt.Sprintf("I'd be happy to schedule a call, %s. I'll send over a calendar invite so we can dive deeper into how Borg can optimize your LLM orchestration.", salesCtx.Contact.Name), nil
	case IntentFollowUp:
		return fmt.Sprintf("Thanks for following up, %s. I'm currently reviewing the technical requirements we discussed and will have a detailed proposal for you shortly.", salesCtx.Contact.Name), nil
	default:
		return "Thank you for your message. How can I assist you further with Borg?", nil
	}
}
