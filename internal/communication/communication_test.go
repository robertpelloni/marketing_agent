package communication

import (
	"context"
	"testing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestIntentClassifier_Classify(t *testing.T) {
	classifier := &MockIntentClassifier{}
	intent, _ := classifier.Classify(context.Background(), "How much does TormentNexus cost?")
	if intent != IntentPricing { t.Errorf("Expected IntentPricing, got %s", intent) }
}

func TestSalesEngine_Decide(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, nil)
	t.Run("Enterprise Pricing Intent", func(t *testing.T) {
		salesCtx := SalesContext{LatestIntent: IntentPricing, Company: db.Company{MarketCapTier: "Enterprise"}, Deal: db.Deal{ID: 1, CurrentState: db.StateResearched}}
		action, _ := engine.Decide(context.Background(), salesCtx)
		if action != ActionRespond { t.Errorf("Expected ActionRespond, got %s", action) }
	})
	t.Run("Should Advance State", func(t *testing.T) {
		// Enterprise (50) + 10 interactions (20) = 70. QualifyLead = 70/2 + 20 = 55.
		// Let's just give it a technical dossier with "bottleneck" (30)
		salesCtx := SalesContext{
			LatestIntent: IntentMeetingRequest,
			Interactions: []db.Interaction{{Direction: "Inbound"}, {Direction: "Inbound"}, {Direction: "Inbound"}},
			Deal: db.Deal{ID: 2, CurrentState: db.StateEngaged, TechnicalDossier: "bottleneck detected"},
			Company: db.Company{MarketCapTier: "Enterprise"},
		}
		// Score: 50 (Ent) + 30 (Dossier) + 6 (Inter) = 86.
		// Qual: 86/2 (43) + 20 (Inbound > 2) + 25 (MeetingRequest) = 88.
		action, _ := engine.Decide(context.Background(), salesCtx)
		if action != ActionAdvanceState { t.Errorf("Expected ActionAdvanceState, got %s", action) }
	})
}
