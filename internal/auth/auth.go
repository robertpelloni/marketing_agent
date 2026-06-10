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

// Authenticator handles session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
	sessions          map[string]time.Time
	mu                sync.RWMutex
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator() *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin" // Default for development
	}

	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
		sessions:          make(map[string]time.Time),
	}
}

func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback_session_id_" + time.Now().String()
	}
	return hex.EncodeToString(b)
}

// Login verifies the password and sets a session cookie.
func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != a.adminPasswordHash {
		return "", errors.New("invalid password")
	}

	sessionID := generateSessionID()

	a.mu.Lock()
	a.sessions[sessionID] = time.Now().Add(24 * time.Hour)
	a.mu.Unlock()

	return sessionID, nil
}

// Middleware provides an HTTP middleware to protect routes.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and webhook endpoints
		// SECURITY: /api/v1/test/simulate_inbound is now protected by auth in production-hardening phase.
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
		expiry, ok := a.sessions[cookie.Value]
		a.mu.RUnlock()

		if !ok || time.Now().After(expiry) {
			if ok {
				a.mu.Lock()
				delete(a.sessions, cookie.Value)
				a.mu.Unlock()
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandleLogin processes the login form submission.
func (a *Authenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Sales Bot Login</title></head>
			<body style="font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f4f4f9;">
				<form method="POST" style="background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
					<h1>Login</h1>
					<input type="password" name="password" placeholder="Password" style="display: block; width: 100%; padding: 10px; margin-bottom: 20px;">
					<button type="submit" style="width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">Login</button>
				</form>
			</body>
			</html>
		`))
		return
	}

	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		token, err := a.Login(password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     a.sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
