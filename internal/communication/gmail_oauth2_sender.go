package communication

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// gmailAuthBreakerWindow is how long Send suppresses attempts after the
// provider rejects the credentials themselves (OAuth client disabled /
// refresh token revoked). Every caller path (cadence, comm manager)
// otherwise retries per deal and logs the same ERROR each time; the
// breaker collapses that to one trip log per window while still
// surfacing an error to every caller.
const gmailAuthBreakerWindow = 15 * time.Minute

// GmailOAuth2Sender implements EmailSender using Gmail SMTP with OAuth2 XOAUTH2.
type GmailOAuth2Sender struct {
	clientID     string
	clientSecret string
	refreshToken string
	fromEmail    string
	fromName     string

	mu            sync.Mutex
	authDownUntil time.Time
	lastAuthErr   error
	now           func() time.Time
	sendFn        func(to string, msg []byte) error
}

// NewGmailOAuth2Sender creates a new Gmail OAuth2 email sender.
func NewGmailOAuth2Sender() *GmailOAuth2Sender {
	s := &GmailOAuth2Sender{
		clientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		clientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		refreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN"),
		fromEmail:    os.Getenv("SMTP_FROM"),
		fromName:     os.Getenv("SMTP_FROM_NAME"),
		now:          time.Now,
	}
	s.sendFn = s.sendViaSMTP
	return s
}

// Send sends an email using Gmail SMTP with OAuth2 XOAUTH2.
func (s *GmailOAuth2Sender) Send(ctx context.Context, msg EmailMessage) error {
	if msg.To == "" {
		return fmt.Errorf("gmail oauth2: recipient address is empty")
	}

	// Build RFC 2822 message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	emailMsg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\nMIME-Version: 1.0\r\n\r\n%s",
		from, msg.To, msg.Subject, msg.Body)

	// Auth breaker: after a provider-side credential rejection, fail
	// fast for the window instead of re-fetching a token (and logging
	// the same ERROR) for every deal on every caller path.
	s.mu.Lock()
	if s.lastAuthErr != nil && s.now().Before(s.authDownUntil) {
		downErr, until := s.lastAuthErr, s.authDownUntil
		s.mu.Unlock()
		return fmt.Errorf("gmail oauth2: auth breaker open until %s: %w", until.Format(time.RFC3339), downErr)
	}
	s.mu.Unlock()

	// Send via SMTP with OAuth2
	err := s.sendFn(msg.To, []byte(emailMsg))
	if err != nil {
		if isProviderAuthError(err) {
			s.mu.Lock()
			tripped := s.lastAuthErr == nil || !s.now().Before(s.authDownUntil)
			s.lastAuthErr, s.authDownUntil = err, s.now().Add(gmailAuthBreakerWindow)
			s.mu.Unlock()
			if tripped {
				slog.Error("GmailOAuth2: provider auth rejected — auth breaker open", "until", s.authDownUntil.Format(time.RFC3339), "error", err)
			}
			return err
		}
		slog.Error("GmailOAuth2: Failed to send email", "to", msg.To, "error", err)
		return err
	}

	s.mu.Lock()
	s.lastAuthErr, s.authDownUntil = nil, time.Time{}
	s.mu.Unlock()

	slog.Info("GmailOAuth2: Email sent", "to", msg.To, "subject", msg.Subject)
	return nil
}

// sendViaSMTP sends email using SMTP with OAuth2 XOAUTH2 mechanism
func (s *GmailOAuth2Sender) sendViaSMTP(to string, msg []byte) error {
	// Get OAuth2 token
	config := &oauth2.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://mail.google.com/"},
	}

	token := &oauth2.Token{
		RefreshToken: s.refreshToken,
		Expiry:       time.Now().Add(-time.Minute),
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Connect to Gmail SMTP (port 587 uses STARTTLS, not direct TLS)
	addr := "smtp.gmail.com:587"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// STARTTLS
	tlsConfig := &tls.Config{ServerName: "smtp.gmail.com"}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// XOAUTH2 auth string
	authStr := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", s.fromEmail, newToken.AccessToken)
	auth := xoauth2Auth(authStr)

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	// Set sender
	if err = client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Send body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("write body failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}

	return client.Quit()
}

// xoauth2Auth implements XOAUTH2 authentication
type xoauth2Auth string

func (a xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(a), nil
}

func (a xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server response: %s", string(fromServer))
	}
	return nil, nil
}

// HealthCheck verifies Gmail OAuth2 credentials are configured.
func (s *GmailOAuth2Sender) HealthCheck(ctx context.Context) error {
	if s.clientID == "" || s.clientSecret == "" || s.refreshToken == "" {
		return fmt.Errorf("gmail OAuth2 not configured: set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REFRESH_TOKEN")
	}
	slog.Info("GmailOAuth2: Health check passed (credentials configured)")
	return nil
}
