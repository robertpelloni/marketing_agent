DROP TABLE IF EXISTS templates;
ALTER TABLE interactions DROP COLUMN IF EXISTS template_id;
ALTER TABLE interactions DROP COLUMN IF EXISTS response_id;

