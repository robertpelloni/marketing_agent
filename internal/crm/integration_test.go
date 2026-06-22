package crm_test

import (
	"context"
	"os"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func setupCRMTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestCRMWorker_Integration_SyncsNegotiatingDeals(t *testing.T) {
	database := setupCRMTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create a company and deal in Negotiating state
	company := &db.Company{
		Name:          "CRM Test Corp",
		Domain:        "crmtest.io",
		MarketCapTier: "Mid-Market",
	}
	err := database.CreateCompany(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	deal := &db.Deal{
		CompanyID:    company.ID,
		CurrentState: db.StateNegotiating,
	}
	err = database.CreateDeal(ctx, deal)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Create CRM worker with mock CRM client
	mockCRM := crm.NewMockCRMClient()
	worker := crm.NewWorker(database, mockCRM)

	// Run sync cycle
	worker.ExecuteSync(ctx)

	// Verify no panic occurred and deal still exists
	updatedDeal, err := database.GetDealByCompanyID(ctx, company.ID)
	if err != nil {
		t.Fatalf("Failed to get deal: %v", err)
	}
	if updatedDeal == nil {
		t.Error("Expected deal to still exist after sync")
	}
}

func TestCRMWorker_Integration_SkipsNilDB(t *testing.T) {
	mockCRM := crm.NewMockCRMClient()
	worker := crm.NewWorker(nil, mockCRM)

	ctx := context.Background()
	worker.ExecuteSync(ctx)
	// If this doesn't panic and returns cleanly, the nil check works
<<<<<<< HEAD
}
=======
}
=======
package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestRestCRMClient_RetryLogic_Integration(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		if atomic.LoadInt32(&attempts) < 3 {
			http.Error(w, "Temporary Server Error", http.StatusInternalServerError)
			return
func TestRestCRMClient_DetailedError(t *testing.T) {
	expectedBody := `{"error": "invalid custom requirements"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	err := client.PushDeal(context.Background(), db.Deal{ID: 1}, db.Company{Name: "TestCorp"}, "test-route")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), expectedBody) {
		t.Errorf("expected error to contain %q, got: %v", expectedBody, err)
	}
}

func TestRestCRMClient_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	err := client.SyncContacts(context.Background(), 1, []db.Contact{{Name: "Test"}})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "429") {
		t.Errorf("expected error to contain 429 status code, got: %v", err)
	}
}

func TestRestCRMClient_DataIntegrity(t *testing.T) {
	var capturedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deals" {
			t.Errorf("expected path /deals, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if err := json.Unmarshal(body, &capturedPayload); err != nil {
			t.Fatalf("failed to decode json: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: The client itself doesn't have retry logic internally yet,
	// it's handled at the worker/manager level. Let's verify we can capture the error.
	err := client.PushDeal(ctx, db.Deal{ID: 1}, db.Company{Name: "RetryCorp"}, "integration-test")
	if err == nil {
		t.Error("Expected error on first attempt")
	}

	// We'll reset and verify success on the 3rd attempt if called manually
	atomic.StoreInt32(&attempts, 2)
	err = client.PushDeal(ctx, db.Deal{ID: 1}, db.Company{Name: "RetryCorp"}, "integration-test")
	if err != nil {
		t.Fatalf("Expected success on attempt 3, got: %v", err)
	}
}

func TestRestCRMClient_Timeout_Integration(t *testing.T) {

	deal := db.Deal{
		ID:                 42,
		CurrentState:       db.StateNegotiating,
		QuotedPricing:      15000.50,
		TechnicalDossier:   "Test Dossier",
	}
	company := db.Company{
		Name: "Acme Corp",
	}

	err := client.PushDeal(context.Background(), deal, company, "inbound")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPayload["deal_id"].(float64) != 42 {
		t.Errorf("expected deal_id 42, got %v", capturedPayload["deal_id"])
	}
	if capturedPayload["company"].(string) != "Acme Corp" {
		t.Errorf("expected company Acme Corp, got %v", capturedPayload["company"])
	}
	if capturedPayload["status"].(string) != string(db.StateNegotiating) {
		t.Errorf("expected status %s, got %v", db.StateNegotiating, capturedPayload["status"])
	}
	if capturedPayload["pricing"].(float64) != 15000.50 {
		t.Errorf("expected pricing 15000.50, got %v", capturedPayload["pricing"])
	}
	if capturedPayload["technical_dossier"].(string) != "Test Dossier" {
		t.Errorf("expected technical_dossier 'Test Dossier', got %v", capturedPayload["technical_dossier"])
	}
	if capturedPayload["route"].(string) != "inbound" {
		t.Errorf("expected route inbound, got %v", capturedPayload["route"])
	}
}

func TestRestCRMClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := client.SyncInteraction(ctx, 1, "Testing timeout")
	if err == nil {
		t.Error("Expected context deadline exceeded error")
	}
}

func TestRestCRMClient_ErrorBody_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid_deal_id"))
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "test-key")
	err := client.PushDeal(context.Background(), db.Deal{ID: 999}, db.Company{Name: "ErrorCorp"}, "integration-test")
	if err == nil {
		t.Fatal("Expected error for 400 Bad Request")
	}

	expected := "crm api error (400): invalid_deal_id"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}
	client.HTTPClient.Timeout = 10 * time.Millisecond

	_, err := client.GetLeadUpdates(context.Background())
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "Timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("expected timeout error message, got: %v", err)
	}
}
>>>>>>> origin/main
