package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("hubspot api error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *HubSpotCRMClient) GetNewInteractions(ctx context.Context) ([]db.Interaction, error) {
	// HubSpot Communications API
	// We include hs_communication_sender_email to identify the contact
	url := fmt.Sprintf("%s/crm/v3/objects/communications?limit=10&properties=hs_communication_body,hs_communication_channel_type,hs_communication_sender_email", c.BaseURL)
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
		return nil, fmt.Errorf("hubspot api error (%d)", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
				Body        string `json:"hs_communication_body"`
				ChannelType string `json:"hs_communication_channel_type"`
				SenderEmail string `json:"hs_communication_sender_email"`
			} `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

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
		}
	}

	return updates, nil
}

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
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
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
		}
	}
	return nil
}

func (c *HubSpotCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=%s,%s", c.BaseURL, dealID, c.Mapping.DealStageProperty, c.Mapping.DealAmountProperty)
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
