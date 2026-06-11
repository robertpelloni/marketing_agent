package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestSalesforceSendEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/v54.0/sobjects/EmailMessage" {
			t.Errorf("expected path /services/data/v54.0/sobjects/EmailMessage, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewSalesforceCRMClient(server.URL, "test-token", "", "", "")
	contact := db.Contact{Email: "test@example.com"}
	err := client.SendEmail(context.Background(), contact, "Subject", "Body")
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}
}
