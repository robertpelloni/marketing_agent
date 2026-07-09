package communication_test

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/communication"
	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

func TestTechnicalAuthority_Routing(t *testing.T) {
	// Initialize test DB if available or mock RAG generator directly
	ctx := context.Background()

	// 1. Create a company & contact
	company := &db.Company{
		Name:          "Code Ninjas",
		Domain:        "codeninjas.dev",
		MarketCapTier: "Startup",
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateOutreachSent,
		TechnicalDossier: "Frequent GitHub contributor. Highly skeptical of black-box orchestration frameworks.",
	}

	contact := &db.Contact{
		CompanyID:        company.ID,
		Name:             "Ninja Dev",
		Email:            "ninja@codeninjas.dev",
		GitHubHandle:     "ninjadev",
		PreferredChannel: "github",
	}

	mockLLM := &llm.MockLLMProvider{}
	responder := communication.NewRAGResponseGenerator(nil, mockLLM)
	engine := communication.NewLearningSalesEngine(nil, nil, mockLLM)

	salesCtx := communication.SalesContext{
		Company: *company,
		Contact: *contact,
		Deal:    *deal,
		LatestIntent: communication.IntentTechnical,
		Interactions: []db.Interaction{{Direction: "Inbound"}},
	}

	// 2. Generate response
	resp, err := responder.Generate(ctx, salesCtx, communication.ActionRespond)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(resp, "TormentNexus") {
		t.Errorf("Expected TormentNexus persona for technical github contact, got %s", resp)
	}

	action, _ := engine.Decide(ctx, salesCtx)
	if action != communication.ActionRespond {
		t.Errorf("Expected engine to decide on ActionRespond for a technical question, got %s", action)
	}

	route := engine.RouteLead(salesCtx)
	if route != "Technical Sales Engineer" {
		t.Errorf("Expected route Technical Sales Engineer, got %s", route)
	}
}
