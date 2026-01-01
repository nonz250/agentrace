-- Add status column to plan_documents
ALTER TABLE plan_documents ADD COLUMN status TEXT NOT NULL DEFAULT 'planning';

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_plan_documents_status ON plan_documents(status);
