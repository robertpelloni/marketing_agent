package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestSalesforceCRMClient_PushDeal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/v54.0/sobjects/Opportunity" {
			t.Errorf("Expected path /services/data/v54.0/sobjects/Opportunity, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewSalesforceCRMClient(server.URL, "test-token", "", "", "")
	err := client.PushDeal(context.Background(), db.Deal{ID: 1}, db.Company{Name: "TestCorp"}, "test")
	if err != nil {
		t.Fatalf("PushDeal failed: %v", err)
	}
}

func TestSalesforceCRMClient_GetLeadUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"records": [{"Id": "SF123", "StageName": "Closed Won"}]}`))
	}))
	defer server.Close()

	client := NewSalesforceCRMClient(server.URL, "test-token", "", "", "")
	updates, err := client.GetLeadUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetLeadUpdates failed: %v", err)
	}

	if len(updates) != 1 || updates[0].ID != "SF123" {
		t.Errorf("Unexpected updates: %+v", updates)
	}
}
