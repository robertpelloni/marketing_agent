package crm
import (
	"context"
	"log/slog"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)
type MockCRMClient struct{}
func NewMockCRMClient() *MockCRMClient { return &MockCRMClient{} }
func (m *MockCRMClient) PushDeal(ctx context.Context, deal db.Deal, company db.Company, route string) error {
	slog.Info("CRM Pushing deal", "deal_id", deal.ID, "company", company.Name, "route", route); return nil
}
func (m *MockCRMClient) GetLeadUpdates(ctx context.Context) ([]LeadUpdate, error) { return []LeadUpdate{}, nil }
func (m *MockCRMClient) SyncContacts(ctx context.Context, companyID int64, contacts []db.Contact) error {
	slog.Info("CRM Syncing contacts", "company_id", companyID, "count", len(contacts)); return nil
}
func (m *MockCRMClient) FetchDealDetails(ctx context.Context, dealID int64) (*DealDetails, error) {
	return &DealDetails{ID: dealID, QuotedPricing: 50000}, nil
}
func (m *MockCRMClient) SyncInteraction(ctx context.Context, companyID int64, note string) error {
	slog.Info("CRM Syncing interaction", "company_id", companyID); return nil
}
func (m *MockCRMClient) ValidateAccount(ctx context.Context, domain string) (bool, error) { return true, nil }
