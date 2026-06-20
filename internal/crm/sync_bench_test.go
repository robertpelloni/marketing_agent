package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func BenchmarkCRMSync_PushDeal(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "bench-key")
	ctx := context.Background()
	deal := db.Deal{ID: 1, CurrentState: db.StateNegotiating, QuotedPricing: 10000}
	company := db.Company{ID: 1, Name: "BenchCorp"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := client.PushDeal(ctx, deal, company, "Benchmark")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCRMSync_SyncContacts(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestCRMClient(server.URL, "bench-key")
	ctx := context.Background()
	contacts := []db.Contact{
		{Name: "Contact 1", Email: "c1@example.com"},
		{Name: "Contact 2", Email: "c2@example.com"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := client.SyncContacts(ctx, 1, contacts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
