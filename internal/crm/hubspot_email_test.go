package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestHubSpotSendEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/crm/v3/objects/communications" {
			t.Errorf("expected path /crm/v3/objects/communications, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewHubSpotCRMClient("test-token")
	client.BaseURL = server.URL

	contact := db.Contact{Email: "test@example.com"}
	err := client.SendEmail(context.Background(), contact, "Subject", "Body")
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}
}
