package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"golang.org/x/time/rate"
)

func TestAuthenticationIntegration(t *testing.T) {
	// 1. Setup Server with Authenticator and Mock Handlers
	os.Setenv("ADMIN_PASSWORD", "testpass")
	defer os.Unsetenv("ADMIN_PASSWORD")

	authenticator := auth.NewAuthenticator()
	mux := http.NewServeMux()

	// Register a dummy handler for the root that doesn't need a DB
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Dashboard Content"))
	})

	// Register the real login handler
	mux.HandleFunc("/login", authenticator.HandleLogin)

	// Register public handlers
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &Server{
		auth:    authenticator,
		mux:     mux,
		limiter: rate.NewLimiter(10, 20),
	}

	// 2. Test Unauthenticated Access to Dashboard (Should Redirect)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status SeeOther (303) for unauthenticated dashboard, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("Expected redirect to /login, got %s", loc)
	}

	// 3. Test Successful Login
	loginData := url.Values{}
	loginData.Set("password", "testpass")
	req = httptest.NewRequest("POST", "/login", strings.NewReader(loginData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status SeeOther (303) after login, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "sales_bot_session" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie 'sales_bot_session' not found")
	}

	// 4. Test Authenticated Access to Dashboard
	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK (200) for authenticated dashboard, got %d", w.Code)
	}

	// 5. Test Public Endpoints (Should not redirect)
	publicEndpoints := []string{"/health"}
	for _, ep := range publicEndpoints {
		req = httptest.NewRequest("GET", ep, nil)
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		if w.Code == http.StatusSeeOther {
			t.Errorf("Public endpoint %s triggered redirect to login", ep)
		}
	}
}
