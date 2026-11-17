package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Operator data endpoints (leads, outreach export, metrics) were mounted
// without auth while the service binds *:8084 and nginx proxies /api/v1/
// publicly — anyone could dump the outreach log or create leads. They
// must require the dashboard session; public funnel endpoints (quote,
// checkout, health, outreach stats aggregates) stay open.
func TestOperatorEndpointsRequireAuth(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)

	for _, path := range []string{"/api/v1/leads", "/api/v1/outreach/export", "/api/v1/metrics", "/metrics"} {
		req, _ := http.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("%s without session: expected 303 redirect to /login, got %d", path, rr.Code)
		}
	}

	// The old static magic value must no longer authenticate (v0.7.11
	// signed sessions replaced it).
	reqLegacy, _ := http.NewRequest("GET", "/api/v1/leads", nil)
	reqLegacy.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: "authorized_admin_session"})
	rrLegacy := httptest.NewRecorder()
	server.ServeHTTP(rrLegacy, reqLegacy)
	if rrLegacy.Code != http.StatusSeeOther {
		t.Errorf("legacy static cookie: expected 303, got %d", rrLegacy.Code)
	}

	// With a real signed session cookie the request must reach the
	// handler (nil DB then fails as 500/503, proving auth passed it).
	token, err := server.auth.Login("admin")
	if err != nil {
		t.Fatalf("mint admin session: %v", err)
	}
	for _, path := range []string{"/api/v1/leads", "/api/v1/metrics"} {
		req, _ := http.NewRequest("GET", path, nil)
		req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: token})
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code == http.StatusSeeOther {
			t.Errorf("%s with session: must not redirect to /login", path)
		}
	}
}

// Public funnel endpoints must stay reachable without a session.
func TestPublicEndpointsStayOpen(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)

	for _, path := range []string{"/health", "/health/detailed", "/api/v1/outreach/stats", "/api/v1/quote?company_size=11-50"} {
		req, _ := http.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code == http.StatusSeeOther {
			t.Errorf("%s must stay public, got redirect to /login", path)
		}
	}
}
