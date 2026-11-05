package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebHandlers_NoDB(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)

	t.Run("Dashboard GET nil DB", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		// It should probably return 500 or handle it gracefully
		if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusOK && rr.Code != http.StatusSeeOther {
			t.Errorf("Unexpected status code: %v", rr.Code)
		}
	})

	t.Run("GitHub Webhook Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/webhook/github", bytes.NewBuffer([]byte(`{}`)))
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			t.Errorf("Webhook handler returned 200 OK without valid signature")
		}
	})

    t.Run("Webhook endpoint routing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/webhook/github", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		// Should not be allowed
		if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusForbidden {
			t.Errorf("Expected MethodNotAllowed or Forbidden for GET on webhook, got %v", rr.Code)
		}
	})

    t.Run("Leads API POST nil DB", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/leads", bytes.NewBuffer([]byte(`{"name":"test"}`)))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		// With a nil DB, this might return an error, but shouldn't panic
		if rr.Code == http.StatusOK {
			t.Errorf("Expected failure with nil DB")
		}
	})

	t.Run("Reddit Devvit Content Endpoints", func(t *testing.T) {
		for _, endpoint := range []string{"/api/v1/social/reddit", "/api/social/reddit"} {
			req, _ := http.NewRequest("GET", endpoint, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("Expected 200 OK from %s, got %d", endpoint, rr.Code)
			}
			body := rr.Body.String()
			if !bytes.Contains(rr.Body.Bytes(), []byte("thumbnail_url")) {
				t.Errorf("Expected thumbnail_url in response from %s, got: %s", endpoint, body)
			}
			if !bytes.Contains(rr.Body.Bytes(), []byte("preview_url")) {
				t.Errorf("Expected preview_url in response from %s, got: %s", endpoint, body)
			}
		}
	})

	t.Run("Metrics Endpoints", func(t *testing.T) {
		// Prometheus text format
		req, _ := http.NewRequest("GET", "/metrics", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK from /metrics, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("marketing_agent_db_open_connections")) {
			t.Errorf("Expected Prometheus metrics in body, got: %s", rr.Body.String())
		}

		// JSON format
		reqJSON, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
		rrJSON := httptest.NewRecorder()
		server.ServeHTTP(rrJSON, reqJSON)

		if rrJSON.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK from /api/v1/metrics, got %d", rrJSON.Code)
		}
		if !bytes.Contains(rrJSON.Body.Bytes(), []byte("db_open_connections")) {
			t.Errorf("Expected JSON metrics in body, got: %s", rrJSON.Body.String())
		}
	})
}
