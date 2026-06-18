ALTER TABLE deals ADD COLUMN IF NOT EXISTS cadence_step INTEGER DEFAULT 0;

CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    channel TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE interactions ADD COLUMN IF NOT EXISTS template_id TEXT;
ALTER TABLE interactions ADD COLUMN IF NOT EXISTS response_id TEXT;

CREATE TABLE IF NOT EXISTS template_metrics (
    template_id TEXT PRIMARY KEY,
    impressions INTEGER DEFAULT 0,
    successes INTEGER DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
