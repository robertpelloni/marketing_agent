ALTER TABLE deals
ADD COLUMN IF NOT EXISTS meddpicc_metrics TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_eco_buyer TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_decision_criteria TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_decision_process TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_paper_process TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_identify_pain TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_champion TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS meddpicc_competition TEXT DEFAULT '';
