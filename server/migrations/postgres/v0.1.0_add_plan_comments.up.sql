-- v0.1.0: Add plan_comments table for inline comments on plan documents
CREATE TABLE IF NOT EXISTS plan_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_document_id UUID NOT NULL REFERENCES plan_documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),

    target_text TEXT NOT NULL,
    context_before TEXT DEFAULT '',
    context_after TEXT DEFAULT '',
    original_body_hash VARCHAR(64) NOT NULL,

    content TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plan_comments_plan_document_id ON plan_comments(plan_document_id);
CREATE INDEX IF NOT EXISTS idx_plan_comments_user_id ON plan_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_plan_comments_status ON plan_comments(status);
