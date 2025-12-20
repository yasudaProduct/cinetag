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

---

### 4. 認証ユーザー系エンドポイント

#### 4.1 GET `/api/v1/me` **[未実装]**

- **概要**: ログイン中ユーザー（`users` テーブル）の情報を取得する。
- **認証**: 必須
- **レスポンス例（200）**

```json
{
  "id": "b1e4f0e8-1234-5678-9012-abcdefabcdef",
  "clerk_user_id": "user_2aBcDeFgHiJk",
  "username": "cinephile_jane",
  "display_name": "cinephile_jane",
  "email": "jane@example.com",
  "avatar_url": "https://images.example.com/avatar.jpg",
  "bio": "映画が好きです",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### 4.2 GET `/api/v1/me/tags` **[未実装]**

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

#### 4.3 GET `/api/v1/me/following-tags` **[未実装]**

- **概要**: ログインユーザーがフォローしているタグ一覧を取得する。
- **認証**: 必須
- **クエリ / レスポンス形式**: `/api/v1/me/tags` と同様（`items` の中身が「フォロー中タグ」となる）。

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
- **認証**: なし（ただし非公開タグアクセス時は `AuthMiddleware` で作成者チェックを行う想定）
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
  "movie_count": 32,
  "follower_count": 120,
  "owner": {
    "id": "user-uuid-1",
    "username": "cinephile_jane",
    "display_name": "cinephile_jane"
  },
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
  "is_public": true
}
```

- **バリデーション**
  - `title`: 1〜100文字
  - `description`: 最大500文字

- **レスポンス例（201）**

```json
{
  "id": "tag-uuid-1",
  "title": "ジブリの名作",
  "description": "スタジオジブリの不朽の名作アニメ。",
  "cover_image_url": "https://example.com/cover.jpg",
  "is_public": true,
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
  "is_public": false
}
```

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

#### 6.4 DELETE `/api/v1/tags/:tagId/movies/:tagMovieId` **[未実装]**

- **概要**: タグから映画を外す。
- **認証**: 必須
- **レスポンス**: `204 No Content`

---

### 7. フォロー（Tag Followers）エンドポイント

#### 7.1 POST `/api/v1/tags/:tagId/follow` **[未実装]**

- **概要**: 指定タグをフォローする。
- **認証**: 必須
- **挙動**
  - `tag_followers` に `(tag_id, user_id)` レコードを作成（既に存在する場合は何もしない idempotent な処理）。

- **レスポンス例（200 or 201）**

```json
{
  "tag_id": "tag-uuid-1",
  "user_id": "user-uuid-1",
  "created_at": "2025-01-01T12:00:00Z"
}
```

#### 7.2 DELETE `/api/v1/tags/:tagId/follow` **[未実装]**

- **概要**: 指定タグのフォローを解除する。
- **認証**: 必須
- **レスポンス**: `204 No Content`

#### 7.3 GET `/api/v1/tags/:tagId/followers` **[未実装]**

- **概要**: タグのフォロワー一覧、またはフォロワー数を取得する。
- **認証**: 任意（一覧を隠したい場合は必須にする）
- **クエリ例**
  - `?page=1&page_size=20`（一覧）
  - `?summary=1`（件数だけ返すなど、将来拡張用）

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


