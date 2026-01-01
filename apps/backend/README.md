## cinetag backend（Go / Gin）

`cinetag` のバックエンド API 実装です。フロントエンド（Next.js, `apps/frontend`）からのリクエストを受け、映画・タグなどのドメインを扱う REST API を提供します。

---

## 技術スタック

- **言語**: Go (Go 1.25 系)
- **Web フレームワーク**: Gin (`github.com/gin-gonic/gin`)
- **ORM**: GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- **DB**: PostgreSQL（開発環境では Docker を想定）

詳細なアーキテクチャやレイヤー構成は `docs/backend-architecture.md` を参照してください。

---

## ディレクトリ構成（抜粋）

```text
apps/backend/
├── src/
│   ├── cmd/
│   │   ├── main.go              # API サーバーのエントリーポイント
│   │   ├── docs/                # Swagger ドキュメント生成用
│   │   └── migrate/
│   │       └── main.go          # DB マイグレーション用コマンド
│   ├── internal/
│   │   ├── handler/             # HTTP ハンドラー
│   │   ├── service/             # ビジネスロジック
│   │   ├── model/               # ドメインモデル
│   │   └── middleware/          # カスタムミドルウェア
│   └── router/
│       └── router.go            # ルーティング定義
├── go.mod
└── go.sum
```

より詳細な責務分担についても `docs/backend-architecture.md` を参照してください。

---

## 開発環境の準備

### 1. 依存パッケージのインストール

プロジェクトルートで以下を実行します（`apps/backend` 配下でも可）:

```bash
cd apps/backend
go mod tidy
```

### 2. データベース（PostgreSQL）の起動

プロジェクトルートにある `compose.yml` を利用して PostgreSQL を起動します:

```bash
cd /Users/yuta/Develop/cinetag
docker compose up -d postgres
```

> ポート `5432` で `postgres/postgres` ユーザー・パスワードの DB が立ち上がります。

### 3. 環境変数の設定

`.env.example` ファイルをコピーして `.env` ファイルを作成し、必要な環境変数を設定します:

```bash
cd apps/backend
cp .env.example .env
```

`.env` ファイルを編集して、以下の環境変数を設定してください:

- `DATABASE_URL` - PostgreSQL 接続文字列
  - **ローカル実行時**: `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`
  - **Docker Compose実行時**: `compose.yml`で自動的に`postgres:5432`に上書きされます
- `CLERK_JWKS_URL` - Clerk JWKS エンドポイント（必須）
- `CLERK_ISSUER`, `CLERK_AUDIENCE` - JWT 検証用（任意）
- `TMDB_API_KEY` - TMDB API キー（映画データ取得用）
- `PORT` - サーバーポート（デフォルト: 8080）

> **注意**: CORSで許可するオリジンは `src/router/router.go` に直接設定されています。新しいフロントエンドURLを追加する場合は、該当ファイルを編集してください。

### 4. DB マイグレーションの実行

データベーススキーマを作成するため、マイグレーションコマンドを実行します:

```bash
cd apps/backend
go run ./src/cmd/migrate
```

> 注意: マイグレーション実行時は既存のテーブルが全て削除され、スキーマが再作成されます。開発環境でのみ使用してください。

---

## サーバーの起動

### Docker Composeで起動（推奨）

プロジェクトルートから以下を実行します:

```bash
cd /Users/yuta/Develop/cinetag
docker compose up backend
```

または、バックグラウンドで起動:

```bash
docker compose up -d backend
```

この方法では:
- `.env` ファイルが自動的に読み込まれます
- PostgreSQL との接続が自動的に設定されます（`postgres` サービス名で接続）
- ポート `8080` で起動します

### 通常起動（ローカル）

`apps/backend` ディレクトリで以下を実行します:

```bash
cd apps/backend
go run ./src/cmd
```

デフォルトではポート `8080` で起動する想定です（実際のポートや環境変数の仕様は `cmd/main.go` を参照してください）。

> **注意**: ローカル実行時は `.env` の `DATABASE_URL` が `localhost:5432` で接続されます。事前に `docker compose up -d postgres` で PostgreSQL を起動しておく必要があります。

### ホットリロード（開発用）

ファイル変更を自動検知して再ビルド・再起動するホットリロード機能を使用する場合は、`air` を使用します。

1. **`air` のインストール**

   ```bash
   # Go 1.16以降の場合
   go install github.com/air-verse/air@latest
   
   # または、Homebrew (macOS)
   brew install air-verse/air/air
   ```

2. **ホットリロードでサーバー起動**

   ```bash
   cd apps/backend
   air
   ```

   `air` は `.air.toml` 設定ファイルを読み込み、`src/` ディレクトリ内の `.go` ファイルの変更を監視します。ファイルを保存すると自動的に再ビルド・再起動されます。

   > **注意**: 初回実行時に `tmp/` ディレクトリが作成され、ビルド済みバイナリが保存されます。このディレクトリは `.gitignore` に追加することを推奨します。

---

## テストの実行

### 通常（unit）

```bash
cd apps/backend
go test ./...
```

任意:

```bash
# レース検知
go test -race ./...

# カバレッジ
go test ./... -coverprofile=coverage.out
```

### integration（DBあり）

`internal/repository` の integration テストは build tag `integration` で分離しています。

1. PostgreSQL を起動（プロジェクトルート）

```bash
cd /Users/yuta/Develop/cinetag
docker compose up -d postgres
```

2. `DATABASE_URL` を設定して実行

```bash
cd apps/backend
DATABASE_URL="postgres://postgres:postgres@localhost:5433/cinetag_test?sslmode=disable" go test -tags=integration ./...
```

---

## DB マイグレーション（スキーマ更新）

本リポジトリでは **マイグレーションファイル（差分SQL）の自動生成は行いません**。
そのため、**テーブル定義の変更が発生した場合は「全テーブル削除 → migrate 実行」でスキーマを作り直す**運用にします。

> 注意: この手順は **開発環境向け**であり、実行すると **DB内のデータは全て消えます**。

### 手順（全テーブル削除 → スキーマ再作成）

1. **migrate コマンドを実行（GORM AutoMigrate）**

```bash
cd apps/backend
go run ./src/cmd/migrate
```

### seed（開発用の初期データ自動投入）

`ENV=develop` のときは、migrate 実行後に **開発用seedデータ（ユーザー/タグ/タグ内映画/フォロー）**を自動投入します。

> seed の定義は `src/internal/seed/dev_seed.go` を参照してください。

### GitHub Actions からのマイグレーション実行

`develop` ブランチへの push 時に、GitHub Actions が自動的に開発用Neonデータベースへマイグレーションを実行します。

#### 設定手順

1. **GitHub Secrets の設定**

   GitHub リポジトリの Settings → Secrets and variables → Actions で、以下のシークレットを追加:

   - `NEON_DATABASE_URL`: 開発用Neonデータベースの接続文字列
     - 例: `postgres://user:password@ep-xxx-xxx.region.aws.neon.tech/dbname?sslmode=require`
     - Neonダッシュボードの「Connection Details」から取得

2. **ワークフローの動作**

   `.github/workflows/ci-develop.yml` の `backend-migrate` ジョブが以下を実行:
   - `ENV=develop` を設定してマイグレーション実行
   - 全テーブル削除 → スキーマ再作成 → seedデータ投入

3. **実行タイミング**

   - `develop` ブランチへの push 時に自動実行
   - 他のCIジョブ（テストなど）と並列実行

#### 手動でのマイグレーション実行

ローカル環境や手動実行が必要な場合は、以下のコマンドを使用:

```bash
cd apps/backend

# 開発用Neonデータベースへのマイグレーション
export DATABASE_URL="<開発用Neon接続文字列>"
ENV=develop go run ./src/cmd/migrate
```

### 補足（AutoMigrate の注意点）

- **GORM の `AutoMigrate` は「削除系」を自動反映しません**（カラム削除、制約/インデックス削除など）
- `./src/cmd/migrate/main.go`ではAutoMigrate前に全テーブルを削除します。

---

## 認証（Clerk JWT 検証）に必要な環境変数

- **`CLERK_JWKS_URL`（必須）**: Clerk の JWKS エンドポイント（例: `https://<your-domain>/.well-known/jwks.json`）
- **`CLERK_ISSUER`（任意）**: 期待する `iss`（設定時のみ検証します）
- **`CLERK_AUDIENCE`（任意）**: 期待する `aud`（設定時のみ検証します）

---

## ngrok を使った公開 URL の作成（例: 外部 Webhook 連携）

Clerk などの外部サービスからローカル開発環境に Webhook を受けたい場合は、`ngrok` を使ってローカルの 8080 ポートをインターネットに公開します。

1. **ngrok で 8080 を公開**

   別ターミナルで以下を実行します:

   ```bash
   ngrok http 8080
   ```

2. **発行された HTTPS URL を外部サービスに設定**

   - ngrok のコンソールに表示される `https://xxxxx.ngrok.io` のような URL をコピーし、
   - Clerk やその他 Webhook 送信元の「エンドポイント URL」として設定します  
     （例: `https://xxxxx.ngrok.io/api/webhooks/clerk` など）

ngrok 経由の URL はセッションごとに変わるため、**ngrok を再起動した場合は Webhook 設定側の URL も更新**してください。

---

## API 仕様

提供しているエンドポイントやレスポンス形式などの詳細は、`docs/api-spec.md` を参照してください。

---

## 今後の拡張メモ

- `internal/config` を追加し、環境変数・設定値の読み込みを集約する
- ログ出力の統一（`log` / `slog` など）とログフォーマットの整理
- 認証・認可（Clerk など）まわりのハンドラー・ミドルウェアの充実


