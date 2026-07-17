package enrichment

import (
	"context"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/crm"
)

// Negative tests for enrichment/worker.go
func TestExecuteEnrichment_NilDB(t *testing.T) {
	// Should skip enrichment cycle without panic
	e := NewEnricher(nil, nil, nil)
	e.ExecuteEnrichment(context.Background())
}

func TestExecuteEnrichment_NilSource(t *testing.T) {
	mockCRM := crm.NewMockCRMClient()
	e := NewEnricher(nil, []EnrichmentSource{nil}, mockCRM)
	e.ExecuteEnrichment(context.Background())
}
