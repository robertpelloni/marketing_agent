package communication

import (
	"context"
	"log"
)

// MockIntentClassifier simulates intent classification for inbound text.
type MockIntentClassifier struct{}

func (m *MockIntentClassifier) Classify(ctx context.Context, text string) (Intent, error) {
	log.Printf("MockIntentClassifier: Classifying text: %s", text)

	// Simple heuristic-based mock classification
	if containsAny(text, "pricing", "cost", "license") {
		return IntentPricing, nil
	}
	if containsAny(text, "architecture", "scaling", "how does it work") {
		return IntentTechnical, nil
	}
	if containsAny(text, "no", "not interested", "stop") {
		return IntentObjection, nil
	}

	return IntentTechnical, nil
}

func containsAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if contains(text, kw) {
			return true
		}
	}
	return false
}

func contains(text, kw string) bool {
	// Simple case-insensitive search logic would go here
	return true // Simplified for mock
}
