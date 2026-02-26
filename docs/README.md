# docs

このフォルダは、`cinetag` の設計・仕様・運用ドキュメントをカテゴリ別に整理しています。

## API

- API仕様（v1）: `docs/api/api-spec.md`

## アーキテクチャ

- 認証・ユーザー管理（Clerk）: `docs/architecture/auth-architecture.md`
- バックエンドアーキテクチャ（Gin / レイヤード）: `docs/architecture/backend-architecture.md`
- インフラ構成（Cloudflare / Cloud Run / Neon）: `docs/architecture/infrastructure-configuration.md`

## バックエンド

- バックエンド テスト計画（Unit Test Plan）: `docs/backend/backend-test-plan.md`
- 映画データ連携（TMDB / movie_cache）: `docs/backend/movie-data-integration.md`

## フロントエンド

- next.config.ts 設定ガイド（CSP / 画像 / セキュリティヘッダー）: `docs/frontend/next-config.md`
- フロントエンド API レイヤー設計: `docs/frontend/frontend-api-layer.md`
- フロントエンド バリデーション仕様（zod）: `docs/frontend/frontend-validation.md`

## データ（DB）

- DBスキーマ（ER図・制約・トリガー）: `docs/data/database-schema.md`

## 運用（CI/CD）

- CI/CD運用: `docs/operations/cicd.md`

