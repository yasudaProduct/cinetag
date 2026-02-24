# 映画タグ共有サービス - テーブル定義

## ER図

```mermaid
erDiagram
    users ||--o{ tags : "creates"
    users ||--o{ tag_movies : "adds"
    users ||--o{ tag_followers : "follows"
    users ||--o{ user_followers : "follows"
    users ||--o{ user_followers : "followed_by"
    tags ||--o{ tag_movies : "contains"
    tags ||--o{ tag_followers : "has"

    users {
        uuid id PK
        text clerk_user_id UK "Clerk認証ID"
        text username "ユーザー名"
        text display_id UK "表示用ID（URL用）"
        text display_name "表示名"
        text email "メールアドレス"
        text avatar_url "アバター画像URL"
        text bio "自己紹介"
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at "退会日時（論理削除）"
    }

    tags {
        uuid id PK
        uuid user_id FK "作成者"
        text title "タイトル"
        text description "説明"
        text cover_image_url "カバー画像URL"
        boolean is_public "公開フラグ"
        text add_movie_policy "映画追加ポリシー"
        timestamptz created_at
        timestamptz updated_at
    }

    tag_movies {
        uuid id PK
        uuid tag_id FK
        integer tmdb_movie_id "TMDb映画ID"
        uuid added_by_user_id FK "追加したユーザー"
        text note "メモ"
        integer position "表示順"
        timestamptz created_at
    }

    tag_followers {
        uuid tag_id PK_FK
        uuid user_id PK_FK
        timestamptz created_at
    }

    user_followers {
        uuid follower_id PK_FK "フォローする側"
        uuid followee_id PK_FK "フォローされる側"
        timestamptz created_at
    }

    movie_cache {
        integer tmdb_movie_id PK
        text title
        text original_title
        text poster_path
        text backdrop_path
        date release_date
        numeric vote_average
        text overview
        jsonb genres
        integer runtime
        timestamptz cached_at
        timestamptz expires_at
    }
```

---

## テーブル一覧

| テーブル名 | 説明 | 主キー |
|-----------|------|--------|
| `users` | ユーザー情報 | `id` (UUID) |
| `tags` | 映画タグ（プレイリスト） | `id` (UUID) |
| `tag_movies` | タグと映画の関連 | `id` (UUID) |
| `tag_followers` | タグのフォロー関係 | `(tag_id, user_id)` |
| `user_followers` | ユーザーのフォロー関係 | `(follower_id, followee_id)` |
| `movie_cache` | TMDb映画情報キャッシュ | `tmdb_movie_id` (INTEGER) |

---

## カラム詳細

### users（ユーザー）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | ユーザーID |
| `clerk_user_id` | TEXT | NO | - | Clerk認証ID（一意） |
| `username` | TEXT | NO | - | ユーザー名 |
| `display_id` | TEXT | NO | - | 表示用ID（URL用、一意、3-20文字） |
| `display_name` | TEXT | NO | - | 表示名 |
| `email` | TEXT | NO | - | メールアドレス |
| `avatar_url` | TEXT | YES | - | アバター画像URL |
| `bio` | TEXT | YES | - | 自己紹介文 |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |
| `updated_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 更新日時 |
| `deleted_at` | TIMESTAMPTZ | YES | - | 退会日時（論理削除） |

### tags（タグ）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | タグID |
| `user_id` | UUID | NO | - | 作成者のユーザーID |
| `title` | TEXT | NO | - | タイトル（1-100文字） |
| `description` | TEXT | YES | - | 説明（最大500文字） |
| `cover_image_url` | TEXT | YES | - | カバー画像URL |
| `is_public` | BOOLEAN | NO | `true` | 公開フラグ |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |
| `updated_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 更新日時 |

### tag_movies（タグ内映画）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | レコードID |
| `tag_id` | UUID | NO | - | タグID |
| `tmdb_movie_id` | INTEGER | NO | - | TMDb映画ID |
| `added_by_user_id` | UUID | NO | - | 追加したユーザーID |
| `note` | TEXT | YES | - | メモ（最大280文字） |
| `position` | INTEGER | NO | `0` | 表示順 |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 追加日時 |

### tag_followers（フォロー）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `tag_id` | UUID | NO | - | タグID（複合PK） |
| `user_id` | UUID | NO | - | ユーザーID（複合PK） |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | フォロー日時 |

### user_followers（ユーザーフォロー）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `follower_id` | UUID | NO | - | フォローする側ユーザーID（複合PK） |
| `followee_id` | UUID | NO | - | フォローされる側ユーザーID（複合PK） |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | フォロー日時 |

### movie_cache（映画キャッシュ）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `tmdb_movie_id` | INTEGER | NO | - | TMDb映画ID（PK） |
| `title` | TEXT | NO | - | 映画タイトル |
| `original_title` | TEXT | YES | - | 原題 |
| `poster_path` | TEXT | YES | - | ポスター画像パス |
| `backdrop_path` | TEXT | YES | - | 背景画像パス |
| `release_date` | DATE | YES | - | 公開日 |
| `vote_average` | NUMERIC(3,1) | YES | - | 評価スコア |
| `overview` | TEXT | YES | - | あらすじ |
| `genres` | JSONB | YES | - | ジャンル |
| `runtime` | INTEGER | YES | - | 上映時間（分） |
| `cached_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | キャッシュ作成日時 |
| `expires_at` | TIMESTAMPTZ | NO | `+7 days` | 有効期限 |

---

## トリガー一覧

| トリガー名 | テーブル | イベント | 用途 |
|-----------|---------|---------|------|
| `update_users_updated_at` | users | BEFORE UPDATE | updated_at自動更新 |
| `update_tags_updated_at` | tags | BEFORE UPDATE | updated_at自動更新 |

> 注: トリガーはドキュメント上の設計であり、現在はGORMが `updated_at` を自動管理しているため、DBトリガーは未適用です。今後のマイグレーションで段階的に追加予定。

---

## 制約一覧

| テーブル | 制約名 | 種類 | 内容 |
|---------|--------|------|------|
| users | `users_clerk_user_id_key` | UNIQUE | clerk_user_idの一意性 |
| users | `users_display_id_key` | UNIQUE | display_idの一意性 |
| tags | `tags_title_length` | CHECK | タイトル1-100文字 |
| tags | `tags_description_length` | CHECK | 説明500文字以下 |
| tag_movies | `tag_movies_unique` | UNIQUE | (tag_id, tmdb_movie_id)の一意性 |
| tag_movies | `tag_movies_note_length` | CHECK | メモ280文字以下 |
| tag_movies | `tag_movies_position_positive` | CHECK | position >= 0 |

> 注: CHECK制約はドキュメント上の設計であり、現在はアプリケーション層でバリデーションしています。FK制約も同様に未適用です。今後のマイグレーションで段階的に追加予定。

---

## スキーマ管理

スキーマ変更は **goose** によるバージョン管理型マイグレーションで管理しています。

- マイグレーションファイル: `apps/backend/src/internal/migration/migrations/*.sql`
- 実行コマンド: `go run ./src/cmd/migrate up` / `down` / `status` / `reset`
- 詳細: `docs/operations/cicd.md` の「6. マイグレーション運用」を参照

---
