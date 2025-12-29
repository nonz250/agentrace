# Agentrace 機能追加計画

## 概要

初期設定体験の改善、認証機能の強化、CLIコマンドの追加を行う。

## 変更内容

### 1. 初期設定体験の改善

**目標**: `npx agentrace init` の体験を簡略化し、ブラウザ経由でスムーズにセットアップできるようにする

#### 1.1 URLパラメータ対応

```bash
# 現在
npx agentrace init
# Server URL: http://localhost:8080  ← 手動入力
# API Key: agtr_xxxx  ← 手動入力

# 変更後
npx agentrace init url=http://localhost:8080
# ブラウザが開く → 登録/ログイン → 自動的にAPIキー取得
```

#### 1.2 ブラウザ連携フロー

```text
┌─────────────────────────────────────────────────────────────┐
│ CLI: npx agentrace init url=http://server:8080              │
├─────────────────────────────────────────────────────────────┤
│  1. CLIがワンタイムトークンを生成                            │
│  2. ブラウザで http://server:8080/setup?token=xxx を開く     │
│  3. CLIはローカルでHTTPサーバーを起動して待機                │
│     (例: http://localhost:19283/callback)                   │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Web: /setup?token=xxx                                       │
├─────────────────────────────────────────────────────────────┤
│  未ログインの場合:                                          │
│    → 登録/ログイン画面を表示                                │
│    → 登録/ログイン完了後、セットアップ画面へ                │
│                                                             │
│  ログイン済みの場合:                                        │
│    → セットアップ画面を直接表示                             │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ Web: セットアップ完了画面                                    │
├─────────────────────────────────────────────────────────────┤
│  1. 新しいAPIキーを自動生成（名前: "CLI Setup - hostname"）  │
│  2. CLIのコールバックURLにAPIキーをPOST                      │
│     POST http://localhost:19283/callback                    │
│     { "api_key": "agtr_xxxxx" }                             │
│  3. 画面に「Setup complete! You can close this tab.」表示   │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ CLI: コールバック受信                                        │
├─────────────────────────────────────────────────────────────┤
│  1. APIキーを受信                                           │
│  2. ~/.agentrace/config.json に保存                         │
│  3. ~/.claude/settings.json に hooks 追加                   │
│  4. 「✓ Setup complete!」を表示                             │
└─────────────────────────────────────────────────────────────┘
```

#### 1.3 実装タスク

**CLI側:**
- [ ] `init` コマンドに `url=` パラメータ追加
- [ ] ローカルHTTPサーバー起動（コールバック受信用）
- [ ] ブラウザを自動で開く
- [ ] コールバック受信後の設定保存処理
- [ ] タイムアウト処理（5分程度）

**Server側:**
- [ ] GET `/setup` - セットアップページ（CLIからのリダイレクト先）
- [ ] POST `/api/setup/complete` - セットアップ完了（APIキー生成 + コールバック）

**Web側:**
- [ ] SetupPage コンポーネント作成
- [ ] 未ログイン時のリダイレクト処理
- [ ] セットアップ完了後のコールバック実行

---

### 2. 認証機能の強化

**目標**: パスワード認証をデフォルトにし、GitHub OAuth にも対応する

#### 2.1 パスワード認証

**変更点:**
- ユーザー登録時にパスワードを必須にする
- ログイン時にメールアドレス + パスワードで認証
- 認証情報はユーザーテーブルとは別テーブルで管理（将来のOAuth対応のため）

**データモデル変更:**

```go
// domain/user.go
type User struct {
    ID          string
    Email       string
    DisplayName string     // オプション（空の場合はEmailを表示）
    CreatedAt   time.Time
}

func (u *User) GetDisplayName() string {
    if u.DisplayName != "" {
        return u.DisplayName
    }
    return u.Email
}

// domain/password_credential.go
type PasswordCredential struct {
    ID           string
    UserID       string
    PasswordHash string     // bcrypt hash
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**API変更:**

| Method | Path | 変更内容 |
| ------ | ---- | -------- |
| POST | `/auth/register` | `email`, `password` パラメータのみ（nameは不要） |
| POST | `/auth/login` | `email`, `password` での認証に変更（APIキー認証は `/auth/login/apikey` に移動） |

#### 2.2 GitHub OAuth

**環境変数:**

| 変数名 | 説明 |
| ------ | ---- |
| `GITHUB_CLIENT_ID` | GitHub App Client ID |
| `GITHUB_CLIENT_SECRET` | GitHub App Client Secret |

**API追加:**

| Method | Path | 説明 |
| ------ | ---- | ---- |
| GET | `/auth/github` | GitHub認証開始（リダイレクト） |
| GET | `/auth/github/callback` | GitHub認証コールバック |

**データモデル:**

```go
// domain/oauth_connection.go
type OAuthConnection struct {
    ID         string
    UserID     string
    Provider   string  // "github", "google"
    ProviderID string  // GitHubのユーザーID
    CreatedAt  time.Time
}
```

**フロー:**

```text
1. ユーザーが「Login with GitHub」をクリック
2. GET /auth/github → GitHub認証画面にリダイレクト
3. GitHub認証完了 → GET /auth/github/callback
4. GitHubユーザー情報を取得
5. 既存ユーザーならログイン、新規ならアカウント作成
6. セッションCookie発行、ダッシュボードへリダイレクト
```

#### 2.3 実装タスク

**パスワード認証:** ✅ 完了
- [x] User モデル変更（name → display_name、email追加）
- [x] PasswordCredential モデル追加（password_credentials テーブル）
- [x] 各Repository（memory, sqlite, postgres, mongodb）の更新
- [x] Migration ファイル更新（users, password_credentials テーブル）
- [x] POST `/auth/register` の変更（email + password のみ）
- [x] POST `/auth/login` の変更
- [x] POST `/auth/login/apikey` 追加（旧APIキーログイン）
- [x] Web: RegisterPage（email + password のみ、name入力なし）
- [x] Web: LoginPage をメール/パスワード認証に変更
- [x] Web: Header に display_name || email を表示

**GitHub OAuth:**
- [ ] OAuthConnection ドメインモデル追加
- [ ] OAuthConnectionRepository 追加
- [ ] GET `/auth/github` ハンドラ
- [ ] GET `/auth/github/callback` ハンドラ
- [ ] Web: LoginPage に「Login with GitHub」ボタン追加
- [ ] Web: RegisterPage に「Sign up with GitHub」ボタン追加

**将来対応（スコープ外）:**
- Google OAuth

---

### 3. CLIコマンド追加

**目標**: hooksの有効/無効を切り替えられるようにする

#### 3.1 on/off コマンド

```bash
# hooks無効化（認証情報は保持）
npx agentrace off
# ✓ Hooks disabled. Your credentials are still saved.
# Run 'npx agentrace on' to re-enable.

# hooks有効化
npx agentrace on
# ✓ Hooks enabled. Session data will be sent to http://localhost:8080
```

**動作:**

| コマンド | hooksの状態 | config.jsonの状態 |
| -------- | ----------- | ----------------- |
| `off` | 削除 | 保持 |
| `on` | 追加 | 保持 |
| `uninstall` | 削除 | 削除 |

#### 3.2 実装タスク ✅ 完了

- [x] `on` コマンド追加
- [x] `off` コマンド追加
- [x] 既存の `uninstallHooks()` 関数を活用（分離不要だった）

---

## 実装順序

### Phase 1: パスワード認証 ✅ 完了

1. User モデル変更（name → display_name、email追加）
2. PasswordCredential モデル追加（パスワードは別テーブルで管理）
3. Repository更新（全DB: memory, sqlite, postgres, mongodb）
4. Migration ファイル更新
5. API変更（register: email+password のみ、login）
6. Web UI更新（登録/ログインフォーム、Header表示）

### Phase 2: CLI on/off コマンド ✅ 完了

1. `off` コマンド実装（cli/src/commands/off.ts）
2. `on` コマンド実装（cli/src/commands/on.ts）
3. index.ts にコマンド登録

### Phase 3: 初期設定体験改善

1. CLI: ローカルサーバー + ブラウザ連携
2. Server: `/setup` エンドポイント
3. Web: SetupPage実装

### Phase 4: GitHub OAuth

1. OAuthConnection モデル追加
2. GitHub OAuth API実装
3. Web: ソーシャルログインボタン追加

---

## 互換性への配慮

- 既存ユーザーはパスワード未設定でも動作継続
- APIキーでのログインは `/auth/login/apikey` で引き続きサポート
- 既存の `init` コマンド（手動入力）も引き続きサポート
