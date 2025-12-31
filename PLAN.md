# PlanDocument リソース実装計画

## 概要

Claude Codeが実装や変更の計画を記録・管理するための新しいリソース「PlanDocument」を追加する。

## データモデル

### PlanDocument
| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | UUID |
| description | string | 短い説明 |
| body | string | 計画のMarkdownテキスト |
| git_remote_url | string | 作成時のgitリポジトリURL |
| created_at | time | 作成日時 |
| updated_at | time | 更新日時 |

※ collaborators は PlanDocumentEvent の user_id から動的に集計（APIレスポンス時）

### PlanDocumentEvent
| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | UUID |
| plan_document_id | string | FK |
| session_id | *string | nullable - Sessionとの関連 |
| user_id | *string | 変更を行ったユーザー |
| patch | string | diff-match-patch形式のパッチ |
| created_at | time | 作成日時 |

## MCP Tools（CLI統合）

`npx agentrace mcp-server` で起動、stdioでMCPプロトコル処理

| Tool | 説明 | 引数 |
|------|------|------|
| list_plans | リポジトリのPlan一覧取得 | git_remote_url |
| read_plan | Plan読み込み | id |
| create_plan | Plan作成 | description, body, git_remote_url |
| update_plan | Plan更新（パッチ生成） | id, body, session_id? |

## API エンドポイント

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| GET | `/api/plans` | Bearer/Session | 一覧（?git_remote_url=でフィルタ） |
| GET | `/api/plans/:id` | Bearer/Session | 詳細 |
| GET | `/api/plans/:id/events` | Bearer/Session | 変更履歴 |
| POST | `/api/plans` | Bearer | 作成 |
| PATCH | `/api/plans/:id` | Bearer | 更新 |
| DELETE | `/api/plans/:id` | Bearer | 削除 |

## Web 画面

- `/plans` - PlanDocument一覧ページ
- `/plans/:id` - 詳細ページ（Markdown表示 + 変更履歴タブ）

---

## 実装ファイル一覧

### Phase 1: Server ドメイン・Repository

**新規作成:**
- `server/internal/domain/plan_document.go`
- `server/internal/domain/plan_document_event.go`
- `server/internal/repository/memory/plan_document.go`
- `server/internal/repository/memory/plan_document_event.go`
- `server/internal/repository/sqlite/plan_document.go`
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/mongodb/plan_document.go`
- `server/internal/repository/mongodb/plan_document_event.go`
- `server/migrations/sqlite/002_plan_documents.sql`
- `server/migrations/postgres/002_plan_documents.up.sql`

**修正:**
- `server/internal/repository/interface.go` - インターフェース追加
- `server/internal/repository/memory/repositories.go` - Repository追加
- `server/internal/repository/sqlite/repositories.go` - Repository追加
- `server/internal/repository/sqlite/db.go` - マイグレーション追加
- `server/internal/repository/postgres/repositories.go` - Repository追加
- `server/internal/repository/postgres/db.go` - マイグレーション追加
- `server/internal/repository/mongodb/repositories.go` - Repository追加
- `server/migrations/embed.go` - SQL埋め込み

### Phase 2: Server API

**新規作成:**
- `server/internal/api/plan_document.go`

**修正:**
- `server/internal/api/router.go` - ルート追加

### Phase 3: CLI MCP Server

**新規作成:**
- `cli/src/commands/mcp-server.ts`
- `cli/src/mcp/plan-document-client.ts`

**修正:**
- `cli/package.json` - 依存追加（@modelcontextprotocol/server, zod, diff-match-patch-es）
- `cli/src/index.ts` - コマンド追加
- `cli/src/hooks/installer.ts` - MCP サーバー設定のインストール処理追加

**init時のMCPサーバー設定:**

`~/.claude/settings.json` に以下を追加:
```json
{
  "mcpServers": {
    "agentrace": {
      "command": "npx",
      "args": ["agentrace", "mcp-server"]
    }
  }
}
```

開発モード（--dev）の場合:
```json
{
  "mcpServers": {
    "agentrace": {
      "command": "npx",
      "args": ["tsx", "/path/to/cli/src/index.ts", "mcp-server"]
    }
  }
}
```

### Phase 4: Web フロントエンド

**新規作成:**
- `web/src/types/plan-document.ts`
- `web/src/api/plan-documents.ts`
- `web/src/pages/PlansPage.tsx`
- `web/src/pages/PlanDetailPage.tsx`
- `web/src/components/plans/PlanList.tsx`
- `web/src/components/plans/PlanCard.tsx`
- `web/src/components/plans/PlanEventHistory.tsx`

**修正:**
- `web/src/App.tsx` - ルート追加
- `web/src/components/layout/Header.tsx` - ナビゲーション追加

---

## 技術詳細

### collaborators集計
- APIレスポンス時に PlanDocumentEvent の user_id を DISTINCT で取得
- User情報と結合してレスポンスに含める

### パッチ形式
- CLI: `diff-match-patch-es` パッケージ
- パッチ生成はCLI側で実行、Server側はパッチを保存のみ

### 依存パッケージ（CLI）
```json
{
  "@modelcontextprotocol/server": "^2.0.0",
  "zod": "^3.25.0",
  "diff-match-patch-es": "^1.0.0"
}
```

---

## 実装順序

1. **Phase 1**: Server ドメイン・Repository（memory + sqlite） ✅ 完了
2. **Phase 2**: Server API ハンドラー・ルーティング
3. **Phase 3**: CLI MCP Server コマンド
4. **Phase 4**: Web フロントエンド

各Phase完了後に動作確認を行う。

---

## 実装ログ

### Phase 1 完了

**作成したファイル:**
- `server/internal/domain/plan_document.go`
- `server/internal/domain/plan_document_event.go`
- `server/internal/repository/memory/plan_document.go`
- `server/internal/repository/memory/plan_document_event.go`
- `server/internal/repository/sqlite/plan_document.go`
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/mongodb/plan_document.go`
- `server/internal/repository/mongodb/plan_document_event.go`
- `server/migrations/sqlite/002_plan_documents.sql`
- `server/migrations/postgres/002_plan_documents.up.sql`

**修正したファイル:**
- `server/internal/repository/interface.go`
- `server/internal/repository/memory/repositories.go`
- `server/internal/repository/sqlite/repositories.go`
- `server/internal/repository/sqlite/db.go`
- `server/internal/repository/postgres/repositories.go`
- `server/internal/repository/postgres/db.go`
- `server/internal/repository/mongodb/repositories.go`
- `server/migrations/embed.go`

**確認:** `go build ./...` 成功
