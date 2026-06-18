DROP TABLE IF EXISTS template_metrics;
ALTER TABLE interactions DROP COLUMN IF EXISTS response_id;
ALTER TABLE interactions DROP COLUMN IF EXISTS template_id;
DROP TABLE IF EXISTS templates;
ALTER TABLE deals DROP COLUMN IF EXISTS cadence_step;
