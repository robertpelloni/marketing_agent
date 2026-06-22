package auth

import (
<<<<<<< HEAD
	"crypto/rand"
=======
	"crypto/sha256"
>>>>>>> origin/main
	"encoding/hex"
	"errors"
	"net/http"
	"os"
<<<<<<< HEAD
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Authenticator handles session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
	sessions          map[string]time.Time
	mu                sync.RWMutex
=======
	"time"
)

// Authenticator handles simple session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
>>>>>>> origin/main
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator() *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin" // Default for development
	}

<<<<<<< HEAD
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &Authenticator{
		adminPasswordHash: string(hash),
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
	err := bcrypt.CompareHashAndPassword([]byte(a.adminPasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	sessionID := generateSessionID()

	a.mu.Lock()
	a.sessions[sessionID] = time.Now().Add(24 * time.Hour)
	a.mu.Unlock()

	return sessionID, nil
=======
	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
	}
}

// Login verifies the password and sets a session cookie.
func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != a.adminPasswordHash {
		return "", errors.New("invalid password")
	}

	// In a real system, we'd generate a secure random session ID and store it in a DB/Redis.
	// For this module, we use a simple static session token for the admin.
	return "authorized_admin_session", nil
>>>>>>> origin/main
}

// Middleware provides an HTTP middleware to protect routes.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
<<<<<<< HEAD
		// Skip auth for health, webhook, and test endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/health/detailed" || r.URL.Path == "/api/v1/webhook/github" || r.URL.Path == "/api/v1/test/simulate_inbound" || r.URL.Path == "/login" {
=======
		// Skip auth for health and webhook endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/health/detailed" || r.URL.Path == "/api/v1/webhook/github" || r.URL.Path == "/login" {
>>>>>>> origin/main
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(a.sessionCookieName)
<<<<<<< HEAD
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
=======
		if err != nil || cookie.Value != "authorized_admin_session" {
>>>>>>> origin/main
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
<<<<<<< HEAD
		w.Write([]byte(`
=======
		if _, err := w.Write([]byte(`
>>>>>>> origin/main
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
<<<<<<< HEAD
		`))
=======
		`)); err != nil {
			// Log the error if the write fails (e.g. broken pipe)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
>>>>>>> origin/main
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
<<<<<<< HEAD
=======
			Secure:   true,
>>>>>>> origin/main
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
