package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDraftsEndpointsRequireAuth(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil)

	for _, path := range []string{"/api/v1/drafts", "/api/v1/drafts/approve", "/api/v1/drafts/reject"} {
		req, _ := http.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("%s without session: expected 303, got %d", path, rr.Code)
		}
	}

	token, err := server.auth.Login("admin")
	if err != nil {
		t.Fatalf("mint session: %v", err)
	}
	req, _ := http.NewRequest("GET", "/api/v1/drafts", nil)
	req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: token})
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("drafts with session + nil DB: expected 503, got %d", rr.Code)
	}
}
