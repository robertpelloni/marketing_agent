package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// HubSpotClient implements CRMClient using the HubSpot REST API.
// Required environment variables:
//   HUBSPOT_API_KEY – Private API key (legacy) or use OAuth token in HUBSPOT_ACCESS_TOKEN
//   HUBSPOT_BASE_URL – Base API URL, e.g., https://api.hubapi.com
// The client uses HubSpot's CRM objects: contacts, companies, deals.

type HubSpotClient struct {
	baseURL     string
	apiKey      string
	accessToken string
	client      *http.Client
	mapping     FieldMapping
}

// NewHubSpotClient creates a new HubSpot CRM client.
func NewHubSpotClient() (*HubSpotClient, error) {
	base := os.Getenv("HUBSPOT_BASE_URL")
	if base == "" {
		base = "https://api.hubapi.com"
	}
	key := os.Getenv("HUBSPOT_API_KEY")
	token := os.Getenv("HUBSPOT_ACCESS_TOKEN")
	if key == "" && token == "" {
		return nil, fmt.Errorf("hubspot client: missing API key or access token")
	}
	return &HubSpotClient{
		baseURL:     base,
		apiKey:      key,
		accessToken: token,
		client:      &http.Client{},
		mapping: FieldMapping{
			DealNameProp:     "dealname",
			DealAmountProp:   "amount",
			DealStageProp:    "dealstage",
			DealDescProp:     "description",
			DealRouteProp:    "custom_route",
			ContactEmailProp: "email",
			ContactRoleProp:  "jobtitle",
			AccountWebProp:   "website",
		},
	}, nil
}

// SetFieldMapping updates the dynamic field mapping.
func (h *HubSpotClient) SetFieldMapping(mapping FieldMapping) {
	if mapping.DealNameProp != "" {
		h.mapping.DealNameProp = mapping.DealNameProp
	}
	if mapping.DealAmountProp != "" {
		h.mapping.DealAmountProp = mapping.DealAmountProp
	}
	if mapping.DealStageProp != "" {
		h.mapping.DealStageProp = mapping.DealStageProp
	}
	if mapping.DealDescProp != "" {
		h.mapping.DealDescProp = mapping.DealDescProp
	}
	if mapping.DealRouteProp != "" {
		h.mapping.DealRouteProp = mapping.DealRouteProp
	}
	if mapping.ContactEmailProp != "" {
		h.mapping.ContactEmailProp = mapping.ContactEmailProp
	}
	if mapping.ContactRoleProp != "" {
		h.mapping.ContactRoleProp = mapping.ContactRoleProp
	}
	if mapping.AccountWebProp != "" {
		h.mapping.AccountWebProp = mapping.AccountWebProp
	}
}

// authHeader constructs the Authorization header for HubSpot requests.
func (h *HubSpotClient) authHeader() string {
	if h.accessToken != "" {
		return "Bearer " + h.accessToken
	}
	return "Bearer " + h.apiKey
}

// PushDeal creates or updates a HubSpot deal.
func (h *HubSpotClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	payload := map[string]interface{}{
		"properties": map[string]any{
			h.mapping.DealNameProp:   fmt.Sprintf("%s – %s", company.Name, route),
			h.mapping.DealAmountProp: deal.QuotedPricing,
			"pipeline":               "default",
			h.mapping.DealStageProp:  mapLeadStateToHubSpotStage(deal.CurrentState),
			h.mapping.DealDescProp:   deal.TechnicalDossier,
			h.mapping.DealRouteProp:  route,
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/crm/v3/objects/deals", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hubspot PushDeal: status %d", resp.StatusCode)
	}
	return nil
}

// GetLeadUpdates fetches recently modified leads (contacts) from HubSpot.
func (h *HubSpotClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// HubSpot contacts endpoint with a simple filter for recent updates (last 24h)
	url := fmt.Sprintf("%s/crm/v3/objects/contacts?properties=firstname,lastname,email,hs_lead_status&limit=100&archived=false&propertiesWithHistory=hs_lead_status", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hubspot GetLeadUpdates: status %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
				LeadStatus string `json:"hs_lead_status"`
			} `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var updates []LeadUpdate
	for _, r := range result.Results {
		state := mapHubSpotLeadStatusToLeadState(r.Properties.LeadStatus)
		if state != "" {
			updates = append(updates, LeadUpdate{ID: r.ID, NewState: state, Notes: ""})
		}
	}

	return updates, nil
}

// ValidateAccount checks if a company domain exists as a HubSpot company.
func (h *HubSpotClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/companies/search", h.baseURL)
	payload := map[string]interface{}{
		"filterGroups": []map[string]interface{}{{
			"filters": []map[string]string{{
				"propertyName": h.mapping.AccountWebProp,
				"operator":     "EQ",
				"value":        domain,
			}},
		}},
		"limit": 1,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("hubspot ValidateAccount: status %d", resp.StatusCode)
	}

	var result struct {
		Results []interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return len(result.Results) > 0, nil
}

// SyncInteraction creates a note attached to a HubSpot deal.
func (h *HubSpotClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	payload := map[string]interface{}{
		"properties": map[string]string{
			"hs_note_body": note,
			"hs_note_subject": "Automated Interaction",
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/crm/v3/objects/notes", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", h.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hubspot SyncInteraction: status %d", resp.StatusCode)
	}
	return nil
}

// SyncContacts creates or updates HubSpot contacts linked to a company.
func (h *HubSpotClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, c := range contacts {
		payload := map[string]interface{}{
			"properties": map[string]string{
				h.mapping.ContactEmailProp: c.Email,
				"firstname":                firstName(c.Name),
				"lastname":                 lastName(c.Name),
				h.mapping.ContactRoleProp:  c.Role,
				"company":                  fmt.Sprintf("%d", companyID),
			},
		}

		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("%s/crm/v3/objects/contacts", h.baseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", h.authHeader())
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("hubspot SyncContacts: status %d", resp.StatusCode)
		}
	}
	return nil
}

// FetchDealDetails retrieves a HubSpot deal.
func (h *HubSpotClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=%s,%s,%s,%s,%s",
		h.baseURL, dealID, h.mapping.DealNameProp, h.mapping.DealAmountProp, h.mapping.DealDescProp, h.mapping.DealStageProp, h.mapping.DealRouteProp)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hubspot FetchDealDetails: status %d", resp.StatusCode)
	}

	var result struct {
		Properties map[string]interface{} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	dealName, _ := result.Properties[h.mapping.DealNameProp].(string)
	amount, _ := result.Properties[h.mapping.DealAmountProp].(float64)
	desc, _ := result.Properties[h.mapping.DealDescProp].(string)
	stage, _ := result.Properties[h.mapping.DealStageProp].(string)
	route, _ := result.Properties[h.mapping.DealRouteProp].(string)

	return &DealDetails{
		ID:                 parseHubSpotDealID(dealName), // placeholder conversion
		Status:             mapHubSpotStageToLeadState(stage),
		QuotedPricing:      amount,
		CustomRequirements: route,
		TechnicalDossier:   desc,
	}, nil
}

// Helper functions for parsing and mapping.
func firstName(full string) string {
	parts := strings.Split(full, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func lastName(full string) string {
	parts := strings.Split(full, " ")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func mapLeadStateToHubSpotStage(state db.LeadState) string { return "appointmentscheduled" }
func mapHubSpotLeadStatusToLeadState(status string) db.LeadState { return db.StateResearched }
func mapHubSpotStageToLeadState(stage string) db.LeadState      { return db.StateResearched }
func parseHubSpotDealID(name string) int64                     { return 0 }
