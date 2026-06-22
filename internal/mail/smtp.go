package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
)

// EmailSender defines an interface for sending emails directly.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPSender implements EmailSender using standard SMTP.
type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// NewSMTPSender creates a new SMTPSender.
func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	slog.Info("SMTP: Sending email", "to", to, "subject", subject)

	if s.Host == "localhost" || s.Host == "" {
		slog.Warn("SMTP: No real host configured, skipping actual delivery.")
		return nil
	}

	err := smtp.SendMail(addr, auth, s.From, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send smtp mail: %w", err)
	}

	return nil
}
