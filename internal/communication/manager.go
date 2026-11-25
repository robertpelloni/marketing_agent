package communication

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
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
	IntentGeneral        Intent = "General"
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
	db         *db.DB
	classifier IntentClassifier
	responder  ResponseGenerator
	strategy   SalesStrategy
	processor  OrderProcessor
	sender     EmailSender // nil = no email sending (log only)
	objections *ObjectionLibrary
}

// NewManager creates a new communication Manager.
// sender is optional — if nil, outbound emails are logged but not sent.
// objections is optional — if nil, objection handling is disabled.
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, sender EmailSender) *Manager {
	return &Manager{
		db:         database,
		classifier: classifier,
		responder:  responder,
		strategy:   strategy,
		processor:  processor,
		sender:     sender,
	}
}

// SetObjectionLibrary attaches the objection handling library to this manager.
// If set, ProcessInbound will automatically detect and counter objections.
func (m *Manager) SetObjectionLibrary(lib *ObjectionLibrary) {
	m.objections = lib
}

// Run starts the periodic inbound communication processing loop.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("Communication Manager: Background poller started (interval: %v)...", interval))

	// Run immediately on startup
	m.pollAndProcess(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Communication Manager: Background poller stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			m.pollAndProcess(ctx)
		}
	}
}

// ExecutePoll manually triggers a poll and process cycle (exported for testing).
func (m *Manager) ExecutePoll(ctx context.Context) {
	m.pollAndProcess(ctx)
}

func (m *Manager) pollAndProcess(ctx context.Context) {
	if m.db == nil {
		slog.Info("Comm Manager: DB unavailable, skipping poll cycle")
		return
	}

	// Process Researched deals: initiate first outreach
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil {
		slog.Info(fmt.Sprintf("Comm Manager: Error polling deals: %v", err))
	} else {
		for _, deal := range deals {
			contacts, err := m.db.ListContactsByCompany(ctx, deal.CompanyID)
			if err != nil || len(contacts) == 0 {
				continue
			}

			primaryContact, hasValidEmail := db.SelectOutreachContact(contacts)
			if !hasValidEmail {
				continue
			}

			interactions, _ := m.db.ListInteractionsByContact(ctx, primaryContact.ID)
			hasOutbound := false
			for _, i := range interactions {
				if i.Direction == "Outbound" {
					hasOutbound = true
					break
				}
			}

			if !hasOutbound && strings.TrimSpace(primaryContact.Email) != "" {
				slog.Info(fmt.Sprintf("Comm Manager: Initiating autonomous outreach for deal %d to %s", deal.ID, primaryContact.Email))
				if _, err := m.ProcessInbound(ctx, primaryContact, "START_OUTREACH"); err != nil {
					slog.Info(fmt.Sprintf("Comm Manager Error: Failed to initiate outreach for deal %d: %v", deal.ID, err))
					continue
				}
				// In HITL mode the artifact is a Draft, not a send — the
				// deal must not claim OutreachSent until a human approves.
				if AutonomousSendEnabled() {
					if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil {
						slog.Info(fmt.Sprintf("Comm Manager Error: Failed to update deal state to OutreachSent for deal %d: %v", deal.ID, err))
					}
				}
			}
		}
	}

	// Retry OutreachSent deals that never got a reply (spaced by >=24h so we never
	// double-send alongside CadenceAwareManager within the same day)
	retryDeals, err := m.db.ListDealsByState(ctx, db.StateOutreachSent)
	if err != nil {
		slog.Info(fmt.Sprintf("Comm Manager: Error listing retry deals: %v", err))
	} else {
		for _, deal := range retryDeals {
			contacts, err := m.db.ListContactsByCompany(ctx, deal.CompanyID)
			if err != nil || len(contacts) == 0 {
				continue
			}

			primaryContact, hasValidEmail := db.SelectOutreachContact(contacts)
			if !hasValidEmail {
				continue
			}

			interactions, _ := m.db.ListInteractionsByContact(ctx, primaryContact.ID)
			hasReply := false
			outboundCount := 0
			hasCadenceStep := false
			var lastOutbound time.Time
			for _, i := range interactions {
				if i.Direction == "Inbound" && i.RawText != "START_OUTREACH" {
					hasReply = true
					break
				}
				if i.TemplateID != "" {
					hasCadenceStep = true
				}
				if i.Direction == "Outbound" {
					outboundCount++
					if i.CreatedAt.After(lastOutbound) {
						lastOutbound = i.CreatedAt
					}
				}
			}

			// Defer follow-up touches to CadenceAwareManager when templates are active;
			// only retry here if zero outbound touches ever succeeded or >=24h elapsed without template tracking.
			shouldRetry := false
			if !hasReply && !hasCadenceStep && strings.TrimSpace(primaryContact.Email) != "" {
				if outboundCount == 0 {
					shouldRetry = true
				} else if outboundCount < 10 && time.Since(lastOutbound) > 24*time.Hour {
					shouldRetry = true
				}
			}

			if shouldRetry {
				slog.Info(fmt.Sprintf("Comm Manager: Retrying outreach for deal %d to %s (attempt %d)", deal.ID, primaryContact.Email, outboundCount+1))
				if _, err := m.ProcessInbound(ctx, primaryContact, "START_OUTREACH"); err != nil {
					slog.Info(fmt.Sprintf("Comm Manager Error: Retry failed for deal %d: %v", deal.ID, err))
				}
				// Don't change state — keep in OutreachSent
			}
		}
	}
}

// DefaultChannelForContact returns the channel to use when communicating with a contact.
// If the contact has a preferred channel set, that is used. Otherwise defaults to "email".
func DefaultChannelForContact(contact db.Contact) string {
	if contact.PreferredChannel != "" {
		return contact.PreferredChannel
	}
	return string(db.ChannelEmail)
}

// ProcessInbound handles a new inbound message from a contact (or internal
// proactive trigger when text == "START_OUTREACH").
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	channel := DefaultChannelForContact(contact)
	isSyntheticOutreach := text == "START_OUTREACH"

	if isSyntheticOutreach && channel == string(db.ChannelEmail) && db.IsAutomatedCommitAuthor(contact.Name, contact.Email) {
		return "", fmt.Errorf("skipping automated or unroutable contact %q <%s>", contact.Name, contact.Email)
	}

	// 1. Persist genuine inbound interactions only (never pollute DB with fake
	// Inbound "START_OUTREACH" triggers).
	if !isSyntheticOutreach {
		inbound := db.Interaction{
			ContactID: contact.ID,
			Channel:   channel,
			Direction: "Inbound",
			RawText:   text,
		}
		if err := m.db.CreateInteraction(ctx, &inbound); err != nil {
			return "", err
		}
	}

	freshText := text
	if !isSyntheticOutreach {
		freshText = ExtractFreshInboundText(text)
	}

	// 2b. Check for email bounce (delivery failure notifications) BEFORE classification
	// so mailer-daemon / DSN payloads never advance a deal to Engaged or trigger replies.
	if !isSyntheticOutreach && isBounce(text) {
		slog.Info(fmt.Sprintf("Comm Manager: Email bounce detected for %s — marking contact as invalid", contact.Email))
		deal, dealErr := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if dealErr == nil {
			_ = m.db.UpdateDealState(ctx, deal.ID, db.StateClosedLost)
		}
		_ = m.db.UpdateContactEmail(ctx, contact.ID, "")
		return "", nil
	}

	// 2c. Check for unsubscribe/opt-out signals on the sender's fresh message text.
	if !isSyntheticOutreach && isUnsubscribe(freshText) {
		slog.Info(fmt.Sprintf("Comm Manager: Unsubscribe detected from %s — closing deal %d", contact.Email, contact.CompanyID))
		deal, dealErr := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if dealErr == nil {
			_ = m.db.UpdateDealState(ctx, deal.ID, db.StateClosedLost)
		}
		return "You have been unsubscribed from further communications. No additional emails will be sent.", nil
	}

	// 2d. Respect polite declines / non-commercial hobby-project replies immediately:
	// close the deal so CadenceAwareManager stops follow-ups without sending pushy rebuttals.
	if !isSyntheticOutreach && isPoliteDecline(freshText) {
		slog.Info(fmt.Sprintf("Comm Manager: Polite decline detected from %s — closing deal for company %d without follow-up", contact.Email, contact.CompanyID))
		deal, dealErr := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if dealErr == nil {
			_ = m.db.UpdateDealState(ctx, deal.ID, db.StateClosedLost)
		}
		return "", nil
	}

	// 2. Classify intent (gracefully fall back to heuristic MockIntentClassifier if LLM is unavailable)
	intent := IntentGeneral
	if !isSyntheticOutreach {
		var err error
		intent, err = m.classifier.Classify(ctx, freshText)
		if err != nil {
			slog.Warn("Comm Manager: Primary intent classifier failed, falling back to heuristic classifier", "error", err)
			intent, _ = (&MockIntentClassifier{}).Classify(ctx, freshText)
		}
	}

	// 2a. Decide next action using strategy engine
	company, err := m.db.GetCompanyByID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company for strategy: %w", err)
	}

	interactions, err := m.db.ListInteractionsByContact(ctx, contact.ID)
	if err != nil {
		slog.Info(fmt.Sprintf("Warning: failed to list interactions for strategy: %v", err))
	}

	deal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
	if err != nil {
		return "", fmt.Errorf("failed to get deal for strategy: %w", err)
	}

	if !isSyntheticOutreach && (deal.CurrentState == db.StateResearched || deal.CurrentState == db.StateOutreachSent) {
		if err := m.db.UpdateDealState(ctx, deal.ID, db.StateEngaged); err != nil {
			slog.Info(fmt.Sprintf("Comm Manager: Failed to update deal state to Engaged: %v", err))
		} else {
			deal.CurrentState = db.StateEngaged
		}
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
		slog.Info(fmt.Sprintf("UI: Deal %d escalated for human review.", salesCtx.Deal.ID))
		return "I've flagged this for our technical lead to review. We will get back to you shortly.", nil
	}

	// 3. Generate response
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil {
		return "", err
	}

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
					}
				}
			}

			if m.processor != nil {
				slog.Info(fmt.Sprintf("Comm Manager: Triggering order processor for won deal %d", updatedDeal.ID))
				if err := m.processor.ProcessOrder(ctx, *updatedDeal); err != nil {
					slog.Info(fmt.Sprintf("Comm Manager Error: Order processing failed: %v", err))
				}
			}
		}
	}

	signoff := "HyperNexus Team"
	if !IsCorporate(contact.Email, company.Domain) {
		signoff = "TormentNexus Team"
	}
	replyText = SanitizeOutboundText(replyText, signoff)

	// 4. Send the communication first when email sender is configured so failed
	// deliveries do not record phantom Outbound interactions. In HITL
	// consensus mode (default) nothing is sent: the reply is persisted
	// as a Draft below for the approval queue.
	drafted := false
	if !AutonomousSendEnabled() && channel == "email" && contact.Email != "" {
		drafted = true
		slog.Info(fmt.Sprintf("Comm Manager: HITL draft queued for %s — awaiting human approval (no autonomous send)", contact.Email))
	} else if m.sender != nil && contact.Email != "" && channel == "email" {
		subject := fmt.Sprintf("Re: %s — HyperNexus", company.Name)
		if isSyntheticOutreach {
			subject = fmt.Sprintf("HyperNexus for %s — Quick Question", company.Name)
		}
		subject = SanitizeOutboundText(subject, signoff)

		emailMsg := EmailMessage{
			To:      contact.Email,
			Subject: subject,
			Body:    replyText,
		}

		if err := m.sender.Send(ctx, emailMsg); err != nil {
			slog.Info(fmt.Sprintf("Comm Manager: Email send failed to %s: %v", contact.Email, err))
			return "", fmt.Errorf("email send failed: %w", err)
		}
		slog.Info(fmt.Sprintf("Comm Manager: Email sent to %s (%s)", contact.Email, subject))
	} else if channel != "email" {
		slog.Info(fmt.Sprintf("Comm Manager: Channel %q requires future implementation for %s — reply logged", channel, contact.Email))
	} else if m.sender == nil {
		slog.Info(fmt.Sprintf("Comm Manager: No email sender configured — reply logged but not sent to %s", contact.Email))
	}

	// 5. Persist outbound interaction (Direction Draft in HITL mode)
	summary := fmt.Sprintf("Reply to intent: %s", intent)
	if isSyntheticOutreach {
		summary = "Initial outreach via email"
	}
	direction := "Outbound"
	if drafted {
		direction = "Draft"
		summary = "DRAFT awaiting approval — " + summary
	}
	outbound := db.Interaction{
		ContactID:  contact.ID,
		Channel:    channel,
		Direction:  direction,
		RawText:    replyText,
		Summary:    summary,
		ResponseID: responseID,
	}
	if err := m.db.CreateInteraction(ctx, &outbound); err != nil {
		return "", err
	}

	return replyText, nil
}

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
}

// isUnsubscribe detects opt-out, unsubscribe, or removal requests in inbound text.
// Returns true if the text indicates the recipient wants no further contact.
func isUnsubscribe(text string) bool {
	lower := strings.ToLower(text)
	unsubPhrases := []string{
		"unsubscribe", "opt out", "opt-out", "remove me",
		"take me off", "stop emailing", "do not contact",
		"don't contact", "please remove", "stop sending",
		"no more emails", "leave me alone", "mark as spam",
		"report spam", "stop spamming",
	}
	for _, phrase := range unsubPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// ExtractFreshInboundText strips an optional leading "Subject: ...\n\n" envelope
// wrapper (when body text is non-empty) and removes quoted reply history ("On ... wrote:",
// "> ..." lines, "-----Original Message-----") so opt-out, polite-decline, and intent
// classifiers inspect only what the prospect actually wrote.
func ExtractFreshInboundText(text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "Subject:") {
		if idx := strings.Index(trimmed, "\n\n"); idx != -1 {
			bodyPart := strings.TrimSpace(trimmed[idx+2:])
			if bodyPart != "" {
				trimmed = bodyPart
			}
		}
	}
	stripped := StripQuotedReply(trimmed)
	if stripped != "" {
		return stripped
	}
	return trimmed
}

// StripQuotedReply removes standard email reply quote blocks from plain-text bodies.
func StripQuotedReply(body string) string {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	var kept []string
	for i, line := range lines {
		trimLine := strings.TrimSpace(line)
		lowerLine := strings.ToLower(trimLine)

		if strings.HasPrefix(trimLine, "-----Original Message-----") ||
			strings.HasPrefix(trimLine, "________________________________") ||
			(strings.HasPrefix(lowerLine, "from:") && i+1 < len(lines) && strings.HasPrefix(strings.ToLower(strings.TrimSpace(lines[i+1])), "sent:")) {
			break
		}

		// Match single-line or two-line wrapped "On <date>, <sender> wrote:" attribution headers.
		if strings.HasPrefix(lowerLine, "on ") && strings.HasSuffix(lowerLine, "wrote:") {
			break
		}
		if strings.HasPrefix(lowerLine, "on ") && i+1 < len(lines) {
			nextLower := strings.ToLower(strings.TrimSpace(lines[i+1]))
			if strings.HasSuffix(nextLower, "wrote:") {
				break
			}
		}

		if strings.HasPrefix(trimLine, ">") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// isPoliteDecline detects clear non-interest or non-commercial/hobby-project replies
// so the deal transitions cleanly to Closed_Lost without pushy follow-up cadences.
func isPoliteDecline(text string) bool {
	lower := strings.ToLower(text)
	declinePhrases := []string{
		"not interested",
		"no thank you",
		"no thanks",
		"not currently exploring",
		"not exploring new",
		"pass for now",
		"hard pass",
		"not in the market",
		"just a personal project",
		"project for fun",
		"not anywhere in the production",
		"not in production",
		"we're all set",
		"we are all set",
		"not looking for",
		"close my file",
	}
	for _, phrase := range declinePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// isBounce detects email delivery failure notifications.
// Matches common bounce formats: mailer-daemon, delivery failure, etc.
func isBounce(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "delivery status notification (delay)") || strings.Contains(lower, "will retry for") {
		return false
	}
	bouncePhrases := []string{
		"mailer-daemon", "mail delivery", "delivery status notification",
		"undelivered mail", "returned mail", "delivery failure",
		"could not be delivered", "couldn't be delivered", "wasn't delivered",
		"address not found", "message blocked", "address rejected",
		"no such user", "mailbox full", "inbox is full", "user unknown",
		"recipient rejected", "550 5.1.1", "550 5.2.1", "550 5.4.1", "552 5.2.2",
	}
	for _, phrase := range bouncePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
