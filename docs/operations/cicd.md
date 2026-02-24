## CI/CD（継続的インテグレーション / 継続的デリバリー）運用ドキュメント

このドキュメントは、`cinetag`における **CI（テスト/静的解析の自動化）** と **CD（デプロイの自動化）** の方針・手順をまとめます。

---

## 1. GitHub Actions

### 1-1. **トリガー**:`main` `develop` ブランチへの pull_request
- **ジョブ**
  - **`backend-unit-test`**: `apps/backend` で `go test ./...`
  - **`backend-migration-check`**: PostgreSQLサービスコンテナでマイグレーションの up/down/up 往復テスト
  - **`frontend-lint`**: `apps/frontend` で `npm run lint`
  - **`frontend-build`** `apps/frontend` で `npm run build`
  - **`backend-vulncheck`** `apps/backend` で `govulncheck ./...`
  - **`frontend-dependency-audit`** `apps/frontend`で`npm audit --audit-level=low`

> 補足: PR 向けのCI（lint/build/test の品質ゲート）は `/.github/workflows/ci-pr.yml` で実行します（deploy/migrate は行いません）。

### 1-2. **トリガー**:`develop` ブランチへの push（`ci-develop.yml`）
- **ジョブ**
  - **`backend-db-migrate`**: `apps/backend` で `go run ./src/cmd/migrate up`（goose による差分マイグレーション、Neon向け）
    - シードデータ投入: `go run ./src/cmd/seed`（`ENV=develop` で実行）
    - 手動トリガー時に `reset_db=true` を指定すると、全マイグレーションをリセットして再適用
  - **`backend-deploy`**: `apps/backend` を Cloud Run `cinetag-backend-develop` へデプロイ（`backend-db-migrate` 完了後に実行）
  - **`frontend-deploy`**: `apps/frontend` で `opennextjs-cloudflare deploy --env develop`（Cloudflare Workers `cinetag-frontend-develop` へ）

### 1-3. **トリガー**:`main` ブランチへの push（`ci-main.yml`）
- **ジョブ**
  - **`backend-db-migrate`**: `apps/backend` で `go run ./src/cmd/migrate up`（goose による差分マイグレーション、本番Neon向け）
  - **`backend-deploy`**: `apps/backend` を Cloud Run `cinetag-backend` 本番へデプロイ（`backend-db-migrate` 完了後に実行）
  - **`frontend-deploy`**: `apps/frontend` で `opennextjs-cloudflare deploy`（Cloudflare Workers `cinetag-frontend` 本番へ）

---

## 2. 前提（構成とコマンド）

### 2.1 バックエンド（Go）

- **場所**: `apps/backend`
- **主要コマンド**
  - unit: `go test ./...`
  - integration（DBあり）: `go test -tags=integration ./...`
  - migrate: `go run ./src/cmd/migrate up`（差分マイグレーション適用）
  - seed: `ENV=develop go run ./src/cmd/seed`（開発用シードデータ投入）

### 2.2 フロントエンド（Next.js）

- **場所**: `apps/frontend`
- **主要コマンド**
  - lint: `npm run lint`
  - build: `npm run build`
  - Cloudflare向け:
    - preview: `npm run preview`（`opennextjs-cloudflare build && opennextjs-cloudflare preview`）
    - deploy: `npm run deploy`（`opennextjs-cloudflare build && opennextjs-cloudflare deploy`）

---

## 3. CI（継続的インテグレーション）設計

### 3.1 共通方針

- **トリガー（推奨）**
  - PR: `main`, `develop` 向け（`ci-pr.yml`）
  - push: `develop` → 開発環境デプロイ（`ci-develop.yml`）、`main` → 本番デプロイ（`ci-main.yml`）
- **失敗時の扱い**: 1つでも失敗したらPRをブロック
- **キャッシュ**
  - Go: `~/go/pkg/mod`, `~/.cache/go-build`
  - Node: npm キャッシュ（`~/.npm`）または `node_modules` キャッシュ（推奨はnpm cache）

### 3.2 バックエンドCI（推奨ジョブ）

- **test（必須）**
  - `go test ./...`
- **migration-check（必須）**
  - PostgreSQLサービスコンテナを使用
  - `go run ./src/cmd/migrate up` → `down` → `up` の往復テスト
  - マイグレーションSQLの構文エラーとロールバックの正常性を検証
- **integration（任意/段階導入）**
  - `docker compose up -d postgres-test`
  - `DATABASE_URL="postgres://postgres:postgres@localhost:5433/cinetag_test?sslmode=disable" go test -tags=integration ./...`

> integration は実行時間と安定性の観点から、まずは **nightly** や **手動トリガー** から始めるのがおすすめです。

### 4.3 フロントエンドCI（推奨ジョブ）

- **install（必須）**
  - `npm ci`（CIでは `npm install` より再現性が高い）
- **lint（必須）**
  - `npm run lint`
- **build（必須）**
  - `npm run build`

> Next.js の `build` は、通常 TypeScript の型チェックも含むため、まずは `build` を品質ゲートとして扱うのが安全です。

---

## 5. CD（継続的デリバリー）設計

### 5.1 secrets / 環境変数（一覧）

#### バックエンド（現状ワークフローで使用）

- **`backend-db-migrate`**
  - `NEON_DATABASE_URL`（develop用、GitHub Actions Secrets）
  - `NEON_DATABASE_URL_PROD`（本番用、GitHub Actions Secrets）
- **`backend-deploy`**（Cloud Run）
  - `GCP_PROJECT_ID` - GCP プロジェクトID
  - `GCP_REGION` - Cloud Run のリージョン（例: `asia-northeast1`）
  - `GCP_SA_KEY` - サービスアカウントの JSON キー（Cloud Run Admin, Artifact Registry Writer, Service Account User 等のロールが必要）
  - `NEON_DATABASE_URL`（develop）/ `NEON_DATABASE_URL_PROD`（本番）
  - `CLERK_JWKS_URL`
  - `TMDB_API_KEY`

> **GCP サービスアカウントの権限**: Cloud Run Admin, Artifact Registry Writer（または Admin）, Service Account User, Cloud Build Editor, Storage Admin 等が必要です。詳細は [Deploying from source code](https://cloud.google.com/run/docs/deploying-source-code#permissions_required_to_deploy) を参照してください。

#### バックエンド（アプリ実行時の例）

- **必須候補**
  - `DATABASE_URL`
  - `CLERK_JWKS_URL`
  - `TMDB_API_KEY`
- **任意**
  - `CLERK_ISSUER`
  - `CLERK_AUDIENCE`
  - `PORT`
  - `MAINTENANCE_MODE` - `true` でメンテナンスモード有効化（全APIが503を返す）

#### フロントエンド（現状ワークフローで使用）

- **`frontend-deploy`**（`ci-develop.yml` / `ci-main.yml` 共通）
  - `CLOUDFLARE_API_TOKEN`
  - `NEXT_PUBLIC_BACKEND_API_BASE`
  - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
  - `CLERK_SECRET_KEY`
  - `NEXT_PUBLIC_CLERK_SIGN_UP_URL`
  - `NEXT_PUBLIC_CLERK_SIGN_UP_FALLBACK_REDIRECT_URL`
  - `NEXT_PUBLIC_CLERK_SIGN_IN_FALLBACK_REDIRECT_URL`

> 本番と開発でバックエンドURLが異なる場合は、`NEXT_PUBLIC_BACKEND_API_BASE_PROD` 等の環境別 secrets を追加し、ワークフローで条件分岐する必要があります。

### 5.2 フロントエンドのデプロイ（Cloudflare / OpenNext）

このリポジトリは `apps/frontend/wrangler.jsonc` を含み、OpenNext（`@opennextjs/cloudflare`）で **Cloudflare Workers にデプロイ**する構成です。

- **環境分離**（`wrangler.jsonc` の `env`）
  - 本番: `cinetag-frontend`（`wrangler deploy` または `npm run deploy`）
  - 開発: `cinetag-frontend-develop`（`wrangler deploy --env develop`）
- **ローカル実行**
  - `cd apps/frontend`
  - 本番: `npm run deploy`
  - 開発: `npx opennextjs-cloudflare build && npx opennextjs-cloudflare deploy --env develop`
- **CIから実行**
  - `develop` push: `opennextjs-cloudflare deploy --env develop`（`ci-develop.yml`）
  - `main` push: `opennextjs-cloudflare deploy`（`ci-main.yml`）
  - `CLOUDFLARE_API_TOKEN` をsecretsとして注入

### 5.3 バックエンドのデプロイ（Cloud Run）

バックエンドは `apps/backend/Dockerfile` を同梱し、GitHub Actions から **Cloud Run** へデプロイします。

- **環境分離**
  - 開発: `cinetag-backend-develop`（`ci-develop.yml`、`develop` push）
  - 本番: `cinetag-backend`（`ci-main.yml`、`main` push）
- **デプロイ方式**
  - `google-github-actions/deploy-cloudrun` を使用し、`source: apps/backend` からソースビルド
  - Cloud Build が Dockerfile をビルドし、Artifact Registry 経由で Cloud Run にデプロイ
- **実行順序**（develop / main 共通）
  - `backend-db-migrate`（goose up）完了後に `backend-deploy` を実行（スキーマ変更の整合性のため）

---

## 6. マイグレーション運用

### 6.1 ツール

**goose**（`github.com/pressly/goose/v3`）をGoライブラリとして使用。SQLファイルは `embed.FS` でバイナリに埋め込み。

- マイグレーションファイル: `apps/backend/src/internal/migration/migrations/*.sql`
- 埋め込み定義: `apps/backend/src/internal/migration/embed.go`
- 実行コマンド: `apps/backend/src/cmd/migrate/main.go`

### 6.2 コマンド

```bash
# apps/backend/ から実行
go run ./src/cmd/migrate up       # 未適用のマイグレーションを全て適用
go run ./src/cmd/migrate down     # 最新1つをロールバック
go run ./src/cmd/migrate status   # 適用状況を表示
ENV=develop go run ./src/cmd/migrate reset  # 全リセット→再適用（develop環境のみ）
```

### 6.3 マイグレーションファイルの書き方

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS username;
```

- `-- +goose Up` と `-- +goose Down` の両方を必ず記述する
- ファイル名: `NNNNN_説明.sql`（例: `00002_add_foreign_keys.sql`）
- PRの `backend-migration-check` ジョブで up/down/up の往復テストが自動実行される

### 6.4 デプロイ順序

```
goose up（マイグレーション） → backend deploy → frontend deploy
```

マイグレーションが先に走り、新しいコードが後からデプロイされる。

### 6.5 Expand-Contract パターン（破壊的変更時）

非破壊的変更（カラム追加、テーブル追加）はそのまま適用可能。破壊的変更（カラム削除、リネーム等）は2段階で実施する:

1. **Expand（PR1）**: 新カラム追加 + コードを新旧両方に対応
2. **Contract（PR2）**: 旧カラム削除のマイグレーション

### 6.6 メンテナンスモード

破壊的マイグレーション実行時に `MAINTENANCE_MODE=true` を Cloud Run の環境変数に設定すると、`/health` 以外の全APIが503を返す。

---

## 7. リリース/ロールバック（推奨）

- **リリース手順（例）**
  - `develop` へマージ → 開発環境へ自動デプロイ（マイグレーション含む）
  - `staging` へマージ → ステージングへ自動デプロイ（任意）
  - `main` へマージ → 本番デプロイ（マイグレーション → デプロイの順で自動実行）
- **ロールバック**
  - マイグレーション: `go run ./src/cmd/migrate down` で最新1つをロールバック
  - アプリケーション: 直前の正常コミットへrevertし再デプロイ（Gitベース）
  - DB変更が絡む場合: Expand-Contract パターンにより後方互換な変更にしておくことが推奨

---
