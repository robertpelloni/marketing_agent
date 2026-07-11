-- Rollback subscriptions tables (soft)
ALTER TABLE subscriptions DISABLE TRIGGER update_subscriptions_updated_at;
