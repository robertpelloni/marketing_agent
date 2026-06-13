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
	Mapping      FieldMapping
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
		Mapping: FieldMapping{
			DealNameProperty:     "Name",
			DealStageProperty:    "StageName",
			DealAmountProperty:   "Amount",
			DealDossierProperty:  "Description",
			ContactEmailProperty: "Email",
		},
	}
}

func (c *SalesforceCRMClient) SetFieldMapping(mapping FieldMapping) {
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

func (c *SalesforceCRMClient) GetNewInteractions(ctx context.Context) ([]db.Interaction, error) {
	// Querying EmailMessage to get real-time inbound emails
	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=SELECT+TextBody,Subject,FromAddress+FROM+EmailMessage+WHERE+Incoming=true+LIMIT+10", c.BaseURL)
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
		return nil, fmt.Errorf("salesforce api error (%d)", resp.StatusCode)
	}

	var result struct {
		Records []struct {
			TextBody    string `json:"TextBody"`
			Subject     string `json:"Subject"`
			FromAddress string `json:"FromAddress"`
		} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	interactions := make([]db.Interaction, len(result.Records))
	for i, r := range result.Records {
		interactions[i] = db.Interaction{
			RawText: r.TextBody,
			Summary: r.FromAddress, // Use sender address for identification
			Channel: "Salesforce Email",
		}
	}

	return interactions, nil
}

func (c *SalesforceCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		c.Mapping.DealNameProperty:    fmt.Sprintf("%s - %d", company.Name, deal.ID),
		c.Mapping.DealStageProperty:   string(deal.CurrentState),
		c.Mapping.DealAmountProperty:  deal.QuotedPricing,
		c.Mapping.DealDossierProperty: deal.TechnicalDossier,
		"CloseDate":                   "2026-12-31", // Placeholder
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
			"LastName":                     contact.Name,
			c.Mapping.ContactEmailProperty: contact.Email,
			"Title":                        contact.Role,
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
	query := fmt.Sprintf("SELECT Id,%s FROM Opportunity LIMIT 10", c.Mapping.DealStageProperty)
	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=%s", c.BaseURL, url.QueryEscape(query))
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
		Records []map[string]interface{} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	updates := make([]LeadUpdate, len(result.Records))
	for i, r := range result.Records {
		id, _ := r["Id"].(string)
		stage, _ := r[c.Mapping.DealStageProperty].(string)
		updates[i] = LeadUpdate{
			ID:       id,
			NewState: db.LeadState(stage),
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

	var r map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	idStr, _ := r["Id"].(string)
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	stage, _ := r[c.Mapping.DealStageProperty].(string)
	amount, _ := r[c.Mapping.DealAmountProperty].(float64)

	return &DealDetails{
		ID:            id,
		Status:        db.LeadState(stage),
		QuotedPricing: amount,
	}, nil
}

func (c *SalesforceCRMClient) SendEmail(ctx context.Context, contact db.Contact, subject, body string) error {
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/EmailMessage", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"Subject":      subject,
		"TextBody":     body,
		"ToAddress":    contact.Email,
		"Status":       "3", // Sent
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
		return fmt.Errorf("salesforce api error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
