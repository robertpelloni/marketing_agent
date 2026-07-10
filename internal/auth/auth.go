package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Authenticator handles simple session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
	db                *sql.DB
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(db *sql.DB) *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin" // Default for development
	}

	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
		db:                db,
	}
}

// Login verifies the password and sets a session cookie.
func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) == a.adminPasswordHash {
		return "authorized_admin_session", nil
	}

	// Check if this is a paid company logging in using domain, name, or Stripe customer ID
	if a.db != nil {
		var companyID int64
		var tier string
		var state string
		query := `
			SELECT c.id, s.tier, s.state 
			FROM companies c
			JOIN subscriptions s ON c.id = s.company_id
			WHERE (c.domain = $1 OR c.name = $1 OR s.stripe_customer_id = $1)
			  AND s.state IN ('active', 'trialing')
			LIMIT 1`
		err := a.db.QueryRow(query, password).Scan(&companyID, &tier, &state)
		if err == nil {
			return fmt.Sprintf("company_%d", companyID), nil
		}
	}

	return "", errors.New("invalid password or inactive company subscription")
}

// Middleware provides an HTTP middleware to protect routes.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and webhook endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/health/detailed" || r.URL.Path == "/api/v1/webhook/github" || r.URL.Path == "/login" || r.URL.Path == "/demo" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(a.sessionCookieName)
		if err != nil || (cookie.Value != "authorized_admin_session" && !strings.HasPrefix(cookie.Value, "company_")) {
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
		_, _ = w.Write([]byte(`
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
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
