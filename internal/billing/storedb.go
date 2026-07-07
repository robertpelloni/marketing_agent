package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBAdapter wraps a sql.DB connection to implement SubscriptionStore.
// This avoids circular imports between the billing and db packages.
type DBAdapter struct {
	DB *sql.DB
}

type subscriptionRow struct {
	ID                int64
	CompanyID         int64
	StripeSubID       string
	StripeCustomerID  string
	Tier              string
	State             string
	CurrentRate       float64
	GrandfatheredRate *float64
	Seats             int
	TrialEnd          *time.Time
	PeriodStart       *time.Time
	PeriodEnd         *time.Time
	CanceledAt        *time.Time
	CreatedAt         time.Time
}

func (r *subscriptionRow) toInfo() *SubscriptionInfo {
	return &SubscriptionInfo{
		ID:                r.ID,
		CompanyID:         r.CompanyID,
		StripeSubID:       r.StripeSubID,
		StripeCustomerID:  r.StripeCustomerID,
		Tier:              Tier(r.Tier),
		State:             r.State,
		CurrentRate:       r.CurrentRate,
		GrandfatheredRate: r.GrandfatheredRate,
		Seats:             r.Seats,
		TrialEnd:          r.TrialEnd,
		PeriodEnd:         r.PeriodEnd,
		CanceledAt:        r.CanceledAt,
		CreatedAt:         r.CreatedAt,
	}
}

func scanSubscriptionRow(row *sql.Row, r *subscriptionRow) error {
	return row.Scan(
		&r.ID, &r.CompanyID, &r.StripeSubID, &r.StripeCustomerID,
		&r.Tier, &r.State, &r.CurrentRate, &r.GrandfatheredRate,
		&r.Seats, &r.TrialEnd, &r.PeriodStart, &r.PeriodEnd,
		&r.CanceledAt, &r.CreatedAt,
	)
}

const subCols = `id, company_id, stripe_subscription_id, stripe_customer_id, tier, state, current_rate, grandfathered_rate, seats, trial_end, current_period_start, current_period_end, canceled_at, created_at`

func (a *DBAdapter) CreateSubscription(ctx context.Context, companyID int64, tier Tier, stripeSubID, stripeCustomerID string, rate float64, seats int, trialEnd *time.Time) (*SubscriptionInfo, error) {
	var r subscriptionRow
	err := scanSubscriptionRow(
		a.DB.QueryRowContext(ctx,
			`INSERT INTO subscriptions (company_id, stripe_subscription_id, stripe_customer_id, tier, current_rate, seats, trial_end)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING `+subCols,
			companyID, stripeSubID, stripeCustomerID, string(tier), rate, seats, trialEnd),
		&r,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return r.toInfo(), nil
}

func (a *DBAdapter) GetSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*SubscriptionInfo, error) {
	var r subscriptionRow
	err := scanSubscriptionRow(
		a.DB.QueryRowContext(ctx,
			`SELECT `+subCols+` FROM subscriptions WHERE stripe_subscription_id = $1`, stripeSubID),
		&r,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription by stripe ID: %w", err)
	}
	return r.toInfo(), nil
}

func (a *DBAdapter) GetSubscriptionByCompanyID(ctx context.Context, companyID int64) (*SubscriptionInfo, error) {
	var r subscriptionRow
	err := scanSubscriptionRow(
		a.DB.QueryRowContext(ctx,
			`SELECT `+subCols+` FROM subscriptions WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, companyID),
		&r,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription by company ID: %w", err)
	}
	return r.toInfo(), nil
}

func (a *DBAdapter) UpdateSubscriptionState(ctx context.Context, stripeSubID, state string) error {
	_, err := a.DB.ExecContext(ctx,
		`UPDATE subscriptions SET state = $1, updated_at = CURRENT_TIMESTAMP WHERE stripe_subscription_id = $2`,
		state, stripeSubID,
	)
	return err
}

func (a *DBAdapter) UpdateSubscriptionPeriod(ctx context.Context, stripeSubID string, periodStart, periodEnd time.Time) error {
	_, err := a.DB.ExecContext(ctx,
		`UPDATE subscriptions SET current_period_start = $1, current_period_end = $2, updated_at = CURRENT_TIMESTAMP WHERE stripe_subscription_id = $3`,
		periodStart, periodEnd, stripeSubID,
	)
	return err
}

func (a *DBAdapter) CancelSubscription(ctx context.Context, stripeSubID string, at time.Time) error {
	_, err := a.DB.ExecContext(ctx,
		`UPDATE subscriptions SET canceled_at = $1, updated_at = CURRENT_TIMESTAMP WHERE stripe_subscription_id = $2`,
		at, stripeSubID,
	)
	return err
}

func (a *DBAdapter) SetGrandfatheredRate(ctx context.Context, stripeSubID string, rate float64) error {
	_, err := a.DB.ExecContext(ctx,
		`UPDATE subscriptions SET grandfathered_rate = $1, updated_at = CURRENT_TIMESTAMP WHERE stripe_subscription_id = $2`,
		rate, stripeSubID,
	)
	return err
}

func (a *DBAdapter) RecordPriceChange(ctx context.Context, subID int64, prevRate, newRate float64) error {
	_, err := a.DB.ExecContext(ctx,
		`INSERT INTO subscription_price_history (subscription_id, previous_rate, new_rate) VALUES ($1, $2, $3)`,
		subID, prevRate, newRate,
	)
	return err
}
