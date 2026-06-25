package web

import (
	"strings"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

// generateCSRFToken generates a secure random token.
func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// csrfMiddleware handles setting and verifying CSRF tokens for state-changing requests.
func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for APIs and webhooks (which use Bearer/HMAC auth)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Ensure every response gets a CSRF cookie
		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value == "" {
			token := generateCSRFToken()
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    token,
				Path:     "/",
				HttpOnly: false, // Must be readable by JS if submitting via fetch, but we use forms
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			// Since we just set it, populate it for the current request context
			r.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
		}

		// Verify token on state-changing requests
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			cookieToken := ""
			if c, err := r.Cookie("csrf_token"); err == nil {
				cookieToken = c.Value
			}

			// Check form value or header
			formToken := r.FormValue("csrf_token")
			if formToken == "" {
				formToken = r.Header.Get("X-CSRF-Token")
			}

			if cookieToken == "" || formToken == "" || cookieToken != formToken {
				http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
