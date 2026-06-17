package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestDispatcher_Dispatch(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		if payload.Event != EventDealStateChange {
			t.Errorf("Expected event %s, got %s", EventDealStateChange, payload.Event)
		}

		if payload.DealID != 123 {
			t.Errorf("Expected deal ID 123, got %d", payload.DealID)
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(server.URL)
	err := dispatcher.Dispatch(context.Background(), 123, db.StateOutreachSent)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
}

func TestDispatcher_RetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewDispatcher(server.URL)
	// Override client for faster tests
	dispatcher.Client.Timeout = 1 * time.Second

	err := dispatcher.Dispatch(context.Background(), 1, db.StateClosedWon)
	if err != nil {
		t.Fatalf("Dispatch should have succeeded on retry: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}
