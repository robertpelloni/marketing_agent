package crm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

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
	client.HTTPClient.Timeout = 10 * time.Millisecond

	_, err := client.GetLeadUpdates(context.Background())
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "Timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("expected timeout error message, got: %v", err)
	}
}
