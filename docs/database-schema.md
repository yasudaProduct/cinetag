# 映画タグ共有サービス - テーブル定義

## ER図

```mermaid
erDiagram
    users ||--o{ tags : "creates"
    users ||--o{ tag_movies : "adds"
    users ||--o{ tag_followers : "follows"
    tags ||--o{ tag_movies : "contains"
    tags ||--o{ tag_followers : "has"

    users {
        uuid id PK
        text clerk_user_id UK "Clerk認証ID"
        text username "ユーザー名"
        text display_name "表示名"
        text email "メールアドレス"
        text avatar_url "アバター画像URL"
        text bio "自己紹介"
        timestamptz created_at
        timestamptz updated_at
    }

    tags {
        uuid id PK
        uuid user_id FK "作成者"
        text title "タイトル"
        text description "説明"
        text cover_image_url "カバー画像URL"
        boolean is_public "公開フラグ"
        text add_movie_policy "映画追加ポリシー"
        integer movie_count "映画数"
        integer follower_count "フォロワー数"
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
| `movie_cache` | TMDb映画情報キャッシュ | `tmdb_movie_id` (INTEGER) |

---

## カラム詳細

### users（ユーザー）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | ユーザーID |
| `clerk_user_id` | TEXT | NO | - | Clerk認証ID（一意） |
| `username` | TEXT | NO | - | ユーザー名 |
| `display_name` | TEXT | NO | - | 表示名 |
| `email` | TEXT | NO | - | メールアドレス |
| `avatar_url` | TEXT | YES | - | アバター画像URL |
| `bio` | TEXT | YES | - | 自己紹介文 |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |
| `updated_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 更新日時 |

### tags（タグ）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | タグID |
| `user_id` | UUID | NO | - | 作成者のユーザーID |
| `title` | TEXT | NO | - | タイトル（1-100文字） |
| `description` | TEXT | YES | - | 説明（最大500文字） |
| `cover_image_url` | TEXT | YES | - | カバー画像URL |
| `is_public` | BOOLEAN | NO | `true` | 公開フラグ |
| `movie_count` | INTEGER | NO | `0` | 映画数（非正規化） |
| `follower_count` | INTEGER | NO | `0` | フォロワー数（非正規化） |
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
| `trigger_update_tag_movie_count` | tag_movies | AFTER INSERT/DELETE | movie_count更新 |
| `trigger_update_tag_follower_count` | tag_followers | AFTER INSERT/DELETE | follower_count更新 |

---

## 制約一覧

| テーブル | 制約名 | 種類 | 内容 |
|---------|--------|------|------|
| users | `users_clerk_user_id_key` | UNIQUE | clerk_user_idの一意性 |
| tags | `tags_title_length` | CHECK | タイトル1-100文字 |
| tags | `tags_description_length` | CHECK | 説明500文字以下 |
| tags | `tags_movie_count_positive` | CHECK | movie_count >= 0 |
| tags | `tags_follower_count_positive` | CHECK | follower_count >= 0 |
| tag_movies | `tag_movies_unique` | UNIQUE | (tag_id, tmdb_movie_id)の一意性 |
| tag_movies | `tag_movies_note_length` | CHECK | メモ280文字以下 |
| tag_movies | `tag_movies_position_positive` | CHECK | position >= 0 |

---