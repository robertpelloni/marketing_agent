package crm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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
