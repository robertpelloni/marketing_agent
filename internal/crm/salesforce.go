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
	mapping     FieldMapping
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
		mapping: FieldMapping{
			DealNameProp:     "Name",
			DealAmountProp:   "Amount",
			DealStageProp:    "StageName",
			DealDescProp:     "Description",
			DealRouteProp:    "Custom_Field__c",
			ContactEmailProp: "Email",
			ContactRoleProp:  "Title",
			AccountWebProp:   "Website",
		},
	}, nil
}

// SetFieldMapping updates the dynamic field mapping.
func (s *SalesforceClient) SetFieldMapping(mapping FieldMapping) {
	if mapping.DealNameProp != "" {
		s.mapping.DealNameProp = mapping.DealNameProp
	}
	if mapping.DealAmountProp != "" {
		s.mapping.DealAmountProp = mapping.DealAmountProp
	}
	if mapping.DealStageProp != "" {
		s.mapping.DealStageProp = mapping.DealStageProp
	}
	if mapping.DealDescProp != "" {
		s.mapping.DealDescProp = mapping.DealDescProp
	}
	if mapping.DealRouteProp != "" {
		s.mapping.DealRouteProp = mapping.DealRouteProp
	}
	if mapping.ContactEmailProp != "" {
		s.mapping.ContactEmailProp = mapping.ContactEmailProp
	}
	if mapping.ContactRoleProp != "" {
		s.mapping.ContactRoleProp = mapping.ContactRoleProp
	}
	if mapping.AccountWebProp != "" {
		s.mapping.AccountWebProp = mapping.AccountWebProp
	}
}

// PushDeal creates or updates a Salesforce Opportunity representing the deal.
func (s *SalesforceClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	// Salesforce Opportunity fields mapping
	payload := map[string]interface{}{
		s.mapping.DealNameProp:   fmt.Sprintf("%s – %s", company.Name, route),
		"AccountId":              s.accountIDFromDomain(company.Domain),
		s.mapping.DealStageProp:  mapLeadStateToStage(deal.CurrentState),
		"CloseDate":              timeNowISO8601(),
		s.mapping.DealAmountProp: deal.QuotedPricing,
		s.mapping.DealDescProp:   deal.TechnicalDossier,
		s.mapping.DealRouteProp:  route,
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("salesforce PushDeal: status %d", resp.StatusCode)
	}

	return nil
}

// GetLeadUpdates fetches recent lead status changes from Salesforce.
// It queries the Lead object for recent "IsConverted" or "StageName" changes.
func (s *SalesforceClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) {
	// Simple SOQL query for recently modified leads (last 24h)
	soql := fmt.Sprintf("SELECT Id, %s FROM Lead WHERE LastModifiedDate = LAST_N_DAYS:1", s.mapping.DealStageProp)
	url := fmt.Sprintf("%s/services/data/%s/query?q=%s", s.instanceURL, s.apiVersion, urlEncode(soql))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("salesforce GetLeadUpdates: status %d", resp.StatusCode)
	}

	var result struct {
		Records []map[string]interface{} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var updates []LeadUpdate
	for _, r := range result.Records {
		id, _ := r["Id"].(string)
		stage, _ := r[s.mapping.DealStageProp].(string)
		state := mapStageToLeadState(stage)
		if state != "" {
			updates = append(updates, LeadUpdate{ID: id, NewState: state, Notes: ""})
		}
	}

	return updates, nil
}

// ValidateAccount verifies if a given domain has a Salesforce Account.
func (s *SalesforceClient) ValidateAccount(ctx context.Context, domain string) (bool, error) {
	soql := fmt.Sprintf("SELECT Id FROM Account WHERE %s = '%s' LIMIT 1", s.mapping.AccountWebProp, domain)
	url := fmt.Sprintf("%s/services/data/%s/query?q=%s", s.instanceURL, s.apiVersion, urlEncode(soql))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("salesforce ValidateAccount: status %d", resp.StatusCode)
	}

	var result struct {
		TotalSize int `json:"totalSize"`
	}
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
			"FirstName":              strings.Split(c.Name, " ")[0],
			"LastName":               strings.Split(c.Name, " ")[1],
			s.mapping.ContactEmailProp: c.Email,
			s.mapping.ContactRoleProp:  c.Role,
			"AccountId":              accountID,
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
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Opportunity/%d?fields=Id,%s,%s,%s,%s",
		s.instanceURL, s.apiVersion, dealID,
		s.mapping.DealStageProp, s.mapping.DealAmountProp, s.mapping.DealDescProp, s.mapping.DealRouteProp)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("salesforce FetchDealDetails: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	id, _ := result["Id"].(string)
	stage, _ := result[s.mapping.DealStageProp].(string)
	amount, _ := result[s.mapping.DealAmountProp].(float64)
	desc, _ := result[s.mapping.DealDescProp].(string)
	route, _ := result[s.mapping.DealRouteProp].(string)

	return &DealDetails{
		ID:                 parseID(id),
		Status:             mapStageToLeadState(stage),
		QuotedPricing:      amount,
		CustomRequirements: route,
		TechnicalDossier:   desc,
	}, nil
}

// Helper functions (placeholders for actual implementations)
func (s *SalesforceClient) accountIDFromDomain(domain string) string { return "" }
func mapLeadStateToStage(state db.LeadState) string       { return "Prospecting" }
func mapStageToLeadState(stage string) db.LeadState      { return db.StateResearched }
func timeNowISO8601() string                               { return "2026-06-14" }
func urlEncode(s string) string                           { return s }
func parseID(s string) int64                               { return 0 }
