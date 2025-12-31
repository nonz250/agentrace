# API エンドポイント

## データ受信（CLI用）

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/api/ingest` | Bearer | transcript行を受信 |
| GET | `/api/sessions` | Bearer/Session | セッション一覧 |
| GET | `/api/sessions/:id` | Bearer/Session | セッション詳細（イベント含む） |
| GET | `/health` | なし | ヘルスチェック |

## PlanDocument

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| GET | `/api/plans` | Bearer/Session | Plan一覧（?git_remote_url=でフィルタ） |
| GET | `/api/plans/:id` | Bearer/Session | Plan詳細 |
| GET | `/api/plans/:id/events` | Bearer/Session | Plan変更履歴 |
| POST | `/api/plans` | Bearer | Plan作成 |
| PATCH | `/api/plans/:id` | Bearer | Plan更新 |
| DELETE | `/api/plans/:id` | Bearer | Plan削除 |

## 認証

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/auth/register` | なし | ユーザー登録 |
| POST | `/auth/login` | なし | メール/パスワードでログイン |
| POST | `/auth/login/apikey` | なし | APIキーでログイン |
| GET | `/auth/session` | なし | トークンでログイン（CLI経由） |
| GET | `/auth/config` | なし | OAuth設定確認（有効なプロバイダー） |
| GET | `/auth/github` | なし | GitHub OAuth開始（リダイレクト） |
| GET | `/auth/github/callback` | なし | GitHub OAuthコールバック |
| POST | `/api/auth/web-session` | Bearer | Webログイントークン発行 |
| POST | `/api/auth/logout` | Session | ログアウト |
| GET | `/api/me` | Session | 自分の情報 |
| PATCH | `/api/me` | Session | 自分の情報を更新 |
| GET | `/api/users` | Session | ユーザー一覧 |
| GET | `/api/keys` | Session | 自分のAPIキー一覧 |
| POST | `/api/keys` | Session | 新しいAPIキー発行 |
| DELETE | `/api/keys/:id` | Session | APIキー削除 |

## リクエスト/レスポンス形式

### POST /api/ingest

**リクエスト**

```json
{
  "session_id": "string",
  "transcript_lines": [{"type": "...", ...}],
  "cwd": "string (作業ディレクトリ)",
  "git_remote_url": "string (git remote origin URL)",
  "git_branch": "string (現在のブランチ名)"
}
```

**レスポンス**

```json
{
  "ok": true,
  "events_created": 10
}
```

### GET /api/sessions

**クエリパラメータ**

| パラメータ | 説明 | デフォルト |
|-----------|------|-----------|
| `limit` | 取得件数 | 100 |
| `offset` | オフセット | 0 |

**レスポンス**

```json
{
  "sessions": [
    {
      "id": "string",
      "user_id": "string | null",
      "user_name": "string | null",
      "claude_session_id": "string",
      "project_path": "string",
      "git_remote_url": "string",
      "git_branch": "string",
      "started_at": "ISO 8601",
      "ended_at": "ISO 8601 | null",
      "event_count": 42,
      "created_at": "ISO 8601"
    }
  ],
  "total": 100
}
```

### GET /api/sessions/:id

**レスポンス**

```json
{
  "id": "string",
  "user_id": "string | null",
  "user_name": "string | null",
  "claude_session_id": "string",
  "project_path": "string",
  "git_remote_url": "string",
  "git_branch": "string",
  "started_at": "ISO 8601",
  "ended_at": "ISO 8601 | null",
  "event_count": 42,
  "created_at": "ISO 8601",
  "events": [
    {
      "id": "string",
      "session_id": "string",
      "event_type": "user | assistant",
      "payload": { ... },
      "created_at": "ISO 8601"
    }
  ]
}
```

## エラーレスポンス

```json
{
  "error": "エラーメッセージ"
}
```

| ステータス | 用途 |
|-----------|------|
| 400 | 入力検証エラー |
| 401 | 認証失敗 |
| 403 | 権限不足 |
| 404 | リソース未検索 |
| 409 | 競合（email重複等） |
| 500 | サーバーエラー |

## ミドルウェア

| エンドポイント | ミドルウェア |
|---------------|-------------|
| `/api/ingest` | AuthenticateBearer |
| `/api/auth/web-session` | AuthenticateBearer |
| `/api/me`, `/api/keys`, `/api/users` | AuthenticateSession |
| `/api/sessions` | AuthenticateBearerOrSession |
| `/api/plans` (GET) | AuthenticateBearerOrSession |
| `/api/plans` (POST), `/api/plans/:id` (PATCH, DELETE) | AuthenticateBearer |

## PlanDocument API詳細

### GET /api/plans

**クエリパラメータ**

| パラメータ | 説明 | デフォルト |
|-----------|------|-----------|
| `git_remote_url` | リポジトリURLでフィルタ | なし（全件） |
| `limit` | 取得件数 | 100 |
| `offset` | オフセット | 0 |

**レスポンス**

```json
{
  "plans": [
    {
      "id": "string",
      "description": "string",
      "body": "string (Markdown)",
      "git_remote_url": "string",
      "collaborators": [
        {
          "id": "string",
          "display_name": "string"
        }
      ],
      "created_at": "ISO 8601",
      "updated_at": "ISO 8601"
    }
  ]
}
```

### GET /api/plans/:id

**レスポンス**

```json
{
  "id": "string",
  "description": "string",
  "body": "string (Markdown)",
  "git_remote_url": "string",
  "collaborators": [
    {
      "id": "string",
      "display_name": "string"
    }
  ],
  "created_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

### GET /api/plans/:id/events

**レスポンス**

```json
{
  "events": [
    {
      "id": "string",
      "plan_document_id": "string",
      "session_id": "string | null",
      "user_id": "string | null",
      "user_name": "string | null",
      "patch": "string (diff-match-patch形式)",
      "created_at": "ISO 8601"
    }
  ]
}
```

### POST /api/plans

**リクエスト**

```json
{
  "description": "string (必須)",
  "body": "string (Markdown)",
  "git_remote_url": "string",
  "session_id": "string | null (Claude Codeセッションとの関連付け)"
}
```

**レスポンス**: `201 Created` + Plan詳細（GET /api/plans/:id と同形式）

### PATCH /api/plans/:id

**リクエスト**

```json
{
  "description": "string | null (更新する場合のみ)",
  "body": "string | null (更新する場合のみ)",
  "patch": "string | null (diff-match-patch形式、履歴用)",
  "session_id": "string | null (Claude Codeセッションとの関連付け)"
}
```

**レスポンス**: `200 OK` + Plan詳細（GET /api/plans/:id と同形式）

### DELETE /api/plans/:id

**レスポンス**: `204 No Content`
