# Web 実装計画

## 概要

React + Vite + Tailwind CSS で実装するフロントエンド。ユーザー登録/ログインとセッション一覧・詳細の表示を行う。

## 技術スタック

| カテゴリ | 技術 | 理由 |
| -------- | ---- | ---- |
| ビルドツール | Vite | 高速なHMR、ESM対応 |
| UIライブラリ | React 18 | コンポーネントベース |
| 言語 | TypeScript | 型安全性 |
| スタイリング | Tailwind CSS | ユーティリティファースト |
| ルーティング | React Router v6 | 標準的なルーティング |
| 状態管理/データ取得 | TanStack Query (React Query) | サーバー状態管理、キャッシュ |
| フォーム | React Hook Form | パフォーマンス、バリデーション |
| 日時処理 | date-fns | 軽量、Tree-shaking対応 |
| アイコン | Lucide React | 軽量、一貫性のあるアイコン |
| コード表示 | react-syntax-highlighter | シンタックスハイライト |

## ディレクトリ構成

```text
web/
├── src/
│   ├── main.tsx                  # エントリーポイント
│   ├── App.tsx                   # ルーティング
│   ├── index.css                 # Tailwind directives
│   │
│   ├── api/                      # API クライアント
│   │   ├── client.ts             # fetch ラッパー
│   │   ├── auth.ts               # 認証 API
│   │   ├── sessions.ts           # セッション API
│   │   └── keys.ts               # APIキー API
│   │
│   ├── hooks/                    # カスタム Hooks
│   │   └── useAuth.ts            # 認証状態管理
│   │
│   ├── pages/                    # ページコンポーネント
│   │   ├── WelcomePage.tsx       # 初期画面（登録/ログイン選択）
│   │   ├── RegisterPage.tsx      # ユーザー登録
│   │   ├── LoginPage.tsx         # ログイン
│   │   ├── SessionListPage.tsx   # セッション一覧
│   │   ├── SessionDetailPage.tsx # セッション詳細
│   │   └── SettingsPage.tsx      # 設定（APIキー管理）
│   │
│   ├── components/               # 共通コンポーネント
│   │   ├── layout/
│   │   │   ├── Layout.tsx        # 共通レイアウト
│   │   │   ├── Header.tsx        # ヘッダー
│   │   │   └── UserMenu.tsx      # ユーザーメニュー
│   │   │
│   │   ├── sessions/
│   │   │   ├── SessionCard.tsx   # セッションカード
│   │   │   └── SessionList.tsx   # セッション一覧
│   │   │
│   │   ├── timeline/
│   │   │   ├── Timeline.tsx      # タイムライン
│   │   │   ├── EventCard.tsx     # イベントカード
│   │   │   ├── UserMessage.tsx   # ユーザーメッセージ
│   │   │   ├── AssistantMessage.tsx # アシスタントメッセージ
│   │   │   └── ToolUse.tsx       # ツール使用表示
│   │   │
│   │   ├── settings/
│   │   │   ├── ApiKeyList.tsx    # APIキー一覧
│   │   │   └── ApiKeyForm.tsx    # 新規キー作成フォーム
│   │   │
│   │   └── ui/                   # 汎用UIコンポーネント
│   │       ├── Button.tsx
│   │       ├── Input.tsx
│   │       ├── Card.tsx
│   │       ├── Modal.tsx
│   │       ├── Spinner.tsx
│   │       └── CopyButton.tsx
│   │
│   ├── lib/                      # ユーティリティ
│   │   ├── utils.ts              # 共通ユーティリティ
│   │   └── cn.ts                 # clsx + tailwind-merge
│   │
│   └── types/                    # 型定義
│       ├── auth.ts
│       ├── session.ts
│       └── event.ts
│
├── index.html
├── package.json
├── vite.config.ts
├── tailwind.config.ts
├── postcss.config.js
└── tsconfig.json
```

## 依存パッケージ

```json
{
  "dependencies": {
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "react-router-dom": "^6.28.0",
    "@tanstack/react-query": "^5.60.0",
    "react-hook-form": "^7.54.0",
    "date-fns": "^4.1.0",
    "lucide-react": "^0.460.0",
    "react-syntax-highlighter": "^15.6.1",
    "clsx": "^2.1.0",
    "tailwind-merge": "^2.6.0"
  },
  "devDependencies": {
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "@types/react-syntax-highlighter": "^15.5.0",
    "@vitejs/plugin-react": "^4.3.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.6.0",
    "vite": "^6.0.0"
  }
}
```

## Tailwind 設定

```ts
// tailwind.config.ts
import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // カスタムカラーパレット
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Menlo', 'monospace'],
      },
    },
  },
  plugins: [],
} satisfies Config
```

## 画面構成

### 1. 初期画面（WelcomePage）

初回アクセス時に表示。登録済みならダッシュボードへリダイレクト。

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                        ◇ Agentrace                          │
│                                                             │
│          Track and review Claude Code sessions              │
│                    with your team.                          │
│                                                             │
│    ┌─────────────────┐    ┌─────────────────────────┐      │
│    │    Register     │    │   Login with API Key    │      │
│    └─────────────────┘    └─────────────────────────┘      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2. ユーザー登録（RegisterPage）

名前を入力するだけでAPIキーが発行される。

**入力画面:**
```
┌─────────────────────────────────────────────────────────────┐
│  ← Back                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                     Create Account                          │
│                                                             │
│     Your Name                                               │
│     ┌───────────────────────────────────────────────┐      │
│     │                                               │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
│     ┌───────────────────────────────────────────────┐      │
│     │              Create Account                   │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**成功画面:**
```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                  ✓ Account Created!                         │
│                                                             │
│     Your API Key                                            │
│     ┌───────────────────────────────────────────────┐      │
│     │ agtr_xxxxxxxxxxxxxxxxxxxxxxxx          [📋]  │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
│     ⚠ Save this key - it won't be shown again.             │
│                                                             │
│     ────────────────────────────────────────────────       │
│                                                             │
│     Set up CLI:                                             │
│     $ npx agentrace init                                    │
│                                                             │
│     ┌───────────────────────────────────────────────┐      │
│     │              Go to Dashboard                  │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3. ログイン（LoginPage）

APIキーを入力してログイン。

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                         Login                               │
│                                                             │
│     API Key                                                 │
│     ┌───────────────────────────────────────────────┐      │
│     │ agtr_                                         │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
│     ┌───────────────────────────────────────────────┐      │
│     │                   Login                       │      │
│     └───────────────────────────────────────────────┘      │
│                                                             │
│     Don't have an account? Register                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 4. セッション一覧（SessionListPage）

全ユーザーのセッションを表示。

```
┌─────────────────────────────────────────────────────────────┐
│  ◇ Agentrace                           Taro ▼    Settings  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Sessions                                                   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  📁 /path/to/project                                │   │
│  │  👤 Taro  •  🕐 2025-12-28 10:30  •  42 events      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  📁 /path/to/another-project                        │   │
│  │  👤 Hanako  •  🕐 2025-12-28 09:15  •  28 events    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  📁 /path/to/third-project                          │   │
│  │  👤 Taro  •  🕐 2025-12-27 15:45  •  156 events     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 5. セッション詳細（SessionDetailPage）

タイムライン形式でイベント表示。

```
┌─────────────────────────────────────────────────────────────┐
│  ← Sessions                                                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  📁 /path/to/project                                        │
│  👤 Taro  •  Started 2025-12-28 10:30                       │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  Timeline                                                   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 👤 User                               10:30:05      │   │
│  │                                                     │   │
│  │ Add a function to calculate fibonacci numbers       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🤖 Assistant                          10:30:10      │   │
│  │                                                     │   │
│  │ I'll create a fibonacci function for you...         │   │
│  │                                                     │   │
│  │ ┌───────────────────────────────────────────────┐   │   │
│  │ │ function fibonacci(n: number): number {       │   │   │
│  │ │   if (n <= 1) return n;                       │   │   │
│  │ │   return fibonacci(n - 1) + fibonacci(n - 2); │   │   │
│  │ │ }                                             │   │   │
│  │ └───────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🔧 Tool: Write                        10:30:15      │   │
│  │                                                     │   │
│  │ file_path: /src/utils/math.ts                       │   │
│  │ ▼ Show content                                      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 6. 設定（SettingsPage）

APIキーの管理画面。

```
┌─────────────────────────────────────────────────────────────┐
│  ◇ Agentrace                           Taro ▼    Settings  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Settings                                                   │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  API Keys                                                   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  💻 MacBook Pro                                     │   │
│  │  agtr_xxxx...  •  Last used: 1 hour ago             │   │
│  │                                            [Delete] │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  💻 Work PC                                         │   │
│  │  agtr_yyyy...  •  Last used: 3 days ago             │   │
│  │                                            [Delete] │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Create New API Key                                         │
│                                                             │
│  Name                                                       │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ e.g. Work Laptop                                      │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                 Create API Key                        │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## コンポーネント設計

### ユーティリティ関数

```ts
// src/lib/cn.ts
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

### 汎用UIコンポーネント

```tsx
// src/components/ui/Button.tsx
import { cn } from '@/lib/cn'
import { Loader2 } from 'lucide-react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
}

export function Button({
  className,
  variant = 'primary',
  size = 'md',
  loading,
  disabled,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-lg font-medium transition-colors',
        'focus:outline-none focus:ring-2 focus:ring-offset-2',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        {
          'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500':
            variant === 'primary',
          'bg-gray-100 text-gray-900 hover:bg-gray-200 focus:ring-gray-500':
            variant === 'secondary',
          'text-gray-600 hover:text-gray-900 hover:bg-gray-100':
            variant === 'ghost',
          'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500':
            variant === 'danger',
        },
        {
          'px-3 py-1.5 text-sm': size === 'sm',
          'px-4 py-2 text-sm': size === 'md',
          'px-6 py-3 text-base': size === 'lg',
        },
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
      {children}
    </button>
  )
}
```

```tsx
// src/components/ui/Input.tsx
import { cn } from '@/lib/cn'
import { forwardRef } from 'react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label className="block text-sm font-medium text-gray-700">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={cn(
            'block w-full rounded-lg border border-gray-300 px-4 py-2',
            'text-gray-900 placeholder:text-gray-400',
            'focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20',
            'disabled:bg-gray-50 disabled:text-gray-500',
            error && 'border-red-500 focus:border-red-500 focus:ring-red-500/20',
            className
          )}
          {...props}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
      </div>
    )
  }
)
```

```tsx
// src/components/ui/Card.tsx
import { cn } from '@/lib/cn'

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  hover?: boolean
}

export function Card({ className, hover, ...props }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-xl border border-gray-200 bg-white p-4 shadow-sm',
        hover && 'cursor-pointer transition-shadow hover:shadow-md',
        className
      )}
      {...props}
    />
  )
}
```

### セッションコンポーネント

```tsx
// src/components/sessions/SessionCard.tsx
import { Card } from '@/components/ui/Card'
import { Folder, User, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { ja } from 'date-fns/locale'
import type { Session } from '@/types/session'

interface SessionCardProps {
  session: Session
  onClick: () => void
}

export function SessionCard({ session, onClick }: SessionCardProps) {
  return (
    <Card hover onClick={onClick}>
      <div className="flex items-start gap-3">
        <Folder className="mt-0.5 h-5 w-5 text-gray-400" />
        <div className="flex-1 min-w-0">
          <p className="font-mono text-sm text-gray-900 truncate">
            {session.projectPath}
          </p>
          <div className="mt-1 flex items-center gap-4 text-sm text-gray-500">
            <span className="flex items-center gap-1">
              <User className="h-4 w-4" />
              {session.userName}
            </span>
            <span className="flex items-center gap-1">
              <Clock className="h-4 w-4" />
              {formatDistanceToNow(new Date(session.startedAt), {
                addSuffix: true,
                locale: ja,
              })}
            </span>
            <span>{session.eventCount} events</span>
          </div>
        </div>
      </div>
    </Card>
  )
}
```

### タイムラインコンポーネント

#### タイムスタンプの取り扱い

イベントの時刻は `payload.timestamp` を優先して使用する（Claude Codeが記録したオリジナルのタイムスタンプ）。
`created_at` はサーバーでの保存時刻なのでフォールバックとしてのみ使用。

- **サーバー側**: イベントを `payload.timestamp` で昇順ソート（会話順）
- **フロントエンド**: EventCardで `payload.timestamp` を表示

```tsx
// src/components/timeline/EventCard.tsx
import { cn } from '@/lib/cn'
import { User, Bot, Wrench, ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import type { Event } from '@/types/event'
import { UserMessage } from './UserMessage'
import { AssistantMessage } from './AssistantMessage'
import { ToolUse } from './ToolUse'

interface EventCardProps {
  event: Event
}

export function EventCard({ event }: EventCardProps) {
  const [expanded, setExpanded] = useState(true)

  // payload.timestamp を優先、なければ created_at
  const timestamp = (event.payload?.timestamp as string) || event.created_at

  const icon = {
    user: <User className="h-4 w-4" />,
    assistant: <Bot className="h-4 w-4" />,
    tool_use: <Wrench className="h-4 w-4" />,
    tool_result: <Wrench className="h-4 w-4" />,
  }[event.eventType] || null

  const label = {
    user: 'User',
    assistant: 'Assistant',
    tool_use: `Tool: ${event.payload?.name || 'Unknown'}`,
    tool_result: 'Tool Result',
  }[event.eventType] || event.eventType

  return (
    <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
      <button
        className={cn(
          'w-full flex items-center justify-between px-4 py-3',
          'text-left hover:bg-gray-50 transition-colors'
        )}
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          <span className={cn(
            'flex items-center justify-center w-6 h-6 rounded-full',
            event.eventType === 'user' && 'bg-blue-100 text-blue-600',
            event.eventType === 'assistant' && 'bg-green-100 text-green-600',
            (event.eventType === 'tool_use' || event.eventType === 'tool_result')
              && 'bg-orange-100 text-orange-600'
          )}>
            {icon}
          </span>
          <span className="font-medium text-gray-900">{label}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span>{format(new Date(timestamp), 'HH:mm:ss')}</span>
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </div>
      </button>

      {expanded && (
        <div className="px-4 pb-4 border-t border-gray-100">
          {event.eventType === 'user' && <UserMessage payload={event.payload} />}
          {event.eventType === 'assistant' && <AssistantMessage payload={event.payload} />}
          {(event.eventType === 'tool_use' || event.eventType === 'tool_result') && (
            <ToolUse payload={event.payload} isResult={event.eventType === 'tool_result'} />
          )}
        </div>
      )}
    </div>
  )
}
```

```tsx
// src/components/timeline/AssistantMessage.tsx
import { useState } from 'react'
import { ChevronDown, ChevronRight, Brain } from 'lucide-react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface AssistantMessageProps {
  payload: Record<string, unknown>
}

// Claude Codeの "thinking" ブロックを折りたたみ可能なUIで表示
function ThinkingBlock({ thinking }: { thinking: string }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="rounded-lg border border-purple-200 bg-purple-50">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-purple-700 hover:bg-purple-100"
      >
        <Brain className="h-4 w-4" />
        <span>Thinking</span>
        {expanded ? (
          <ChevronDown className="ml-auto h-4 w-4" />
        ) : (
          <ChevronRight className="ml-auto h-4 w-4" />
        )}
      </button>
      {expanded && (
        <div className="border-t border-purple-200 px-3 py-2">
          <p className="whitespace-pre-wrap text-sm text-purple-900">
            {thinking}
          </p>
        </div>
      )}
    </div>
  )
}

export function AssistantMessage({ payload }: AssistantMessageProps) {
  const message = payload?.message as Record<string, unknown> | undefined
  const content = message?.content

  if (!content) {
    return (
      <pre className="mt-3 text-sm text-gray-600 whitespace-pre-wrap">
        {JSON.stringify(payload, null, 2)}
      </pre>
    )
  }

  // contentが配列の場合（テキスト+コードブロック+thinkingブロック等）
  if (Array.isArray(content)) {
    return (
      <div className="mt-3 space-y-3">
        {content.map((block, i) => {
          // テキストブロック
          if (block.type === 'text') {
            return (
              <p key={i} className="text-gray-700 whitespace-pre-wrap">
                {block.text}
              </p>
            )
          }
          // 思考ブロック（Claude Codeのinterleaved thinking）
          if (block.type === 'thinking' && typeof block.thinking === 'string') {
            return <ThinkingBlock key={i} thinking={block.thinking} />
          }
          // ツール使用ブロック
          if (block.type === 'tool_use') {
            return (
              <div key={i} className="rounded-lg bg-gray-50 p-3">
                <p className="text-sm font-medium text-gray-600 mb-2">
                  Tool: {block.name}
                </p>
                <SyntaxHighlighter
                  language="json"
                  style={oneLight}
                  customStyle={{ fontSize: '0.875rem', borderRadius: '0.5rem' }}
                >
                  {JSON.stringify(block.input, null, 2)}
                </SyntaxHighlighter>
              </div>
            )
          }
          // ツール結果ブロック
          if (block.type === 'tool_result') {
            return (
              <div key={i} className="rounded-lg bg-gray-50 p-3">
                <p className="text-sm font-medium text-gray-600 mb-2">Tool Result</p>
                <pre className="whitespace-pre-wrap text-sm text-gray-700">
                  {typeof block.content === 'string'
                    ? block.content
                    : JSON.stringify(block.content, null, 2)}
                </pre>
              </div>
            )
          }
          // 未知のブロックタイプはJSONで表示
          if (block.type) {
            return (
              <div key={i} className="rounded-lg bg-gray-100 p-3">
                <p className="mb-2 text-xs font-medium text-gray-500">{block.type}</p>
                <pre className="whitespace-pre-wrap text-sm text-gray-600">
                  {JSON.stringify(block, null, 2)}
                </pre>
              </div>
            )
          }
          return null
        })}
      </div>
    )
  }

  return (
    <p className="mt-3 text-gray-700 whitespace-pre-wrap">{String(content)}</p>
  )
}
```

## API クライアント

```ts
// src/api/client.ts
const BASE_URL = import.meta.env.VITE_API_URL || ''

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

export async function fetchAPI<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const message = await res.text().catch(() => 'Unknown error')
    throw new ApiError(res.status, message)
  }

  // No content
  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}
```

```ts
// src/api/auth.ts
import { fetchAPI } from './client'
import type { User } from '@/types/auth'

export async function register(name: string): Promise<{ user: User; api_key: string }> {
  return fetchAPI('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function login(apiKey: string): Promise<{ user: User }> {
  return fetchAPI('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ api_key: apiKey }),
  })
}

export async function logout(): Promise<void> {
  return fetchAPI('/api/auth/logout', { method: 'POST' })
}

export async function getMe(): Promise<User> {
  return fetchAPI('/api/me')
}

export async function getUsers(): Promise<{ users: User[] }> {
  return fetchAPI('/api/users')
}
```

```ts
// src/api/sessions.ts
import { fetchAPI } from './client'
import type { Session, SessionDetail } from '@/types/session'

export async function getSessions(): Promise<{ sessions: Session[] }> {
  return fetchAPI('/api/sessions')
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`)
}
```

```ts
// src/api/keys.ts
import { fetchAPI } from './client'
import type { ApiKey } from '@/types/auth'

export async function getKeys(): Promise<{ keys: ApiKey[] }> {
  return fetchAPI('/api/keys')
}

export async function createKey(name: string): Promise<{ key: ApiKey; api_key: string }> {
  return fetchAPI('/api/keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function deleteKey(id: string): Promise<void> {
  return fetchAPI(`/api/keys/${id}`, { method: 'DELETE' })
}
```

## 型定義

```ts
// src/types/auth.ts
export interface User {
  id: string
  name: string
  created_at: string
}

export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  last_used_at: string | null
  created_at: string
}
```

```ts
// src/types/session.ts
export interface Session {
  id: string
  user_id: string | null
  user_name: string | null
  claude_session_id: string
  project_path: string
  started_at: string
  ended_at: string | null
  event_count: number
}

export interface SessionDetail extends Session {
  events: Event[]
}
```

```ts
// src/types/event.ts
export interface Event {
  id: string
  session_id: string
  event_type: 'user' | 'assistant' | 'tool_use' | 'tool_result' | string
  payload: Record<string, unknown>
  created_at: string
}
```

## TanStack Query 設定

```tsx
// src/main.tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000, // 30秒
      retry: 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>
)
```

```tsx
// src/hooks/useAuth.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import * as authApi from '@/api/auth'

export function useAuth() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { data: user, isLoading, error } = useQuery({
    queryKey: ['me'],
    queryFn: authApi.getMe,
    retry: false,
  })

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
      navigate('/')
    },
  })

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      queryClient.clear()
      navigate('/welcome')
    },
  })

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    login: loginMutation.mutate,
    logout: logoutMutation.mutate,
  }
}
```

## ルーティング

```tsx
// src/App.tsx
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { Layout } from '@/components/layout/Layout'
import { WelcomePage } from '@/pages/WelcomePage'
import { RegisterPage } from '@/pages/RegisterPage'
import { LoginPage } from '@/pages/LoginPage'
import { SessionListPage } from '@/pages/SessionListPage'
import { SessionDetailPage } from '@/pages/SessionDetailPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { Spinner } from '@/components/ui/Spinner'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/welcome" replace />
  }

  return <>{children}</>
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner />
      </div>
    )
  }

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route
        path="/welcome"
        element={
          <PublicRoute>
            <WelcomePage />
          </PublicRoute>
        }
      />
      <Route
        path="/register"
        element={
          <PublicRoute>
            <RegisterPage />
          </PublicRoute>
        }
      />
      <Route
        path="/login"
        element={
          <PublicRoute>
            <LoginPage />
          </PublicRoute>
        }
      />

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<SessionListPage />} />
        <Route path="sessions/:id" element={<SessionDetailPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
```

## 環境変数

| 変数名 | 説明 | デフォルト |
| ------ | ---- | ---------- |
| `VITE_API_URL` | バックエンドAPI URL | '' (同一オリジン) |

## Vite 設定

```ts
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/auth': 'http://localhost:8080',
    },
  },
})
```

## 実装順序

### Step 3: 基本UI ✅ 完了

1. **プロジェクトセットアップ** ✅
   - Vite + React + TypeScript 初期化
   - Tailwind CSS 設定
   - ディレクトリ構造作成
   - パスエイリアス設定

2. **汎用コンポーネント** ✅
   - Button, Input, Card, Modal, Spinner, CopyButton
   - cn() ユーティリティ（clsx + tailwind-merge）
   - Layout, Header コンポーネント

3. **認証ページ** ✅
   - WelcomePage（初期画面）
   - RegisterPage（ユーザー登録、APIキー表示）
   - LoginPage（APIキーでログイン）
   - useAuth フック（TanStack Query）

4. **セッション機能** ✅
   - SessionListPage（一覧）- StartedAt降順でソート
   - SessionCard コンポーネント
   - SessionDetailPage（詳細）
   - Timeline, EventCard コンポーネント

5. **設定ページ** ✅
   - SettingsPage
   - ApiKeyList, ApiKeyForm コンポーネント

6. **メッセージ表示** ✅
   - UserMessage: ユーザー入力の表示
   - AssistantMessage: アシスタント応答の表示
     - テキストブロック
     - Thinkingブロック（折りたたみ可能、紫色のUI）
     - ツール使用ブロック
     - ツール結果ブロック
     - 未知のブロックタイプ（JSONで表示）
   - ToolUse: ツール使用/結果の表示

7. **ソートとタイムスタンプ** ✅
   - セッション一覧: StartedAt降順（新しい順）
   - イベント一覧: payload.timestamp昇順（会話順）
   - 時刻表示: payload.timestampを優先（created_atフォールバック）
