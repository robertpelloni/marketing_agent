package mail

import (
	"context"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// InboundProcessor handles the autonomous response for an inbound email.
type InboundProcessor interface {
	ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error)
}

// IMAPPoller periodically checks an IMAP server for new inbound emails.
type IMAPPoller struct {
	db      *db.DB
	comm    InboundProcessor
	address string
	user    string
	pass    string
}

// NewIMAPPoller creates a new IMAP poller.
func NewIMAPPoller(database *db.DB, comm InboundProcessor, address, user, pass string) *IMAPPoller {
	return &IMAPPoller{
		db:      database,
		comm:    comm,
		address: address,
		user:    user,
		pass:    pass,
	}
}

// Run starts the periodic IMAP polling loop.
func (p *IMAPPoller) Run(ctx context.Context, interval time.Duration) {
	if p.address == "" || p.address == "localhost" {
		slog.Warn("IMAP: No real server configured, skipping IMAP poller")
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("IMAP: Poller started", "address", p.address, "interval", interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("IMAP: Poller stopping")
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *IMAPPoller) poll(ctx context.Context) {
	slog.Debug("IMAP: Executing poll cycle...")
	// In a real implementation, we would use a library like 'github.com/emersion/go-imap'
	// to connect, authenticate, search for UNSEEN messages, and parse them.
	// For now, we provide the architectural hook for live integration.
}
