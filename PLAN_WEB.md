# Web 実装計画

## 概要

React + Vite で実装するフロントエンド。セッション一覧・詳細の表示とAPIキー管理を行う。

## ディレクトリ構成

```text
web/
├── src/
│   ├── main.tsx              # エントリーポイント
│   ├── App.tsx               # ルーティング
│   ├── api/                  # API クライアント
│   │   ├── client.ts         # fetch ラッパー
│   │   ├── auth.ts           # 認証 API
│   │   ├── sessions.ts       # セッション API
│   │   └── apikeys.ts        # APIキー API
│   ├── hooks/                # カスタム Hooks
│   │   ├── useAuth.ts        # 認証状態管理
│   │   ├── useSessions.ts    # セッション一覧
│   │   └── useWebSocket.ts   # WebSocket 接続
│   ├── pages/                # ページコンポーネント
│   │   ├── LoginPage.tsx
│   │   ├── SessionListPage.tsx
│   │   ├── SessionDetailPage.tsx
│   │   └── SettingsPage.tsx
│   ├── components/           # 共通コンポーネント
│   │   ├── Layout.tsx
│   │   ├── SessionList.tsx
│   │   ├── SessionTimeline.tsx
│   │   ├── EventCard.tsx
│   │   └── ApiKeyManager.tsx
│   └── styles/
│       └── global.css
├── index.html
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 画面構成

### 1. ログイン/登録

- メール/パスワード
- OAuth ボタン（GitHub, Google）

### 2. セッション一覧

- 日時順でソート
- ユーザー、プロジェクトでフィルタ
- 検索機能

### 3. セッション詳細

- タイムライン形式でイベント表示
- 各イベントの展開/折りたたみ
- ツール入出力の表示

### 4. 設定

- プロフィール編集
- APIキー管理（発行/削除）

## コンポーネント設計

### SessionList

```tsx
interface SessionListProps {
  sessions: Session[];
  onSelect: (session: Session) => void;
}

function SessionList({ sessions, onSelect }: SessionListProps) {
  return (
    <div className="session-list">
      {sessions.map(session => (
        <div key={session.id} onClick={() => onSelect(session)}>
          <div className="session-project">{session.projectPath}</div>
          <div className="session-time">{formatTime(session.startedAt)}</div>
          <div className="session-user">{session.userName}</div>
        </div>
      ))}
    </div>
  );
}
```

### SessionTimeline

```tsx
interface SessionTimelineProps {
  events: Event[];
}

function SessionTimeline({ events }: SessionTimelineProps) {
  return (
    <div className="timeline">
      {events.map(event => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  );
}
```

### EventCard

```tsx
interface EventCardProps {
  event: Event;
}

function EventCard({ event }: EventCardProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="event-card">
      <div className="event-header" onClick={() => setExpanded(!expanded)}>
        <span className="event-type">{event.eventType}</span>
        <span className="tool-name">{event.toolName}</span>
        <span className="event-time">{formatTime(event.createdAt)}</span>
      </div>
      {expanded && (
        <div className="event-detail">
          <div className="event-input">
            <h4>Input</h4>
            <pre>{JSON.stringify(event.payload.tool_input, null, 2)}</pre>
          </div>
          <div className="event-output">
            <h4>Output</h4>
            <pre>{JSON.stringify(event.payload.tool_response, null, 2)}</pre>
          </div>
        </div>
      )}
    </div>
  );
}
```

### ApiKeyManager

```tsx
function ApiKeyManager() {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [newKeyName, setNewKeyName] = useState('');

  const createKey = async () => {
    const key = await api.createApiKey({ name: newKeyName });
    // 新規作成時のみ key.rawKey が返される（一度だけ表示）
    alert(`New API Key: ${key.rawKey}`);
    setKeys([...keys, key]);
  };

  const deleteKey = async (id: string) => {
    await api.deleteApiKey(id);
    setKeys(keys.filter(k => k.id !== id));
  };

  return (
    <div className="api-key-manager">
      <h3>API Keys</h3>
      <ul>
        {keys.map(key => (
          <li key={key.id}>
            <span>{key.name}</span>
            <span>{key.keyPrefix}...</span>
            <button onClick={() => deleteKey(key.id)}>Delete</button>
          </li>
        ))}
      </ul>
      <div className="new-key">
        <input
          value={newKeyName}
          onChange={e => setNewKeyName(e.target.value)}
          placeholder="Key name (e.g., My MacBook)"
        />
        <button onClick={createKey}>Create New Key</button>
      </div>
    </div>
  );
}
```

## API クライアント

```ts
// src/api/client.ts

const BASE_URL = import.meta.env.VITE_API_URL || '';

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    throw new Error(`API Error: ${res.status}`);
  }

  return res.json();
}

// src/api/sessions.ts

export async function getSessions(workspaceId: string): Promise<Session[]> {
  return fetchAPI(`/api/sessions?workspace_id=${workspaceId}`);
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`);
}
```

## WebSocket 接続

```ts
// src/hooks/useWebSocket.ts

export function useWebSocket(workspaceId: string) {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    const ws = new WebSocket(`${WS_URL}/ws/live?workspace_id=${workspaceId}`);

    ws.onmessage = (e) => {
      const event = JSON.parse(e.data);
      setEvents(prev => [...prev, event]);
    };

    return () => ws.close();
  }, [workspaceId]);

  return events;
}
```

## 依存パッケージ

- `react` + `react-dom`
- `react-router-dom` - ルーティング
- `@tanstack/react-query` - データフェッチング（オプション）
- `date-fns` - 日時フォーマット

## 実装順序

### Step 4: 基本UI

1. Vite + React セットアップ
2. ログイン画面
3. セッション一覧
4. セッション詳細（タイムライン）

### Step 5: リアルタイム機能

1. WebSocket 接続
2. 新規イベントのリアルタイム表示

## 環境変数

| 変数名 | 説明 |
| ------ | ---- |
| `VITE_API_URL` | バックエンドAPI URL |
