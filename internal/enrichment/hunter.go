package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type HunterSource struct {
	APIKey     string
	HTTPClient *http.Client
}

type hunterDomainResponse struct {
	Data struct {
		Emails []struct {
			Value       string  `json:"value"`
			FirstName   string  `json:"first_name"`
			LastName    string  `json:"last_name"`
			Position    string  `json:"position"`
			Confidence  float64 `json:"confidence"`
			Seniority   string  `json:"seniority"`
			Department  string  `json:"department"`
			LinkedInURL string  `json:"linkedin"`
		} `json:"emails"`
	} `json:"data"`
	Errors []string `json:"errors"`
}

func NewHunterSource(apiKey string) *HunterSource {
	return &HunterSource{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *HunterSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	if h.APIKey == "" {
		return nil, fmt.Errorf("hunter: API key not configured")
	}

	domain := cleanDomain(company.Domain)
	if domain == "" {
		return nil, fmt.Errorf("hunter: invalid domain %q", company.Domain)
	}

	slog.Info("HunterSource: Searching for contacts", "name", company.Name, "domain", domain)

	contacts, err := h.domainSearch(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("hunter: domain search failed for %s: %w", domain, err)
	}

	if len(contacts) == 0 {
		slog.Info("HunterSource: No contacts found", "domain", domain)
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

func (h *HunterSource) domainSearch(ctx context.Context, domain string) ([]db.Contact, error) {
	apiURL := fmt.Sprintf("https://api.hunter.io/v2/domain-search?domain=%s&api_key=%s&limit=10&type=personal",
		url.QueryEscape(domain), url.QueryEscape(h.APIKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hunter: API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var hunterResp hunterDomainResponse
	if err := json.NewDecoder(resp.Body).Decode(&hunterResp); err != nil {
		return nil, err
	}

	var contacts []db.Contact
	for _, e := range hunterResp.Data.Emails {
		name := e.FirstName + " " + e.LastName
		if name == " " { name = e.Value }
		contacts = append(contacts, db.Contact{
			Name:        name,
			Role:        e.Position,
			Email:       e.Value,
			LinkedInURL: e.LinkedInURL,
		})
	}
	return contacts, nil
}

func (h *HunterSource) HealthCheck(ctx context.Context) error {
	apiURL := fmt.Sprintf("https://api.hunter.io/v2/account?api_key=%s", h.APIKey)
	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	resp, err := h.HTTPClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { return fmt.Errorf("hunter health: failed") }
	return nil
}
