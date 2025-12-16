## /cmd/db-migrate

このスラッシュコマンドは、`apps/backend/README.md` の **「DB マイグレーション（スキーマ更新）」** に従って、
バックエンドDBのスキーマを **再作成（開発向け・破壊的）** します。

### ゴール

- `apps/backend/src/cmd/migrate` を実行して、GORM の `AutoMigrate` によりスキーマを最新化する。
- （開発環境のみ）必要に応じて **public スキーマをDROP/再作成**し、テーブル定義を作り直す。

### 重要な注意（必読）

- **この操作はDBのデータを消します**（開発環境向け）。
- `apps/backend/src/cmd/migrate/main.go` は **`ENV=develop` のときだけ** `DROP SCHEMA public CASCADE; CREATE SCHEMA public;` を実行します。
  - つまり README の「全テーブル削除 → migrate」は、実質 **`ENV=develop` を付けて実行する**のが前提です。

### 前提

- `DATABASE_URL` が設定されている（`postgres://...`）
- `apps/backend` の依存関係が解決できている（`go mod tidy` 済み）

### 実行手順

1. **（推奨）実行対象のDBを確認する**
   - `DATABASE_URL` を確認し、破壊してよいDB（開発用）であることを確認する。

2. **マイグレーション実行（開発向け・全削除あり）**

```bash
cd apps/backend
ENV=develop go run ./src/cmd/migrate
```

### 期待される結果

- 標準出力に `migration completed successfully` が表示される
- DBに `users`, `tags`, `tag_movies`, `tag_followers`, `movie_cache` が作成/更新される

### トラブルシュート

- `DATABASE_URL is not set`:
  - 環境変数 `DATABASE_URL` を設定して再実行する
- `.envファイルの情報が取得できません`:
  - `apps/backend/.env` を用意する（例: `apps/backend/.env.example` を参考）
