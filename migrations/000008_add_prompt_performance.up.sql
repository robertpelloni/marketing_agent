CREATE TABLE IF NOT EXISTS prompt_performance (
    id SERIAL PRIMARY KEY,
    deal_id INTEGER NOT NULL REFERENCES deals(id),
    prompt_type VARCHAR(50) NOT NULL,
    context_injected BOOLEAN NOT NULL DEFAULT FALSE,
    response_quality_score FLOAT DEFAULT 0.0,
    successful_outcome BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_prompt_performance_deal_id ON prompt_performance(deal_id);
