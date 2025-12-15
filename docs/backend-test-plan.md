## バックエンド テスト計画（Unit Test Plan）

このドキュメントは `cinetag` のバックエンド（Go / Gin / GORM）の **単体テスト（unit test）を中心**に、どこをどうテストしていくかの方針と、実施順序をまとめます。

---

## タスク分析

- **主要タスク**
  - バックエンドの単体テストを継続的に追加できるように、レイヤー別のテスト方針・優先順位・粒度・ツール・運用を定義する。
- **技術スタック制約**
  - Go（`go test`）
  - Gin（HTTP ハンドラのテストは `net/http/httptest` + Gin のテストコンテキスト）
  - GORM（DBアクセス層は“純粋な単体”で完結しにくい）
  - Clerk JWT（`internal/middleware/clerk_jwt.go` に最小実装の検証器がある）
- **重要要件 / 制約**
  - 単体テストの主目的は **ビジネスロジック・入力バリデーション・エラー変換の安定化**。
  - 外部依存（DB、JWKS取得、外部API）は **基本は stub/mock** で分離。
  - DB を叩く必要がある箇所は **結合テスト（integration）として別枠**に切る。
- **潜在的な課題**
  - `NewAuthMiddleware` が環境変数と JWKS フェッチを内包しており、テストが少し重くなりがち。
  - `TagService.AddMovieToTag` に goroutine によるベストエフォート処理があり、テストで非決定性が出やすい。
  - repository（GORM）を unit だけで完結させるには mock が難しく、テスト設計の切り分けが必要。
- **実施ステップ（推奨順）**
  - **Phase 0**: テスト土台（共通ヘルパ、fake/mock 方針、`go test` コマンド、CI想定）
  - **Phase 1**: `service/` の unit test（最優先）
  - **Phase 2**: `handler/` の unit test（HTTP 入出力・バリデーション・エラーマッピング）
  - **Phase 3**: `middleware/` の unit test（認証ヘッダ処理、ユーザー同期、JWT検証器）
  - **Phase 4**: `repository/` のテスト（必要なら integration として DB ありで）

---

## 1. テストのゴール / 非ゴール

- **ゴール**
  - 仕様の破壊的変更を早期に検知する（回帰防止）
  - API が返すステータスコード/エラー形式の一貫性を担保する
  - 主要ユースケース（タグ作成・公開タグ一覧・タグへの映画追加・ユーザー同期）を安全に改修できる状態にする

- **非ゴール**
  - E2E（フロントエンド含むブラウザテスト）は本計画の対象外
  - Clerk 本番環境や外部 API への実通信を伴うテストは原則行わない

---

## 2. テストレベル（テストピラミッド）

- **Unit（最優先）**
  - `service/`（ビジネスロジック）
  - `handler/`（入力バリデーション、HTTP I/O、エラーマッピング）
  - `middleware/`（ヘッダ処理、コンテキスト設定、ユーザー同期呼び出し）

- **Integration（必要最小限）**
  - `repository/`（GORM + PostgreSQL の実挙動検証）
  - migration（`src/cmd/migrate` の結果の検証）

---

## 3. レイヤー別テスト方針

### 3.1 `internal/service`（最優先）

- **狙い**: 仕様の中心（入力チェック、ドメイン制約、エラー変換）を unit で固定する。
- **依存の扱い**
  - `repository.*` は fake/mock（インターフェースが用意されているため差し替え可能）
  - `MovieService` は fake（もしくは `nil` 注入で非同期処理を発火させない）
- **主要対象**
  - `TagService`
    - `CreateTag`: `TagRepository.Create` 呼び出しと戻り値
    - `AddMovieToTag`: 入力バリデーション、`FindByID` エラー→`ErrTagNotFound` 変換、重複→`ErrTagMovieAlreadyExists` 変換、`IncrementMovieCount` の呼び出し
    - `ListPublicTags`: page/pageSize 正規化、`ListPublicTags` の filter 生成、画像生成（`movieService` が nil の場合/ある場合）
  - `UserService`（ユーザー同期）
    - `EnsureUser`: 既存ユーザー/新規作成/DBエラーの分岐

- **注意（非同期処理）**
  - `AddMovieToTag` の goroutine は **unit test では原則発火させない**（`movieService=nil` でテスト）。
  - 非同期まで検証したい場合は「チャネルで同期できる fake」を使い、`t.Cleanup` で待ち合わせしてから検証する。

### 3.2 `internal/handler`

- **狙い**: HTTP 入力（path/query/body）→ Service 呼び出し → HTTP 出力（status/json）の仕様を固定する。
- **方法**
  - `net/http/httptest` + Gin ルータ（もしくは handler を直接呼ぶ）
  - service は fake/mock で差し替え
- **主要対象（例）**
  - `TagHandler.CreateTag`
    - JSON の binding エラー → `400`
    - `user` が context にない → `401`
    - タイトル長/説明長の制約 → `400`
    - service エラー → `500`
    - 正常系 → `201` + 返却フィールド
  - `TagHandler.AddMovieToTag`
    - `tmdb_movie_id` / `position` バリデーション → `400`
    - service の `ErrTagNotFound` → `404` / `ErrTagPermissionDenied` → `403` / `ErrTagMovieAlreadyExists` → `409`

### 3.3 `internal/middleware`

- **狙い**: 認証ヘッダ処理と、認証済みユーザーのコンテキスト格納を固定する。
- **主要対象**
  - `NewAuthMiddleware`
    - Authorization ヘッダなし/形式不正 → `401`
    - token 文字列空 → `401`
    - Verify エラー → `401`
    - claims から `sub`/`email` が取れない → `401`
    - `EnsureUser` エラー → `500`
    - 正常系 → `c.Set("user", ...)` が設定され後続に進む

- **JWT 検証器（`clerk_jwt.go`）の unit test 方針**
  - **JWKS は `httptest.Server` でローカル配信**し、ネットワーク依存を排除する。
  - テスト用に RSA 鍵を生成して JWKS（`n`/`e`）を返し、対応する秘密鍵で RS256 署名した JWT を生成する。
  - 検証観点
    - `alg != RS256` / `kid` なし / 署名不一致
    - `exp`/`nbf` の挙動
    - `iss`/`aud` 指定時の一致/不一致
    - JWKS のキャッシュ（TTL 内の再取得抑制）は“過剰に厳密”にやらず、主要分岐のみ

> 補足: `NewAuthMiddleware` は環境変数を読むため、テストでは `t.Setenv("CLERK_JWKS_URL", server.URL)` を使用して制御する。

### 3.4 `internal/repository`（原則 integration として扱う）

- **理由**: GORM の挙動（WHERE/ORDER/Unique制約など）は mock で再現するより **実DBで検証した方が安く確実**。
- **方針**
  - unit（mock）に固執せず、repository は **PostgreSQL を起動しての integration test** を基本とする
  - テストはトランザクションで巻き戻す（もしくはテーブル truncate）
- **実行方法候補**
  - Docker（`compose.yml` の postgres）をテスト実行前に起動しておく運用
  - もしくは将来的に `testcontainers-go` 等で自動起動（導入は後回しでOK）

---

## 4. テストの配置ルール / 命名

- **ファイル**: `*_test.go`
- **パッケージ**
  - 基本は `package <対象>`（同一パッケージ）で開始
  - 依存の切り方を厳密にしたい場合のみ `package <対象>_test` を検討
- **スタイル**
  - テーブル駆動テスト（table-driven tests）
  - 重要な分岐は `t.Run` で名前を明確化
  - `t.Parallel()` は共有状態がないテストのみ

---

## 5. 依存の差し替え（fake/mock）方針

- **第一選択**: 手書き fake（小さく、読みやすい）
- **第二選択**: `go.uber.org/mock` による生成（必要になった時点で導入）
- **禁止**: repository の“振る舞い”を unit で過剰に mock して、実DBとの差分が大きくなること

---

## 6. 実行コマンド（ローカル/CI想定）

- **全体**
  - `go test ./...`
- **レース検知（余裕がある時）**
  - `go test -race ./...`
- **カバレッジ**
  - `go test ./... -coverprofile=coverage.out`

---

## 7. 優先順位（最初に書くべきテスト）

- **P0（最優先）**
  - `TagService.AddMovieToTag` の主要分岐
  - `TagHandler.CreateTag` のバリデーション/認可（context user）
  - `AuthMiddleware` の Authorization 分岐（401/500/成功）

- **P1**
  - `TagService.ListPublicTags`（ページング/ソート/画像組み立て）
  - `UserService.EnsureUser`（同期の根幹）

- **P2**
  - `ClerkJWTValidator` の主要分岐（JWKS + 署名 + exp/nbf + iss/aud）

---

## 8. 進め方（小さく確実に増やす運用）

- まず `service/` を unit で固める（最も ROI が高い）
- 次に `handler/` で API の入力/出力仕様を固定する
- 最後に `middleware/` と `repository/`（必要なら integration）を追加する

---

## 9. 完了条件（Definition of Done）

- `apps/backend` で `go test ./...` が安定して通る
- P0 の対象に対して **成功/失敗の主要分岐**がテストで覆われている
- 新規のユースケース追加時に、同じ型のテストを迷わず追加できる（配置・命名・fake方針が共有されている）
