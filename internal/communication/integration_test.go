package communication_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type mockClassifier struct{}
func (m *mockClassifier) Classify(ctx context.Context, text string) (communication.Intent, error) {
	return communication.IntentGeneral, nil
}

type mockResponder struct{}
func (m *mockResponder) Generate(ctx context.Context, salesCtx communication.SalesContext, action communication.Action) (string, error) {
	return "mock reply", nil
}

type mockStrategy struct{}
func (m *mockStrategy) Decide(ctx context.Context, salesCtx communication.SalesContext) (communication.Action, error) {
	return communication.ActionAdvanceState, nil
}

type mockOrderProcessor struct{}
func (m *mockOrderProcessor) ProcessOrder(ctx context.Context, deal db.Deal) error { return nil }

type mockEmailSender struct{}
func (m *mockEmailSender) Send(ctx context.Context, msg communication.EmailMessage) error { return nil }

func setupCommTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" { t.Skip("DATABASE_URL not set") }
	database, err := db.NewDB(dbURL)
	if err != nil { t.Fatalf("Failed to connect: %v", err) }
	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}
	return database
}

func TestCommunicationManager_Integration_OutreachForResearchedDeal(t *testing.T) {
	database := setupCommTestDB(t)
	defer database.Close()
	ctx := context.Background()

	company := &db.Company{Name: "Comm Test Corp", Domain: "commtest.io", MarketCapTier: "Mid-Market"}
	_ = database.CreateCompany(ctx, company)
	_ = database.CreateDeal(ctx, &db.Deal{CompanyID: company.ID, CurrentState: db.StateResearched})
	_ = database.CreateContact(ctx, &db.Contact{CompanyID: company.ID, Name: "Jane Doe", Email: "jane@commtest.io"})

	mgr := communication.NewManager(database, &mockClassifier{}, &mockResponder{}, &mockStrategy{}, &mockOrderProcessor{}, &mockEmailSender{}, nil, nil, nil)
	mgr.ExecutePoll(ctx)

	contacts, _ := database.ListContactsByCompany(ctx, company.ID)
	if len(contacts) == 0 { t.Error("Expected contact") }
}

func TestCommunicationManager_Integration_SkipsNilDB(t *testing.T) {
	mgr := communication.NewManager(nil, &mockClassifier{}, &mockResponder{}, &mockStrategy{}, &mockOrderProcessor{}, nil, nil, nil, nil)
	mgr.ExecutePoll(context.Background())
}
