package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"

	"github.com/robertpelloni/marketing_agent/internal/db"
)

// InvoiceStatus represents the current state of a deal's billing.
type InvoiceStatus string

const (
	InvoicePending InvoiceStatus = "Pending"
	InvoiceSent    InvoiceStatus = "Sent"
	InvoicePaid    InvoiceStatus = "Paid"
	InvoiceFailed  InvoiceStatus = "Failed"
)

// Tier represents a pricing tier.
type Tier string

const (
	TierCommunity    Tier = "community"
	TierProfessional Tier = "professional"
	TierEnterprise   Tier = "enterprise"
)

// SubscriptionInfo holds the state of a subscription.
type SubscriptionInfo struct {
	ID                int64      `json:"id"`
	CompanyID         int64      `json:"company_id"`
	StripeSubID       string     `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID  string     `json:"stripe_customer_id,omitempty"`
	Tier              Tier       `json:"tier"`
	State             string     `json:"state"`
	CurrentRate       float64    `json:"current_rate"`
	GrandfatheredRate *float64   `json:"grandfathered_rate,omitempty"`
	Seats             int        `json:"seats"`
	TrialEnd          *time.Time `json:"trial_end,omitempty"`
	PeriodEnd         *time.Time `json:"period_end,omitempty"`
	CanceledAt        *time.Time `json:"canceled_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// BillingClient defines the interface for financial operations.
type BillingClient interface {
	CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error)
	GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error)
	CreateCheckoutSession(ctx context.Context, companyID int64, tier Tier, successURL, cancelURL string) (string, error)
	GetSubscription(ctx context.Context, subID string) (*SubscriptionInfo, error)
	CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error
	UpdateSubscriptionSeats(ctx context.Context, subID string, seats int) error
	HandleWebhook(ctx context.Context, payload []byte, sigHeader string) (string, error)
}

// StripeBillingClient implements BillingClient using the Stripe API.
type StripeBillingClient struct {
	APIKey              string
	WebhookSecret       string
	PriceIDCommunity    string
	PriceIDProfessional string
	PriceIDEnterprise   string
	db                  SubscriptionStore
}

// SubscriptionStore is the database interface for subscription storage.
type SubscriptionStore interface {
	CreateSubscription(ctx context.Context, companyID int64, tier Tier, stripeSubID, stripeCustomerID string, rate float64, seats int, trialEnd *time.Time) (*SubscriptionInfo, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*SubscriptionInfo, error)
	GetSubscriptionByCompanyID(ctx context.Context, companyID int64) (*SubscriptionInfo, error)
	UpdateSubscriptionState(ctx context.Context, stripeSubID, state string) error
	UpdateSubscriptionPeriod(ctx context.Context, stripeSubID string, periodStart, periodEnd time.Time) error
	CancelSubscription(ctx context.Context, stripeSubID string, at time.Time) error
	SetGrandfatheredRate(ctx context.Context, stripeSubID string, rate float64) error
	RecordPriceChange(ctx context.Context, subID int64, prevRate, newRate float64) error
}

// NewStripeBillingClient creates a new Stripe-based billing client.
func NewStripeBillingClient(apiKey, webhookSecret string, priceIDs map[Tier]string, store SubscriptionStore) *StripeBillingClient {
	return &StripeBillingClient{
		APIKey:              apiKey,
		WebhookSecret:       webhookSecret,
		PriceIDCommunity:    priceIDs[TierCommunity],
		PriceIDProfessional: priceIDs[TierProfessional],
		PriceIDEnterprise:   priceIDs[TierEnterprise],
		db:                  store,
	}
}

func (s *StripeBillingClient) stripe() { stripe.Key = s.APIKey }

// --- Invoice Operations ---

func (s *StripeBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	s.stripe()
	params := &stripe.InvoiceParams{
		Customer:         stripe.String(company.Domain),
		AutoAdvance:      stripe.Bool(true),
		CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodSendInvoice)),
		DaysUntilDue:     stripe.Int64(30),
	}
	inv, err := invoice.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe invoice creation failed: %w", err)
	}
	return inv.ID, nil
}

func (s *StripeBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	s.stripe()
	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return InvoiceFailed, fmt.Errorf("stripe invoice retrieval failed: %w", err)
	}
	switch inv.Status {
	case stripe.InvoiceStatusPaid:
		return InvoicePaid, nil
	case stripe.InvoiceStatusOpen:
		return InvoiceSent, nil
	case stripe.InvoiceStatusVoid, stripe.InvoiceStatusUncollectible:
		return InvoiceFailed, nil
	default:
		return InvoicePending, nil
	}
}

// --- Subscription Operations ---

func (s *StripeBillingClient) priceIDForTier(tier Tier) string {
	switch tier {
	case TierCommunity:
		return s.PriceIDCommunity
	case TierProfessional:
		return s.PriceIDProfessional
	case TierEnterprise:
		return s.PriceIDEnterprise
	default:
		return ""
	}
}

func (s *StripeBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier Tier, successURL, cancelURL string) (string, error) {
	s.stripe()

	priceID := s.priceIDForTier(tier)
	if priceID == "" {
		return "", fmt.Errorf("unknown tier: %s", tier)
	}

	// Determine if the price is recurring (subscription) or one-time (payment)
	// by fetching the price from Stripe.
	price, err := getPrice(ctx, priceID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch price %s: %w", priceID, err)
	}

	mode := stripe.CheckoutSessionModePayment
	if price.Recurring != nil {
		mode = stripe.CheckoutSessionModeSubscription
	}

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(mode)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"company_id": fmt.Sprintf("%d", companyID),
			"tier":       string(tier),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout session creation failed: %w", err)
	}
	return sess.URL, nil
}

// getPrice fetches a Stripe price object to check if it's recurring or one-time.
func getPrice(_ context.Context, priceID string) (*stripe.Price, error) {
	return price.Get(priceID, nil)
}

func (s *StripeBillingClient) GetSubscription(ctx context.Context, subID string) (*SubscriptionInfo, error) {
	s.stripe()
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe subscription retrieval failed: %w", err)
	}

	rate := 0.0
	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		rate = float64(sub.Items.Data[0].Price.UnitAmount) / 100
	}

	info := &SubscriptionInfo{
		StripeSubID:      sub.ID,
		StripeCustomerID: sub.Customer.ID,
		State:            string(sub.Status),
		CurrentRate:      rate,
		Seats:            int(sub.Items.Data[0].Quantity),
	}

	if sub.TrialEnd > 0 {
		t := time.Unix(sub.TrialEnd, 0)
		info.TrialEnd = &t
	}
	if sub.CurrentPeriodEnd > 0 {
		t := time.Unix(sub.CurrentPeriodEnd, 0)
		info.PeriodEnd = &t
	}
	if sub.CanceledAt > 0 {
		t := time.Unix(sub.CanceledAt, 0)
		info.CanceledAt = &t
	}

	// Enrich with local DB state for grandfathering info
	dbSub, err := s.db.GetSubscriptionByStripeID(ctx, sub.ID)
	if err == nil && dbSub != nil {
		info.GrandfatheredRate = dbSub.GrandfatheredRate
		info.CompanyID = dbSub.CompanyID
		info.Tier = dbSub.Tier
		info.ID = dbSub.ID
	}

	return info, nil
}

func (s *StripeBillingClient) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error {
	s.stripe()
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(atPeriodEnd),
	}
	_, err := subscription.Update(subID, params)
	if err != nil {
		return fmt.Errorf("stripe subscription cancellation failed: %w", err)
	}
	return nil
}

func (s *StripeBillingClient) UpdateSubscriptionSeats(ctx context.Context, subID string, seats int) error {
	s.stripe()
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if len(sub.Items.Data) == 0 {
		return fmt.Errorf("subscription has no items")
	}

	itemID := sub.Items.Data[0].ID
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(itemID),
				Quantity: stripe.Int64(int64(seats)),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}
	_, err = subscription.Update(subID, params)
	if err != nil {
		return fmt.Errorf("stripe subscription seat update failed: %w", err)
	}
	return nil
}

// --- Webhook Handling ---

func (s *StripeBillingClient) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) (string, error) {
	s.stripe()

	event, err := webhook.ConstructEvent(payload, sigHeader, s.WebhookSecret)
	if err != nil {
		return "", fmt.Errorf("stripe webhook signature verification failed: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event)
	case "invoice.paid":
		return s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	default:
		return fmt.Sprintf("unhandled event type: %s", event.Type), nil
	}
}

func (s *StripeBillingClient) handleCheckoutCompleted(ctx context.Context, event stripe.Event) (string, error) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		return "", fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	companyIDStr, ok := sess.Metadata["company_id"]
	if !ok {
		return "no company_id in metadata", nil
	}
	tierStr, ok := sess.Metadata["tier"]
	if !ok {
		tierStr = "professional"
	}

	var companyID int64
	fmt.Sscanf(companyIDStr, "%d", &companyID)

	// Create Stripe customer if not exists
	custParams := &stripe.CustomerParams{
		Email:    stripe.String(sess.CustomerDetails.Email),
		Name:     stripe.String(sess.CustomerDetails.Name),
		Metadata: map[string]string{"company_id": companyIDStr},
	}
	cust, err := customer.New(custParams)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Get the subscription from the session
	if sess.Subscription == nil {
		return "no subscription in session", nil
	}

	sub, err := subscription.Get(sess.Subscription.ID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get subscription: %w", err)
	}

	rate := 0.0
	seats := 1
	if len(sub.Items.Data) > 0 {
		rate = float64(sub.Items.Data[0].Price.UnitAmount) / 100
		seats = int(sub.Items.Data[0].Quantity)
	}

	var trialEnd *time.Time
	if sub.TrialEnd > 0 {
		t := time.Unix(sub.TrialEnd, 0)
		trialEnd = &t
	}

	_, err = s.db.CreateSubscription(ctx, companyID, Tier(tierStr), sub.ID, cust.ID, rate, seats, trialEnd)
	if err != nil {
		return "", fmt.Errorf("failed to save subscription: %w", err)
	}

	slog.Info("subscription created via checkout",
		"company_id", companyID,
		"tier", tierStr,
		"stripe_sub_id", sub.ID,
		"rate", rate,
	)

	return fmt.Sprintf("subscription created: %s", sub.ID), nil
}

func (s *StripeBillingClient) handleInvoicePaid(ctx context.Context, event stripe.Event) (string, error) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return "", fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	if inv.Subscription != nil {
		if err := s.db.UpdateSubscriptionState(ctx, inv.Subscription.ID, "active"); err != nil {
			slog.Warn("failed to update subscription state", "error", err)
		}
		if inv.PeriodStart > 0 && inv.PeriodEnd > 0 {
			start := time.Unix(inv.PeriodStart, 0)
			end := time.Unix(inv.PeriodEnd, 0)
			if err := s.db.UpdateSubscriptionPeriod(ctx, inv.Subscription.ID, start, end); err != nil {
				slog.Warn("failed to update subscription period", "error", err)
			}
		}
	}

	return fmt.Sprintf("invoice paid: %s", inv.ID), nil
}

func (s *StripeBillingClient) handlePaymentFailed(ctx context.Context, event stripe.Event) (string, error) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return "", fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	if inv.Subscription != nil {
		if err := s.db.UpdateSubscriptionState(ctx, inv.Subscription.ID, "past_due"); err != nil {
			slog.Warn("failed to update subscription state", "error", err)
		}
	}

	return fmt.Sprintf("payment failed for invoice: %s", inv.ID), nil
}

func (s *StripeBillingClient) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) (string, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return "", fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	if err := s.db.UpdateSubscriptionState(ctx, sub.ID, string(sub.Status)); err != nil {
		slog.Warn("failed to update subscription state", "error", err)
	}

	if sub.CanceledAt > 0 {
		t := time.Unix(sub.CanceledAt, 0)
		if err := s.db.CancelSubscription(ctx, sub.ID, t); err != nil {
			slog.Warn("failed to record cancellation", "error", err)
		}
	}

	// Check for price change → apply grandfathering
	dbSub, _ := s.db.GetSubscriptionByStripeID(ctx, sub.ID)
	if dbSub != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		newRate := float64(sub.Items.Data[0].Price.UnitAmount) / 100
		if newRate > dbSub.CurrentRate && dbSub.GrandfatheredRate == nil {
			// Rate increased, freeze the old rate as grandfathered
			if err := s.db.SetGrandfatheredRate(ctx, sub.ID, dbSub.CurrentRate); err != nil {
				slog.Warn("failed to set grandfathered rate", "error", err)
			}
			if err := s.db.RecordPriceChange(ctx, dbSub.ID, dbSub.CurrentRate, newRate); err != nil {
				slog.Warn("failed to record price change", "error", err)
			}
			slog.Info("grandfathered rate applied",
				"subscription", sub.ID,
				"old_rate", dbSub.CurrentRate,
				"new_rate", newRate,
			)
		}
	}

	return fmt.Sprintf("subscription updated: %s", sub.ID), nil
}

func (s *StripeBillingClient) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) (string, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return "", fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	if err := s.db.UpdateSubscriptionState(ctx, sub.ID, "canceled"); err != nil {
		slog.Warn("failed to update subscription state", "error", err)
	}

	return fmt.Sprintf("subscription deleted: %s", sub.ID), nil
}

// MockBillingClient implements BillingClient for testing and offline development.
type MockBillingClient struct {
	db SubscriptionStore
}

func NewMockBillingClient(store SubscriptionStore) *MockBillingClient {
	return &MockBillingClient{db: store}
}

func (m *MockBillingClient) CreateInvoice(ctx context.Context, deal db.Deal, company db.Company) (string, error) {
	return "mock_invoice_id", nil
}

func (m *MockBillingClient) GetInvoiceStatus(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	return InvoicePaid, nil
}

func (m *MockBillingClient) CreateCheckoutSession(ctx context.Context, companyID int64, tier Tier, successURL, cancelURL string) (string, error) {
	return "http://localhost:8087/checkout/success", nil
}

func (m *MockBillingClient) GetSubscription(ctx context.Context, subID string) (*SubscriptionInfo, error) {
	return &SubscriptionInfo{
		StripeSubID: subID,
		State:       "active",
		Tier:        TierProfessional,
		CurrentRate: 49.00,
	}, nil
}

func (m *MockBillingClient) CancelSubscription(ctx context.Context, subID string, atPeriodEnd bool) error {
	return nil
}

func (m *MockBillingClient) UpdateSubscriptionSeats(ctx context.Context, subID string, seats int) error {
	return nil
}

func (m *MockBillingClient) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) (string, error) {
	return "mock webhook processed", nil
}

