# Plan Document Web操作機能 実装計画

**ステータス: 実装完了**

## 概要
Web画面からPlan Documentの新規作成・編集・ステータス変更を行えるようにする。
ステータス変更もPlan Document Eventとして保存する。

---

## Phase 1: Server側変更

### 1.1 PlanDocumentEvent ドメインモデル変更
**ファイル**: `server/internal/domain/plan_document_event.go`

- `EventType` フィールド追加（`body_change` | `status_change`）
- 既存データ互換性: EventTypeが空の場合は`body_change`として扱う

```go
type PlanDocumentEventType string

const (
    PlanDocumentEventTypeBodyChange   PlanDocumentEventType = "body_change"
    PlanDocumentEventTypeStatusChange PlanDocumentEventType = "status_change"
)
```

### 1.2 DBマイグレーション追加
**ファイル**: `server/migrations/sqlite/002_add_event_type.sql`, `server/migrations/postgres/002_add_event_type.sql`

```sql
ALTER TABLE plan_document_events ADD COLUMN event_type TEXT NOT NULL DEFAULT 'body_change';
```

### 1.3 Repository層変更
**ファイル**:
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/memory/plan_document_event.go`

- `Create`: event_typeカラム追加
- `scanEvent`/FindByXXX: event_type読み取り追加

### 1.4 SetStatusでイベント作成
**ファイル**: `server/internal/api/plan_document.go:432-479`

```go
func (h *PlanDocumentHandler) SetStatus(...) {
    oldStatus := doc.Status
    // ステータス更新後...

    // イベント作成を追加
    event := &domain.PlanDocumentEvent{
        PlanDocumentID: doc.ID,
        EventType:      domain.PlanDocumentEventTypeStatusChange,
        Patch:          fmt.Sprintf("%s -> %s", oldStatus, status),
    }
    if userID != "" {
        event.UserID = &userID
    }
    h.repos.PlanDocumentEvent.Create(ctx, event)
}
```

### 1.5 CreatePlanDocumentRequest にProjectID追加
**ファイル**: `server/internal/api/plan_document.go:65-69`

```go
type CreatePlanDocumentRequest struct {
    Description     string  `json:"description"`
    Body            string  `json:"body"`
    ProjectID       *string `json:"project_id"`       // 追加
    ClaudeSessionID *string `json:"claude_session_id"`
}
```

Create handler (line 275-340) で `req.ProjectID` があればそれを使用するよう変更。

### 1.6 認証ミドルウェア変更
**ファイル**: `server/internal/api/router.go:40-43`

Plan操作エンドポイントを `AuthenticateBearer` から `AuthenticateBearerOrSession` に変更。
- **CLI**: Bearer認証（APIキー）で引き続き利用可能
- **Web**: Session認証（ログイン必須）で利用可能
- **未認証**: 操作不可（閲覧のみ可能）

```go
// 変更前
apiBearer.HandleFunc("/plans", planDocumentHandler.Create).Methods("POST")

// 変更後: 新しいサブルーターを作成（または既存apiBearerのルートを移動）
apiBearerOrSession := r.PathPrefix("/api").Subrouter()
apiBearerOrSession.Use(mw.AuthenticateBearerOrSession)
apiBearerOrSession.HandleFunc("/plans", planDocumentHandler.Create).Methods("POST")
apiBearerOrSession.HandleFunc("/plans/{id}", planDocumentHandler.Update).Methods("PATCH")
apiBearerOrSession.HandleFunc("/plans/{id}", planDocumentHandler.Delete).Methods("DELETE")
apiBearerOrSession.HandleFunc("/plans/{id}/status", planDocumentHandler.SetStatus).Methods("PATCH")
```

### 1.7 Project一覧API追加
**ファイル**: `server/internal/api/project.go` (新規), `server/internal/api/router.go`

```go
// GET /api/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
    projects, _ := h.repos.Project.FindAll(ctx, limit, offset)
    // JSON返却
}
```

ProjectRepository の `FindAll` メソッドが必要であれば追加。

### 1.8 EventレスポンスにEventType追加
**ファイル**: `server/internal/api/plan_document.go:49-57`

```go
type PlanDocumentEventResponse struct {
    // 既存フィールド...
    EventType       string  `json:"event_type"` // 追加
}
```

---

## Phase 2: Web側変更

### 2.1 型定義更新
**ファイル**: `web/src/types/plan-document.ts`

```typescript
export type PlanDocumentEventType = 'body_change' | 'status_change'

export interface PlanDocumentEvent {
    // 既存フィールド...
    event_type: PlanDocumentEventType
}
```

### 2.2 APIクライアント拡張
**ファイル**: `web/src/api/plan-documents.ts`

```typescript
export async function createPlan(params: {
    description: string
    body: string
    project_id?: string
}): Promise<PlanDocument>

export async function updatePlan(id: string, params: {
    description?: string
    body?: string
    patch?: string
}): Promise<PlanDocument>

export async function deletePlan(id: string): Promise<void>
```

**ファイル**: `web/src/api/projects.ts` (新規)

```typescript
export async function getProjects(): Promise<{ projects: Project[] }>
```

### 2.3 基本UIコンポーネント追加
**ファイル**:
- `web/src/components/ui/Select.tsx` - ドロップダウン選択
- `web/src/components/ui/Textarea.tsx` - 複数行テキスト入力

### 2.4 Plan作成モーダル
**ファイル**: `web/src/components/plans/CreatePlanModal.tsx`

- Projectドロップダウン（Project一覧から選択）
- Description入力（Input）
- Body入力（Textarea）
- 作成ボタン

### 2.5 PlansPage変更
**ファイル**: `web/src/pages/PlansPage.tsx`

- 「Create Plan」ボタン追加
- CreatePlanModal表示制御

### 2.6 PlanDetailPage変更
**ファイル**: `web/src/pages/PlanDetailPage.tsx`

- ステータス変更ドロップダウン追加（ヘッダー部分）
- 編集モード切替ボタン追加
- description/bodyの編集フォーム（編集モード時）
- 保存時にpatch計算（diff-match-patch）

### 2.7 History表示拡張
**ファイル**: `web/src/components/plans/PlanEventHistory.tsx`

- `event_type === 'status_change'` の場合は「Status changed: draft → planning」のように表示

---

## 実装順序

1. Server: ドメインモデル変更 (plan_document_event.go)
2. Server: マイグレーション追加
3. Server: Repository層変更（memory, sqlite, postgres）
4. Server: SetStatusでイベント作成
5. Server: CreateRequestにProjectID追加
6. Server: 認証ミドルウェア変更
7. Server: Project一覧API追加
8. Server: EventレスポンスにEventType追加
9. Web: 型定義・APIクライアント
10. Web: 基本UIコンポーネント
11. Web: Plan作成モーダル
12. Web: PlansPage変更
13. Web: ステータス変更UI (PlanDetailPage)
14. Web: 編集機能 (PlanDetailPage)
15. Web: History表示拡張

---

## 修正対象ファイル一覧

### Server
- `server/internal/domain/plan_document_event.go`
- `server/migrations/sqlite/002_add_event_type.sql` (新規)
- `server/migrations/postgres/002_add_event_type.sql` (新規)
- `server/internal/repository/interface.go` (FindAll追加の場合)
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/memory/plan_document_event.go`
- `server/internal/api/plan_document.go`
- `server/internal/api/router.go`
- `server/internal/api/project.go` (新規)

### Web
- `web/src/types/plan-document.ts`
- `web/src/api/plan-documents.ts`
- `web/src/api/projects.ts` (新規)
- `web/src/components/ui/Select.tsx` (新規)
- `web/src/components/ui/Textarea.tsx` (新規)
- `web/src/components/plans/CreatePlanModal.tsx` (新規)
- `web/src/components/plans/PlanEventHistory.tsx`
- `web/src/pages/PlansPage.tsx`
- `web/src/pages/PlanDetailPage.tsx`
