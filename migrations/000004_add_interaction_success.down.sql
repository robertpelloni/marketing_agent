-- Remove prompt optimization loop columns from interactions
ALTER TABLE interactions DROP COLUMN IF EXISTS success;
ALTER TABLE interactions DROP COLUMN IF EXISTS response_id;
