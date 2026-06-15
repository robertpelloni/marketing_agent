package communication

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// LinkedInSenderInterface defines the operations for LinkedIn outreach.
type LinkedInSenderInterface interface {
	Send(ctx context.Context, msg LinkedInMessage) error
	SendConnectionRequest(ctx context.Context, profileURL, note string) error
	HealthCheck(ctx context.Context) error
}

// LinkedInSender sends messages via LinkedIn messaging.
// Note: LinkedIn does not provide a public API for direct messaging.
// This implementation uses:
// - Simulation fallback when credentials are not configured
// - Placeholder for future headless browser automation (rod/chromedp)
type LinkedInSender struct {
	Username string
	Password string
}

// LinkedInMessage represents a message to send via LinkedIn.
type LinkedInMessage struct {
	ProfileURL string // LinkedIn profile URL of the recipient
	Subject    string // Subject line (note subject)
	Body       string // Message body content
}

// NewLinkedInSender creates a new LinkedInSender.
func NewLinkedInSender() *LinkedInSender {
	return &LinkedInSender{
		Username: os.Getenv("LINKEDIN_USERNAME"),
		Password: os.Getenv("LINKEDIN_PASSWORD"),
	}
}

// Send sends a LinkedIn message to a recipient.
// Falls back to simulation when credentials are not configured.
func (l *LinkedInSender) Send(ctx context.Context, msg LinkedInMessage) error {
	if l.Username == "" || l.Password == "" {
		log.Println("LinkedInSender: No LINKEDIN_USERNAME/PASSWORD configured, logging message (simulation)")
		return l.simulateSend(ctx, msg)
	}

	// Future: Implement real LinkedIn message sending via headless browser
	log.Println("LinkedInSender: Credentials configured but browser automation not yet implemented. Logging message.")
	return l.simulateSend(ctx, msg)
}

// simulateSend logs the message that would be sent.
func (l *LinkedInSender) simulateSend(ctx context.Context, msg LinkedInMessage) error {
	log.Printf(`LinkedInSender [SIMULATION] — Would send LinkedIn message:
  To: %s
  Subject: %s
  Body: %s
  Sent at: %s`,
		msg.ProfileURL,
		msg.Subject,
		msg.Body,
		time.Now().Format(time.RFC3339),
	)
	return nil
}

// HealthCheck verifies LinkedIn credentials are configured.
func (l *LinkedInSender) HealthCheck(ctx context.Context) error {
	if l.Username == "" || l.Password == "" {
		return fmt.Errorf("LINKEDIN_USERNAME and LINKEDIN_PASSWORD must be configured for LinkedIn messaging")
	}
	log.Println("LinkedInSender: Health check passed (credentials configured)")
	return nil
}

// SendConnectionRequest sends a LinkedIn connection request with a note.
// Future implementation will use headless browser automation.
func (l *LinkedInSender) SendConnectionRequest(ctx context.Context, profileURL, note string) error {
	if l.Username == "" || l.Password == "" {
		log.Printf("LinkedInSender [SIMULATION] — Would send connection request to %s with note: %s", profileURL, note)
		return nil
	}

	log.Printf("LinkedInSender: Connection request to %s — browser automation pending", profileURL)
	return nil
}

// MockLinkedInSender is a no-op LinkedIn sender for testing.
type MockLinkedInSender struct {
	SentMessages []LinkedInMessage
	Connections  []string
}

func (m *MockLinkedInSender) Send(ctx context.Context, msg LinkedInMessage) error {
	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

func (m *MockLinkedInSender) SendConnectionRequest(ctx context.Context, profileURL, note string) error {
	m.Connections = append(m.Connections, profileURL)
	return nil
}

func (m *MockLinkedInSender) HealthCheck(ctx context.Context) error {
	return nil
}
