package auth
import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
	"golang.org/x/crypto/bcrypt"
)
type Authenticator struct {
	adminPasswordHash string; sessionCookieName string; sessions map[string]time.Time; mu sync.RWMutex
}
func NewAuthenticator() *Authenticator {
	password := os.Getenv("ADMIN_PASSWORD"); if password == "" { password = "admin" }
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &Authenticator{adminPasswordHash: string(hash), sessionCookieName: "sales_bot_session", sessions: make(map[string]time.Time)}
}
func (a *Authenticator) Login(password string) (string, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(a.adminPasswordHash), []byte(password)); err != nil { return "", errors.New("invalid password") }
	b := make([]byte, 32); rand.Read(b); token := hex.EncodeToString(b)
	a.mu.Lock(); a.sessions[token] = time.Now().Add(24 * time.Hour); a.mu.Unlock(); return token, nil
}
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/health" || p == "/health/detailed" || p == "/metrics" || p == "/api/v1/webhook/github" || p == "/api/v1/test/simulate_inbound" || p == "/login" { next.ServeHTTP(w, r); return }
		cookie, err := r.Cookie(a.sessionCookieName)
		if err != nil { http.Redirect(w, r, "/login", 303); return }
		a.mu.RLock(); exp, ok := a.sessions[cookie.Value]; a.mu.RUnlock()
		if !ok || time.Now().After(exp) { http.Redirect(w, r, "/login", 303); return }
		next.ServeHTTP(w, r)
	})
}
func (a *Authenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" { w.Header().Set("Content-Type", "text/html"); w.Write([]byte(`<html><body><form method="POST"><h1>Login</h1><input type="password" name="password"><button type="submit">Login</button></form></body></html>`)) }
	if r.Method == "POST" {
		if token, err := a.Login(r.FormValue("password")); err == nil {
			http.SetCookie(w, &http.Cookie{Name: a.sessionCookieName, Value: token, Path: "/", HttpOnly: true, Expires: time.Now().Add(24 * time.Hour)})
			http.Redirect(w, r, "/", 303)
		} else { http.Error(w, "Unauthorized", 401) }
	}
}
