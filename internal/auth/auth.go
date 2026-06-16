package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
)

type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
	csrfTokens        map[string]string // session -> token
	mu                sync.RWMutex
}

func NewAuthenticator() *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin"
	}
	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
		csrfTokens:        make(map[string]string),
	}
}

func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != a.adminPasswordHash {
		return "", errors.New("invalid password")
	}

	// Generate a secure random session ID
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	sessionID := hex.EncodeToString(b)

	// Generate a CSRF token for this session
	t := make([]byte, 32)
	_, _ = rand.Read(t)
	csrfToken := hex.EncodeToString(t)

	a.mu.Lock()
	a.csrfTokens[sessionID] = csrfToken
	a.mu.Unlock()

	return sessionID, nil
}

func (a *Authenticator) GetCSRFToken(r *http.Request) string {
	cookie, err := r.Cookie(a.sessionCookieName)
	if err != nil {
		return ""
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.csrfTokens[cookie.Value]
}

func (a *Authenticator) ValidateCSRF(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	token := r.FormValue("csrf_token")
	expected := a.GetCSRFToken(r)
	return token != "" && token == expected
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/health/detailed" || r.URL.Path == "/api/v1/webhook/github" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(a.sessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		a.mu.RLock()
		_, ok := a.csrfTokens[cookie.Value]
		a.mu.RUnlock()

		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !a.ValidateCSRF(r) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html><head><title>Sales Bot Login</title></head>
			<body style="font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f4f4f9;">
				<form method="POST" style="background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
					<h1>Login</h1>
					<input type="password" name="password" placeholder="Password" style="display: block; width: 100%; padding: 10px; margin-bottom: 20px;">
					<button type="submit" style="width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">Login</button>
				</form>
			</body></html>
		`)) // #nosec G104
		return
	}
	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		sessionID, err := a.Login(password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     a.sessionCookieName,
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
