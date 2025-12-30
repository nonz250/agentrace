# Agentrace

Claude Codeのやりとりをチームでレビューできるサービス

## クイックスタート

```bash
# 起動（ポート9080、データは./dataに保存）
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data satetsu888/agentrace:latest
```

http://localhost:9080 にアクセス

### CLIセットアップ

```bash
npx agentrace init --url http://localhost:9080
```

ブラウザが開くので、ユーザー登録/ログイン後「Setup CLI」をクリック

### 停止

```bash
docker stop agentrace && docker rm agentrace
```

## 環境変数

| 変数                    | デフォルト          | 説明             |
| ----------------------- | ------------------- | ---------------- |
| `DB_TYPE`               | sqlite              | データベース種類 |
| `DATABASE_URL`          | /data/agentrace.db  | DBパス           |
| `DEV_MODE`              | false               | デバッグログ     |
| `GITHUB_CLIENT_ID`      | (空)                | GitHub OAuth     |
| `GITHUB_CLIENT_SECRET`  | (空)                | GitHub OAuth     |

```bash
# 例: デバッグモードで起動
docker run -d -p 9080:9080 -v $(pwd)/data:/data -e DEV_MODE=true satetsu888/agentrace:latest
```

## 詳細ドキュメント

詳細は [CLAUDE.md](./CLAUDE.md) を参照
