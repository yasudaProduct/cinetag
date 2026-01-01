# cinetag - 映画タグ共有サービス

## 1. サービス概要

映画に対してユーザーが自由に「タグ（プレイリストのようなテーマ）」を作成し、他ユーザーもタグに映画を登録できる共有サービス。

## 2. 環境構築

### 前提条件

- Go 1.25 以上
- Node.js 18 以上
- Docker と Docker Compose
- ngrok（Webhook 受信用、任意）

### 全体の起動手順

1. **リポジトリのクローン**

   ```bash
   git clone <repository-url>
   cd cinetag
   ```

2. **PostgreSQL の起動**

   プロジェクトルートで Docker Compose を使用して PostgreSQL を起動します。

   ```bash
   docker compose up -d postgres
   ```

3. **バックエンドのセットアップと起動**

   詳細は [apps/backend/README.md](apps/backend/README.md) を参照してください。

4. **フロントエンドのセットアップと起動**

   詳細は [apps/frontend/README.md](apps/frontend/README.md) を参照してください。

### 詳細

- バックエンドの詳細: [apps/backend/README.md](apps/backend/README.md)
- フロントエンドの詳細: [apps/frontend/README.md](apps/frontend/README.md)
- API 仕様: [docs/api-spec.md](docs/api-spec.md)
- アーキテクチャ: [docs/backend-architecture.md](docs/backend-architecture.md)
- CI/CD: [docs/cicd.md](docs/cicd.md)

