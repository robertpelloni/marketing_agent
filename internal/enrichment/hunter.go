package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type HunterSource struct {
	APIKey     string
	HTTPClient *http.Client
}

type hunterDomainResponse struct {
	Data struct {
		Emails []struct {
			Email        string  `json:"email"`
			FirstName    string  `json:"first_name"`
			LastName     string  `json:"last_name"`
			Position     string  `json:"position"`
			Seniority    string  `json:"seniority"`
			LinkedInURL  string  `json:"linkedin"`
		} `json:"emails"`
	} `json:"data"`
	Errors []string `json:"errors"`
}

func NewHunterSource(apiKey string) *HunterSource {
	return &HunterSource{APIKey: apiKey, HTTPClient: &http.Client{}}
}

func (h *HunterSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	url := fmt.Sprintf("https://api.hunter.io/v2/domain-search?domain=%s&api_key=%s", company.Domain, h.APIKey)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := h.HTTPClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	var hunterResp hunterDomainResponse
	if err := json.NewDecoder(resp.Body).Decode(&hunterResp); err != nil { return nil, err }

	var contacts []db.Contact
	for _, e := range hunterResp.Data.Emails {
		if isDecisionMaker(e.Position) {
			contacts = append(contacts, db.Contact{
				CompanyID: company.ID,
				Name: fmt.Sprintf("%s %s", e.FirstName, e.LastName),
				Role: e.Position,
				Email: e.Email,
				LinkedInURL: e.LinkedInURL,
			})
		}
	}
	return contacts, nil
}

func (h *HunterSource) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("https://api.hunter.io/v2/account?api_key=%s", h.APIKey)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := h.HTTPClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { return fmt.Errorf("failed") }
	return nil
}
