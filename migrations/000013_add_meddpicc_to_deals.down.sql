ALTER TABLE deals
DROP COLUMN IF EXISTS meddpicc_metrics,
DROP COLUMN IF EXISTS meddpicc_eco_buyer,
DROP COLUMN IF EXISTS meddpicc_decision_criteria,
DROP COLUMN IF EXISTS meddpicc_decision_process,
DROP COLUMN IF EXISTS meddpicc_paper_process,
DROP COLUMN IF EXISTS meddpicc_identify_pain,
DROP COLUMN IF EXISTS meddpicc_champion,
DROP COLUMN IF EXISTS meddpicc_competition;
