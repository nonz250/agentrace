# Agentrace スクリプト

GitHub Pages 用の静的デモサイトを生成するためのスクリプト。

## セットアップ

```bash
cd scripts/generate-static-site
npm install
```

## 静的サイト生成 (`generate-static-site.ts`)

起動中のサーバーとWebアプリから各ページをクロールし、静的HTMLを生成します。

```bash
# サーバーとWebアプリを起動した状態で実行
npm run generate-static
```

**前提条件:**
1. サーバーが起動していること
2. Webアプリが起動していること
3. 表示したいデータがDBに存在すること

**環境変数:**

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `AGENTRACE_SERVER_URL` | サーバーURL | `http://localhost:8080` |
| `AGENTRACE_WEB_URL` | WebアプリURL | `http://localhost:5173` |
| `GITHUB_PAGES_BASE` | GitHub Pagesのベースパス | `/agentrace` |

## 静的サイト生成の流れ

1. **サーバーを起動**
   ```bash
   cd server
   DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./demo.db WEB_URL=http://localhost:5173 go run ./cmd/server
   ```

2. **Webアプリを起動**
   ```bash
   cd web
   npm install && npm run dev
   ```

3. **静的サイトを生成**
   ```bash
   cd scripts/generate-static-site
   npm install
   npm run generate-static
   ```

4. **生成されたサイトを確認**
   ```bash
   cd ../..
   npx serve docs
   # http://localhost:3000/agentrace/ でプレビュー
   ```

## 出力ファイル

生成されたファイルは `docs/` ディレクトリに保存されます：

```
docs/
├── index.html                      # トップページ
├── 404.html                        # SPAルーティング用
├── .nojekyll                       # GitHub Pages用
├── api/                            # 静的APIレスポンス
│   ├── projects.json
│   ├── sessions.json
│   ├── plans.json
│   ├── projects/
│   │   ├── {id}.json
│   │   ├── {id}-sessions.json
│   │   └── {id}-plans.json
│   ├── sessions/
│   │   └── {id}.json
│   └── plans/
│       ├── {id}.json
│       └── {id}-events.json
├── assets/                         # CSS/JS
└── projects/
    └── {project-id}/
        ├── index.html
        ├── sessions/
        │   ├── index.html
        │   └── {session-id}/
        │       └── index.html
        └── plans/
            ├── index.html
            └── {plan-id}/
                └── index.html
```

## GitHub Pages へのデプロイ

1. `docs/` ディレクトリをリポジトリにコミット・プッシュ

2. リポジトリの Settings > Pages で:
   - Source: "Deploy from a branch"
   - Branch: `main` (または対象ブランチ)
   - Folder: `/docs`

3. https://username.github.io/agentrace/ でアクセス可能に

## 注意事項

- 静的サイトはレンダリング済みのスナップショットです
  - APIレスポンスは静的JSONファイルとして保存され、fetchがオーバーライドされて読み込まれます
  - ローカルで動作するインタラクション（絞り込み、折りたたみ等）は動作します
  - 書き込み操作（ログイン、データ作成等）は動作しません
- Puppeteer を使用するため、初回実行時に Chromium がダウンロードされます
- ベースパスを変更する場合は `GITHUB_PAGES_BASE` 環境変数を設定してください
