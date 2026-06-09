package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// SalesforceCRMClient implements CRMClient for Salesforce.
type SalesforceCRMClient struct {
	BaseURL      string
	AccessToken  string
	ClientID     string
	ClientSecret string
	AuthURL      string
	HTTPClient   *http.Client
}

// NewSalesforceCRMClient creates a new Salesforce CRM client.
func NewSalesforceCRMClient(baseURL, accessToken, clientID, clientSecret, authURL string) *SalesforceCRMClient {
	return &SalesforceCRMClient{
		BaseURL:      baseURL,
		AccessToken:  accessToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      authURL,
		HTTPClient:   &http.Client{},
	}
}

func (c *SalesforceCRMClient) RefreshToken(ctx context.Context) error {
	if c.ClientID == "" || c.ClientSecret == "" || c.AuthURL == "" {
		return nil // Skip if not configured for OAuth
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", c.AuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth error: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.AccessToken = result.AccessToken
	return nil
}

func (c *SalesforceCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"Name":          fmt.Sprintf("%s - %d", company.Name, deal.ID),
		"StageName":     string(deal.CurrentState),
		"Amount":        deal.QuotedPricing,
		"Description":   deal.TechnicalDossier,
		"CloseDate":     "2026-12-31", // Placeholder
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
		return fmt.Errorf("salesforce api error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *SalesforceCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	for _, contact := range contacts {
		url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Contact", c.BaseURL)
		payload, _ := json.Marshal(map[string]interface{}{
			"LastName":  contact.Name,
			"Email":     contact.Email,
			"Title":     contact.Role,
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

func (c *SalesforceCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// Simplified: Querying Opportunities
	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=SELECT+Id,StageName+FROM+Opportunity+LIMIT+10", c.BaseURL)
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
		return nil, fmt.Errorf("salesforce api error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Records []struct {
			Id        string `json:"Id"`
			StageName string `json:"StageName"`
		} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	updates := make([]LeadUpdate, len(result.Records))
	for i, r := range result.Records {
		updates[i] = LeadUpdate{
			ID:       r.Id,
			NewState: db.LeadState(r.StageName),
		}
	}

	return updates, nil
}

func (c *SalesforceCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=SELECT+Id+FROM+Account+WHERE+Website+LIKE+'%%%s%%'+LIMIT+1", c.BaseURL, domain)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false, nil
	}

	var result struct {
		TotalSize int `json:"totalSize"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.TotalSize > 0, nil
}

func (c *SalesforceCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	// In Salesforce, notes are often attached via ContentNote or Task
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Task", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"Description": note,
		"Status":      "Completed",
		"Priority":     "Normal",
		"Subject":      "Autonomous Sales Interaction",
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

	return nil
}

func (c *SalesforceCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity/%d", c.BaseURL, dealID)
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
		return nil, fmt.Errorf("salesforce api error (%d)", resp.StatusCode)
	}

	var r struct {
		Id        string  `json:"Id"`
		StageName string  `json:"StageName"`
		Amount    float64 `json:"Amount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	var id int64
	fmt.Sscanf(r.Id, "%d", &id)

	return &DealDetails{
		ID:            id,
		Status:        db.LeadState(r.StageName),
		QuotedPricing: r.Amount,
	}, nil
}
