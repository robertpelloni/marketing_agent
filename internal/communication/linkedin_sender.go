package communication

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type LinkedInSender struct {
	Username string
	Password string
}

type LinkedInMessage struct {
	ProfileURL string
	Subject    string
	Body       string
}

func NewLinkedInSender() *LinkedInSender {
	return &LinkedInSender{
		Username: os.Getenv("LINKEDIN_USERNAME"),
		Password: os.Getenv("LINKEDIN_PASSWORD"),
	}
}

func (l *LinkedInSender) Send(ctx context.Context, msg LinkedInMessage) error {
	if l.Username == "" || l.Password == "" {
		slog.Info("LinkedInSender: No credentials, logging message (simulation)")
		return l.simulateSend(ctx, msg)
	}
	slog.Info("LinkedInSender: Credentials configured, browser automation pending. Logging message.")
	return l.simulateSend(ctx, msg)
}

func (l *LinkedInSender) SendMessage(ctx context.Context, profileURL, message string) error {
	return l.Send(ctx, LinkedInMessage{ProfileURL: profileURL, Body: message})
}

func (l *LinkedInSender) simulateSend(ctx context.Context, msg LinkedInMessage) error {
	slog.Info("LinkedInSender [SIMULATION] — Would send message",
		"to", msg.ProfileURL,
		"subject", msg.Subject,
		"body", msg.Body,
		"at", time.Now().Format(time.RFC3339))
	return nil
}

func (l *LinkedInSender) HealthCheck(ctx context.Context) error {
	if l.Username == "" || l.Password == "" {
		return fmt.Errorf("LinkedIn credentials not configured")
	}
	return nil
}

func (l *LinkedInSender) SendConnectionRequest(ctx context.Context, profileURL, note string) error {
	slog.Info("LinkedInSender: Connection request (simulation)", "to", profileURL, "note", note)
	return nil
}
