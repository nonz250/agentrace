-- サブエージェント関連フィールドの追加
ALTER TABLE sessions ADD COLUMN parent_session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE;
ALTER TABLE sessions ADD COLUMN agent_id TEXT;
ALTER TABLE sessions ADD COLUMN is_sidechain INTEGER NOT NULL DEFAULT 0;

-- 親セッションID検索用インデックス
CREATE INDEX IF NOT EXISTS idx_sessions_parent_session_id ON sessions(parent_session_id);

-- サブエージェント除外した一覧取得用インデックス
CREATE INDEX IF NOT EXISTS idx_sessions_is_sidechain ON sessions(is_sidechain);
