package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestIntentClassifier_Classify(t *testing.T) {
	classifier := &MockIntentClassifier{}
	intent, err := classifier.Classify(context.Background(), "How much does Borg cost?")
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}
	if intent != IntentPricing {
		t.Errorf("Expected IntentPricing, got %s", intent)
	}
}

func TestSalesEngine_Decide(t *testing.T) {
	engine := NewLearningSalesEngine(nil) // Mock DB not needed for pure logic
	salesCtx := SalesContext{
		LatestIntent: IntentPricing,
		Company: db.Company{
			MarketCapTier: "Enterprise",
		},
		Deal: db.Deal{
			CurrentState: db.StateResearched,
		},
	}

	action, err := engine.Decide(context.Background(), salesCtx)
	if err != nil {
		t.Fatalf("Decision failed: %v", err)
	}
	if action != ActionRespond {
		t.Errorf("Expected ActionRespond for Enterprise pricing intent, got %s", action)
	}
}
