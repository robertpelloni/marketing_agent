ALTER TABLE interactions ADD COLUMN IF NOT EXISTS template_id TEXT;
ALTER TABLE interactions ADD COLUMN IF NOT EXISTS response_id TEXT;
