## CI/CD（継続的インテグレーション / 継続的デリバリー）運用ドキュメント

このドキュメントは、`cinetag`における **CI（テスト/静的解析の自動化）** と **CD（デプロイの自動化）** の方針・手順をまとめます。

---

## 1. 目的 / スコープ

- **目的**
  - PR/Pushごとに品質ゲート（lint/test/build）を自動実行し、回帰を早期検知する
  - デプロイ手順を定型化し、手作業ミスを削減する
  - secrets/環境変数、マイグレーション、ロールバックを運用として明確化する
- **スコープ**
  - バックエンド（`apps/backend`）
  - フロントエンド（`apps/frontend`）
  - 依存サービス（PostgreSQL、Clerk、TMDB）
  - 推奨デプロイ先: Cloudflare（`docs/infrastructure-configuration.md` の方針に準拠）

---

## 2. 現状の GitHub Actions（`ci-develop.yml`）

`/.github/workflows/ci-develop.yml` の要点:

- **トリガー**: `develop` ブランチへの push
- **concurrency**
  - `group: ci-develop-${{ github.ref }}`
  - `cancel-in-progress: true`
- **ジョブ**
  - **`backend-unit-test`**: `apps/backend` で `go test ./...`
  - **`backend-migrate`**: `apps/backend` で `ENV=develop` を付けて `go run ./src/cmd/migrate`（Neon向け）
  - **`frontend-deploy`**: `apps/frontend` で `npm run deploy`（Cloudflare向け）
  - `backend-vulncheck` / `frontend-dependency-audit` / `frontend-check` は現状コメントアウト

> 補足: PR 向けのCI（lint/build/test の品質ゲート）は `/.github/workflows/ci-pr.yml` で実行します（deploy/migrate は行いません）。

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

## 3. ブランチ / 環境戦略（推奨）

`docs/infrastructure-configuration.md` の環境分離に合わせ、以下を推奨します。

- **`main`**: 本番（Production）
- **`staging`**: ステージング（任意）
- **`develop`**: 開発（Develop）

運用の基本:

- **PRは必ずCIを通す**（`main`/`develop` へのマージ前に必須）
- **CDはブランチに紐付けて自動化**（例: `develop` へpush→開発環境へデプロイ）
- **本番は自動デプロイより「手動承認（approval）」を挟む**（誤デプロイ防止）

---

## 4. CI（継続的インテグレーション）設計

### 4.1 共通方針

- **トリガー（推奨）**
  - PR: `main`, `develop` 向け（※現状は未実装）
  - push: `main`, `develop`（※現状は `develop` のみ）
- **失敗時の扱い**: 1つでも失敗したらPRをブロック
- **キャッシュ**
  - Go: `~/go/pkg/mod`, `~/.cache/go-build`
  - Node: npm キャッシュ（`~/.npm`）または `node_modules` キャッシュ（推奨はnpm cache）

### 4.2 バックエンドCI（推奨ジョブ）

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

- **`backend-migrate`**
  - `NEON_DATABASE_URL`（GitHub Actions Secrets）

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

- **`frontend-deploy`**
  - `CLOUDFLARE_API_TOKEN`
  - `NEXT_PUBLIC_BACKEND_API_BASE`
  - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
  - `CLERK_SECRET_KEY`
  - `NEXT_PUBLIC_CLERK_SIGN_UP_URL`
  - `NEXT_PUBLIC_CLERK_SIGN_UP_FALLBACK_REDIRECT_URL`
  - `NEXT_PUBLIC_CLERK_SIGN_IN_FALLBACK_REDIRECT_URL`

> どの値が「ビルド時に必要」か「実行時に必要」かは、デプロイ方式（Pages/Workers）や設定により変わります。まずは **CIで `npm run build` が通ること** を最小ゴールにして、必要に応じて追加してください。

### 5.2 フロントエンドのデプロイ（Cloudflare / OpenNext）

このリポジトリは `apps/frontend/wrangler.jsonc` を含み、OpenNext（`@opennextjs/cloudflare`）で **Cloudflare Workers にデプロイ**する構成です。

- **ローカル実行**
  - `cd apps/frontend`
  - `npm run deploy`
- **CIから実行（推奨）**
  - `npm ci`
  - `npm run deploy`
  - `CLOUDFLARE_API_TOKEN` をsecretsとして注入

### 5.3 バックエンドのデプロイ（コンテナ）

バックエンドは `apps/backend/Dockerfile` を同梱しています。

- **推奨方針**
  - CDでは **コンテナイメージをビルド**し、デプロイ先（例: Cloudflare Containers 等）へ反映する
  - ただし、現時点で Cloudflare Containers 向けの `wrangler.toml` 等はリポジトリ内にないため、
    - 「どこへ」「どのCLIで」デプロイするか（Cloudflare / 他PaaS）を決めた上で、環境ごとの設定ファイルを追加する

---

## 6. マイグレーション運用（重要）

バックエンドの migrate は **全テーブル削除 → 再作成** の方針です（`apps/backend/README.md` 参照）。

- **開発/検証環境**: 自動実行してよい（ただしデータは消える）
- **本番環境**: 原則 **自動実行しない**（データが消えるため）

本番でのスキーマ変更が必要になった場合は、差分マイグレーション戦略（DDLの段階適用）を別途設計してください。

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

## 8. 導入チェックリスト

- **CI**
  - [ ] `apps/backend` の `go test ./...` がCI上で成功する
  - [ ] `apps/frontend` の `npm run lint` / `npm run build` がCI上で成功する
  - [ ] キャッシュが有効化され、実行時間が許容範囲
- **CD（フロント）**
  - [ ] `CLOUDFLARE_API_TOKEN` を設定し、CIから `npm run deploy` が成功する
  - [ ] `NEXT_PUBLIC_*` / `CLERK_SECRET_KEY` を適切に注入できる
- **CD（バック）**
  - [ ] デプロイ先を決め、必要な設定（例: Cloudflare Containers の設定）を追加する
  - [ ] `DATABASE_URL` など必須環境変数を安全に管理できる


