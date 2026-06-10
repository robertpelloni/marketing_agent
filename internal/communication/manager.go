package communication
import (
	"context"

	"log/slog"
	"time"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/metrics"
)
type Intent string
const (
	IntentTechnical Intent = "Technical"; IntentPricing Intent = "Pricing"; IntentObjection Intent = "Objection"
	IntentMeetingRequest Intent = "MeetingRequest"; IntentFollowUp Intent = "FollowUp"; IntentSpam Intent = "Spam"; IntentUnknown Intent = "Unknown"
)
type IntentClassifier interface { Classify(ctx context.Context, text string) (Intent, error) }
type ResponseGenerator interface { Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) }
type OrderProcessor interface { ProcessOrder(ctx context.Context, deal db.Deal) error }
type Manager struct {
	db *db.DB; classifier IntentClassifier; responder ResponseGenerator; strategy SalesStrategy; processor OrderProcessor; crmClient crm.CRMClient
}
func NewManager(database *db.DB, classifier IntentClassifier, responder ResponseGenerator, strategy SalesStrategy, processor OrderProcessor, crmClient crm.CRMClient) *Manager {
	return &Manager{db: database, classifier: classifier, responder: responder, strategy: strategy, processor: processor, crmClient: crmClient}
}
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval); defer ticker.Stop()
	slog.Info("Communication Manager Background poller started", "interval", interval)
	for {
		select {
		case <-ctx.Done(): slog.Info("Communication Manager Background poller stopping"); return
		case <-ticker.C: m.pollAndProcess(ctx)
		}
	}
}
func (m *Manager) pollAndProcess(ctx context.Context) {
	deals, err := m.db.ListDealsByState(ctx, db.StateResearched)
	if err != nil { slog.Error("Comm Manager Error polling deals", "error", err); return }
	for _, deal := range deals {
		contacts, err := m.db.ListContactsByCompany(ctx, deal.CompanyID)
		if err != nil || len(contacts) == 0 { continue }
		interactions, _ := m.db.ListInteractionsByContact(ctx, contacts[0].ID)
		hasOutbound := false
		for _, i := range interactions { if i.Direction == "Outbound" { hasOutbound = true; break } }
		if !hasOutbound {
			slog.Info("Comm Manager Initiating autonomous outreach", "deal_id", deal.ID, "email", contacts[0].Email)
			if _, err := m.ProcessInbound(ctx, contacts[0], "START_OUTREACH"); err != nil { slog.Error("Comm Manager Error Failed to initiate outreach", "deal_id", deal.ID, "error", err) }
			if err := m.db.UpdateDealState(ctx, deal.ID, db.StateOutreachSent); err != nil { slog.Error("Comm Manager Error Failed to update deal state", "deal_id", deal.ID, "error", err) }
		}
	}
}
func (m *Manager) ProcessInbound(ctx context.Context, contact db.Contact, text string) (string, error) {
	inbound := db.Interaction{ContactID: contact.ID, Channel: "Email", Direction: "Inbound", RawText: text}
	if err := m.db.CreateInteraction(ctx, &inbound); err != nil { return "", err }
	metrics.InteractionsProcessed.WithLabelValues("Inbound", "Email").Inc()
	intent, err := m.classifier.Classify(ctx, text)
	if err != nil { return "", err }
	company, _ := m.db.GetCompanyByID(ctx, contact.CompanyID)
	interactions, _ := m.db.ListInteractionsByContact(ctx, contact.ID)
	deal, _ := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
	salesCtx := SalesContext{Company: *company, Deal: *deal, Contact: contact, Interactions: interactions, LatestIntent: intent}
	action, err := m.strategy.Decide(ctx, salesCtx)
	if err != nil { return "", err }
	if action == ActionEscalate { slog.Info("UI Deal escalated for human review", "deal_id", salesCtx.Deal.ID); return "I've flagged this for review.", nil }
	replyText, err := m.responder.Generate(ctx, salesCtx, action)
	if err != nil { return "", err }
	if action == ActionAdvanceState {
		updatedDeal, err := m.db.GetDealByCompanyID(ctx, contact.CompanyID)
		if err == nil && updatedDeal.CurrentState == db.StateClosedWon {
			metrics.DealsWon.Inc()
			if m.processor != nil { m.processor.ProcessOrder(ctx, *updatedDeal) }
		}
	}
	outbound := db.Interaction{ContactID: contact.ID, Channel: "Email", Direction: "Outbound", RawText: replyText}
	m.db.CreateInteraction(ctx, &outbound)
	metrics.InteractionsProcessed.WithLabelValues("Outbound", "Email").Inc()
	return replyText, nil
}
