-- Remove technical_dossier column from deals table
ALTER TABLE deals DROP COLUMN IF EXISTS technical_dossier;
