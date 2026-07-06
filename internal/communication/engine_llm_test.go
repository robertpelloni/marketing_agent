package communication

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

func TestRAGResponseGenerator_Generate(t *testing.T) {
	// Create a mock LLM provider that returns predictable responses
	mockLLM := &llm.MockLLMProvider{}

	generator := NewRAGResponseGenerator(nil, mockLLM)

	salesCtx := SalesContext{
		Company: db.Company{
			Name: "Acme Corp",
			MarketCapTier: "Enterprise",
		},
		Contact: db.Contact{
			Name: "John Doe",
		},
		Deal: db.Deal{
			CurrentState: db.StateOutreachSent,
			TechnicalDossier: "They are using a legacy monolithic architecture and facing scaling issues.",
		},
		LatestIntent: IntentTechnical,
	}

	// Test Generate method with a question intent
	resp, err := generator.Generate(context.Background(), salesCtx, ActionRespond)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp == "" {
		t.Error("Expected non-empty response")
	}

	// The mock LLM provider includes the prompt in its output
	if !strings.Contains(resp, "MOCK LLM RESPONSE") {
		t.Errorf("Expected response to contain 'MOCK LLM RESPONSE', got: %s", resp)
	}

	// The prompt should contain context about the intent
	if !strings.Contains(resp, string(IntentTechnical)) {
		t.Errorf("Expected response to contain intent string, got: %s", resp)
	}
}
