# Docker デプロイ

## Docker Hub

ビルド済みイメージを公開している。

- **リポジトリ**: https://hub.docker.com/r/satetsu888/agentrace
- **対応アーキテクチャ**: `linux/amd64`, `linux/arm64`

## イメージのビルド

```bash
# ローカルビルド
docker build -t agentrace:latest .

# マルチアーキテクチャビルド & push
docker buildx build --platform linux/amd64,linux/arm64 -t satetsu888/agentrace:latest --push .
```

## 起動

```bash
# 基本的な起動（ポート9080、SQLiteデータは./dataに保存）
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data agentrace:latest

# docker-composeを使う場合
docker-compose up -d
```

アクセス: http://localhost:9080

## 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `DB_TYPE` | sqlite | データベース種類 |
| `DATABASE_URL` | /data/agentrace.db | DBパス |
| `DEV_MODE` | false | デバッグログ |
| `GITHUB_CLIENT_ID` | (空) | GitHub OAuth |
| `GITHUB_CLIENT_SECRET` | (空) | GitHub OAuth |

```bash
# 環境変数を上書きする例
docker run -d -p 9080:9080 -v $(pwd)/data:/data \
  -e DEV_MODE=true \
  agentrace:latest
```

## 停止・削除

```bash
docker stop agentrace && docker rm agentrace
# または
docker-compose down
```

## Docker 構成

```
agentrace/
├── Dockerfile              # マルチステージビルド（node→go→runtime）
├── docker-compose.yml      # 簡易起動用
└── docker/
    ├── nginx.conf          # nginx設定（:9080で静的ファイル+APIプロキシ）
    ├── supervisord.conf    # プロセス管理（nginx + Go server）
    └── entrypoint.sh       # 起動スクリプト
```

## 開発環境での利用

Docker を使って開発環境をセットアップする場合:

```bash
# ビルド
docker build -t agentrace:latest .

# 起動
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data -e DEV_MODE=true agentrace:latest

# CLI初期化
npx agentrace init --url http://localhost:9080
```
