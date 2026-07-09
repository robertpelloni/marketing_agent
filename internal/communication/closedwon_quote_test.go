package communication_test

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/communication"
	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

func TestClosedWonQuote_Injection(t *testing.T) {
	ctx := context.Background()

	company := &db.Company{
		Name:          "Won Corp",
		Domain:        "won.corp",
		MarketCapTier: "Enterprise", // Forces an expected pricing calculation
	}

	deal := &db.Deal{
		ID:           42,
		CompanyID:    company.ID,
		CurrentState: db.StateClosedWon,
	}

	contact := &db.Contact{
		CompanyID: company.ID,
		Name:      "Winner",
		Email:     "win@won.corp",
	}

	mockLLM := &llm.MockLLMProvider{}
	responder := communication.NewRAGResponseGenerator(nil, mockLLM)

	salesCtx := communication.SalesContext{
		Company: *company,
		Contact: *contact,
		Deal:    *deal,
		LatestIntent: communication.IntentPricing,
	}

	resp, err := responder.Generate(ctx, salesCtx, communication.ActionRespond)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 1. Verify LLM Provider received the request
	if !strings.Contains(resp, "MOCK LLM RESPONSE") {
		t.Errorf("Expected mock LLM response, got: %s", resp)
	}

	// 2. Verify pricing logic was executed and injected
	if !strings.Contains(resp, "Pricing Context: Annual subscription is approximately $") {
		t.Errorf("Expected pricing context in prompt, got: %s", resp)
	}
}
