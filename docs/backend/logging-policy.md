# バックエンド ログ方針

このドキュメントは、cinetagバックエンドにおけるログ出力の方針を定義する。

## 1. 概要

### 1.1 ログの目的

| 目的 | 説明 |
|------|------|
| **障害調査** | 本番環境での問題発生時の原因特定 |
| **監視・アラート** | エラー率の上昇など異常検知 |
| **デバッグ** | 開発中の動作確認 |

### 1.2 使用ライブラリ

- **Go標準 `log/slog`** を使用
- 構造化ログ（JSON/テキスト形式）をサポート
- 実装: `internal/logger/logger.go`

## 2. 環境変数

| 変数名 | 値 | 説明 |
|--------|-----|------|
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | ログレベル（デフォルト: `info`） |
| `GIN_MODE` | `release` / `debug` | `release` 時はJSON形式、それ以外はテキスト形式 |

### 2.1 環境別の推奨設定

| 環境 | LOG_LEVEL | GIN_MODE | 出力形式 |
|------|-----------|----------|---------|
| 開発 | `debug` | `debug` | テキスト |
| 本番 | `info` | `release` | JSON |

## 3. ログレベルの使い分け

| レベル | 用途 | 例 |
|--------|------|-----|
| **DEBUG** | 開発時のみ必要な詳細情報 | Service/Repository層の開始ログ、変数の値 |
| **INFO** | 正常系の重要なイベント | Handler層の開始ログ、リクエストログ |
| **WARN** | 問題ではないが注意が必要な状況 | 非推奨APIの使用、リトライ発生 |
| **ERROR** | 処理が失敗した状況 | DB接続エラー、外部API失敗、パニック |

## 4. ログに含める情報

### 4.1 必須フィールド

すべてのログに以下のフィールドを含める:

| フィールド | 説明 |
|-----------|------|
| `time` | タイムスタンプ（自動付与） |
| `level` | ログレベル（自動付与） |
| `msg` | ログメッセージ |
| `request_id` | リクエスト識別子（UUID） |

### 4.2 コンテキスト情報

状況に応じて以下のフィールドを追加する:

| フィールド | 条件 | 説明 |
|-----------|------|------|
| `user_id` | 認証済みの場合 | 操作を行ったユーザーのID |
| `tag_id` | タグ操作時 | 対象タグのID |
| `error` | エラー発生時 | エラー内容 |

### 4.3 出力例

**開発環境（テキスト形式）:**

```
time=2026-01-13T10:00:00.000+09:00 level=INFO msg="handler.GetTagDetail started" request_id=abc-123 user_id=user_456 tag_id=tag_789
time=2026-01-13T10:00:00.001+09:00 level=DEBUG msg="service.GetTagDetail started" request_id=abc-123 tag_id=tag_789
time=2026-01-13T10:00:00.015+09:00 level=INFO msg=request request_id=abc-123 method=GET path=/api/v1/tags/tag_789 status=200 latency=15.234ms
```

**本番環境（JSON形式）:**

```json
{"time":"2026-01-13T01:00:00.000Z","level":"INFO","msg":"handler.GetTagDetail started","request_id":"abc-123","user_id":"user_456","tag_id":"tag_789"}
{"time":"2026-01-13T01:00:00.015Z","level":"INFO","msg":"request","request_id":"abc-123","method":"GET","path":"/api/v1/tags/tag_789","status":200,"latency":"15.234ms"}
```

## 5. セキュリティ

### 5.1 ログに含めてはいけない情報

以下の情報は **絶対にログに出力しない**:

| 種類 | 例 |
|------|-----|
| **認証情報** | JWTトークン、パスワード、APIキー |
| **個人情報** | メールアドレス、電話番号 |
| **センシティブデータ** | クレジットカード番号 |

### 5.2 注意が必要な情報

以下の情報は必要最小限に留める:

| 種類 | 対応 |
|------|------|
| Clerk User ID | `user_id` としてログ出力可（内部ID） |
| リクエストボディ | 原則ログ出力しない |
| クエリパラメータ | パスに含める場合は注意 |

## 6. レイヤー別の方針

### 6.1 Middleware層

| 種類 | レベル | 内容 |
|------|--------|------|
| リクエストログ | INFO | method, path, status, latency, request_id, client_ip |
| パニックリカバリー | ERROR | エラー内容、スタックトレース |

実装: `internal/middleware/request_logger.go`, `recovery.go`

### 6.2 Handler層

| タイミング | レベル | 内容 |
|-----------|--------|------|
| 開始時 | **INFO** | メソッド名、request_id、user_id（認証済みの場合）、主要パラメータ |
| エラー時 | - | ログ出力しない（発生源でログ出力済み） |

**例:**

```go
func (h *TagHandler) GetTagDetail(c *gin.Context) {
    requestID := middleware.GetRequestID(c)
    tagID := c.Param("tagId")
    
    // 開始ログ（INFO）
    h.logger.Info("handler.GetTagDetail started",
        slog.String("request_id", requestID),
        slog.String("tag_id", tagID),
    )
    
    // 認証済みの場合は user_id も含める
    if user := getUserFromContext(c); user != nil {
        h.logger.Info("handler.GetTagDetail started",
            slog.String("request_id", requestID),
            slog.String("user_id", user.ID),
            slog.String("tag_id", tagID),
        )
    }
    // ...
}
```

### 6.3 Service層

| タイミング | レベル | 内容 |
|-----------|--------|------|
| 開始時 | **DEBUG** | メソッド名、主要パラメータ |
| エラー時 | - | ログ出力しない（発生源でログ出力済み） |

**例:**

```go
func (s *tagService) GetTagDetail(ctx context.Context, tagID string, viewerUserID *string) (*TagDetail, error) {
    // 開始ログ（DEBUG）
    s.logger.Debug("service.GetTagDetail started",
        slog.String("tag_id", tagID),
    )
    // ...
}
```

### 6.4 Repository層

| タイミング | レベル | 内容 |
|-----------|--------|------|
| 開始時 | **DEBUG** | メソッド名、主要パラメータ |
| エラー時 | **ERROR** | エラー内容（発生源としてログ出力） |

**例:**

```go
func (r *tagRepository) FindByID(ctx context.Context, id string) (*model.Tag, error) {
    // 開始ログ（DEBUG）
    r.logger.Debug("repository.FindByID started",
        slog.String("tag_id", id),
    )
    
    var tag model.Tag
    if err := r.db.First(&tag, "id = ?", id).Error; err != nil {
        // エラーログ（ERROR）- 発生源で出力
        r.logger.Error("repository.FindByID failed",
            slog.String("tag_id", id),
            slog.Any("error", err),
        )
        return nil, fmt.Errorf("repository.FindByID: %w", err)
    }
    return &tag, nil
}
```

## 7. エラーログの方針

### 7.1 基本方針: 発生源で出力

エラーは **発生した最初のレイヤーでログ出力** し、上位レイヤーではログ出力しない。

```
[Repository] エラー発生 → ERROR ログ出力 → エラーをラップして返す
     ↓
[Service] エラー受け取り → ログ出力しない → エラーをそのまま返す
     ↓
[Handler] エラー受け取り → ログ出力しない → HTTPエラーレスポンスを返す
```

### 7.2 エラーのラップ

エラーを返す際は `fmt.Errorf` でラップし、発生元を特定できるようにする:

```go
// Good
return nil, fmt.Errorf("repository.FindByID: %w", err)

// Bad
return nil, err
```

### 7.3 センチネルエラーの扱い

ビジネスロジックのエラー（`ErrTagNotFound` など）は **ログ出力しない**。
これらは正常なフローの一部であり、エラーではない。

```go
// Service層
if errors.Is(err, gorm.ErrRecordNotFound) {
    // ログ出力しない（正常なフロー）
    return nil, ErrTagNotFound
}
```

## 8. コード例

### 8.1 良い例

```go
// Handler層: INFO ログ、user_id を含める
func (h *TagHandler) CreateTag(c *gin.Context) {
    requestID := middleware.GetRequestID(c)
    user := getUserFromContext(c)
    
    h.logger.Info("handler.CreateTag started",
        slog.String("request_id", requestID),
        slog.String("user_id", user.ID),
    )
    // ...
}

// Repository層: エラー発生源でログ出力
func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
    if err := r.db.Create(tag).Error; err != nil {
        r.logger.Error("repository.Create failed",
            slog.String("tag_id", tag.ID),
            slog.Any("error", err),
        )
        return fmt.Errorf("repository.Create: %w", err)
    }
    return nil
}
```

### 8.2 悪い例

```go
// Bad: fmt.Println を使用
fmt.Println("tagID", tagID)

// Bad: センシティブ情報をログ出力
h.logger.Info("auth", slog.String("token", authHeader))

// Bad: 複数レイヤーで同じエラーをログ出力
// Repository層
r.logger.Error("failed", slog.Any("error", err))
// Service層（重複！）
s.logger.Error("failed", slog.Any("error", err))
// Handler層（重複！）
h.logger.Error("failed", slog.Any("error", err))
```

## 9. 実装ファイル

| ファイル | 説明 |
|---------|------|
| `internal/logger/logger.go` | ロガー初期化 |
| `internal/middleware/request_logger.go` | リクエストログミドルウェア |
| `internal/middleware/recovery.go` | パニックリカバリーミドルウェア |
| `router/dependencies.go` | ロガーのDI設定 |
