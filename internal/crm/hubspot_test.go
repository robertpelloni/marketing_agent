package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestHubSpotCRMClient_PushDeal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/crm/v3/objects/deals" {
			t.Errorf("Expected path /crm/v3/objects/deals, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected auth header, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewHubSpotCRMClient("test-token")
	client.BaseURL = server.URL // Override for testing

	err := client.PushDeal(context.Background(), db.Deal{ID: 1}, db.Company{Name: "TestCorp"}, "test")
	if err != nil {
		t.Fatalf("PushDeal failed: %v", err)
	}
}

func TestHubSpotCRMClient_GetLeadUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results": [{"id": "123", "properties": {"dealstage": "Negotiating"}}]}`))
	}))
	defer server.Close()

	client := NewHubSpotCRMClient("test-token")
	client.BaseURL = server.URL

	updates, err := client.GetLeadUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetLeadUpdates failed: %v", err)
	}

	if len(updates) != 1 || updates[0].ID != "123" {
		t.Errorf("Unexpected updates: %+v", updates)
	}
}

func TestHubSpotCRMClient_ValidateAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total": 1}`))
	}))
	defer server.Close()

	client := NewHubSpotCRMClient("test-token")
	client.BaseURL = server.URL

	valid, err := client.ValidateAccount(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("ValidateAccount failed: %v", err)
	}

	if !valid {
		t.Error("Expected account to be valid")
	}
}
