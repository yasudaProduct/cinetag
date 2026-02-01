# CLAUDE.md - Backend

Go + Gin APIサーバーのガイダンスです。

## 技術スタック

- **言語**: Go 1.25
- **フレームワーク**: Gin (github.com/gin-gonic/gin)
- **ORM**: GORM with PostgreSQL
- **認証**: Clerk (JWT RS256検証)

## アーキテクチャ

バックエンドは**レイヤードアーキテクチャ**を採用しています:

```
HTTPリクエスト → Router/Middleware → Handler → Service → Repository → Database
```

- **Handlers** (`internal/handler/`): HTTPリクエスト/レスポンス処理、薄いレイヤー
- **Services** (`internal/service/`): ビジネスロジック、ユースケースのオーケストレーション
- **Repositories** (`internal/repository/`): GORMを使ったデータアクセス層
- **Models** (`internal/model/`): 全レイヤーで共有されるドメインエンティティ
- **Middleware** (`internal/middleware/`): 認証、CORSなど

## よく使うコマンド

```bash
# 依存パッケージのインストール
go mod tidy

# PostgreSQLの起動（プロジェクトルートから）
docker compose up -d postgres

# APIサーバーの起動
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

## 認証

- `AuthMiddleware` と `OptionalAuthMiddleware` で **Clerk JWT検証**を実施
- 必須環境変数: `CLERK_JWKS_URL`（その他オプション: `CLERK_ISSUER`, `CLERK_AUDIENCE`）
- "lazy sync"を実装: `UserService.EnsureUser()` がClerkユーザーをローカル`users`テーブルに同期
- `/api/v1/clerk/webhook` でWebhook受信（`user.created`イベント）

## データベース

- **マイグレーション戦略**: 全テーブル削除 + AutoMigrate（差分マイグレーションではない）
- `go run ./src/cmd/migrate` でスキーマをリセット・再作成
- `ENV=develop` の場合、開発用seedデータを自動投入
- 完全なスキーマは `docs/data/database-schema.md` を参照

## 主要な設計パターン

- **依存性注入**: `router/router.go` でコンストラクタベース
- **インターフェースベースのサービス**: テスタビリティと疎結合のため
- **コンテキスト伝搬**: リクエストコンテキストを全レイヤーで渡す
- **エラーハンドリング**: カスタムセンチネルエラー（例: `ErrTagNotFound`, `ErrTagPermissionDenied`）

## TMDB連携

- 映画メタデータを `movie_cache` テーブルにキャッシュ（7日間のTTL）
- キャッシュファースト戦略: DB確認 → 期限切れ/存在しない場合はTMDB取得 → キャッシュをupsert
- 必須環境変数: `TMDB_API_KEY`
- 詳細は `docs/backend/movie-data-integration.md` を参照

## 環境変数

`.env` ファイルに設定:
- `DATABASE_URL` - PostgreSQL接続文字列
- `CLERK_JWKS_URL` - Clerk JWKSエンドポイント（必須）
- `CLERK_ISSUER`, `CLERK_AUDIENCE` - オプショナルなJWT検証
- `TMDB_API_KEY` - 映画データ用TMDB APIキー
- `PORT` - サーバーポート（デフォルト: 8080）

## テスト戦略

- ユニットテスト: `go test ./...`
- インテグレーションテスト: `go test -tags=integration ./...`（テストDB必要）
- カバレッジ: `go test ./... -coverprofile=coverage.out`

## 新規APIエンドポイントの追加

1. `internal/model/` でモデルを定義
2. `internal/repository/` でリポジトリインターフェースと実装を作成
3. `internal/service/` でサービスインターフェースと実装を作成
4. `internal/handler/` でハンドラーを作成
5. `router/router.go` でルートを登録
6. `docs/api/api-spec.md` を更新

## データベース変更

- `internal/model/` のモデル構造体を修正
- `go run ./src/cmd/migrate` でテーブルを削除して再作成
- **警告**: 全データが削除されます - 開発環境のみ
