## 映画情報（TMDB）連携仕様

このドキュメントでは、`cinetag` における **映画情報の取得元としての TMDB API の扱い方** と、バックエンドでのキャッシュ・API連携の設計方針をまとめます。

---

## 1. 概要

- **目的**
  - `cinetag` における映画情報は、TMDB API から取得したデータをもとに表示する。
  - アプリケーション内では、**`movie_cache` テーブルを中心としたキャッシュ戦略**を採用し、TMDB へのアクセス頻度を抑えつつ、十分な鮮度の映画情報を提供する。
- **対象範囲**
  - バックエンド（`apps/backend`）における TMDB API の利用方法。
  - `movie_cache` テーブルの使い方と、既存 API との連携方法。
- **非対象**
  - フロントエンドが直接 TMDB API を呼び出す場合の詳細（必要があれば別ドキュメントで定義）。

---

## 2. 使用する TMDB API

### 2.1 ベース情報

- **ベースURL**
  - TMDB v3 API を前提とする: `https://api.themoviedb.org/3`
- **認証**
  - 環境変数 `TMDB_API_KEY` に設定された API キーを利用する。
  - 実装では、TMDB の仕様に従い `api_key` クエリパラメータ、または認証ヘッダを使用する。

### 2.2 想定する主なエンドポイント

- **映画詳細取得**
  - `GET /movie/{movie_id}`
  - 用途: `movie_cache` の作成・更新。
- **（拡張候補）映画検索**
  - `GET /search/movie`
  - 用途: フロントエンドからの検索要求をバックエンド経由で TMDB に委譲する場合に利用。

> 初期実装では **`/movie/{movie_id}` のみ必須** とし、検索 API は拡張候補とする。

---

## 3. データモデルとマッピング

### 3.1 `movie_cache` テーブル

`docs/data/database-schema.md` で定義済みの `movie_cache` テーブルを、TMDB レスポンスのキャッシュとして利用する。

- **テーブル構造（抜粋）**

| カラム名         | 型                | 説明                           |
|------------------|-------------------|--------------------------------|
| `tmdb_movie_id`  | INTEGER (PK)      | TMDb 映画 ID                   |
| `title`          | TEXT              | 映画タイトル（言語別）         |
| `original_title` | TEXT              | 原題                           |
| `poster_path`    | TEXT              | ポスター画像パス               |
| `backdrop_path`  | TEXT              | 背景画像パス                   |
| `release_date`   | DATE              | 公開日                         |
| `vote_average`   | NUMERIC(3,1)      | TMDB 平均スコア                |
| `overview`       | TEXT              | あらすじ                       |
| `genres`         | JSONB             | ジャンル配列                   |
| `runtime`        | INTEGER           | 上映時間（分）                 |
| `cached_at`      | TIMESTAMPTZ       | キャッシュ作成日時             |
| `expires_at`     | TIMESTAMPTZ       | キャッシュ有効期限（+7 日）   |

### 3.2 TMDB レスポンスとのマッピング

- **TMDB `/movie/{movie_id}` → `movie_cache`**

| TMDB フィールド  | `movie_cache` カラム | 変換ルール                                     |
|------------------|----------------------|------------------------------------------------|
| `id`             | `tmdb_movie_id`      | 整数をそのまま格納                             |
| `title`          | `title`              | `language` パラメータに基づくローカライズ済タイトル |
| `original_title` | `original_title`     | 文字列そのまま                                 |
| `poster_path`    | `poster_path`        | 文字列（先頭の `/` を含むパス）               |
| `backdrop_path`  | `backdrop_path`      | 同上                                           |
| `release_date`   | `release_date`       | `YYYY-MM-DD` → `DATE` へ変換                  |
| `vote_average`   | `vote_average`       | 小数を NUMERIC(3,1) へ                         |
| `overview`       | `overview`           | 文字列そのまま                                 |
| `genres`         | `genres`             | `{id, name}` 配列を JSONB でそのまま保存      |
| `runtime`        | `runtime`            | 分単位の整数                                   |

### 3.3 ローカライズポリシー

- **デフォルト言語**: `ja-JP`
  - TMDB の `language` パラメータに `ja-JP` を指定して取得を試みる。
- **フォールバック**
  - もし `title` や `overview` 等が空、または取得に失敗した場合、
    **フォールバックとして `en-US` でもう一度取得して必要なフィールドを補完**する方針とする。
  - 実装では 2 回呼び出すか、1 回目のレスポンスの不足分を 2 回目で補うかなど、最終的な挙動は実装時に検討。

---

## 4. キャッシュ戦略

### 4.1 TTL と有効期限

- **TTL**
  - `movie_cache.expires_at` は、`cached_at + 7 days` をデフォルトとする（DB 定義に準拠）。
- **有効／無効判定**
  - 現在時刻 `now()` が `expires_at` より前: **有効キャッシュ**
  - `now() >= expires_at`: **期限切れキャッシュ**

### 4.2 キャッシュ取得アルゴリズム（共通）

映画情報が必要な場面では、以下の共通ロジックを利用する。

1. `movie_cache` から `tmdb_movie_id` でレコードを取得する。
2. レコードが存在し、かつ `expires_at > now()` であればそのまま利用する。
3. レコードが存在しない、または `expires_at <= now()` の場合:
   1. TMDB `/movie/{movie_id}` を呼び出す。
   2. 正常なレスポンスが得られた場合:
      - `movie_cache` に **UPSERT**（存在しなければ INSERT、あれば UPDATE）。
      - `cached_at` を現在時刻に更新し、`expires_at` を `cached_at + 7 days` に設定。
   3. TMDB 呼び出しが失敗した場合:
      - 旧キャッシュがあれば **期限切れでも暫定的に利用**するか、利用しないかを API ごとに定義（下記参照）。
      - エラー内容をログに記録する。
      - クライアントへのレスポンスには、映画情報を欠落させるか、簡易なエラーフラグを含める。

> 初期実装では「期限切れキャッシュは利用しない（存在しても `movie` サブオブジェクトを `null` にする）」方針でもよいが、UX とのトレードオフのため後日調整可能とする。

---

## 5. 既存 API への組み込み

`docs/api/api-spec.md` に定義済みのエンドポイントに、TMDB 映画情報をどのように組み込むかを定義する。

### 5.1 `GET /api/v1/tags/:tagId/movies`

- **目的**
  - 指定タグに含まれる映画の一覧を返すと同時に、各映画の TMDB 情報を `movie` サブオブジェクトとして含める。

- **レスポンス（抜粋）**

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
  ]
}
```

- **組み込みフロー**

1. `tag_movies` から対象タグのレコードをページング取得する。
2. 各 `TagMovie.TmdbMovieID` に対して、上記「キャッシュ取得アルゴリズム」を適用する。
3. 取得した `MovieCache` をもとに、レスポンスの `movie` オブジェクトを構築する:
   - `title` ← `MovieCache.Title`
   - `original_title` ← `MovieCache.OriginalTitle`
   - `poster_path` ← `MovieCache.PosterPath`
   - `release_date` ← `MovieCache.ReleaseDate` を `YYYY-MM-DD` 文字列に変換
   - `vote_average` ← `MovieCache.VoteAverage`
4. TMDB 呼び出しに失敗し、かつ `movie_cache` も存在しない場合:
   - `movie` フィールドを `null` または省略として返す。
   - ログには `tag_id`, `tmdb_movie_id`, エラー内容を記録する。

### 5.2 `GET /api/v1/tags`

- **目的**
  - 公開タグ一覧のカードを表示する際に、サムネイル用の映画ポスター画像を `images` 配列で返す。

- **レスポンス（抜粋）**

```json
{
  "items": [
    {
      "id": "tag-uuid-1",
      "title": "90年代SFクラシック",
      "images": [
        "https://image.tmdb.org/t/p/w400/poster1.jpg",
        "https://image.tmdb.org/t/p/w400/poster2.jpg",
        "https://image.tmdb.org/t/p/w400/poster3.jpg",
        "https://image.tmdb.org/t/p/w400/poster4.jpg"
      ]
    }
  ]
}
```

- **組み込みフロー**

1. タグ一覧を取得後、各タグについて `tag_movies` から先頭 N 件（例: 4 件）の `tmdb_movie_id` を取得する。
2. 各 `tmdb_movie_id` について「キャッシュ取得アルゴリズム」を適用する。
3. `MovieCache.PosterPath` を `TMDB_IMAGE_BASE_URL`（例: `https://image.tmdb.org/t/p/w400`）と結合してフル URL を生成する:
   - `TMDB_IMAGE_BASE_URL + poster_path`
4. 有効な `PosterPath` がある映画のみを `images` 配列として返す。
5. TMDB 呼び出しに失敗しても、タグ自体の情報は返却し、`images` が空配列になることを許容する。

### 5.3 `POST /api/v1/tags/:tagId/movies`

- **目的**
  - ユーザーがタグに映画を追加する際、TMDB の映画 ID を登録する。

- **リクエスト（抜粋）**

```json
{
  "tmdb_movie_id": 12345,
  "note": "子どもの頃に観た思い出の作品",
  "position": 1
}
```

- **TMDB データ利用ポリシー**

1. `tag_movies` レコードの作成自体は、**TMDB への問い合わせが失敗しても継続する**（ID ベースの関連なので、柔軟性を優先）。
2. 可能であれば、作成時に非同期で TMDB `/movie/{movie_id}` を呼び出し、`movie_cache` に **先行してキャッシュを作成**する（UX 向上）。
3. 作成時に TMDB 呼び出しがエラーとなった場合:
   - `tag_movies` の作成は成功とし、以降の `GET /tags/:tagId/movies` 呼び出し時に再度キャッシュ取得を試みる。
   - エラーはログのみに記録する。

---

## 6. 環境変数・設定

- **環境変数一覧（例）**
  - `TMDB_API_KEY`: TMDB API キー（必須）
  - `TMDB_BASE_URL`: TMDB API ベース URL（デフォルト: `https://api.themoviedb.org/3`）
  - `TMDB_IMAGE_BASE_URL`: 画像 URL ベース（例: `https://image.tmdb.org/t/p/w400`）
  - `TMDB_DEFAULT_LANGUAGE`: 既定言語（デフォルト: `ja-JP`）

- **読み込み場所**
  - 将来的に `internal/config` パッケージで一元管理する方針（`docs/architecture/backend-architecture.md` の想定に準拠）。

---

## 7. エラーハンドリング・ログ

### 7.1 TMDB API エラー

- **タイムアウト・HTTP 5xx・ネットワークエラーなど**
  - アプリケーションログに `tmdb_error`, `movie_id`, `endpoint`, `status_code` などを記録する。
  - クライアントには「映画情報の一部が取得できなかった」状態として、`movie` フィールドの欠落や空配列などで表現する。

- **HTTP 404（映画が存在しない）**
  - `tag_movies` レコードは残すが、`movie_cache` には作成しない。
  - 必要であれば将来的に「削除済み映画」を表すフラグを追加する拡張も検討する。

### 7.2 レート制限（429）

- 429 応答を受けた場合:
  - 直近の呼び出しは失敗として扱い、キャッシュがあればそれを利用（または映画情報を欠落させる）。
  - ログにレート制限発生を記録し、必要ならアラート対象とする。

---

## 8. セキュリティと利用ポリシー

- TMDB API キーは **バックエンドのみで保持し、フロントエンドやクライアントには一切返さない**。
- TMDB の利用規約に従い、著作権表示やクレジット表記が必要な場合はフロントエンド側で対応する（別ドキュメントで管理）。
- 不正使用を防ぐため、TMDB 呼び出し部分には適切なタイムアウト・リトライ制御を導入する（実装詳細は別途）。

---

## 9. 今後の拡張候補

- **検索 API のバックエンド実装**
  - `GET /api/v1/movies/search?q=...` などで TMDB `/search/movie` をラップし、フロントエンドから直接 TMDB を叩かずに済むようにする。
- **バッチ更新ジョブ**
  - 人気タグに含まれる映画の `movie_cache` を定期的に更新するバッチ処理（例: 1 日 1 回）。
- **詳細情報の拡張**
  - 監督・キャスト・予告編動画 URL など、TMDB の他エンドポイントを利用して情報を拡張する。

---


