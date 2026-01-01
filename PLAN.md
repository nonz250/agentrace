# Project リソース導入設計書

## 概要

新しいリソース `Project` を導入し、Session と PlanDocument の親リソースとして機能させる。これにより、同じ Git リポジトリで作業している Session や PlanDocument をグループ化できる。

## 現状の課題

- Session と PlanDocument がそれぞれ `git_remote_url` を独立して持っている
- 同じリポジトリに関連するデータの紐付けが困難
- Git URL の表記揺れ（SSH vs HTTPS、`.git` suffix）への対応がない

## 設計

### 1. Project ドメインモデル

```go
type Project struct {
    ID                     string     // UUID
    CanonicalGitRepository string     // 正規化されたHTTP形式のGit URL（空文字 = no project）
    CreatedAt              time.Time
}
```

**特別なプロジェクト: "No Project"**
- `CanonicalGitRepository` が空文字のプロジェクト
- Git リポジトリ情報を持たない Session/PlanDocument が所属
- Web UI では「(no project)」と表示

### 2. Git URL の正規化

CLI から送信される `git_remote_url` を正規化して `CanonicalGitRepository` に変換:

| 入力形式 | 正規化後 |
|----------|----------|
| `git@github.com:user/repo.git` | `https://github.com/user/repo` |
| `https://github.com/user/repo.git` | `https://github.com/user/repo` |
| `https://github.com/user/repo` | `https://github.com/user/repo` |
| `ssh://git@github.com/user/repo.git` | `https://github.com/user/repo` |
| (空文字) | (空文字) |

**正規化ロジック**（サーバー側で実装）:
1. SSH形式 (`git@host:path`) を HTTPS 形式に変換
2. `ssh://` プロトコルを `https://` に変換
3. 末尾の `.git` を除去
4. トレイリングスラッシュを除去

### 3. ドメインモデルの変更

**Session**
```go
type Session struct {
    ID              string
    UserID          *string
    ProjectID       string     // 追加: Project への外部キー
    ClaudeSessionID string
    ProjectPath     string
    GitBranch       string     // 維持: ブランチ情報は Session に残す
    StartedAt       time.Time
    EndedAt         *time.Time
    CreatedAt       time.Time
}
// GitRemoteURL は削除
```

**PlanDocument**
```go
type PlanDocument struct {
    ID          string
    ProjectID   string     // 追加: Project への外部キー
    Description string
    Body        string
    Status      PlanDocumentStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
// GitRemoteURL は削除
```

### 4. Repository インターフェース

**ProjectRepository**
```go
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    FindByID(ctx context.Context, id string) (*Project, error)
    FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*Project, error)
    FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*Project, error)
    FindAll(ctx context.Context, limit int, offset int) ([]*Project, error)
    GetDefaultProject(ctx context.Context) (*Project, error) // CanonicalGitRepository が空のプロジェクト
}
```

**SessionRepository の変更**
```go
type SessionRepository interface {
    // 既存メソッドの維持
    Create(ctx context.Context, session *Session) error
    FindByID(ctx context.Context, id string) (*Session, error)
    FindAll(ctx context.Context, limit int, offset int) ([]*Session, error)
    FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*Session, error)
    UpdateUserID(ctx context.Context, id string, userID string) error
    UpdateProjectPath(ctx context.Context, id string, projectPath string) error

    // 変更: GitInfo → ProjectID
    UpdateProjectID(ctx context.Context, id string, projectID string) error
    UpdateGitBranch(ctx context.Context, id string, gitBranch string) error

    // 追加: Project でフィルタリング
    FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*Session, error)
}
```

**PlanDocumentRepository の変更**
```go
type PlanDocumentRepository interface {
    Create(ctx context.Context, doc *PlanDocument) error
    FindByID(ctx context.Context, id string) (*PlanDocument, error)
    FindAll(ctx context.Context, limit int, offset int) ([]*PlanDocument, error)
    Update(ctx context.Context, doc *PlanDocument) error
    Delete(ctx context.Context, id string) error
    SetStatus(ctx context.Context, id string, status PlanDocumentStatus) error

    // 変更: GitRemoteURL → ProjectID
    FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*PlanDocument, error)
}
```

### 5. API の変更

**Ingest API** (`POST /api/ingest`)

リクエストは変更なし（後方互換性維持）:
```json
{
  "session_id": "xxx",
  "transcript_lines": [...],
  "cwd": "/path/to/project",
  "git_remote_url": "git@github.com:user/repo.git",
  "git_branch": "main"
}
```

サーバー側の処理:
1. `git_remote_url` を正規化
2. 正規化した URL で Project を検索/作成
3. Session の `project_id` を設定

**Session API** レスポンス変更

```json
{
  "id": "session-uuid",
  "project": {
    "id": "project-uuid",
    "canonical_git_repository": "https://github.com/user/repo"
  },
  "git_branch": "main",
  ...
}
```

**PlanDocument API**

リクエスト変更（後方互換性維持）:
```json
{
  "description": "...",
  "body": "...",
  "git_remote_url": "https://github.com/user/repo"  // 引き続きサポート
}
```

または新形式:
```json
{
  "description": "...",
  "body": "...",
  "project_id": "project-uuid"
}
```

レスポンス:
```json
{
  "id": "plan-uuid",
  "project": {
    "id": "project-uuid",
    "canonical_git_repository": "https://github.com/user/repo"
  },
  ...
}
```

**Project API** (新規)

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/projects` | Project 一覧 |
| GET | `/api/projects/:id` | Project 詳細 |
| GET | `/api/projects/:id/sessions` | Project 配下の Session 一覧 |
| GET | `/api/projects/:id/plans` | Project 配下の PlanDocument 一覧 |

### 6. データベースマイグレーション

**Phase 1: テーブル追加とカラム追加**

```sql
-- projects テーブル作成
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    canonical_git_repository TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(canonical_git_repository)
);

-- デフォルトプロジェクト作成（canonical_git_repository が空）
INSERT INTO projects (id, canonical_git_repository)
VALUES ('00000000-0000-0000-0000-000000000000', '');

-- sessions に project_id カラム追加
ALTER TABLE sessions ADD COLUMN project_id TEXT REFERENCES projects(id);

-- plan_documents に project_id カラム追加
ALTER TABLE plan_documents ADD COLUMN project_id TEXT REFERENCES projects(id);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_projects_canonical ON projects(canonical_git_repository);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_plan_documents_project ON plan_documents(project_id);
```

**Phase 2: 既存データ移行**

```sql
-- 既存の git_remote_url から projects を作成
INSERT OR IGNORE INTO projects (id, canonical_git_repository)
SELECT
    lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))),
    DISTINCT normalized_url -- 正規化処理が必要
FROM sessions
WHERE git_remote_url IS NOT NULL AND git_remote_url != '';

-- sessions.project_id を設定
UPDATE sessions
SET project_id = (
    SELECT p.id FROM projects p
    WHERE p.canonical_git_repository = normalize_url(sessions.git_remote_url)
)
WHERE git_remote_url IS NOT NULL AND git_remote_url != '';

-- git_remote_url が空の sessions はデフォルトプロジェクトに
UPDATE sessions
SET project_id = '00000000-0000-0000-0000-000000000000'
WHERE project_id IS NULL;

-- plan_documents も同様
```

**Phase 3: 旧カラム削除（将来）**

```sql
-- 十分な移行期間後
ALTER TABLE sessions DROP COLUMN git_remote_url;
ALTER TABLE plan_documents DROP COLUMN git_remote_url;
```

### 7. Web UI の変更

**Project 表示**

Session/PlanDocument の一覧・詳細で Project 情報を表示:
- `canonical_git_repository` があれば: `github.com/user/repo` の形式で表示
- `canonical_git_repository` が空なら: `(no project)` と表示

**新規ページ**

- `/projects` - Project 一覧ページ（オプション）
- `/projects/:id` - Project 詳細ページ（配下の Session/PlanDocument を表示）

### 8. CLI の変更

CLI 側は変更不要（後方互換性維持）:
- 引き続き `git_remote_url` を送信
- サーバー側で正規化・Project 紐付けを実施

### 9. 実装順序

1. **Server: Domain/Repository**
   - Project ドメインモデル追加
   - ProjectRepository インターフェース定義
   - Git URL 正規化ユーティリティ実装
   - SQLite/Memory/Postgres/MongoDB 実装

2. **Server: Migration**
   - マイグレーションスクリプト作成
   - 既存データ移行スクリプト作成

3. **Server: API**
   - Ingest ハンドラー修正（Project 自動作成）
   - Session ハンドラー修正（Project 情報返却）
   - PlanDocument ハンドラー修正（Project 自動作成・情報返却）
   - Project API 新規作成

4. **Web: Types/API**
   - Project 型定義追加
   - Session/PlanDocument 型に Project を追加
   - Project API クライアント追加

5. **Web: Components**
   - ProjectBadge コンポーネント（git_remote_url または "no project" を表示）
   - Session/PlanDocument 一覧・詳細に Project 表示追加

6. **Web: Pages（オプション）**
   - Project 一覧・詳細ページ

### 10. 移行期間の互換性

**API 互換性**
- `git_remote_url` パラメータは引き続きサポート
- レスポンスに `git_remote_url` と `project` の両方を含める（当面）

**データ整合性**
- 新規作成: 必ず Project 紐付け
- 既存データ: マイグレーションで全て Project 紐付け
- `project_id` は NOT NULL（デフォルトプロジェクトがあるため）

## 未決定事項

1. **Project の追加属性**: 現時点では `canonical_git_repository` のみ。将来的に名前やアイコンを追加する可能性
2. **Project の削除**: 配下にデータがある場合の挙動（禁止 or CASCADE）
3. **認可**: Project レベルでのアクセス制御（将来の拡張）
