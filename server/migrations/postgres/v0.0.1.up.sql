-- サブエージェント関連フィールドの追加
ALTER TABLE sessions ADD COLUMN parent_session_id UUID REFERENCES sessions(id) ON DELETE CASCADE;
ALTER TABLE sessions ADD COLUMN agent_id VARCHAR(50);
ALTER TABLE sessions ADD COLUMN is_sidechain BOOLEAN NOT NULL DEFAULT FALSE;

-- 親セッションID検索用インデックス
CREATE INDEX IF NOT EXISTS idx_sessions_parent_session_id ON sessions(parent_session_id);

-- サブエージェント除外した一覧取得用（部分インデックス）
CREATE INDEX IF NOT EXISTS idx_sessions_main_only ON sessions(updated_at DESC) WHERE is_sidechain = FALSE;
CREATE INDEX IF NOT EXISTS idx_sessions_main_only_created ON sessions(created_at DESC) WHERE is_sidechain = FALSE;

-- プロジェクト別一覧でもサブエージェント除外
CREATE INDEX IF NOT EXISTS idx_sessions_project_main_only ON sessions(project_id, updated_at DESC) WHERE is_sidechain = FALSE;
