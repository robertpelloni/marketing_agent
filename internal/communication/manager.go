package communication

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/mail"
	"github.com/robertpelloni/enterprise_sales_bot/internal/metrics"
)

// Intent represents the classified purpose of an inbound message.
type Intent string

const (
	IntentTechnical      Intent = "Technical"
	IntentPricing        Intent = "Pricing"
	IntentObjection      Intent = "Objection"
	IntentMeetingRequest Intent = "MeetingRequest"
	IntentFollowUp       Intent = "FollowUp"
	IntentSpam           Intent = "Spam"
	IntentUnknown        Intent = "Unknown"
)

// IntentClassifier defines the interface for categorizing inbound communication.
type IntentClassifier interface {
	Classify(ctx context.Context, text string) (Intent, error)
}

// ResponseGenerator defines the interface for creating tailored replies.
type ResponseGenerator interface {
	Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error)
}

// OrderProcessor defines the interface for fulfillment after a win.
type OrderProcessor interface {
	ProcessOrder(ctx context.Context, deal db.Deal) error
}

// Manager coordinates the inbound communication state machine.
type Manager struct {
	db          *db.DB
	classifier  IntentClassifier
	responder   ResponseGenerator
	strategy    SalesStrategy
	processor   OrderProcessor
	crmClient   crm.CRMClient
	emailSender mail.EmailSender
}

// NewManager creates a new communication Manager.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, crmClient crm.CRMClient, emailSender mail.EmailSender) *Manager {
	return &Manager{
		db:          database,
		classifier:  classifier,
		responder:   responder,
		strategy:    strategy,
		processor:   processor,
		crmClient:   crmClient,
		emailSender: emailSender,
	}
}

// Run starts the periodic inbound communication processing loop.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Communication Manager: Background poller started", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Communication Manager: Background poller stopping: Draining in-flight work")
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	// In a real scenario, this would poll an IMAP server or Webhook queue.
	// For this architecture, we simulate by checking for 'Researched' deals
	// that haven't had an outbound interaction yet (triggering initial outreach).
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Error("Comm Manager: Error polling deals", "error", err)
		return
	}

	for _, deal := range deals {
		contacts, err := m.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 {
			continue
		}

		// Check if we already sent outreach
		interactions, _ := m.db.ListInteractionsByContact(ctx, contacts[0].ID)
		hasOutbound := false
		for _, i := range interactions {
			if i.Direction == "Outbound" {
				hasOutbound = true
				break
			}
		}

		if !hasOutbound {
			slog.Info("Comm Manager: Initiating autonomous outreach", "deal_id", deal.ID, "contact_email", contacts[0].Email)
			// Trigger outreach
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
				slog.Error("Comm Manager: Failed to initiate outreach", "deal_id", deal.ID, "error", err)
			}
			if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil {
				slog.Error("Comm Manager: Failed to update deal state to OutreachSent", "deal_id", deal.ID, "error", err)
			}
		}
	}
}

// ProcessInbound handles a new inbound message from a contact.
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	// 1. Persist inbound interaction
	inbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   "Email", // Default for now
		Direction: "Inbound",
		RawText:   text,
	}
	err := m.db.CreateInteraction(ctx, &inbound)
	if err == nil {
		metrics.InteractionsProcessed.WithLabelValues("Inbound", "Email").Inc()
	}
	if err != nil {
		return "", err
	}

	// 2. Classify intent
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil {
		return "", err
	}

	// Synchronize inbound interaction with the CRM
	if m.crmClient != nil {
		go m.syncInteractionWithRetry(ctx, contact.CompanyID, fmt.Sprintf("Inbound (%s): %s", intent, text))
	}

	// 2a. Decide next action using strategy engine
	company, err := m.db.GetCompanyByID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company for strategy: %w", err)
	}

	interactions, err := m.db.ListInteractionsByContact(ctx, contact.ID)
	if err != nil {
		slog.Warn("Comm Manager: failed to list interactions for strategy", "error", err)
	}

	deal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get deal for strategy: %w", err)
	}

	salesCtx := SalesContext{
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: interactions,
		LatestIntent: intent,
	}

	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil {
		return "", err
	}

	if action == ActionEscalate {
		slog.Info("UI: Deal escalated for human review", "deal_id", salesCtx.Deal.ID)
		return "I've flagged this for our technical lead to review. We will get back to you shortly.", nil
	}

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil {
		return "", err
	}

	// 3a. Handle order processing and prompt optimization loop if deal was won
	if action == ActionAdvanceState {
		updatedDeal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err == nil && updatedDeal.CurrentState == db.StateClosedWon {
			slog.Info("Comm Manager: Deal won! Flagging past outbound interactions as successful", "deal_id", updatedDeal.ID)
			metrics.DealsWon.Inc()
			// Mark all outbound interactions for this contact as successful to feed back into RAG
			for _, interaction := range interactions {
				if interaction.Direction == "Outbound" {
					if err := m.db.UpdateInteractionSuccess(ctx, interaction.ID, true); err != nil {
						slog.Error("Comm Manager: Failed to mark interaction as successful", "interaction_id", interaction.ID, "error", err)
					}
				}
			}

			if m.processor != nil {
				slog.Info("Comm Manager: Triggering order processor", "deal_id", updatedDeal.ID)
				if err := m.processor.ProcessOrder(ctx, *updatedDeal); err != nil {
					slog.Error("Comm Manager: Order processing failed", "deal_id", updatedDeal.ID, "error", err)
				}
			}
		}
	}

	// 4. Persist outbound interaction
	outbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   "Email",
		Direction: "Outbound",
		RawText:   replyText,
		Summary:   fmt.Sprintf("Reply to intent: %s", intent),
	}
	err = m.db.CreateInteraction(ctx, &outbound)
	if err == nil {
		metrics.InteractionsProcessed.WithLabelValues("Outbound", "Email").Inc()
	}
	if err != nil {
		return "", err
	}

	// Synchronize outbound interaction with the CRM and send the real email
	go func() {
		asyncCtx := context.Background()
		subject := fmt.Sprintf("Follow-up: TormentNexus for %s", salesCtx.Company.Name)

		// 1. Primary: Send via SMTP if configured
		if m.emailSender != nil {
			if err := m.emailSender.Send(asyncCtx, contact.Email, subject, replyText); err != nil {
				slog.Error("Comm Manager: Direct SMTP delivery failed", "contact_email", contact.Email, "error", err)
			}
		}

		// 2. Secondary/Record: Send/Log via CRM
		if m.crmClient != nil {
			if err := m.crmClient.SendEmail(asyncCtx, contact, subject, replyText); err != nil {
				slog.Error("Comm Manager: Failed to record email in CRM", "contact_email", contact.Email, "error", err)
			}
			m.syncInteractionWithRetry(asyncCtx, contact.CompanyID, fmt.Sprintf("Outbound (Reply to %s): %s", intent, replyText))
		}
	}()

	return replyText, nil
}

func (m *Manager) syncInteractionWithRetry(ctx context.Context, companyID int64, note string) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// In a real system, we'd look up the CRM Deal ID.
		// For this integration, we use the local CompanyID as a proxy for the deal ID in the CRM interface.
		if err := m.crmClient.SyncInteraction(ctx, companyID, note); err != nil {
			slog.Warn("Comm Manager: Failed to sync interaction to CRM", "attempt", i+1, "max_retries", maxRetries, "company_id", companyID, "error", err)
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		return
	}
	slog.Error("Comm Manager: CRM interaction synchronization failed after all attempts", "company_id", companyID)
}
