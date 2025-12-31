-- Plan Documents table
CREATE TABLE IF NOT EXISTS plan_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    git_remote_url VARCHAR(1024) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Plan Document Events table
CREATE TABLE IF NOT EXISTS plan_document_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_document_id UUID NOT NULL REFERENCES plan_documents(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    patch TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_plan_documents_git_remote ON plan_documents(git_remote_url);
CREATE INDEX IF NOT EXISTS idx_plan_documents_updated ON plan_documents(updated_at);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_doc ON plan_document_events(plan_document_id);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_session ON plan_document_events(session_id);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_user ON plan_document_events(user_id);
