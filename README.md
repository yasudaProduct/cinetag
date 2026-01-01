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

   ```bash
   cd apps/backend

   # 依存パッケージのインストール
   go mod tidy

   # 環境変数の設定（.env ファイルを作成）
   # CLERK_JWKS_URL、TMDB_API_KEY など

   # DB マイグレーション（開発用 seed データ込み）
   ENV=develop go run ./src/cmd/migrate

   # API サーバーの起動
   go run ./src/cmd
   ```

   サーバーはデフォルトで `http://localhost:8080` で起動します。

4. **フロントエンドのセットアップと起動**

   詳細は [apps/frontend/README.md](apps/frontend/README.md) を参照してください。

   ```bash
   cd apps/frontend
   
   # 依存パッケージのインストール
   npm install
   
   # 環境変数の設定（.env.local ファイルを作成）
   # NEXT_PUBLIC_BACKEND_API_BASE=http://localhost:8080
   # NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY、CLERK_SECRET_KEY など
   
   # 開発サーバーの起動
   npm run dev
   ```
   
   フロントエンドは `http://localhost:3000` で起動します。

### 環境変数

**バックエンド** (`apps/backend/.env`):
- `DATABASE_URL` - PostgreSQL 接続文字列（デフォルト: `postgres://postgres:postgres@localhost:5432/cinetag?sslmode=disable`）
- `CLERK_JWKS_URL` - Clerk JWKS エンドポイント（必須）
- `CLERK_ISSUER`, `CLERK_AUDIENCE` - JWT 検証用（任意）
- `TMDB_API_KEY` - TMDB API キー（映画データ取得用）
- `PORT` - サーバーポート（デフォルト: 8080）

**フロントエンド** (`apps/frontend/.env.local`):
- `NEXT_PUBLIC_BACKEND_API_BASE` - バックエンド API の URL
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` - Clerk 公開キー
- `CLERK_SECRET_KEY` - Clerk シークレットキー

### Webhook のローカル受信（任意）

Clerk などの Webhook をローカル環境で受信する場合は ngrok を使用します。

```bash
# 別ターミナルで
ngrok http 8080

# 表示された HTTPS URL を Clerk ダッシュボードに設定
# 例: https://xxxxx.ngrok.io/api/v1/clerk/webhook
```

### 詳細

- バックエンドの詳細: [apps/backend/README.md](apps/backend/README.md)
- フロントエンドの詳細: [apps/frontend/README.md](apps/frontend/README.md)
- API 仕様: [docs/api-spec.md](docs/api-spec.md)
- アーキテクチャ: [docs/backend-architecture.md](docs/backend-architecture.md)

