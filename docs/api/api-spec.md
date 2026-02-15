## API仕様書（v1）

### 1. 概要

- **ベースURL**
  - ローカル: `http://localhost:8080`
  - APIプレフィックス: `/api/v1`
- **認証方式**
  - フロントエンドは Clerk を利用し、HTTP リクエストにトークンを付与する（`Authorization: Bearer <token>` など）。
  - バックエンドでは Gin の `AuthMiddleware` でトークン検証を行い、認証必須エンドポイントを保護する。
- **レスポンス形式**
  - すべて `application/json`
  - タイムスタンプは原則 ISO 8601 (`YYYY-MM-DDTHH:MM:SSZ`) で返す想定。

---

### 2. 認証ポリシー

- **認証必須（`AuthMiddleware` 適用）**
  - ログインユーザーに紐づくリソースの取得・作成・更新・削除
    - 例: 自分のタグ一覧、タグ作成、タグへの映画追加、フォロー操作 など
- **認証不要**
  - ヘルスチェック (`GET /health`)
  - 公開タグの一覧・詳細取得（将来の方針に応じて変更可能）

> 方針: 「ユーザー固有の状態を扱う API」はすべて `AuthMiddleware` を必須とする。

---

### 3. ヘルスチェック

#### 3.1 GET `/health`

- **概要**: システムの稼働確認用エンドポイント。
- **認証**: 不要
- **リクエスト例**

```http
GET /health HTTP/1.1
Host: localhost:8080
```

- **レスポンス例**

```json
{
  "status": "ok"
}
```

#### 3.2 GET `/swagger/*any`

- **概要**: Swagger UI を表示する（開発用）。
- **認証**: 不要
- **備考**:
  - API 仕様としての JSON を返すエンドポイントではなく、HTML を返す。
  - 本番公開の可否は運用方針に従う。

---

### 4. 認証ユーザー系エンドポイント

#### 4.1 GET `/api/v1/users/me`

- **概要**: 認証済みユーザー自身のプロフィール情報を取得する。
- **認証**: 必須
- **レスポンス例（200）**

```json
{
  "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
  "display_id": "cinephile_jane",
  "display_name": "cinephile_jane",
  "avatar_url": "https://images.example.com/avatar.jpg",
  "bio": "映画が好きです"
}
```

#### 4.2 PATCH `/api/v1/users/me`

- **概要**: 認証済みユーザー自身のプロフィール情報を更新する（部分更新）。
- **認証**: 必須
- **リクエストボディ**

```json
{
  "display_name": "新しい表示名"
}
```

- **備考**
  - `display_name` は任意。省略した場合は更新されない。

- **レスポンス例（200）**: 更新後のユーザープロフィール。

```json
{
  "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
  "display_id": "cinephile_jane",
  "display_name": "新しい表示名",
  "avatar_url": "https://images.example.com/avatar.jpg",
  "bio": "映画が好きです"
}
```

- **レスポンス例（400）**

```json
{
  "error": "invalid request body"
}
```

- **レスポンス例（401）**

```json
{
  "error": "unauthorized"
}
```

#### 4.3 GET `/api/v1/users/:displayId`

- **概要**: 指定ユーザー（`displayId`）のユーザー情報を取得する。
- **認証**: 不要
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **レスポンス例（200）**

```json
{
  "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
  "display_id": "cinephile_jane",
  "display_name": "cinephile_jane",
  "avatar_url": "https://images.example.com/avatar.jpg",
  "bio": "映画が好きです"
}
```

- **レスポンス例（404）**

```json
{
  "error": "user not found"
}
```

#### 4.4 GET `/api/v1/users/:displayId/tags`

- **概要**: 指定ユーザー（`displayId`）が作成したタグの一覧を取得する。
- **認証**: 任意（公開/非公開タグの扱いに応じて将来変更の可能性あり）
- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                                             |
|-------------|-----|------|--------------------------------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **備考**
  - 未認証、または閲覧者が本人以外の場合は **公開タグのみ** を返す。
  - 認証済みで閲覧者が本人の場合は **非公開タグも含めて** 返す。

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "tag-uuid-1",
      "title": "90年代SFクラシック",
      "description": "黄金時代の象徴的なSF作品。",
      "author": "retro_future",
      "author_display_id": "retro_future",
      "cover_image_url": null,
      "is_public": true,
      "movie_count": 55,
      "follower_count": 2100,
      "images": [
        "https://image.tmdb.org/t/p/w400/poster1.jpg",
        "https://image.tmdb.org/t/p/w400/poster2.jpg",
        "https://image.tmdb.org/t/p/w400/poster3.jpg",
        "https://image.tmdb.org/t/p/w400/poster4.jpg"
      ],
      "created_at": "2025-01-01T12:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 4.5 POST `/api/v1/users/:displayId/follow`

- **概要**: 指定ユーザー（`displayId`）をフォローする。
- **認証**: 必須
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **レスポンス例（200）**

```json
{
  "message": "successfully followed"
}
```

- **レスポンス例（400）**

```json
{
  "error": "cannot follow yourself"
}
```

- **レスポンス例（404）**

```json
{
  "error": "user not found"
}
```

- **レスポンス例（409）**

```json
{
  "error": "already following"
}
```

#### 4.6 DELETE `/api/v1/users/:displayId/follow`

- **概要**: 指定ユーザー（`displayId`）のフォローを解除する。
- **認証**: 必須
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **レスポンス例（200）**

```json
{
  "message": "successfully unfollowed"
}
```

- **レスポンス例（404）**

```json
{
  "error": "user not found"
}
```

- **レスポンス例（409）**

```json
{
  "error": "not following"
}
```

#### 4.7 GET `/api/v1/users/:displayId/following`

- **概要**: 指定ユーザー（`displayId`）がフォローしているユーザー一覧を取得する。
- **認証**: 不要
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                                             |
|-------------|-----|------|--------------------------------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
      "display_id": "cinephile_jane",
      "display_name": "cinephile_jane",
      "avatar_url": "https://images.example.com/avatar.jpg",
      "bio": "映画が好きです"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 4.8 GET `/api/v1/users/:displayId/followers`

- **概要**: 指定ユーザー（`displayId`）をフォローしているユーザー一覧を取得する。
- **認証**: 不要
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                                             |
|-------------|-----|------|--------------------------------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
      "display_id": "cinephile_jane",
      "display_name": "cinephile_jane",
      "avatar_url": "https://images.example.com/avatar.jpg",
      "bio": "映画が好きです"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 4.9 GET `/api/v1/users/:displayId/follow-stats`

- **概要**: 指定ユーザー（`displayId`）のフォロー数・フォロワー数を取得する。
- **認証**: 任意
- **備考**
  - 認証済みの場合、閲覧者がこのユーザーをフォローしているか（`is_following`）も返す。
  - 未認証の場合は `is_following: false` を返す。
- **パスパラメータ**

| 名前         | 型   | 説明 |
|--------------|------|------|
| `displayId`  | text | ユーザーの表示ID（`display_id`） |

- **レスポンス例（200）**

```json
{
  "following_count": 12,
  "followers_count": 34,
  "is_following": false
}
```

#### 4.10 GET `/api/v1/me/tags` **[未実装]**

- **概要**: ログインユーザーが作成したタグの一覧を取得する。
- **認証**: 必須
- **クエリパラメータ**

| 名前        | 型    | 必須 | 説明                          |
|-------------|-------|------|-------------------------------|
| `page`      | int   | 任意 | ページ番号（デフォルト: 1）   |
| `page_size` | int   | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "tag-uuid-1",
      "title": "ジブリの名作",
      "description": "スタジオジブリの不朽の名作アニメ。",
      "cover_image_url": "https://example.com/cover.jpg",
      "is_public": true,
      "movie_count": 32,
      "follower_count": 120,
      "created_at": "2025-01-01T12:00:00Z",
      "updated_at": "2025-01-01T12:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 4.11 GET `/api/v1/me/following-tags`

- **概要**: ログインユーザーがフォローしているタグ一覧を取得する。
- **認証**: 必須
- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                                             |
|-------------|-----|------|--------------------------------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "tag-uuid-1",
      "title": "90年代SFクラシック",
      "description": "黄金時代の象徴的なSF作品。",
      "author": "retro_future",
      "author_display_id": "retro_future",
      "cover_image_url": null,
      "is_public": true,
      "movie_count": 55,
      "follower_count": 2100,
      "images": [
        "https://image.tmdb.org/t/p/w400/poster1.jpg",
        "https://image.tmdb.org/t/p/w400/poster2.jpg",
        "https://image.tmdb.org/t/p/w400/poster3.jpg",
        "https://image.tmdb.org/t/p/w400/poster4.jpg"
      ],
      "created_at": "2025-01-01T12:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 4.12 POST `/api/v1/clerk/webhook`

- **概要**: Clerk Webhook を受信し、`user.created` および `user.deleted` イベントをローカル `users` テーブルに同期する。
- **認証**: 不要
- **注意**
  - 現時点では **svix署名検証は未実装**（将来追加予定）。
  - サポートするイベントタイプ: `user.created`, `user.deleted`
  - これら以外のイベントは `200 OK` で無視する。

- **リクエストボディ（`user.created` の例）**

```json
{
  "type": "user.created",
  "data": {
    "id": "user_2aBcDeFgHiJk",
    "username": "cinephile_jane",
    "first_name": "Jane",
    "last_name": "Doe",
    "image_url": "https://images.example.com/avatar.jpg",
    "email_addresses": [
      {
        "email_address": "jane@example.com"
      }
    ]
  }
}
```

- **リクエストボディ（`user.deleted` の例）**

```json
{
  "type": "user.deleted",
  "data": {
    "id": "user_2aBcDeFgHiJk"
  }
}
```

- **レスポンス**
  - `200 OK`（ボディなし）

- **レスポンス例（400）**

```json
{
  "error": "invalid webhook payload"
}
```

- **レスポンス例（500）**

```json
{
  "error": "failed to sync user"
}
```

---

### 5. タグ（Tags）エンドポイント

#### 5.1 GET `/api/v1/tags`

- **概要**: 公開タグを一覧・検索する。
- **認証**: なし
- **クエリパラメータ**

| 名前        | 型    | 必須 | 説明                                             |
|-------------|-------|------|--------------------------------------------------|
| `q`         | text  | 任意 | タイトル等の全文検索キーワード                  |
| `sort`      | text  | 任意 | `popular` / `recent` / `movie_count` など       |
| `page`      | int   | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int   | 任意 | 1ページあたり件数（デフォルト: 20, 上限例: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "tag-uuid-1",
      "title": "90年代SFクラシック",
      "description": "黄金時代の象徴的なSF作品。",
      "author": "retro_future",
      "author_display_id": "retro_future",
      "cover_image_url": null,
      "is_public": true,
      "movie_count": 55,
      "follower_count": 2100,
      "images": [
        "https://image.tmdb.org/t/p/w400/poster1.jpg",
        "https://image.tmdb.org/t/p/w400/poster2.jpg",
        "https://image.tmdb.org/t/p/w400/poster3.jpg",
        "https://image.tmdb.org/t/p/w400/poster4.jpg"
      ],
      "created_at": "2025-01-01T12:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 5.2 GET `/api/v1/tags/:tagId`

- **概要**: 指定タグの詳細情報を取得する。
- **認証**: 任意（未認証時は公開タグのみ参照可能。非公開タグの参照権限はサーバー側で判定する）
- **パスパラメータ**

| 名前     | 型    | 説明          |
|----------|-------|---------------|
| `tagId`  | UUID  | タグのID      |

- **レスポンス例（200）**

```json
{
  "id": "tag-uuid-1",
  "title": "ジブリの名作",
  "description": "スタジオジブリの不朽の名作アニメ。",
  "cover_image_url": "https://example.com/cover.jpg",
  "is_public": true,
  "add_movie_policy": "everyone",
  "movie_count": 32,
  "follower_count": 120,
  "owner": {
    "id": "user-uuid-1",
    "username": "cinephile_jane",
    "display_id": "cinephile_jane",
    "display_name": "cinephile_jane",
    "avatar_url": "https://images.example.com/avatar.jpg"
  },
  "can_edit": true,
  "can_add_movie": true,
  "participant_count": 120,
  "participants": [],
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### 5.3 POST `/api/v1/tags`

- **概要**: 新しいタグを作成する。
- **認証**: 必須
- **リクエストボディ**

```json
{
  "title": "ジブリの名作",
  "description": "スタジオジブリの不朽の名作アニメ。",
  "cover_image_url": "https://example.com/cover.jpg",
  "is_public": true,
  "add_movie_policy": "everyone"
}
```

- **バリデーション**
  - `title`: 1〜100文字
  - `description`: 最大500文字
  - `add_movie_policy`: `everyone` / `owner_only`（省略時: `everyone`）

- **備考**
  - `is_public` を省略した場合は `true` として作成される。

- **レスポンス例（201）**

```json
{
  "id": "tag-uuid-1",
  "title": "ジブリの名作",
  "description": "スタジオジブリの不朽の名作アニメ。",
  "cover_image_url": "https://example.com/cover.jpg",
  "is_public": true,
  "add_movie_policy": "everyone",
  "movie_count": 0,
  "follower_count": 0,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### 5.4 PATCH `/api/v1/tags/:tagId`

- **概要**: 既存タグのメタ情報を更新する（作成者のみ）。
- **認証**: 必須
- **ボディ（部分更新）**

```json
{
  "title": "ジブリの名作（日本語版）",
  "description": "説明を更新しました",
  "cover_image_url": null,
  "is_public": false,
  "add_movie_policy": "owner_only"
}
```

- **備考**
  - 省略したフィールドは更新されない（部分更新）。
  - `description` / `cover_image_url` は `null` を送ることで値をクリアできる。

- **レスポンス例（200）**: 更新後のタグオブジェクト。

#### 5.5 DELETE `/api/v1/tags/:tagId` **[未実装]**

- **概要**: タグを削除する（物理削除 or ソフトデリートは実装方針による）。
- **認証**: 必須（作成者のみ）
- **レスポンス**
  - `204 No Content`（ボディなし）を想定。

---

### 6. タグ内映画（Tag Movies）エンドポイント

#### 6.1 GET `/api/v1/tags/:tagId/movies`

- **概要**: 指定タグに含まれる映画一覧を取得する。
- **認証**: 任意（タグの公開設定に応じて制御）
- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                        |
|-------------|-----|------|-----------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1） |
| `page_size` | int | 任意 | 1ページ件数                 |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "tag-movie-uuid-1",
      "tag_id": "tag-uuid-1",
      "tmdb_movie_id": 12345,
      "note": "子どもの頃に観た思い出の作品",
      "position": 1,
      "added_by_user_id": "user-uuid-1",
      "can_delete": true,
      "created_at": "2025-01-01T12:00:00Z",
      "movie": {
        "title": "千と千尋の神隠し",
        "original_title": "Spirited Away",
        "poster_path": "/path/to/poster.jpg",
        "release_date": "2001-07-20",
        "vote_average": 8.7
      }
    }
  ],
  "page": 1,
  "page_size": 50,
  "total_count": 1
}
```

#### 6.2 POST `/api/v1/tags/:tagId/movies`

- **概要**: タグに映画を追加する。
- **認証**: 必須
- **リクエストボディ**

```json
{
  "tmdb_movie_id": 12345,
  "note": "子どもの頃に観た思い出の作品",
  "position": 1
}
```

- **制約**
  - `(tag_id, tmdb_movie_id)` は一意（`tag_movies_unique`）。

- **レスポンス例（201）**

```json
{
  "id": "tag-movie-uuid-1",
  "tag_id": "tag-uuid-1",
  "tmdb_movie_id": 12345,
  "note": "子どもの頃に観た思い出の作品",
  "position": 1,
  "added_by_user_id": "user-uuid-1",
  "created_at": "2025-01-01T12:00:00Z"
}
```

- **レスポンス例（409）**

```json
{
  "error": "movie already added to tag"
}
```

#### 6.3 PATCH `/api/v1/tags/:tagId/movies/:tagMovieId` **[未実装]**

- **概要**: タグ内の映画レコード（メモ・表示順など）を更新する。
- **認証**: 必須
- **ボディ例**

```json
{
  "note": "メモを更新しました",
  "position": 2
}
```

#### 6.4 DELETE `/api/v1/tags/:tagId/movies/:tagMovieId`

- **概要**: タグから映画を削除する。
- **認証**: 必須
- **権限**:
  - タグ作成者は全ての映画を削除可能
  - タグの `add_movie_policy` が `owner_only` の場合、作成者のみ削除可能
  - それ以外の場合、ユーザーは自分が追加した映画のみ削除可能
- **パスパラメータ**

| 名前          | 型   | 説明              |
|---------------|------|-------------------|
| `tagId`       | UUID | タグのID          |
| `tagMovieId`  | UUID | タグ映画レコードのID |

- **レスポンス例（204）**: `No Content`（ボディなし）

- **レスポンス例（403）**

```json
{
  "error": "forbidden"
}
```

- **レスポンス例（404）**

```json
{
  "error": "tag movie not found"
}
```

---

### 7. フォロー（Tag Followers）エンドポイント

#### 7.1 POST `/api/v1/tags/:tagId/follow`

- **概要**: 指定タグをフォローする。
- **認証**: 必須
- **挙動**
  - `tag_followers` に `(tag_id, user_id)` レコードを作成。
  - 既にフォロー済みの場合は 409 Conflict を返す。

- **レスポンス例（200）**

```json
{
  "message": "successfully followed"
}
```

- **レスポンス例（409）**

```json
{
  "error": "already following"
}
```

#### 7.2 DELETE `/api/v1/tags/:tagId/follow`

- **概要**: 指定タグのフォローを解除する。
- **認証**: 必須

- **レスポンス例（200）**

```json
{
  "message": "successfully unfollowed"
}
```

- **レスポンス例（409）**

```json
{
  "error": "not following"
}
```

#### 7.3 GET `/api/v1/tags/:tagId/followers`

- **概要**: タグのフォロワー一覧を取得する。
- **認証**: 不要
- **クエリパラメータ**

| 名前        | 型  | 必須 | 説明                                             |
|-------------|-----|------|--------------------------------------------------|
| `page`      | int | 任意 | ページ番号（デフォルト: 1）                     |
| `page_size` | int | 任意 | 1ページあたり件数（デフォルト: 20, 上限: 100） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "id": "user-uuid-1",
      "display_id": "cinephile_jane",
      "display_name": "cinephile_jane",
      "avatar_url": "https://images.example.com/avatar.jpg"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 1
}
```

#### 7.4 GET `/api/v1/tags/:tagId/follow-status`

- **概要**: 認証ユーザーがタグをフォローしているかチェックする。
- **認証**: 必須

- **レスポンス例（200）**

```json
{
  "is_following": true
}
```

---

### 8. 映画（Movies）エンドポイント

#### 8.1 GET `/api/v1/movies/search`

- **概要**: TMDB の検索結果（映画の候補一覧）を返す。
- **認証**: 不要
- **クエリパラメータ**

| 名前   | 型   | 必須 | 説明 |
|--------|------|------|------|
| `q`    | text | 任意 | 検索キーワード。空の場合は `items: []` / `total_count: 0` を返す |
| `page` | int  | 任意 | ページ番号（デフォルト: 1） |

- **リクエスト例**

```http
GET /api/v1/movies/search?q=spirited&page=1 HTTP/1.1
Host: localhost:8080
```

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "tmdb_movie_id": 129,
      "title": "千と千尋の神隠し",
      "original_title": "Spirited Away",
      "poster_path": "/path/to/poster.jpg",
      "release_date": "2001-07-20",
      "vote_average": 8.5
    }
  ],
  "page": 1,
  "total_count": 1
}
```

- **レスポンス例（500）**

```json
{
  "error": "failed to search movies"
}
```

#### 8.2 GET `/api/v1/movies/:tmdbMovieId`

- **概要**: 指定した TMDB 映画IDの詳細情報を取得する（キャッシュを内部で確保する）。
- **認証**: 不要
- **パスパラメータ**

| 名前           | 型  | 説明            |
|----------------|-----|-----------------|
| `tmdbMovieId`  | int | TMDB の映画ID   |

- **レスポンス例（200）**

```json
{
  "tmdb_movie_id": 129,
  "title": "千と千尋の神隠し",
  "original_title": "Spirited Away",
  "poster_path": "/path/to/poster.jpg",
  "release_date": "2001-07-20",
  "vote_average": 8.5,
  "overview": "不思議な世界に迷い込んだ少女の物語。",
  "genres": [
    { "id": 16, "name": "Animation" },
    { "id": 14, "name": "Fantasy" }
  ],
  "runtime": 125,
  "production_countries": [
    { "iso_3166_1": "JP", "name": "Japan" }
  ],
  "directors": ["Hayao Miyazaki"],
  "cast": [
    { "name": "Rumi Hiiragi", "character": "Chihiro Ogino" }
  ]
}
```

- **レスポンス例（400）**

```json
{
  "error": "invalid tmdb_movie_id"
}
```

- **レスポンス例（500）**

```json
{
  "error": "failed to get movie detail"
}
```

#### 8.3 GET `/api/v1/movies/:tmdbMovieId/tags`

- **概要**: 指定した映画が含まれる **公開タグ** の一覧を取得する（フォロワー数順）。
- **認証**: 不要
- **パスパラメータ**

| 名前           | 型  | 説明          |
|----------------|-----|---------------|
| `tmdbMovieId`  | int | TMDB の映画ID |

- **クエリパラメータ**

| 名前     | 型  | 必須 | 説明 |
|----------|-----|------|------|
| `limit`  | int | 任意 | 返却件数（デフォルト: 10、0以下の場合も 10 として扱う） |

- **レスポンス例（200）**

```json
{
  "items": [
    {
      "tag_id": "tag-uuid-1",
      "title": "ジブリの名作",
      "follower_count": 120,
      "movie_count": 32
    }
  ]
}
```

- **レスポンス例（400）**

```json
{
  "error": "invalid tmdb_movie_id"
}
```

- **レスポンス例（500）**

```json
{
  "error": "failed to get movie tags"
}
```

---

### 9. 今後の拡張候補

- **フィード系エンドポイント**
  - `GET /api/v1/feed/tags/popular` : フォロワー数順の人気タグ
  - `GET /api/v1/feed/tags/recent`  : 作成日時順の新着タグ
  - `GET /api/v1/feed/me`           : 自分がフォローしているタグの更新情報
- **通知・アクティビティ**
  - タグに映画が追加されたときの通知や、フォローされたときの通知を返すエンドポイント。
- **管理系（Admin）**
  - スパム報告、タグの非公開化／BANなどを行う管理者用 API。


