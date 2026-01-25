-- v0.1.0: Add plan comment thread tables
-- Create plan_comment_threads table (anchored to specific text in plan)
CREATE TABLE IF NOT EXISTS plan_comment_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_document_id UUID NOT NULL REFERENCES plan_documents(id) ON DELETE CASCADE,

    target_text TEXT NOT NULL,
    context_before TEXT DEFAULT '',
    context_after TEXT DEFAULT '',
    original_body_hash VARCHAR(64) NOT NULL,

    status VARCHAR(32) NOT NULL DEFAULT 'active',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plan_comment_threads_plan_document_id ON plan_comment_threads(plan_document_id);
CREATE INDEX IF NOT EXISTS idx_plan_comment_threads_status ON plan_comment_threads(status);

-- Create plan_comment_messages table (messages within a thread)
CREATE TABLE IF NOT EXISTS plan_comment_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES plan_comment_threads(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),

    content TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plan_comment_messages_thread_id ON plan_comment_messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_plan_comment_messages_user_id ON plan_comment_messages(user_id);
