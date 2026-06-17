package enrichment

import (
	"context"
	"net/http"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type ApolloSource struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewApolloSource(apiKey string) *ApolloSource {
	return &ApolloSource{APIKey: apiKey, HTTPClient: &http.Client{}}
}

func (a *ApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	return nil, nil // Real implementation would go here
}

type MockApolloSource struct{}
func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	return []db.Contact{{Name: "Sarah Chen", Role: "CTO", Email: "sarah@example.com"}}, nil
}

func (a *ApolloSource) HealthCheck(ctx context.Context) error {
	return nil
}
