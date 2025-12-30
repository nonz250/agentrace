# Server 設定

## 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 | memory |
| `DATABASE_URL` | DB接続文字列 | - |
| `DEV_MODE` | デバッグログ有効化 | false |
| `WEB_URL` | フロントエンドURL（開発時のリダイレクト用） | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth Client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Client Secret | - |

## データベース設定

### 対応データベース

| DB | DB_TYPE | 利用シーン |
|----|---------|------------|
| オンメモリ | `memory` | 開発・テスト |
| SQLite3 | `sqlite` | ローカル/小規模運用 |
| PostgreSQL | `postgres` | イントラネット/本番運用 |
| MongoDB | `mongodb` | AWS (DocumentDB) 環境 |

### DATABASE_URL の形式

| DB_TYPE | DATABASE_URL 例 |
|---------|-----------------|
| sqlite | `./data/agentrace.db` |
| postgres | `postgres://user:pass@localhost:5432/agentrace?sslmode=disable` |
| mongodb | `mongodb://user:pass@localhost:27017/agentrace` |

### 実装切り替え

```go
switch cfg.DBType {
case "memory":  repos = memory.NewRepositories()
case "sqlite":  repos = sqlite.NewRepositories(db)
case "postgres": repos = postgres.NewRepositories(db)
case "mongodb": repos = mongodb.NewRepositories(db)
}
```

## マイグレーション

### SQLite / PostgreSQL

- `migrations/embed.go` で SQL ファイルを Go コードに埋め込み
- DB 接続時に自動実行

```
migrations/
├── embed.go
├── sqlite/001_initial.sql
└── postgres/001_initial.up.sql
```

### MongoDB

- `mongodb.Open()` 時に `db.createIndexes()` でインデックス作成

## 開発時の起動例

```bash
cd server
DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./dev.db WEB_URL=http://localhost:5173 go run ./cmd/server
```

- `DEV_MODE=true`: リクエストログを出力
- `DB_TYPE=sqlite DATABASE_URL=./dev.db`: SQLiteを使用
- `WEB_URL`: フロントエンドURL（CLI init時のリダイレクト先）
