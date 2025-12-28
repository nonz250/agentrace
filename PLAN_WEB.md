# Web 実装計画

## 概要

React + Vite で実装するフロントエンド。ユーザー登録/ログインとセッション一覧・詳細の表示を行う。

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
│   │   └── keys.ts           # APIキー API
│   ├── hooks/                # カスタム Hooks
│   │   ├── useAuth.ts        # 認証状態管理
│   │   ├── useSessions.ts    # セッション一覧
│   │   └── useWebSocket.ts   # WebSocket 接続（Step 5）
│   ├── pages/                # ページコンポーネント
│   │   ├── WelcomePage.tsx   # 初期画面（登録/ログイン選択）
│   │   ├── RegisterPage.tsx  # ユーザー登録
│   │   ├── LoginPage.tsx     # ログイン
│   │   ├── SessionListPage.tsx
│   │   ├── SessionDetailPage.tsx
│   │   └── SettingsPage.tsx  # 設定（APIキー管理）
│   ├── components/           # 共通コンポーネント
│   │   ├── Layout.tsx
│   │   ├── SessionList.tsx
│   │   ├── SessionTimeline.tsx
│   │   ├── EventCard.tsx
│   │   └── ApiKeyManager.tsx # APIキー管理
│   └── styles/
│       └── global.css
├── index.html
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 画面構成

### 1. 初期画面（WelcomePage）

初回アクセス時に表示。登録済みならダッシュボードへリダイレクト。

```
┌─────────────────────────────────────────┐
│ Agentrace                               │
├─────────────────────────────────────────┤
│                                         │
│ Track and review Claude Code sessions   │
│ with your team.                         │
│                                         │
│     [Register]  [Login with API Key]    │
│                                         │
└─────────────────────────────────────────┘
```

### 2. ユーザー登録（RegisterPage）

名前を入力するだけでAPIキーが発行される。

```
┌─────────────────────────────────────────┐
│ Create Account                          │
├─────────────────────────────────────────┤
│                                         │
│ Your Name: [__________________]         │
│                                         │
│            [Create Account]             │
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│ Account Created!                        │
├─────────────────────────────────────────┤
│ Your API Key:                           │
│ ┌─────────────────────────────────────┐ │
│ │ agtr_xxxxxxxxxxxxxxxxxxxxxxxx [Copy]│ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ⚠️ Save this key - it won't be shown   │
│    again.                               │
│                                         │
│ Set up CLI:                             │
│ $ npx agentrace init                    │
│                                         │
│            [Go to Dashboard]            │
└─────────────────────────────────────────┘
```

### 3. ログイン（LoginPage）

APIキーを入力してログイン。

```
┌─────────────────────────────────────────┐
│ Login                                   │
├─────────────────────────────────────────┤
│                                         │
│ API Key: [__________________________]   │
│                                         │
│            [Login]                      │
│                                         │
│ Don't have an account? [Register]       │
└─────────────────────────────────────────┘
```

### 4. セッション一覧（SessionListPage）

全ユーザーのセッションを表示。

```
┌─────────────────────────────────────────┐
│ Agentrace          [Taro ▼] [Logout]    │
├─────────────────────────────────────────┤
│ Sessions                                │
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ /path/to/project                    │ │
│ │ Taro • 2025-12-28 10:30            │ │
│ └─────────────────────────────────────┘ │
│ ┌─────────────────────────────────────┐ │
│ │ /path/to/another-project            │ │
│ │ Hanako • 2025-12-28 09:15          │ │
│ └─────────────────────────────────────┘ │
│ ...                                     │
└─────────────────────────────────────────┘
```

### 5. セッション詳細（SessionDetailPage）

タイムライン形式でイベント表示。

```
┌─────────────────────────────────────────┐
│ ← Back to Sessions                      │
├─────────────────────────────────────────┤
│ /path/to/project                        │
│ Taro • Started 2025-12-28 10:30        │
├─────────────────────────────────────────┤
│ Timeline                                │
│                                         │
│ ● 10:30:05 user                        │
│   ├─ [▼ Show content]                  │
│                                         │
│ ● 10:30:10 assistant                   │
│   ├─ [▼ Show content]                  │
│                                         │
│ ● 10:30:15 tool_use                    │
│   ├─ [▼ Show content]                  │
│                                         │
└─────────────────────────────────────────┘
```

### 6. 設定（SettingsPage）

APIキーの管理画面。複数デバイス用に複数のAPIキーを発行可能。

```
┌─────────────────────────────────────────┐
│ Agentrace          [Taro ▼] [Logout]    │
├─────────────────────────────────────────┤
│ Settings                                │
├─────────────────────────────────────────┤
│ API Keys                                │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ MacBook Pro                         │ │
│ │ agtr_xxxx... • Last used: 1h ago   │ │
│ │                          [Delete]   │ │
│ └─────────────────────────────────────┘ │
│ ┌─────────────────────────────────────┐ │
│ │ Work PC                             │ │
│ │ agtr_yyyy... • Last used: 3d ago   │ │
│ │                          [Delete]   │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ Name: [____________________]        │ │
│ │              [Create New API Key]   │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

**新規キー発行後:**

```
┌─────────────────────────────────────────┐
│ New API Key Created!                    │
├─────────────────────────────────────────┤
│                                         │
│ Your API Key:                           │
│ ┌─────────────────────────────────────┐ │
│ │ agtr_zzzzzzzzzzzzzzzzzzzzzz   [Copy]│ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ⚠️ Save this key - it won't be shown   │
│    again.                               │
│                                         │
│               [Done]                    │
└─────────────────────────────────────────┘
```

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
          <div className="session-meta">
            <span>{session.userName}</span>
            <span>{formatTime(session.startedAt)}</span>
          </div>
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
        <span className="event-time">{formatTime(event.createdAt)}</span>
      </div>
      {expanded && (
        <div className="event-detail">
          <pre>{JSON.stringify(event.payload, null, 2)}</pre>
        </div>
      )}
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

// src/api/auth.ts

export async function register(name: string): Promise<{ user: User; api_key: string }> {
  return fetchAPI('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function login(apiKey: string): Promise<{ user: User }> {
  return fetchAPI('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ api_key: apiKey }),
  });
}

export async function logout(): Promise<void> {
  return fetchAPI('/api/auth/logout', { method: 'POST' });
}

export async function getMe(): Promise<User> {
  return fetchAPI('/api/me');
}

// src/api/sessions.ts

export async function getSessions(): Promise<Session[]> {
  return fetchAPI('/api/sessions');
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`);
}

// src/api/keys.ts

export async function getKeys(): Promise<{ keys: ApiKey[] }> {
  return fetchAPI('/api/keys');
}

export async function createKey(name: string): Promise<{ key: ApiKey; api_key: string }> {
  return fetchAPI('/api/keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function deleteKey(id: string): Promise<void> {
  return fetchAPI(`/api/keys/${id}`, { method: 'DELETE' });
}
```

## 認証フロー

### 初回アクセス

```
1. WelcomePageが表示される
2. GET /api/me を試行
   - 成功 → SessionListPageへリダイレクト
   - 401 → そのまま表示
```

### 登録フロー

```
1. RegisterPageで名前入力
2. POST /auth/register
3. レスポンスでAPIキー表示（+ 自動ログイン）
4. ユーザーがAPIキーをコピー
5. 「Go to Dashboard」でSessionListPageへ
```

### ログインフロー（Web）

```
1. LoginPageでAPIキー入力
2. POST /auth/login
3. セッションCookie設定
4. SessionListPageへリダイレクト
```

### ログインフロー（CLI経由）

```
1. CLI: POST /api/auth/web-session
2. CLI: ブラウザでURL開く
3. ブラウザ: GET /auth/session?token=xxxxx
4. サーバー: セッションCookie設定
5. ブラウザ: SessionListPageへリダイレクト
```

## WebSocket 接続（Step 5）

```ts
// src/hooks/useWebSocket.ts

export function useWebSocket() {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    const ws = new WebSocket(`${WS_URL}/ws/live`);

    ws.onmessage = (e) => {
      const event = JSON.parse(e.data);
      setEvents(prev => [...prev, event]);
    };

    return () => ws.close();
  }, []);

  return events;
}
```

## 依存パッケージ

- `react` + `react-dom`
- `react-router-dom` - ルーティング
- `@tanstack/react-query` - データフェッチング（オプション）
- `date-fns` - 日時フォーマット

## 実装順序

### Step 3: 基本UI

1. Vite + React セットアップ
2. WelcomePage（初期画面）
3. RegisterPage（名前入力→APIキー発行）
4. LoginPage（APIキー入力）
5. SessionListPage（セッション一覧）
6. SessionDetailPage（タイムライン表示）

### Step 5: リアルタイム機能

1. WebSocket 接続
2. 新規イベントのリアルタイム表示

## 環境変数

| 変数名 | 説明 |
| ------ | ---- |
| `VITE_API_URL` | バックエンドAPI URL |
