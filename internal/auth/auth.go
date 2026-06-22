package auth

import (
<<<<<<< HEAD
	"crypto/rand"
=======
>>>>>>> origin/main
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
<<<<<<< HEAD
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
=======
	"time"
)

// Authenticator handles simple session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator() *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin" // Default for development
	}

>>>>>>> origin/main
	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
<<<<<<< HEAD
		csrfTokens:        make(map[string]string),
	}
}

=======
	}
}

// Login verifies the password and sets a session cookie.
>>>>>>> origin/main
func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) != a.adminPasswordHash {
		return "", errors.New("invalid password")
	}

<<<<<<< HEAD
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
=======
	// In a real system, we'd generate a secure random session ID and store it in a DB/Redis.
	// For this module, we use a simple static session token for the admin.
	return "authorized_admin_session", nil
}

// Middleware provides an HTTP middleware to protect routes.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and webhook endpoints
>>>>>>> origin/main
		if r.URL.Path == "/health" || r.URL.Path == "/health/detailed" || r.URL.Path == "/api/v1/webhook/github" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
<<<<<<< HEAD
		cookie, err := r.Cookie(a.sessionCookieName)
		if err != nil {
=======

		cookie, err := r.Cookie(a.sessionCookieName)
		if err != nil || cookie.Value != "authorized_admin_session" {
>>>>>>> origin/main
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

<<<<<<< HEAD
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

=======
>>>>>>> origin/main
		next.ServeHTTP(w, r)
	})
}

<<<<<<< HEAD
func (a *Authenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html><head><title>Sales Bot Login</title></head>
=======
// HandleLogin processes the login form submission.
func (a *Authenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Sales Bot Login</title></head>
>>>>>>> origin/main
			<body style="font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f4f4f9;">
				<form method="POST" style="background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
					<h1>Login</h1>
					<input type="password" name="password" placeholder="Password" style="display: block; width: 100%; padding: 10px; margin-bottom: 20px;">
					<button type="submit" style="width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">Login</button>
				</form>
<<<<<<< HEAD
			</body></html>
		`)) // #nosec G104
		return
	}
	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		sessionID, err := a.Login(password)
=======
			</body>
			</html>
		`)); err != nil {
			// Log the error if the write fails (e.g. broken pipe)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		token, err := a.Login(password)
>>>>>>> origin/main
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
<<<<<<< HEAD
		http.SetCookie(w, &http.Cookie{
			Name:     a.sessionCookieName,
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})
=======

		http.SetCookie(w, &http.Cookie{
			Name:     a.sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			Expires:  time.Now().Add(24 * time.Hour),
		})

>>>>>>> origin/main
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
