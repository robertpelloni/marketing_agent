-- Drop triggers
DROP TRIGGER IF EXISTS update_deals_updated_at ON deals;
DROP TRIGGER IF EXISTS update_contacts_updated_at ON contacts;
DROP TRIGGER IF EXISTS update_companies_updated_at ON companies;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS interactions;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS companies;

-- Drop enum type
DROP TYPE IF EXISTS lead_state;
