package web

import (
	"github.com/robertpelloni/marketing_agent/internal/db"

	"context"

	"os"

	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebHandlers_NoDB(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil)

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
}

func TestHandleDashboard_SocialPostsIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("no database url")
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test, DB not available: %v", err)
	}

	err = database.RunMigrations(context.Background())
	if err != nil {
		t.Skipf("Skipping integration test, migrations failed: %v", err)
	}
	defer database.Close()

	// Insert a test social post
	post := &db.SocialPost{
		Brand: "TormentNexus",
		Platform: "Reddit",
		AccountUsername: "nexus_dev",
		PostContent: "Testing integration!",
		Status: "posted",
	}
	_ = database.CreateSocialPost(context.Background(), post)

	server := NewServer(database, nil, nil, nil, nil, nil)
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Dashboard handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("Testing integration!")) {
		t.Errorf("Expected dashboard to contain social post content")
	}
}
