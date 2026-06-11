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
}

// NewHubSpotCRMClient creates a new HubSpot CRM client.
func NewHubSpotCRMClient(accessToken string) *HubSpotCRMClient {
	return &HubSpotCRMClient{
		AccessToken: accessToken,
		BaseURL:     "https://api.hubapi.com",
		HTTPClient:  &http.Client{},
	}
}

func (c *HubSpotCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	url := fmt.Sprintf("%s/crm/v3/objects/deals", c.BaseURL)
	properties := map[string]string{
		"dealname":          fmt.Sprintf("%s - %d", company.Name, deal.ID),
		"dealstage":         string(deal.CurrentState), // Mapping logic needed for real stages
		"amount":            fmt.Sprintf("%.2f", deal.QuotedPricing),
		"technical_dossier": deal.TechnicalDossier,
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"properties": properties,
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

	return nil
}

func (c *HubSpotCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// Simplified: Polling recent deals for status changes
	url := fmt.Sprintf("%s/crm/v3/objects/deals?limit=10&properties=dealstage", c.BaseURL)
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
			ID         string `json:"id"`
			Properties struct {
				DealStage string `json:"dealstage"`
			} `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	updates := make([]LeadUpdate, len(result.Results))
	for i, r := range result.Results {
		updates[i] = LeadUpdate{
			ID:       r.ID,
			NewState: db.LeadState(r.Properties.DealStage),
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

	// Association with deal would happen here in HubSpot
	return nil
}

func (c *HubSpotCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, contact := range contacts {
		url := fmt.Sprintf("%s/crm/v3/objects/contacts", c.BaseURL)
		payload, _ := json.Marshal(map[string]interface{}{
			"properties": map[string]string{
				"email":     contact.Email,
				"firstname": contact.Name,
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
	url := fmt.Sprintf("%s/crm/v3/objects/deals/%d?properties=dealstage,amount", c.BaseURL, dealID)
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
		ID         string `json:"id"`
		Properties struct {
			DealStage string `json:"dealstage"`
			Amount    string `json:"amount"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	id, _ := strconv.ParseInt(r.ID, 10, 64)
	amount, _ := strconv.ParseFloat(r.Properties.Amount, 64)

	return &DealDetails{
		ID:            id,
		Status:        db.LeadState(r.Properties.DealStage),
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
