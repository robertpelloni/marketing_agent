-- Create subscriptions table with grandfathering support

CREATE TYPE subscription_state AS ENUM (
    'trialing',
    'active',
    'past_due',
    'canceled',
    'incomplete',
    'incomplete_expired'
);

CREATE TYPE pricing_tier AS ENUM (
    'community',
    'professional',
    'enterprise'
);

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    company_id INTEGER REFERENCES companies(id) ON DELETE CASCADE,
    stripe_subscription_id TEXT UNIQUE,
    stripe_customer_id TEXT,
    tier pricing_tier NOT NULL DEFAULT 'community',
    state subscription_state NOT NULL DEFAULT 'trialing',
    current_rate NUMERIC(15, 2) NOT NULL,
    grandfathered_rate NUMERIC(15, 2),
    grandfathered_from TIMESTAMP WITH TIME ZONE,
    currency TEXT NOT NULL DEFAULT 'usd',
    seats INTEGER NOT NULL DEFAULT 1,
    trial_end TIMESTAMP WITH TIME ZONE,
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Track price changes for grandfathering audit trail
CREATE TABLE subscription_price_history (
    id SERIAL PRIMARY KEY,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE CASCADE,
    previous_rate NUMERIC(15, 2),
    new_rate NUMERIC(15, 2),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Track billing events from Stripe webhooks
CREATE TABLE billing_events (
    id SERIAL PRIMARY KEY,
    stripe_event_id TEXT UNIQUE NOT NULL,
    event_type TEXT NOT NULL,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE SET NULL,
    raw_payload JSONB,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Index for fast lookups
CREATE INDEX idx_subscriptions_company ON subscriptions(company_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_state ON subscriptions(state);
CREATE INDEX idx_billing_events_type ON billing_events(event_type);
