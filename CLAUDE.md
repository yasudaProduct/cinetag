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

## バックエンド (Go + Gin)

### 技術スタック

- **言語**: Go 1.25
- **フレームワーク**: Gin (github.com/gin-gonic/gin)
- **ORM**: GORM with PostgreSQL
- **認証**: Clerk (JWT RS256検証)

### アーキテクチャ

バックエンドは**レイヤードアーキテクチャ**を採用しています:

```
HTTPリクエスト → Router/Middleware → Handler → Service → Repository → Database
```

- **Handlers** (`internal/handler/`): HTTPリクエスト/レスポンス処理、薄いレイヤー
- **Services** (`internal/service/`): ビジネスロジック、ユースケースのオーケストレーション
- **Repositories** (`internal/repository/`): GORMを使ったデータアクセス層
- **Models** (`internal/model/`): 全レイヤーで共有されるドメインエンティティ
- **Middleware** (`internal/middleware/`): 認証、CORSなど

### よく使うコマンド

```bash
# 依存パッケージのインストール
cd apps/backend
go mod tidy

# PostgreSQLの起動（プロジェクトルートから）
docker compose up -d postgres

# APIサーバーの起動（apps/backendから）
go run ./src/cmd

# テストの実行
go test ./...

# カバレッジ付きテスト実行
go test ./... -coverprofile=coverage.out

# インテグレーションテストの実行（postgres-testコンテナが必要）
docker compose up -d postgres-test
DATABASE_URL="postgres://postgres:postgres@localhost:5433/cinetag_test?sslmode=disable" go test -tags=integration ./...

# データベースマイグレーション（全テーブル削除して再作成）
go run ./src/cmd/migrate

# 開発用seedデータと一緒に実行
ENV=develop go run ./src/cmd/migrate
```

### 認証

- `AuthMiddleware` と `OptionalAuthMiddleware` で **Clerk JWT検証**を実施
- 必須環境変数: `CLERK_JWKS_URL`（その他オプション: `CLERK_ISSUER`, `CLERK_AUDIENCE`）
- "lazy sync"を実装: `UserService.EnsureUser()` がClerkユーザーをローカル`users`テーブルに同期
- `/api/v1/clerk/webhook` でWebhook受信（`user.created`イベント）

### データベース

- **マイグレーション戦略**: 全テーブル削除 + AutoMigrate（差分マイグレーションではない）
- `go run ./src/cmd/migrate` でスキーマをリセット・再作成
- `ENV=develop` の場合、開発用seedデータを自動投入
- 完全なスキーマは `docs/data/database-schema.md` を参照

### 主要な設計パターン

- **依存性注入**: `router/router.go` でコンストラクタベース
- **インターフェースベースのサービス**: テスタビリティと疎結合のため
- **コンテキスト伝搬**: リクエストコンテキストを全レイヤーで渡す
- **エラーハンドリング**: カスタムセンチネルエラー（例: `ErrTagNotFound`, `ErrTagPermissionDenied`）

### TMDB連携

- 映画メタデータを `movie_cache` テーブルにキャッシュ（7日間のTTL）
- キャッシュファースト戦略: DB確認 → 期限切れ/存在しない場合はTMDB取得 → キャッシュをupsert
- 必須環境変数: `TMDB_API_KEY`
- 詳細は `docs/backend/movie-data-integration.md` を参照

## フロントエンド (Next.js + React)

### 技術スタック

- **フレームワーク**: Next.js 16.0.5 with App Router
- **React**: 19.2.0（React Compiler有効）
- **TypeScript**: 5.x
- **状態管理**: TanStack React Query 5.x
- **認証**: Clerk (@clerk/nextjs)
- **スタイリング**: Tailwind CSS v4 + PostCSS
- **UIコンポーネント**: shadcn/ui + Radix UI
- **バリデーション**: Zod
- **アイコン**: Lucide React

### ディレクトリ構成

```
apps/frontend/src/
├── app/                    # Next.js App Router
│   ├── (auth)/            # 認証ルート（ルートグループ）
│   ├── tags/[tagId]/      # 動的ルート
│   └── layout.tsx         # プロバイダー含むルートレイアウト
├── components/
│   ├── providers/         # React Queryなど
│   └── ui/               # shadcn/uiコンポーネント
└── lib/
    ├── api/              # リソース別に整理されたAPIレイヤー
    │   ├── _shared/      # http.ts（fetchユーティリティ）、auth.ts（トークン）
    │   ├── tags/         # タグAPI関数
    │   └── movies/       # 映画API関数
    ├── validation/       # Zodスキーマ
    └── mock/            # 開発用モックデータ
```

### よく使うコマンド

```bash
# 依存パッケージのインストール
cd apps/frontend
npm install

# 開発サーバーの起動
npm run dev

# プロダクションビルド
npm run build

# プロダクションサーバーの起動
npm start

# リンターの実行
npm run lint
```

### APIレイヤーのパターン

すべてのAPI呼び出しは以下のパターンに従います:

1. `lib/api/_shared/http.ts` の集約されたfetchユーティリティを使用
2. `lib/validation/` のZodスキーマでレスポンスを検証
3. React Query（`useQuery`, `useMutation`）でサーバー状態を管理
4. `lib/api/_shared/auth.ts` の `getBackendTokenOrThrow()` で認証トークンを処理

例:
```typescript
// コンポーネント内
const { data } = useQuery({
  queryKey: ["tags"],
  queryFn: listTags
});

// lib/api/tags/list.ts 内
export async function listTags(): Promise<TagsList> {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags`);
  const body = await safeJson(res);
  if (!res.ok) throw new Error(toApiErrorMessage({...}));
  return TagsListResponseSchema.safeParse(body).data.items;
}
```

### 認証

- `middleware.ts` の `clerkMiddleware()` によるルート保護
- 公開ルート: `/`, `/sign-in`, `/sign-up`
- トークン注入: 認証が必要なAPI呼び出しには `getBackendTokenOrThrow()` を使用
- Clerkテンプレート名: "cinetag-backend"

### スタイリング方針

- **ユーティリティファースト**: インラインTailwindクラス
- **CSS変数**: `:root` にoklch色空間を使用したテーマカラー
- **コンポーネントバリアント**: class-variance-authority (CVA) を使用
- **ダークモード**: CSS変数でサポート

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

### テスト戦略

**バックエンド**:
- ユニットテスト: `go test ./...`
- インテグレーションテスト: `go test -tags=integration ./...`（テストDB必要）
- カバレッジ: `go test ./... -coverprofile=coverage.out`

**フロントエンド**:
- React Query DevToolsでの手動テスト
- リンティング: `npm run lint`

## 重要な規約

### コード構成

- **バックエンド**: レイヤー（`handler/`, `service/`, `repository/`）でグループ化し、次にドメイン別
- **フロントエンド**: `lib/api/` では機能別、共有UIは `components/ui/`

### エラーハンドリング

- **バックエンド**: ドメインエラーにはセンチネルエラーを使用し、適切なHTTPステータスコードを返す
- **フロントエンド**: ランタイム検証にZodを使用し、APIレスポンスからユーザーフレンドリーなエラーメッセージを表示

### データベース変更

- `internal/model/` のモデル構造体を修正
- `go run ./src/cmd/migrate` でテーブルを削除して再作成
- **警告**: 全データが削除されます - 開発環境のみ

### 新規APIエンドポイントの追加

1. `internal/model/` でモデルを定義
2. `internal/repository/` でリポジトリインターフェースと実装を作成
3. `internal/service/` でサービスインターフェースと実装を作成
4. `internal/handler/` でハンドラーを作成
5. `router/router.go` でルートを登録
6. `docs/api/api-spec.md` を更新

### フロントエンドAPI統合の追加

1. `lib/validation/` でZodスキーマを定義
2. `lib/api/{resource}/` でAPI関数を作成
3. コンポーネントで適切なクエリキーを使用してReact Queryを利用
4. UIでローディング/エラー状態を処理

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

このプロジェクトは日本語プロジェクトのため、ユーザーへの応答はすべて**日本語**で行ってください（`.cursor/rules/global.mdc` を参照）。
