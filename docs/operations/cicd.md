## CI/CD（継続的インテグレーション / 継続的デリバリー）運用ドキュメント

このドキュメントは、`cinetag`における **CI（テスト/静的解析の自動化）** と **CD（デプロイの自動化）** の方針・手順をまとめます。

---

## 1. GitHub Actions

### 1-1. **トリガー**:`main` `develop` ブランチへの pull_request
- **ジョブ**
  - **`backend-unit-test`**: `apps/backend` で `go test ./...`
  - **`frontend-lint`**: `apps/frontend` で `npm run lint`
  - **`frontend-build`** `apps/frontend` で `npm run build`
  - **`backend-vulncheck`** `apps/backend` で `govulncheck ./...`
  - **`frontend-dependency-audit`** `apps/frontend`で`npm audit --audit-level=low`

> 補足: PR 向けのCI（lint/build/test の品質ゲート）は `/.github/workflows/ci-pr.yml` で実行します（deploy/migrate は行いません）。

### 1-2. **トリガー**:`develop` ブランチへの push（`ci-develop.yml`）
- **ジョブ**
  - **`backend-db-migrate`**: `apps/backend` で `ENV=develop` を付けて `go run ./src/cmd/migrate`（Neon向け）
  - **`backend-deploy`**: `apps/backend` を Cloud Run `cinetag-backend-develop` へデプロイ（`backend-db-migrate` 完了後に実行）
  - **`frontend-deploy`**: `apps/frontend` で `opennextjs-cloudflare deploy --env develop`（Cloudflare Workers `cinetag-frontend-develop` へ）

### 1-3. **トリガー**:`main` ブランチへの push（`ci-main.yml`）
- **ジョブ**
  - **`backend-db-migrate`**: `apps/backend` で `ENV=production` を付けて `go run ./src/cmd/migrate`（Neon 本番ブランチ向け）
  - **`backend-deploy`**: `apps/backend` を Cloud Run `cinetag-backend` 本番へデプロイ（`backend-db-migrate` 完了後に実行）
  - **`frontend-deploy`**: `apps/frontend` で `opennextjs-cloudflare deploy`（Cloudflare Workers `cinetag-frontend` 本番へ）
---

## 2. 前提（構成とコマンド）

### 2.1 バックエンド（Go）

- **場所**: `apps/backend`
- **主要コマンド**
  - unit: `go test ./...`
  - integration（DBあり）: `go test -tags=integration ./...`
  - migrate（注意: 全テーブル削除→再作成）: `go run ./src/cmd/migrate`

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

- **format（推奨）**
  - `gofmt`（差分が出たら失敗）
- **test（必須）**
  - `go test ./...`
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
  - `NEON_DATABASE_URL`（GitHub Actions Secrets）
- **`backend-deploy`**（Cloud Run）
  - `GCP_PROJECT_ID` - GCP プロジェクトID
  - `GCP_REGION` - Cloud Run のリージョン（例: `asia-northeast1`）
  - `GCP_SA_KEY` - サービスアカウントの JSON キー（Cloud Run Admin, Artifact Registry Writer, Service Account User 等のロールが必要）
  - `NEON_DATABASE_URL`（develop）/ `NEON_DATABASE_URL_PROD`（本番、別DBの場合は別途作成）
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
- **実行順序**（develop）
  - `backend-db-migrate` 完了後に `backend-deploy` を実行（スキーマ変更の整合性のため）

---

## 6. マイグレーション運用（重要）

バックエンドの migrate は **全テーブル削除 → 再作成** の方針です（`apps/backend/README.md` 参照）。

- **開発環境**（`ci-develop.yml`）: 自動実行（`NEON_DATABASE_URL` 向け）
- **本番環境**（`ci-main.yml`）: 自動実行（`NEON_DATABASE_URL_PROD` 向け）

**注意**: 本番マイグレーションは全テーブル削除のため、実行のたびに本番データが消えます。スキーマ変更のみでデータを保持したい場合は、差分マイグレーション戦略（DDLの段階適用）を別途設計してください。

---

## 7. リリース/ロールバック（推奨）

- **リリース手順（例）**
  - `develop` へマージ → 開発環境へ自動デプロイ
  - `staging` へマージ → ステージングへ自動デプロイ（任意）
  - `main` へマージ → 本番デプロイ（手動承認つき）
- **ロールバック**
  - 原則: 直前の正常コミットへrevertし再デプロイ（Gitベース）
  - DB変更が絡む場合: 事前に後方互換な変更（expand/contract）にするか、別途ロールバック手順を準備する

---


