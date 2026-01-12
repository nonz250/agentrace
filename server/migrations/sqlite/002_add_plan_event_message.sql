-- Add message column to plan_document_events table
ALTER TABLE plan_document_events ADD COLUMN message TEXT NOT NULL DEFAULT '';
