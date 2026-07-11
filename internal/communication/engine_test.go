package communication

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)


func TestScoreLead(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, nil)

	ctx := SalesContext{
		Company: db.Company{MarketCapTier: "Enterprise"},
		Deal:    db.Deal{TechnicalDossier: "Contains a BOTTLENECK in scaling."},
		Interactions: []db.Interaction{
			{Direction: "Inbound"},
			{Direction: "Outbound"},
		},
	}

	score := engine.ScoreLead(ctx)
	// 50 (Enterprise) + 30 (Bottleneck) + 4 (2 interactions) = 84
	if score != 84 {
		t.Errorf("Expected score 84, got %d", score)
	}
}

func TestQualifyLead(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, nil)

	ctx := SalesContext{
		Company: db.Company{MarketCapTier: "Enterprise"},
		Deal:    db.Deal{TechnicalDossier: "BOTTLENECK detected."},
		Interactions: []db.Interaction{
			{Direction: "Inbound"},
			{Direction: "Inbound"},
			{Direction: "Inbound"},
		},
		LatestIntent: IntentMeetingRequest,
	}

	qual := engine.QualifyLead(ctx)
	// ScoreLead: 50 (Enterprise) + 30 (Bottleneck) + 6 (3 interactions) = 86
	// Base Qual: 86 / 2 = 43
	// Inbound > 2: +20
	// MeetingRequest: +25
	// Total: 88
	if qual != 88 {
		t.Errorf("Expected qualification 88, got %d", qual)
	}
}

func TestDecide_AdvanceToWon(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, nil)
	ctx := context.Background()

	salesCtx := SalesContext{
		Company: db.Company{MarketCapTier: "Enterprise"},
		Deal: db.Deal{
			CurrentState:     db.StateNegotiating,
			TechnicalDossier: "BOTTLENECK detected.",
		},
		Interactions: []db.Interaction{
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
			{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"},
		},
		LatestIntent: IntentFollowUp,
	}

	qual := engine.QualifyLead(salesCtx)
	t.Logf("Qualification: %d", qual)

	// QualifyLead should be >= 80
	action, err := engine.Decide(ctx, &salesCtx)
	if err != nil {
		t.Fatalf("Decide failed: %v", err)
	}

	if action != ActionAdvanceState {
		t.Errorf("Expected ActionAdvanceState, got %s", action)
	}
}

func TestDecide_RAGIngestion(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, &llm.MockLLMProvider{})
	ctx := context.Background()

	salesCtx := SalesContext{
		Company: db.Company{MarketCapTier: "Enterprise"},
		Deal: db.Deal{
			ID:               1,
			CurrentState:     db.StateEngaged,
			TechnicalDossier: "Initial Dossier.",
		},
		Interactions: []db.Interaction{},
		LatestIntent: IntentPricing,
	}

	_, err := engine.Decide(ctx, &salesCtx)
	if err != nil {
		t.Fatalf("Decide failed: %v", err)
	}

	if !strings.Contains(salesCtx.Deal.TechnicalDossier, "[RAG Ingestion]") {
		t.Errorf("Expected RAG Ingestion in TechnicalDossier, got: %s", salesCtx.Deal.TechnicalDossier)
	}
}
