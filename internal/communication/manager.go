package communication

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"log/slog"
>>>>>>> origin/main
	"time"

<<<<<<< HEAD
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/mail"
	"github.com/robertpelloni/enterprise_sales_bot/internal/metrics"
=======
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
>>>>>>> origin/main
)

// Intent represents the classified purpose of an inbound message.
type Intent string

const (
<<<<<<< HEAD
	IntentTechnical      Intent = "Technical"
	IntentPricing        Intent = "Pricing"
	IntentObjection      Intent = "Objection"
	IntentMeetingRequest Intent = "MeetingRequest"
	IntentFollowUp       Intent = "FollowUp"
	IntentSpam           Intent = "Spam"
	IntentUnknown        Intent = "Unknown"
=======
	IntentTechnical		Intent	= "Technical"
	IntentPricing		Intent	= "Pricing"
	IntentObjection		Intent	= "Objection"
	IntentMeetingRequest	Intent	= "MeetingRequest"
	IntentFollowUp		Intent	= "FollowUp"
	IntentSpam		Intent	= "Spam"
	IntentUnknown		Intent	= "Unknown"
	IntentGeneral		Intent	= "General"
>>>>>>> origin/main
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
<<<<<<< HEAD
	db         *db.DB
	classifier IntentClassifier
	responder  ResponseGenerator
	strategy   SalesStrategy
	processor  OrderProcessor
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

=======
	db		*db.DB
	classifier	IntentClassifier
	responder	ResponseGenerator
	strategy	SalesStrategy
	processor	OrderProcessor
	sender		EmailSender	// nil = no email sending (log only)
	objections	*ObjectionLibrary
}

// NewManager creates a new communication Manager.
// sender is optional — if nil, outbound emails are logged but not sent.
// objections is optional — if nil, objection handling is disabled.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, sender EmailSender) *Manager {
	return &Manager{
		db:		database,
		classifier:	classifier,
		responder:	responder,
		strategy:	strategy,
		processor:	processor,
		sender:		sender,
	}
}

// SetObjectionLibrary attaches the objection handling library to this manager.
// If set, ProcessInbound will automatically detect and counter objections.
func (m *Manager) SetObjectionLibrary(lib *ObjectionLibrary) {
	m.objections = lib
}

>>>>>>> origin/main
// Run starts the periodic inbound communication processing loop.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

<<<<<<< HEAD
	log.Printf("Communication Manager: Background poller started (interval: %v)...", interval)
=======
	slog.Info(fmt.Sprintf("Communication Manager: Background poller started (interval: %v)...", interval))

	// Run immediately on startup
	m.pollAndProcess(ctx)
>>>>>>> origin/main

	for {
		select {
		case <-ctx.Done():
<<<<<<< HEAD
			log.Println("Communication Manager: Background poller stopping: Draining in-flight work...")
=======
			slog.Info("Communication Manager: Background poller stopping: Draining in-flight work...")
>>>>>>> origin/main
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

<<<<<<< HEAD
func (m *Manager) pollAndProcess(ctx context.Context) {
	// In a real scenario, this would poll an IMAP server or Webhook queue.
	// For this architecture, we simulate by checking for 'Researched' deals
	// that haven't had an outbound interaction yet (triggering initial outreach).
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		log.Printf("Comm Manager: Error polling deals: %v", err)
=======
// ExecutePoll manually triggers a poll and process cycle (exported for testing).
func (m *Manager) ExecutePoll(ctx context.Context) {
	m.pollAndProcess(ctx)
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	if m.db == nil {
		slog.Info("Comm Manager: DB unavailable, skipping poll cycle")
		return
	}
	// Check for 'Researched' deals that haven't had an outbound interaction yet
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info(fmt.Sprintf("Comm Manager: Error polling deals: %v", err))
>>>>>>> origin/main
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
<<<<<<< HEAD
			log.Printf("Comm Manager: Initiating autonomous outreach for deal %d to %s", deal.ID, contacts[0].Email)
			// Trigger outreach
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
				log.Printf("Comm Manager Error: Failed to initiate outreach for deal %d: %v", deal.ID, err)
			}
			if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil {
				log.Printf("Comm Manager Error: Failed to update deal state to OutreachSent for deal %d: %v", deal.ID, err)
=======
			slog.Info(fmt.Sprintf("Comm Manager: Initiating autonomous outreach for deal %d to %s", deal.ID, contacts[0].Email))
			// Trigger outreach
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil {
				slog.Info(fmt.Sprintf("Comm Manager Error: Failed to initiate outreach for deal %d: %v", deal.ID, err))
			}
			if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil {
				slog.Info(fmt.Sprintf("Comm Manager Error: Failed to update deal state to OutreachSent for deal %d: %v", deal.ID, err))
>>>>>>> origin/main
			}
>>>>>>> origin/main
		}
	}
}

<<<<<<< HEAD
// ProcessInbound handles a new inbound message from a contact.
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	// 1. Persist inbound interaction
	inbound := db.Interaction{
		ContactID: contact.ID,
		Channel:   "Email", // Default for now
		Direction: "Inbound",
		RawText:   text,
=======
// DefaultChannelForContact returns the channel to use when communicating with a contact.
// If the contact has a preferred channel set, that is used. Otherwise defaults to "email".
func DefaultChannelForContact(contact db.Contact) string {
	if contact.PreferredChannel != "" {
		return contact.PreferredChannel
	}
	return string(db.ChannelEmail)
}

// ProcessInbound handles a new inbound message from a contact.
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	// Determine the channel to use for this contact
	channel := DefaultChannelForContact(contact)

	// 1. Persist inbound interaction
	inbound := db.Interaction{
		ContactID:	contact.ID,
		Channel:	channel,
		Direction:	"Inbound",
		RawText:	text,
>>>>>>> origin/main
	}
	err := m.db.CreateInteraction(ctx, &inbound)
	if err != nil {
		return "", err
>>>>>>> origin/main
	}

	// 2. Classify intent
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil {
		return "", err
	}

<<<<<<< HEAD
	// Synchronize inbound interaction with the CRM
	if m.crmClient != nil {
		go m.syncInteractionWithRetry(ctx, contact.CompanyID, fmt.Sprintf("Inbound (%s): %s", intent, text))
	}

	// 2a. Decide next action using strategy engine
	var company *db.Company
	var deal *db.Deal
	var interactions []db.Interaction

	if m.db != nil && m.db.Conn != nil {
		var err error
		company, err = m.db.GetCompanyByID(ctx, contact.CompanyID)
		if err != nil {
			return "", fmt.Errorf("failed to get company for strategy: %w", err)
		}

		interactions, err = m.db.ListInteractionsByContact(ctx, contact.ID)
		if err != nil {
			slog.Warn("Comm Manager: failed to list interactions for strategy", "error", err)
		}

		deal, err = m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err != nil {
			return "", fmt.Errorf("failed to get deal for strategy: %w", err)
		}
	} else {
		// Mock contexts for verification without DB
		company = &db.Company{Name: "Mock Corp"}
		deal = &db.Deal{CurrentState: db.StateResearched}
	}

	salesCtx := SalesContext{
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: interactions,
		LatestIntent: intent,
=======
	// 2a. Decide next action using strategy engine
	company, err := m.db.GetCompanyByID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company for strategy: %w", err)
	}

	interactions, err := m.db.ListInteractionsByContact(ctx, contact.ID)
	if err != nil {
<<<<<<< HEAD
		log.Printf("Warning: failed to list interactions for strategy: %v", err)
=======
		slog.Info(fmt.Sprintf("Warning: failed to list interactions for strategy: %v", err))
>>>>>>> origin/main
	}

	deal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get deal for strategy: %w", err)
	}

	salesCtx := SalesContext{
<<<<<<< HEAD
		Company:      *company,
		Deal:         *deal,
		Contact:      contact,
		Interactions: interactions,
		LatestIntent: intent,
=======
		Company:	*company,
		Deal:		*deal,
		Contact:	contact,
		Interactions:	interactions,
		LatestIntent:	intent,
>>>>>>> origin/main
	}

	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil {
		return "", err
	}

	if action == ActionEscalate {
<<<<<<< HEAD
		log.Printf("UI: Deal %d escalated for human review.", salesCtx.Deal.ID)
=======
		slog.Info(fmt.Sprintf("UI: Deal %d escalated for human review.", salesCtx.Deal.ID))
>>>>>>> origin/main
		return "I've flagged this for our technical lead to review. We will get back to you shortly.", nil
	}

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil {
		return "", err
	}

<<<<<<< HEAD
	// 3a. Handle order processing and prompt optimization loop if deal was won
	if action == ActionAdvanceState {
		updatedDeal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err == nil && updatedDeal.CurrentState == db.StateClosedWon {
			log.Printf("Comm Manager: Deal %d won! Flagging past outbound interactions as successful.", updatedDeal.ID)
			// Mark all outbound interactions for this contact as successful to feed back into RAG
			for _, interaction := range interactions {
				if interaction.Direction == "Outbound" {
					if err := m.db.UpdateInteractionSuccess(ctx, interaction.ID, true); err != nil {
						log.Printf("Comm Manager Error: Failed to mark interaction %d as successful: %v", interaction.ID, err)
=======
	// 3a. Objection detection and counter-response injection
	var responseID string
	if m.objections != nil {
		sentiment := AnalyzeSentiment(text)
		if sentiment.Sentiment == SentimentNegative || sentiment.Sentiment == SentimentMixed {
			match := m.objections.MatchObjection(ctx, text, sentiment, deal.CurrentState)
			if match != nil {
				slog.Info("Objection detected",
					"category", match.Objection.Category,
					"objection", match.Objection.Title,
					"response_id", match.Response.ID,
					"score", match.Score,
				)
				// Prepend the counter-response to the generated reply
				replyText = match.Response.Text + "\n\n" + replyText

				responseID = match.Response.ID
				// Record usage for A/B testing
				m.objections.RecordOutcome(responseID, false)
			}
		}
	}

	// 3b. Handle order processing and prompt optimization loop if deal was won
	if action == ActionAdvanceState {
		updatedDeal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err == nil && updatedDeal.CurrentState == db.StateClosedWon {
			slog.Info(fmt.Sprintf("Comm Manager: Deal %d won! Flagging past outbound interactions as successful.", updatedDeal.ID))
			for _, interaction := range interactions {
				if interaction.Direction == "Outbound" {
if err := m.db.UpdateInteractionSuccess(ctx, interaction.ID, true); err != nil {
						slog.Info(fmt.Sprintf("Comm Manager Error: Failed to mark interaction %d as successful: %v", interaction.ID, err))
					}
					if interaction.ResponseID != "" && m.objections != nil {
						m.objections.RecordOutcome(interaction.ResponseID, true)
>>>>>>> origin/main
					}
				}
			}

			if m.processor != nil {
<<<<<<< HEAD
				log.Printf("Comm Manager: Triggering order processor for won deal %d", updatedDeal.ID)
				if err := m.processor.ProcessOrder(ctx, *updatedDeal); err != nil {
					log.Printf("Comm Manager Error: Order processing failed: %v", err)
=======
				slog.Info(fmt.Sprintf("Comm Manager: Triggering order processor for won deal %d", updatedDeal.ID))
				if err := m.processor.ProcessOrder(ctx, *updatedDeal); err != nil {
					slog.Info(fmt.Sprintf("Comm Manager Error: Order processing failed: %v", err))
>>>>>>> origin/main
				}
>>>>>>> origin/main
			}
		}
	}

	// 4. Persist outbound interaction
	outbound := db.Interaction{
<<<<<<< HEAD
		ContactID: contact.ID,
		Channel:   "Email",
		Direction: "Outbound",
		RawText:   replyText,
		Summary:   fmt.Sprintf("Reply to intent: %s", intent),
=======
		ContactID:	contact.ID,
		Channel:	channel,
		Direction:	"Outbound",
		RawText:	replyText,
		Summary:	fmt.Sprintf("Reply to intent: %s", intent),
		ResponseID:	responseID,
>>>>>>> origin/main
	}
	err = m.db.CreateInteraction(ctx, &outbound)
	if err != nil {
		return "", err
	}

<<<<<<< HEAD
	return replyText, nil
}
=======
	// 5. Actually send the communication if sender is configured
	if m.sender != nil && contact.Email != "" && channel == "email" {
		subject := fmt.Sprintf("Re: %s — TormentNexus", company.Name)
		if text == "START_OUTREACH" {
			subject = fmt.Sprintf("TormentNexus for %s — Quick Question", company.Name)
		}

		emailMsg := EmailMessage{
			To:		contact.Email,
			Subject:	subject,
			Body:		replyText,
		}

		if err := m.sender.Send(ctx, emailMsg); err != nil {
			slog.Info(fmt.Sprintf("Comm Manager: Email send failed to %s: %v", contact.Email, err))
		} else {
			slog.Info(fmt.Sprintf("Comm Manager: Email sent to %s (%s)", contact.Email, subject))
		}
	} else if channel != "email" {
		slog.Info(fmt.Sprintf("Comm Manager: Channel %q requires future implementation for %s — reply logged", channel, contact.Email))
	} else if m.sender == nil {
		slog.Info(fmt.Sprintf("Comm Manager: No email sender configured — reply logged but not sent to %s", contact.Email))
	}
>>>>>>> origin/main

	return replyText, nil
}

<<<<<<< HEAD
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
=======
// GetDB returns the database connection (used by IMAP receiver for contact lookup).
func (m *Manager) GetDB() *db.DB {
	return m.db
}

// ApproveDeal transitions a high-value deal from PendingApproval to Negotiating,
// signaling that a human has reviewed and authorized further autonomous action.
func (m *Manager) ApproveDeal(ctx context.Context, dealID int64) error {
	if err := m.db.UpdateDealState(ctx, dealID, db.StateNegotiating); err != nil {
		return fmt.Errorf("failed to approve deal %d: %w", dealID, err)
	}
	slog.Info(fmt.Sprintf("Manager: Deal %d approved by human review, now in Negotiating state", dealID))
	return nil
}

// RejectDeal transitions a high-value deal from PendingApproval to ClosedLost,
// signaling that a human has reviewed and decided not to pursue.
func (m *Manager) RejectDeal(ctx context.Context, dealID int64) error {
	if err := m.db.UpdateDealState(ctx, dealID, db.StateClosedLost); err != nil {
		return fmt.Errorf("failed to reject deal %d: %w", dealID, err)
	}
	slog.Info(fmt.Sprintf("Manager: Deal %d rejected by human review, marked as Closed_Lost", dealID))
	return nil
>>>>>>> origin/main
}
>>>>>>> origin/main
