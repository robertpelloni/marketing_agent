package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginIssuesSignedExpiringToken(t *testing.T) {
	a := NewAuthenticator(nil)
	tok, err := a.Login("admin")
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	if tok == "authorized_admin_session" || !strings.HasPrefix(tok, "v1.admin.") {
		t.Fatalf("expected signed v1 token, got %q", tok)
	}
	subject, ok := a.ValidateSession(tok)
	if !ok || subject != "admin" {
		t.Fatalf("issued token must validate as admin, got subject=%q ok=%v", subject, ok)
	}
}

func TestValidateRejectsLegacyAndTampered(t *testing.T) {
	a := NewAuthenticator(nil)

	for _, legacy := range []string{"authorized_admin_session", "company_7", ""} {
		if _, ok := a.ValidateSession(legacy); ok {
			t.Errorf("legacy static value %q must be rejected", legacy)
		}
	}

	tok, _ := a.Login("admin")
	parts := strings.Split(tok, ".")
	if len(parts) != 4 {
		t.Fatalf("unexpected token shape: %q", tok)
	}

	// Subject tampering (admin -> company) must fail the signature.
	tampered := "v1.company_7." + parts[2] + "." + parts[3]
	if _, ok := a.ValidateSession(tampered); ok {
		t.Error("tampered subject must be rejected")
	}

	// Signature corruption must fail.
	corrupt := parts[0] + "." + parts[1] + "." + parts[2] + ".0000" + parts[3][4:]
	if _, ok := a.ValidateSession(corrupt); ok {
		t.Error("corrupted signature must be rejected")
	}
}

func TestValidateRejectsExpired(t *testing.T) {
	a := NewAuthenticator(nil)
	expired := a.sign("admin", time.Now().Add(-time.Hour))
	if _, ok := a.ValidateSession(expired); ok {
		t.Error("expired token must be rejected")
	}
	fresh := a.sign("admin", time.Now().Add(time.Hour))
	if _, ok := a.ValidateSession(fresh); !ok {
		t.Error("fresh token must validate")
	}
}

func TestCompanySubjectRoundtrip(t *testing.T) {
	a := NewAuthenticator(nil)
	tok := a.IssueSession("company_42")

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: tok})
	id, ok := a.CompanyID(req)
	if !ok || id != 42 {
		t.Fatalf("expected company 42, got id=%d ok=%v", id, ok)
	}

	adminReq := httptest.NewRequest("GET", "/", nil)
	adminTok, _ := a.Login("admin")
	adminReq.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: adminTok})
	if _, ok := a.CompanyID(adminReq); ok {
		t.Error("admin session must not yield a company id")
	}
}

func TestMiddlewareSignedSession(t *testing.T) {
	a := NewAuthenticator(nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := a.Middleware(next)

	// No cookie -> redirect to /login
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/private", nil))
	if rr.Code != http.StatusSeeOther {
		t.Errorf("no cookie: expected 303, got %d", rr.Code)
	}

	// Legacy static cookie -> redirect
	rr = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/private", nil)
	req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: "authorized_admin_session"})
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("legacy cookie: expected 303, got %d", rr.Code)
	}

	// Valid signed cookie -> passes through
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/private", nil)
	tok, _ := a.Login("admin")
	req.AddCookie(&http.Cookie{Name: "sales_bot_session", Value: tok})
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("signed cookie: expected 200, got %d", rr.Code)
	}
}

func TestDifferentSecretRejects(t *testing.T) {
	t.Setenv("SESSION_SECRET", "secret-one")
	a := NewAuthenticator(nil)
	tok, _ := a.Login("admin")

	t.Setenv("SESSION_SECRET", "secret-two")
	b := NewAuthenticator(nil)
	if _, ok := b.ValidateSession(tok); ok {
		t.Error("token signed under another secret must be rejected")
	}
}
