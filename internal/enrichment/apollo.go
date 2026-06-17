package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type ApolloSource struct {
	APIKey     string
	HTTPClient *http.Client
}

type apolloSearchRequest struct {
	QOrganizationDomains []string `json:"q_organization_domains"`
	PersonTitles         []string `json:"person_titles,omitempty"`
	Page                 int      `json:"page"`
	PerPage              int      `json:"per_page"`
	RevealPhone          bool     `json:"reveal_phone"`
	RevealPersonalEmails bool     `json:"reveal_personal_emails"`
	RequireEmail         bool     `json:"require_email"`
}

type apolloSearchResponse struct {
	People       []apolloPerson `json:"people"`
	TotalEntries int            `json:"total_entries"`
	Paginated    bool           `json:"paginated"`
}

type apolloPerson struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Email        string `json:"email"`
	LinkedInURL  string `json:"linkedin_url"`
	Organization struct {
		Name string `json:"name"`
	} `json:"organization"`
	Seniority string `json:"seniority"`
}

func NewApolloSource(apiKey string) *ApolloSource {
	return &ApolloSource{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *ApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if a.APIKey == "" {
		return nil, fmt.Errorf("apollo: API key not configured")
	}

	domain := cleanDomain(company.Domain)
	if domain == "" {
		return nil, fmt.Errorf("apollo: invalid domain %q", company.Domain)
	}

	slog.Info("ApolloSource: Searching for contacts", "name", company.Name, "domain", domain)

	contacts, err := a.peopleSearch(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("apollo: people search failed for %s: %w", domain, err)
	}

	if len(contacts) == 0 {
		slog.Info("ApolloSource: No contacts found", "domain", domain)
		return nil, nil
	}

	var filtered []db.Contact
	for _, c := range contacts {
		if isDecisionMaker(c.Role) {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 && len(contacts) > 0 {
		filtered = contacts[:1]
	}

	return filtered, nil
}

func (a *ApolloSource) peopleSearch(ctx context.Context, domain string) ([]db.Contact, error) {
	reqBody := apolloSearchRequest{
		QOrganizationDomains: []string{domain},
		PersonTitles: []string{
			"CTO", "VP", "Director", "Head of", "Principal", "Lead", "Architect",
		},
		Page:         1,
		PerPage:      25,
		RequireEmail: true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.apollo.io/api/v1/people/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("apollo: API returned HTTP %d", resp.StatusCode)
	}

	var result apolloSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var contacts []db.Contact
	for _, p := range result.People {
		if p.Email == "" { continue }
		name := p.Name
		if name == "" { name = p.FirstName + " " + p.LastName }
		contacts = append(contacts, db.Contact{
			Name:        name,
			Role:        p.Title,
			Email:       p.Email,
			LinkedInURL: p.LinkedInURL,
		})
	}

	return contacts, nil
}

func (a *ApolloSource) HealthCheck(ctx context.Context) error {
	if a.APIKey == "" {
		return fmt.Errorf("apollo: API key not configured")
	}
	_, err := a.peopleSearch(ctx, "openai.com")
	return err
}

type MockApolloSource struct{}
func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	return []db.Contact{{Name: "Sarah Chen", Role: "CTO", Email: "sarah@example.com"}}, nil
}
func (m *MockApolloSource) HealthCheck(ctx context.Context) error { return nil }
