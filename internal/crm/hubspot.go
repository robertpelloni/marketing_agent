package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
<<<<<<< HEAD
	"net/http"
	"os"
	"strings"
=======
	"io"
	"net/http"
	"strconv"
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

<<<<<<< HEAD
// HubSpotClient implements CRMClient using the HubSpot REST API.
// Required environment variables:
//   HUBSPOT_API_KEY – Private API key (legacy) or use OAuth token in HUBSPOT_ACCESS_TOKEN
//   HUBSPOT_BASE_URL – Base API URL, e.g., https://api.hubapi.com
// The client uses HubSpot's CRM objects: contacts, companies, deals.

type HubSpotClient struct {
	baseURL    string
	apiKey     string
	accessToken string
	client     *http.Client
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
	return &HubSpotClient{baseURL: base, apiKey: key, accessToken: token, client: &http.Client{}}, nil
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
			"dealname":           fmt.Sprintf("%s – %s", company.Name, route),
			"amount":             deal.QuotedPricing,
			"pipeline":           "default",
			"dealstage":          mapLeadStateToHubSpotStage(deal.CurrentState),
			"description":        deal.TechnicalDossier,
			"custom_route":       route,
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
=======
// HubSpotCRMClient implements CRMClient for HubSpot.
type HubSpotCRMClient struct {
	AccessToken string
	BaseURL     string
	HTTPClient  *http.Client
	Mapping     FieldMapping
}

// NewHubSpotCRMClient creates a new HubSpot CRM client.
func NewHubSpotCRMClient(accessToken string) *HubSpotCRMClient {
	return &HubSpotCRMClient{
		AccessToken: accessToken,
		BaseURL:     "https://api.hubapi.com",
		HTTPClient:  &http.Client{},
		Mapping: FieldMapping{
			DealNameProperty:     "dealname",
			DealStageProperty:    "dealstage",
			DealAmountProperty:   "amount",
			DealDossierProperty:  "technical_dossier",
			ContactEmailProperty: "email",
		},
	}
}

func (c *HubSpotCRMClient) SetFieldMapping(mapping FieldMapping) {
	if mapping.DealNameProperty != "" {
		c.Mapping.DealNameProperty = mapping.DealNameProperty
	}
	if mapping.DealStageProperty != "" {
		c.Mapping.DealStageProperty = mapping.DealStageProperty
	}
	if mapping.DealAmountProperty != "" {
		c.Mapping.DealAmountProperty = mapping.DealAmountProperty
	}
	if mapping.DealDossierProperty != "" {
		c.Mapping.DealDossierProperty = mapping.DealDossierProperty
	}
	if mapping.ContactEmailProperty != "" {
		c.Mapping.ContactEmailProperty = mapping.ContactEmailProperty
	}
}

func (c *HubSpotCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	// 1. Search for existing deal with this ID (HubSpot ID or custom property)
	// For this implementation, we try to UPDATE if dealID > 0, otherwise CREATE
	// Real-world would likely use a search API with a custom 'external_id' property
	method := "POST"
	url := fmt.Sprintf("%s/crm/v3/objects/deals", c.BaseURL)
	if deal.ID > 1000 { // Heuristic for existing CRM IDs vs internal increments
		method = "PATCH"
		url = fmt.Sprintf("%s/crm/v3/objects/deals/%d", c.BaseURL, deal.ID)
	}

	properties := map[string]string{
		c.Mapping.DealNameProperty:    fmt.Sprintf("%s - %d", company.Name, deal.ID),
		c.Mapping.DealStageProperty:   string(deal.CurrentState),
		c.Mapping.DealAmountProperty:  fmt.Sprintf("%.2f", deal.QuotedPricing),
		c.Mapping.DealDossierProperty: deal.TechnicalDossier,
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"properties": properties,
	})

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
<<<<<<< HEAD
		return fmt.Errorf("hubspot PushDeal: status %d", resp.StatusCode)
	}
	return nil
}

// GetLeadUpdates fetches recently modified leads (contacts) from HubSpot.
func (h *HubSpotClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// HubSpot contacts endpoint with a simple filter for recent updates (last 24h)
	url := fmt.Sprintf("%s/crm/v3/objects/contacts?properties=firstname,lastname,email,hs_lead_status&limit=100&archived=false&propertiesWithHistory=hs_lead_status", h.baseURL)
=======
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *HubSpotCRMClient) GetNewInteractions(ctx context.Context) ([]db.Interaction, error) {
	// HubSpot Communications API
	// We include hs_communication_sender_email to identify the contact
	url := fmt.Sprintf("%s/crm/v3/objects/communications?limit=10&properties=hs_communication_body,hs_communication_channel_type,hs_communication_sender_email", c.BaseURL)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
=======
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
<<<<<<< HEAD
		return nil, fmt.Errorf("hubspot GetLeadUpdates: status %d", resp.StatusCode)
=======
		return nil, fmt.Errorf("hubspot api error (%d)", resp.StatusCode)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	}

	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
<<<<<<< HEAD
				LeadStatus string `json:"hs_lead_status"`
=======
				Body        string `json:"hs_communication_body"`
				ChannelType string `json:"hs_communication_channel_type"`
				SenderEmail string `json:"hs_communication_sender_email"`
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
			} `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

<<<<<<< HEAD
	var updates []LeadUpdate
	for _, r := range result.Results {
		state := mapHubSpotLeadStatusToLeadState(r.Properties.LeadStatus)
		if state != "" {
			updates = append(updates, LeadUpdate{ID: r.ID, NewState: state, Notes: ""})
=======
	interactions := make([]db.Interaction, len(result.Results))
	for i, r := range result.Results {
		interactions[i] = db.Interaction{
			RawText: r.Properties.Body,
			Channel: r.Properties.ChannelType,
			Summary: r.Properties.SenderEmail, // Store sender email in Summary for identification
		}
	}

	return interactions, nil
}

func (c *HubSpotCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// Simplified: Polling recent deals for status changes
	url := fmt.Sprintf("%s/crm/v3/objects/deals?limit=10&properties=%s", c.BaseURL, c.Mapping.DealStageProperty)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []struct {
			ID         string            `json:"id"`
			Properties map[string]string `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	updates := make([]LeadUpdate, len(result.Results))
	for i, r := range result.Results {
		updates[i] = LeadUpdate{
			ID:       r.ID,
			NewState: db.LeadState(r.Properties[c.Mapping.DealStageProperty]),
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
		}
	}

	return updates, nil
}

<<<<<<< HEAD
// ValidateAccount checks if a company domain exists as a HubSpot company.
func (h *HubSpotClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/companies/search", h.baseURL)
	payload := map[string]interface{}{
		"filterGroups": []map[string]interface{}{{
			"filters": []map[string]string{{
				"propertyName": "website",
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
=======
func (c *HubSpotCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/companies/search", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"filterGroups": []interface{}{
			map[string]interface{}{
				"filters": []interface{}{
					map[string]interface{}{
						"propertyName": "domain",
						"operator":     "EQ",
						"value":        domain,
					},
				},
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
<<<<<<< HEAD
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
=======
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return false, fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Total int `json:"total"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Total > 0, nil
}

func (c *HubSpotCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	// 1. Create the Note
	url := fmt.Sprintf("%s/crm/v3/objects/notes", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"properties": map[string]string{
			"hs_note_body": note,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
<<<<<<< HEAD
		return fmt.Errorf("hubspot SyncInteraction: status %d", resp.StatusCode)
	}
	return nil
}

// SyncContacts creates or updates HubSpot contacts linked to a company.
func (h *HubSpotClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, c := range contacts {
		payload := map[string]interface{}{
			"properties": map[string]string{
				"email":       c.Email,
				"firstname":   firstName(c.Name),
				"lastname":    lastName(c.Name),
				"jobtitle":    c.Role,
				"company":     fmt.Sprintf("%d", companyID),
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
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("hubspot SyncContacts: status %d", resp.StatusCode)
=======
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	var noteResult struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&noteResult); err != nil {
		return err
	}

	// 2. Associate Note with Deal (Deal ID 204 in HubSpot Note associations)
	assocURL := fmt.Sprintf("%s/crm/v3/associations/notes/deals/batch/create", c.BaseURL)
	assocPayload, _ := json.Marshal(map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{
				"from": map[string]string{"id": noteResult.ID},
				"to":   map[string]string{"id": strconv.FormatInt(dealID, 10)},
				"type": "note_to_deal",
			},
		},
	})

	assocReq, err := http.NewRequestWithContext(ctx, "POST", assocURL, bytes.NewBuffer(assocPayload))
	if err != nil {
		return err
	}
	assocReq.Header.Set("Authorization", "Bearer "+c.AccessToken)
	assocReq.Header.Set("Content-Type", "application/json")

	assocResp, err := c.HTTPClient.Do(assocReq)
	if err == nil {
		assocResp.Body.Close()
	}

	return nil
}

func (c *HubSpotCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, contact := range contacts {
		url := fmt.Sprintf("%s/crm/v3/objects/contacts", c.BaseURL)
		payload, _ := json.Marshal(map[string]interface{}{
			"properties": map[string]string{
				c.Mapping.ContactEmailProperty: contact.Email,
				"firstname":                    contact.Name,
			},
		})

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err == nil {
			resp.Body.Close()
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
		}
	}
	return nil
}

<<<<<<< HEAD
// FetchDealDetails retrieves a HubSpot deal.
func (h *HubSpotClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=dealname,amount,description,hs_deal_stage,custom_route", h.baseURL, dealID)
=======
func (c *HubSpotCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=%s,%s", c.BaseURL, dealID, c.Mapping.DealStageProperty, c.Mapping.DealAmountProperty)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
	req.Header.Set("Authorization", h.authHeader())

	resp, err := h.client.Do(req)
=======
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

<<<<<<< HEAD
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hubspot FetchDealDetails: status %d", resp.StatusCode)
	}

	var result struct {
		Properties struct {
			DealName      string  `json:"dealname"`
			Amount        float64 `json:"amount"`
			Description   string  `json:"description"`
			Stage         string  `json:"hs_deal_stage"`
			CustomRoute   string  `json:"custom_route"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &DealDetails{
		ID:                 parseHubSpotDealID(result.Properties.DealName), // placeholder conversion
		Status:             mapHubSpotStageToLeadState(result.Properties.Stage),
		QuotedPricing:      result.Properties.Amount,
		CustomRequirements: result.Properties.CustomRoute,
		TechnicalDossier:   result.Properties.Description,
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
=======
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	var r struct {
		ID         string            `json:"id"`
		Properties map[string]string `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	id, _ := strconv.ParseInt(r.ID, 10, 64)
	amount, _ := strconv.ParseFloat(r.Properties[c.Mapping.DealAmountProperty], 64)

	return &DealDetails{
		ID:            id,
		Status:        db.LeadState(r.Properties[c.Mapping.DealStageProperty]),
		QuotedPricing: amount,
	}, nil
}

func (c *HubSpotCRMClient) SendEmail(ctx context.Context, contact db.Contact, subject, body string) error {
	// HubSpot Engagements API (Communication type)
	url := fmt.Sprintf("%s/crm/v3/objects/communications", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"properties": map[string]string{
			"hs_communication_channel_type": "EMAIL",
			"hs_communication_logged_from":  "CRM",
			"hs_communication_body":         body,
			"hs_communication_subject":      subject,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
