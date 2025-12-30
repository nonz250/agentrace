# 認証フロー

## 認証方式

| 方式 | 用途 | 有効期間 |
|------|------|---------|
| Bearer 認証 | CLI → Server | 無期限（APIキー） |
| Session 認証 | Web → Server | 7日間（Cookie） |

## Bearer 認証（CLI用）

```
リクエスト:
  Authorization: Bearer agtr_xxxxxxxx

サーバー処理:
  1. APIKeyをbcryptでハッシュ化
  2. DB検索で一致するAPIKeyを探す
  3. 一致すればUserIDを取得
  4. コンテキストにUserを設定
```

## Session 認証（Web用）

```
リクエスト:
  Cookie: session=xxxxx

サーバー処理:
  1. WebSessionテーブルからトークンで検索
  2. 有効期限チェック
  3. UserIDを取得
  4. コンテキストにUserを設定
```

## ユーザー登録（Web）

1. ブラウザで http://server:8080 にアクセス
2. 「Register」→ email + password 入力
3. パスワード: 8文字以上
4. APIキー自動発行（この1回のみ表示）

## CLIセットアップ（ブラウザ連携）

```bash
npx agentrace init --url http://server:8080
```

1. CLIがワンタイムトークンを生成し、ローカルHTTPサーバー起動
2. ブラウザで `/setup?token=xxx&callback=http://localhost:xxxxx/callback` を開く
3. 未ログインなら登録/ログイン画面を経由
4. セットアップ画面で「Setup CLI」ボタン押下
5. WebがAPIキーを生成し、CLIのコールバックURLにPOST
6. CLIがAPIキーを受信、config保存、hooks追加

### セキュリティ

- トークンは `crypto.randomUUID()` で生成（推測困難）
- コールバックURLは `localhost` のみ許可
- タイムアウト5分

## Webログイン

- 方法1: `npx agentrace login` → URL発行 → ブラウザで開く
- 方法2: Webでemail + passwordを入力してログイン
- 方法3: GitHub OAuthでログイン（`GITHUB_CLIENT_ID/SECRET` 設定時のみ）

## GitHub OAuth

環境変数 `GITHUB_CLIENT_ID` と `GITHUB_CLIENT_SECRET` が設定されている場合のみ有効。

```
1. ユーザーが「Continue with GitHub」をクリック
2. GET /auth/github → GitHub認証画面にリダイレクト
3. GitHub認証完了 → GET /auth/github/callback
4. GitHubユーザー情報を取得
5. 既存ユーザー（email一致 or OAuth連携済み）ならログイン
   新規ならアカウント作成
6. セッションCookie発行、ダッシュボードへリダイレクト
```

OAuthConnectionテーブルでGitHubのユーザーIDとローカルユーザーを紐付け管理。

## 複数APIキー

- 各ユーザーは複数のAPIキーを発行可能（別デバイス用など）
- Webの設定画面（/settings）でAPIキーの管理
- キー発行時に名前を付けられる（例: "MacBook Pro", "Work PC"）
