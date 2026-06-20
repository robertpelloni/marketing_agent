package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
<<<<<<< HEAD
	"net/http"
	"os"
=======
	"io"
	"net/http"
	"net/url"
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

<<<<<<< HEAD
// SalesforceClient implements CRMClient using the Salesforce REST API.
// It expects the following environment variables to be set:
//  - SALESFORCE_INSTANCE_URL (e.g., https://yourInstance.my.salesforce.com)
//  - SALESFORCE_ACCESS_TOKEN (OAuth2 Bearer token)
//  - SALESFORCE_API_VERSION (e.g., "v57.0")
// The client uses the standard Salesforce REST endpoints for leads, contacts, and opportunities.

type SalesforceClient struct {
	instanceURL string // Base URL for the instance, without trailing slash
	accessToken string // OAuth2 bearer token
	apiVersion  string // API version, e.g., "v57.0"
	client      *http.Client
}

// NewSalesforceClient creates a new Salesforce CRM client.
func NewSalesforceClient() (*SalesforceClient, error) {
	inst := os.Getenv("SALESFORCE_INSTANCE_URL")
	token := os.Getenv("SALESFORCE_ACCESS_TOKEN")
	ver := os.Getenv("SALESFORCE_API_VERSION")
	if inst == "" || token == "" {
		return nil, fmt.Errorf("salesforce client: missing required env vars (instance URL or access token)")
	}
	if ver == "" {
		ver = "v57.0"
	}
	return &SalesforceClient{
		instanceURL: inst,
		accessToken: token,
		apiVersion:  ver,
		client:      &http.Client{},
	}, nil
}

// PushDeal creates or updates a Salesforce Opportunity representing the deal.
func (s *SalesforceClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	// Salesforce Opportunity fields mapping
	payload := map[string]interface{}{
		"Name":               fmt.Sprintf("%s – %s", company.Name, route),
		"AccountId":          s.accountIDFromDomain(company.Domain),
		"StageName":          mapLeadStateToStage(deal.CurrentState),
		"CloseDate":          timeNowISO8601(),
		"Amount":             deal.QuotedPricing,
		"Description":        deal.TechnicalDossier,
		"Custom_Field__c":    route, // placeholder for custom routing info
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Opportunity", s.instanceURL, s.apiVersion)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
=======
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
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return err
	}
	defer resp.Body.Close()

<<<<<<< HEAD
	if resp.StatusCode >= 400 {
		return fmt.Errorf("salesforce PushDeal: status %d", resp.StatusCode)
	}

	return nil
}

// GetLeadUpdates fetches recent lead status changes from Salesforce.
// It queries the Lead object for recent "IsConverted" or "StageName" changes.
func (s *SalesforceClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// Simple SOQL query for recently modified leads (last 24h)
	soql := "SELECT Id, StageName FROM Lead WHERE LastModifiedDate = LAST_N_DAYS:1"
	url := fmt.Sprintf("%s/services/data/%s/query?q=%s", s.instanceURL, s.apiVersion, urlEncode(soql))

=======
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
	if err := c.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Querying EmailMessage to get real-time inbound emails
	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=SELECT+TextBody,Subject,FromAddress+FROM+EmailMessage+WHERE+Incoming=true+LIMIT+10", c.BaseURL)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
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
		return nil, fmt.Errorf("salesforce GetLeadUpdates: status %d", resp.StatusCode)
=======
		return nil, fmt.Errorf("salesforce api error (%d)", resp.StatusCode)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	}

	var result struct {
		Records []struct {
<<<<<<< HEAD
			Id        string `json:"Id"`
			StageName string `json:"StageName"`
=======
			TextBody    string `json:"TextBody"`
			Subject     string `json:"Subject"`
			FromAddress string `json:"FromAddress"`
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
		} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

<<<<<<< HEAD
	var updates []LeadUpdate
	for _, r := range result.Records {
		state := mapStageToLeadState(r.StageName)
		if state != "" {
			updates = append(updates, LeadUpdate{ID: r.Id, NewState: state, Notes: ""})
=======
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
	if err := c.RefreshToken(ctx); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	method := "POST"
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity", c.BaseURL)
	if deal.ID > 1000 {
		method = "PATCH"
		url = fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity/%015d", c.BaseURL, deal.ID)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		c.Mapping.DealNameProperty:    fmt.Sprintf("%s - %d", company.Name, deal.ID),
		c.Mapping.DealStageProperty:   string(deal.CurrentState),
		c.Mapping.DealAmountProperty:  deal.QuotedPricing,
		c.Mapping.DealDossierProperty: deal.TechnicalDossier,
		"CloseDate":                   "2026-12-31", // Placeholder
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
		return fmt.Errorf("salesforce api error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *SalesforceCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	if err := c.RefreshToken(ctx); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

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
	if err := c.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

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
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
		}
	}

	return updates, nil
}

<<<<<<< HEAD
// ValidateAccount verifies if a given domain has a Salesforce Account.
func (s *SalesforceClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	soql := fmt.Sprintf("SELECT Id FROM Account WHERE Website = '%s' LIMIT 1", domain)
	url := fmt.Sprintf("%s/services/data/%s/query?q=%s", s.instanceURL, s.apiVersion, urlEncode(soql))

=======
func (c *SalesforceCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	if err := c.RefreshToken(ctx); err != nil {
		return false, fmt.Errorf("token refresh failed: %w", err)
	}

	url := fmt.Sprintf("%s/services/data/v54.0/query/?q=SELECT+Id+FROM+Account+WHERE+Website+LIKE+'%%%s%%'+LIMIT+1", c.BaseURL, domain)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
<<<<<<< HEAD
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
=======
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

<<<<<<< HEAD
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("salesforce ValidateAccount: status %d", resp.StatusCode)
=======
	if resp.StatusCode >= 400 {
		return false, nil
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	}

	var result struct {
		TotalSize int `json:"totalSize"`
	}
<<<<<<< HEAD
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.TotalSize > 0, nil
}

// SyncInteraction creates a note on the related Opportunity.
func (s *SalesforceClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	payload := map[string]interface{}{
		"ParentId": fmt.Sprintf("%d", dealID), // Assuming dealID matches Opportunity Id mapping
		"Body":    note,
		"Title":   "Automated Interaction Log",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Task", s.instanceURL, s.apiVersion)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
=======
	json.NewDecoder(resp.Body).Decode(&result)

	return result.TotalSize > 0, nil
}

func (c *SalesforceCRMClient) SyncInteraction(ctx context.Context, dealID int64, note string) error {
	if err := c.RefreshToken(ctx); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// In Salesforce, notes are often attached via ContentNote or Task.
	// We use 'WhatId' to associate the Task with the Opportunity.
	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Task", c.BaseURL)
	payload, _ := json.Marshal(map[string]interface{}{
		"Description": note,
		"Status":      "Completed",
		"Priority":     "Normal",
		"Subject":      "Autonomous Sales Interaction",
		"WhatId":       fmt.Sprintf("%015d", dealID), // Salesforce ID format padding
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

<<<<<<< HEAD
	if resp.StatusCode >= 400 {
		return fmt.Errorf("salesforce SyncInteraction: status %d", resp.StatusCode)
	}

	return nil
}

// SyncContacts creates or updates Salesforce Contact records for a company.
func (s *SalesforceClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	accountID := s.accountIDFromDomain("") // placeholder – in real impl we lookup AccountId by domain

	for _, c := range contacts {
		payload := map[string]interface{}{
			"FirstName": strings.Split(c.Name, " ")[0],
			"LastName":  strings.Split(c.Name, " ")[1],
			"Email":     c.Email,
			"Title":     c.Role,
			"AccountId": accountID,
		}

		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("%s/services/data/%s/sobjects/Contact", s.instanceURL, s.apiVersion)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("salesforce SyncContacts: status %d", resp.StatusCode)
		}
	}
	return nil
}

// FetchDealDetails retrieves an Opportunity from Salesforce.
func (s *SalesforceClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Opportunity/%d", s.instanceURL, s.apiVersion, dealID)
=======
	return nil
}

func (c *SalesforceCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	if err := c.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	url := fmt.Sprintf("%s/services/data/v54.0/sobjects/Opportunity/%d", c.BaseURL, dealID)
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
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
		return nil, fmt.Errorf("salesforce FetchDealDetails: status %d", resp.StatusCode)
	}

	var result struct {
		Id                string  `json:"Id"`
		StageName         string  `json:"StageName"`
		Amount            float64 `json:"Amount"`
		Description       string  `json:"Description"`
		Custom_Field__c    string  `json:"Custom_Field__c"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &DealDetails{
		ID:                 parseID(result.Id),
		Status:             mapStageToLeadState(result.StageName),
		QuotedPricing:      result.Amount,
		CustomRequirements: result.Custom_Field__c,
		TechnicalDossier:   result.Description,
	}, nil
}

// Helper functions (placeholders for actual implementations)
func (s *SalesforceClient) accountIDFromDomain(domain string) string { return "" }
func mapLeadStateToStage(state db.LeadState) string       { return "Prospecting" }
func mapStageToLeadState(stage string) db.LeadState      { return db.StateResearched }
func timeNowISO8601() string                               { return "2026-06-14" }
func urlEncode(s string) string                           { return s }
func parseID(s string) int64                               { return 0 }
=======
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
	if err := c.RefreshToken(ctx); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

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
>>>>>>> origin/jules-phase6-production-hardening-042-863b86a9-12417263503841031080
