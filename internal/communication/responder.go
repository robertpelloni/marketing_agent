package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// LLMResponseGenerator utilizes large language models for hyper-personalized outreach.
type LLMResponseGenerator struct {
	llm llm.LLMProvider
}

// NewLLMResponseGenerator creates a new generator instance.
func NewLLMResponseGenerator(provider llm.LLMProvider) *LLMResponseGenerator {
	return &LLMResponseGenerator{llm: provider}
}

// Generate creates a tailored response using the technical dossier and conversational context.
func (g *LLMResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	log.Printf("LLMResponseGenerator: Generating personalized response for intent: %s", salesCtx.LatestIntent)

	latestMsg := "START_OUTREACH"
	if len(salesCtx.Interactions) > 0 {
		latestMsg = salesCtx.Interactions[0].RawText
	}

	prompt := llm.Prompt{
		System: "You are an expert sales engineer at Borg, a multi-agent LLM orchestration platform. Your goal is to provide hyper-personalized outreach based on the prospect's technical findings.",
		User: fmt.Sprintf("Draft a reply to %s (%s) at %s. Context: %s. Technical Findings: %s. Latest Message: %s. Action: %s.",
			salesCtx.Contact.Name, salesCtx.Contact.Role, salesCtx.Company.Name, salesCtx.LatestIntent, salesCtx.Deal.TechnicalDossier, latestMsg, action),
	}

	return g.llm.Generate(ctx, prompt)
}

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
