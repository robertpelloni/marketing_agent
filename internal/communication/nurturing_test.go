package communication_test

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/communication"
	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

func TestNurturingPrompt_Constraints(t *testing.T) {
	var database *db.DB


	ctx := context.Background()

	company := &db.Company{
		Name:          "Nurture Corp",
		Domain:        "nurture.corp",
		MarketCapTier: "Mid-Market",
	}
	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateEngaged,
		TechnicalDossier: "Evaluating open-source alternatives to current cloud solutions.",
	}
	contact := &db.Contact{
		CompanyID:        company.ID,
		Name:             "Nurture Lead",
		Email:            "lead@nurture.corp",
	}
	mockLLM := &llm.MockLLMProvider{}
	responder := communication.NewRAGResponseGenerator(database, mockLLM)

	salesCtx := communication.SalesContext{
		Company: *company,
		Contact: *contact,
		Deal:    *deal,
		LatestIntent: communication.IntentFollowUp,
	}

	resp, err := responder.Generate(ctx, salesCtx, communication.ActionRespond)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(resp, "CRITICAL CONSTRAINTS: Keep the response extremely brief") {
		t.Errorf("Expected prompt to contain constraints, got %s", resp)
	}

	if !strings.Contains(resp, "absolutely DO NOT use any bracketed placeholders") {
		t.Errorf("Expected prompt to forbid bracketed placeholders, got %s", resp)
	}
}
