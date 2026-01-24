-- v0.1.0: Add plan_comments table for inline comments on plan documents
CREATE TABLE IF NOT EXISTS plan_comments (
    id TEXT PRIMARY KEY,
    plan_document_id TEXT NOT NULL REFERENCES plan_documents(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),

    target_text TEXT NOT NULL,
    context_before TEXT DEFAULT '',
    context_after TEXT DEFAULT '',
    original_body_hash TEXT NOT NULL,

    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',

    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_plan_comments_plan_document_id ON plan_comments(plan_document_id);
CREATE INDEX IF NOT EXISTS idx_plan_comments_user_id ON plan_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_plan_comments_status ON plan_comments(status);
