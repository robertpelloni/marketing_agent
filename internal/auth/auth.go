package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// sessionTTL bounds every issued session (matches the cookie Expires).
const sessionTTL = 24 * time.Hour

// Authenticator handles simple session-based authentication.
type Authenticator struct {
	adminPasswordHash string
	sessionCookieName string
	sessionSecret     []byte
	db                *sql.DB
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(db *sql.DB) *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin" // Default for development
	}

	// Session signing secret: SESSION_SECRET when set, otherwise derived
	// from the admin password so existing deployments keep working with
	// no new required env (rotating either invalidates all sessions,
	// which is the desired response to a credential change).
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = password
	}
	key := sha256.Sum256([]byte("sales-bot-session:" + secret))

	hash := sha256.Sum256([]byte(password))
	return &Authenticator{
		adminPasswordHash: hex.EncodeToString(hash[:]),
		sessionCookieName: "sales_bot_session",
		sessionSecret:     key[:],
		db:                db,
	}
}

// sign mints a session token: v1.<subject>.<expiryUnix>.<hmacHex>.
// The previous static values ("authorized_admin_session", bare
// "company_<id>") were forgeable by anyone who read the source; a token
// now proves possession of the signing secret and expires.
func (a *Authenticator) sign(subject string, expiry time.Time) string {
	payload := fmt.Sprintf("v1.%s.%d", subject, expiry.Unix())
	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

// IssueSession mints a signed session token for a subject ("admin" or
// "company_<id>") expiring after sessionTTL.
func (a *Authenticator) IssueSession(subject string) string {
	return a.sign(subject, time.Now().Add(sessionTTL))
}

// ValidateSession verifies signature and expiry, returning the subject.
func (a *Authenticator) ValidateSession(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 4 || parts[0] != "v1" {
		return "", false
	}
	subject := parts[1]
	if subject != "admin" && !strings.HasPrefix(subject, "company_") {
		return "", false
	}
	exp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return "", false
	}
	payload := strings.Join(parts[:3], ".")
	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[3]), []byte(expected)) {
		return "", false
	}
	return subject, true
}

// Subject returns the validated session subject for a request.
func (a *Authenticator) Subject(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(a.sessionCookieName)
	if err != nil {
		return "", false
	}
	return a.ValidateSession(cookie.Value)
}

// CompanyID returns the tenant company id for a company session.
func (a *Authenticator) CompanyID(r *http.Request) (int64, bool) {
	subject, ok := a.Subject(r)
	if !ok {
		return 0, false
	}
	var id int64
	if _, err := fmt.Sscanf(subject, "company_%d", &id); err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// Login verifies the password and sets a session cookie.
func (a *Authenticator) Login(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	if hex.EncodeToString(hash[:]) == a.adminPasswordHash {
		return a.IssueSession("admin"), nil
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
			return a.IssueSession(fmt.Sprintf("company_%d", companyID)), nil
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

		if _, ok := a.Subject(r); !ok {
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
