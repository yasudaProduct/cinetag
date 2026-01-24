# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## プロジェクト概要

**cinetag** は、ユーザーが映画に対して自由に「タグ（プレイリストのようなテーマ）」を作成し、他のユーザーもそのタグに映画を登録できる共有サービスです。バックエンドにGo、フロントエンドにNext.jsを使用したモノレポ構成です。

## リポジトリ構成

```
cinetag/
├── apps/
│   ├── backend/     # Go + Gin APIサーバー
│   └── frontend/    # Next.js 16 (App Router) + React 19
├── docs/            # アーキテクチャとAPI仕様書
└── compose.yml      # 開発/テスト用PostgreSQLコンテナ
```

## 開発ワークフロー

### フルスタックの起動

```bash
# ターミナル1: PostgreSQLの起動
docker compose up -d postgres

# ターミナル2: バックエンドの起動（apps/backendから）
go run ./src/cmd

# ターミナル3: フロントエンドの起動（apps/frontendから）
npm run dev
```

### 環境変数

**バックエンド** (`apps/backend/` の `.env`):
- `DATABASE_URL` - PostgreSQL接続文字列
- `CLERK_JWKS_URL` - Clerk JWKSエンドポイント（必須）
- `CLERK_ISSUER`, `CLERK_AUDIENCE` - オプショナルなJWT検証
- `TMDB_API_KEY` - 映画データ用TMDB APIキー
- `PORT` - サーバーポート（デフォルト: 8080）

**フロントエンド** (`apps/frontend/` の `.env.local`):
- `NEXT_PUBLIC_BACKEND_API_BASE` - バックエンドAPI URL
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` - Clerk公開キー
- `CLERK_SECRET_KEY` - Clerkシークレットキー

## 重要な規約

### コード構成

- **バックエンド**: レイヤー（`handler/`, `service/`, `repository/`）でグループ化し、次にドメイン別
- **フロントエンド**: `lib/api/` では機能別、共有UIは `components/ui/`

### エラーハンドリング

- **バックエンド**: ドメインエラーにはセンチネルエラーを使用し、適切なHTTPステータスコードを返す
- **フロントエンド**: ランタイム検証にZodを使用し、APIレスポンスからユーザーフレンドリーなエラーメッセージを表示

## ドキュメント

`docs/` 内の主要なドキュメント:
- `api/api-spec.md` - 完全なAPIエンドポイント仕様
- `architecture/auth-architecture.md` - Clerk連携の詳細
- `architecture/backend-architecture.md` - バックエンドの設計判断
- `data/database-schema.md` - ER図付き完全なDBスキーマ
- `frontend/frontend-api-layer.md` - フロントエンドAPI統合パターン
- `backend/movie-data-integration.md` - TMDBキャッシュ戦略

## Webhookのためのngrok

ローカルでClerk webhookを受信するには:

```bash
# 別ターミナルで
ngrok http 8080

# https URLをコピーしてClerkダッシュボードで設定
# 例: https://xxxxx.ngrok.io/api/v1/clerk/webhook
```

## 応答言語

このプロジェクトは日本語プロジェクトのため、ユーザーへの応答はすべて**日本語**で行ってください。
