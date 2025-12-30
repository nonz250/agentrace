# CLI コマンド

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `agentrace init --url <url>` | ブラウザ連携で設定 + hooks インストール |
| `agentrace init --url <url> --dev` | 開発モード（ローカルCLIパス使用） |
| `agentrace login` | WebログインURL発行 |
| `agentrace send` | transcript差分送信（hooks用） |
| `agentrace on` | hooks有効化（認証情報は保持） |
| `agentrace on --dev` | hooks有効化（開発モード） |
| `agentrace off` | hooks無効化（認証情報は保持） |
| `agentrace uninstall` | hooks/config 削除 |

## init コマンド

サーバーの初期設定、API キー取得、hooks インストールを行う。

```bash
npx agentrace init --url http://server:8080
```

### 処理フロー

```
1. トークン生成 → callbackサーバー起動
       ↓
2. ブラウザで /setup?token=xxx&callback=... を開く
       ↓
3. Web側でユーザー登録/ログイン → APIキー生成
       ↓
4. callback URLにAPIキーをPOST
       ↓
5. config保存 → hooks インストール
```

### セキュリティ

- トークンは `crypto.randomUUID()` で生成（推測困難）
- コールバックURLは `localhost` のみ許可
- タイムアウト5分

## send コマンド

Claude Code Stop hook から呼ばれて、transcript 差分をサーバーに送信。

```bash
npx agentrace send
# stdin から JSON を受け取る
```

### 処理フロー

```
1. stdin JSON 読み取り
       ↓ { session_id, transcript_path, cwd }
2. transcript差分抽出
       ↓ getCursor() で前回位置取得
       ↓ JSONL を行ごとにパース
3. Git情報取得（初回のみ）
       ↓ CLAUDE_PROJECT_DIR 環境変数優先
       ↓ git remote -url, git branch --show
4. POST /api/ingest
       ↓ Bearer: config.api_key
       ↓ body: { session_id, transcript_lines, git情報 }
5. カーソル更新
       ↓ saveCursor(sessionId, totalLineCount)
```

### エラーハンドリング

すべてのエラーで `exit(0)` を返す（hooks をブロックしない設計）:
- config 未設定
- stdin 読み取り失敗
- JSON parse 失敗
- API エラー

## login コマンド

Web ダッシュボードへのログイン URL を発行し、ブラウザで開く。

```bash
npx agentrace login
```

### 処理フロー

```
1. POST /api/auth/web-session（Bearer認証）
       ↓
2. 短期トークン（10分）を含むURLを生成
       ↓
3. URLをコンソールに表示
       ↓
4. Enter キーでブラウザ起動
```

## on / off コマンド

hooks の有効化/無効化。config は保持したまま。

```bash
# hooks 有効化
npx agentrace on

# 開発モードで hooks 有効化
npx agentrace on --dev

# hooks 無効化
npx agentrace off
```

## uninstall コマンド

hooks と config を完全に削除。

```bash
npx agentrace uninstall
```

## 開発モード

`--dev` オプションを付けると、hooks コマンドが変わる:

| モード | hooks コマンド |
|--------|---------------|
| 本番 | `npx agentrace send` |
| 開発 | `npx tsx /path/to/cli/src/index.ts send` |

開発時はビルドを待たずに TypeScript を直接実行できる。
