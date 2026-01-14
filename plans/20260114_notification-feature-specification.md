# 通知機能 仕様書

**作成日**: 2026-01-14  
**ステータス**: 仕様策定中

---

## 1. 概要

### 1.1 目的

ユーザーに対して重要なイベント（タグへの映画追加、フォロー通知など）をリアルタイムまたは非同期で通知することで、ユーザーエンゲージメントとアプリケーションの利用価値を向上させる。

### 1.2 スコープ

- **実装対象**
  - 通知設定管理（ユーザーごと、管理者による全体設定）
  - 通知イベント定義と拡張可能な設計
  - 通知配信チャネルの実装（アプリ内通知、メール通知、Push通知）
  - 通知履歴の管理と表示
- **非対象（将来実装）**
  - リアルタイム通知配信（WebSocket/SSE）
  - 通知のグルーピング・要約機能
  - 通知の既読一括操作

### 1.3 設計原則

1. **拡張性**: 通知種別と通知対象イベントは運用中に容易に追加・削除できること
2. **柔軟性**: ユーザーごとに細かく通知設定をカスタマイズできること
3. **パフォーマンス**: 非同期処理により本体処理をブロックしないこと
4. **保守性**: レイヤードアーキテクチャに準拠し、テスタビリティを確保すること

---

## 2. 通知種別（配信チャネル）

### 2.1 通知チャネル一覧

| チャネルID | 名称 | 説明 | 初期実装 |
|-----------|------|------|---------|
| `in_app` | アプリ内通知 | Webアプリ内で通知バッジ・一覧表示 | ✅ Phase 1 |
| `email` | メール通知 | ユーザーのメールアドレスに配信 | ✅ Phase 2 |
| `push` | Push通知 | モバイルアプリへのプッシュ通知 | ⏳ 将来実装 |

### 2.2 チャネル拡張方針

- データベースに `notification_channels` マスターテーブルを設け、動的に追加可能にする
- 各チャネルの実装は `NotificationService` のインターフェースを介して疎結合に保つ
- 新規チャネル追加時は、対応する配信ハンドラーを実装し、DIコンテナに登録する

---

## 3. 通知対象イベント

### 3.1 イベント一覧

| イベントID | 名称 | 説明 | 通知対象者 | 初期実装 |
|-----------|------|------|-----------|---------|
| `tag_movie_added` | タグに映画追加 | 自分のタグまたはフォロー中のタグに、自分以外のユーザーが映画を追加 | タグ作成者 & タグフォロワー（追加者除く） | ✅ Phase 1 |
| `tag_followed` | タグフォロー | 自分の作成したタグがフォローされた | タグ作成者 | ✅ Phase 1 |
| `tag_liked` | タグいいね | 自分の作成したタグがいいねされた | タグ作成者 | ⏳ 将来実装 |
| `user_followed` | ユーザーフォロー | 自分がフォローされた | フォローされたユーザー | ✅ Phase 1 |

### 3.2 イベント拡張方針

- データベースに `notification_event_types` マスターテーブルを設け、動的に追加可能にする
- イベント発生時は汎用的な `NotificationService.CreateNotification()` を呼び出す
- 各イベントの通知生成ロジックは、イベントタイプごとの Factory パターンで実装

### 3.3 イベントペイロード設計

各イベントには、通知文生成に必要な情報をJSONとして格納する。

**例: `tag_movie_added`**
```json
{
  "tag_id": "uuid",
  "tag_title": "アクション映画コレクション",
  "movie_id": 12345,
  "movie_title": "ダークナイト",
  "added_by_user_id": "uuid",
  "added_by_display_name": "山田太郎"
}
```

---

## 4. 通知設定（Notification Settings）

### 4.1 設定レベル

通知設定は以下の2レベルで管理する：

1. **グローバル設定（管理者）**: 全ユーザー共通のデフォルト設定
2. **ユーザー個別設定**: ユーザーごとのカスタム設定（グローバル設定を上書き）

### 4.2 設定項目

各ユーザーは、**イベントタイプ × 配信チャネル** の組み合わせごとに通知の有効/無効を設定できる。

**設定例**:
| イベント | アプリ内 | メール | Push |
|---------|---------|--------|------|
| タグに映画追加 | ✅ ON | ✅ ON | ❌ OFF |
| タグフォロー | ✅ ON | ❌ OFF | ❌ OFF |
| ユーザーフォロー | ✅ ON | ✅ ON | ✅ ON |

### 4.3 デフォルト設定

- 新規ユーザー登録時は、全イベントの**アプリ内通知のみ有効**とする
- メール・Push通知はデフォルト無効（オプトイン方式）

---

## 5. データベース設計

### 5.1 新規テーブル

#### 5.1.1 `notification_channels`（通知チャネルマスター）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | TEXT | NO | - | チャネルID（例: `in_app`, `email`, `push`） |
| `name` | TEXT | NO | - | 表示名（例: "アプリ内通知"） |
| `description` | TEXT | YES | - | 説明 |
| `is_active` | BOOLEAN | NO | `true` | 有効フラグ |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |

**主キー**: `id`

#### 5.1.2 `notification_event_types`（通知イベントタイプマスター）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | TEXT | NO | - | イベントID（例: `tag_movie_added`） |
| `name` | TEXT | NO | - | 表示名（例: "タグに映画が追加されました"） |
| `description` | TEXT | YES | - | 説明 |
| `is_active` | BOOLEAN | NO | `true` | 有効フラグ |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |

**主キー**: `id`

#### 5.1.3 `notification_settings`（ユーザー通知設定）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `user_id` | UUID | NO | - | ユーザーID |
| `event_type_id` | TEXT | NO | - | イベントタイプID |
| `channel_id` | TEXT | NO | - | チャネルID |
| `is_enabled` | BOOLEAN | NO | `true` | 通知有効フラグ |
| `updated_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 更新日時 |

**主キー**: `(user_id, event_type_id, channel_id)`  
**外部キー**:
- `user_id` → `users(id)` ON DELETE CASCADE
- `event_type_id` → `notification_event_types(id)` ON DELETE CASCADE
- `channel_id` → `notification_channels(id)` ON DELETE CASCADE

**インデックス**:
- `idx_notification_settings_user` ON `(user_id)`

#### 5.1.4 `notifications`（通知履歴）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | 通知ID |
| `user_id` | UUID | NO | - | 通知対象ユーザーID |
| `event_type_id` | TEXT | NO | - | イベントタイプID |
| `title` | TEXT | NO | - | 通知タイトル（最大100文字） |
| `message` | TEXT | NO | - | 通知メッセージ（最大500文字） |
| `payload` | JSONB | YES | - | イベントペイロード（追加情報） |
| `link_url` | TEXT | YES | - | 遷移先URL |
| `is_read` | BOOLEAN | NO | `false` | 既読フラグ |
| `read_at` | TIMESTAMPTZ | YES | - | 既読日時 |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |

**主キー**: `id`  
**外部キー**:
- `user_id` → `users(id)` ON DELETE CASCADE
- `event_type_id` → `notification_event_types(id)` ON DELETE RESTRICT

**インデックス**:
- `idx_notifications_user_created` ON `(user_id, created_at DESC)`
- `idx_notifications_user_unread` ON `(user_id, is_read)` WHERE `is_read = false`

#### 5.1.5 `notification_deliveries`（配信履歴）

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| `id` | UUID | NO | `gen_random_uuid()` | 配信ID |
| `notification_id` | UUID | NO | - | 通知ID |
| `channel_id` | TEXT | NO | - | チャネルID |
| `status` | TEXT | NO | `pending` | 配信ステータス（pending/sent/failed） |
| `sent_at` | TIMESTAMPTZ | YES | - | 配信日時 |
| `error_message` | TEXT | YES | - | エラーメッセージ（失敗時） |
| `retry_count` | INTEGER | NO | `0` | リトライ回数 |
| `created_at` | TIMESTAMPTZ | NO | `CURRENT_TIMESTAMP` | 作成日時 |

**主キー**: `id`  
**外部キー**:
- `notification_id` → `notifications(id)` ON DELETE CASCADE
- `channel_id` → `notification_channels(id)` ON DELETE RESTRICT

**インデックス**:
- `idx_notification_deliveries_notification` ON `(notification_id)`
- `idx_notification_deliveries_status` ON `(status)` WHERE `status = 'pending'`

### 5.2 ER図

```mermaid
erDiagram
    users ||--o{ notification_settings : "has"
    users ||--o{ notifications : "receives"
    
    notification_channels ||--o{ notification_settings : "configures"
    notification_channels ||--o{ notification_deliveries : "delivers_via"
    
    notification_event_types ||--o{ notification_settings : "configures"
    notification_event_types ||--o{ notifications : "categorizes"
    
    notifications ||--o{ notification_deliveries : "has"
    
    notification_channels {
        text id PK
        text name
        text description
        boolean is_active
        timestamptz created_at
    }
    
    notification_event_types {
        text id PK
        text name
        text description
        boolean is_active
        timestamptz created_at
    }
    
    notification_settings {
        uuid user_id PK_FK
        text event_type_id PK_FK
        text channel_id PK_FK
        boolean is_enabled
        timestamptz updated_at
    }
    
    notifications {
        uuid id PK
        uuid user_id FK
        text event_type_id FK
        text title
        text message
        jsonb payload
        text link_url
        boolean is_read
        timestamptz read_at
        timestamptz created_at
    }
    
    notification_deliveries {
        uuid id PK
        uuid notification_id FK
        text channel_id FK
        text status
        timestamptz sent_at
        text error_message
        integer retry_count
        timestamptz created_at
    }
```

---

## 6. バックエンドアーキテクチャ

### 6.1 レイヤー構成

レイヤードアーキテクチャに準拠し、以下の構成とする：

```
イベント発生
  ↓
Service層（業務ロジック）
  ↓
NotificationService.CreateNotification()
  ↓
Repository層（データ永続化）
  ↓
配信キュー（非同期処理）
  ↓
配信ワーカー（各チャネル）
```

### 6.2 主要コンポーネント

#### 6.2.1 Repository層

**ファイル配置**: `apps/backend/src/internal/repository/`

| ファイル | 責務 |
|---------|------|
| `notification_channel_repository.go` | 通知チャネルマスターのCRUD |
| `notification_event_type_repository.go` | イベントタイプマスターのCRUD |
| `notification_settings_repository.go` | ユーザー通知設定のCRUD |
| `notification_repository.go` | 通知履歴のCRUD、既読管理 |
| `notification_delivery_repository.go` | 配信履歴のCRUD、ステータス更新 |

#### 6.2.2 Service層

**ファイル配置**: `apps/backend/src/internal/service/`

| ファイル | 責務 |
|---------|------|
| `notification_service.go` | 通知作成・取得・既読管理のビジネスロジック |
| `notification_settings_service.go` | 通知設定の取得・更新ロジック |
| `notification_delivery_service.go` | 配信実行・リトライロジック |

**主要メソッド（`NotificationService`）**:
```go
// CreateNotification は通知を作成し、配信キューに登録
CreateNotification(ctx context.Context, req CreateNotificationRequest) error

// ListNotifications はユーザーの通知一覧を取得（ページネーション対応）
ListNotifications(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]*model.Notification, int64, error)

// MarkAsRead は通知を既読にする
MarkAsRead(ctx context.Context, notificationID, userID string) error

// MarkAllAsRead はユーザーの全通知を既読にする
MarkAllAsRead(ctx context.Context, userID string) error

// GetUnreadCount は未読通知数を取得
GetUnreadCount(ctx context.Context, userID string) (int64, error)
```

**主要メソッド（`NotificationSettingsService`）**:
```go
// GetUserSettings はユーザーの通知設定を取得（全イベント×チャネルのマトリックス）
GetUserSettings(ctx context.Context, userID string) (map[string]map[string]bool, error)

// UpdateUserSettings はユーザーの通知設定を更新
UpdateUserSettings(ctx context.Context, userID string, settings map[string]map[string]bool) error

// InitializeDefaultSettings は新規ユーザーのデフォルト設定を初期化
InitializeDefaultSettings(ctx context.Context, userID string) error
```

#### 6.2.3 Handler層

**ファイル配置**: `apps/backend/src/internal/handler/`

| ファイル | 責務 |
|---------|------|
| `notification_handler.go` | 通知関連エンドポイントのHTTPハンドラー |
| `notification_settings_handler.go` | 通知設定関連エンドポイントのHTTPハンドラー |

### 6.3 非同期配信アーキテクチャ

#### 6.3.1 配信フロー

```mermaid
sequenceDiagram
    participant Service as Service層
    participant Repo as Repository層
    participant Queue as 配信キュー（DB/Redis）
    participant Worker as 配信ワーカー
    participant Channel as 配信チャネル（Email/Push）
    
    Service->>Repo: CreateNotification()
    Repo->>Repo: notifications テーブルに挿入
    Repo->>Repo: notification_deliveries レコード作成（status=pending）
    Repo-->>Service: 完了
    
    Worker->>Repo: FetchPendingDeliveries()
    Repo-->>Worker: 配信待ちリスト
    
    Worker->>Channel: Send()
    Channel-->>Worker: 成功/失敗
    
    Worker->>Repo: UpdateDeliveryStatus(sent/failed)
    
    alt 失敗時
        Worker->>Repo: IncrementRetryCount()
    end
```

#### 6.3.2 配信キュー実装方針

**Phase 1（初期実装）**:
- `notification_deliveries` テーブルの `status = 'pending'` レコードをポーリング
- 定期実行される配信ワーカー（Cron Job または独立プロセス）

**Phase 2（拡張案）**:
- Redis や PostgreSQL の `LISTEN/NOTIFY` を使った非同期ジョブキュー
- または外部サービス（AWS SQS、Cloud Tasks）との統合

### 6.4 配信チャネル実装

#### 6.4.1 アプリ内通知（`in_app`）

- `notifications` テーブルへの挿入のみ（配信ワーカー不要）
- フロントエンドがポーリングまたはWebSocketで取得

#### 6.4.2 メール通知（`email`）

- SMTP経由でメール送信（Go標準 `net/smtp` または SendGrid SDK）
- 環境変数: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`
- テンプレート: `templates/email/` にHTMLテンプレート配置

#### 6.4.3 Push通知（`push`）

- Firebase Cloud Messaging (FCM) を使用（将来実装）
- デバイストークン管理用テーブル `user_device_tokens` を追加

---

## 7. API設計

### 7.1 エンドポイント一覧

#### 7.1.1 通知取得・管理

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| GET | `/api/v1/notifications` | 通知一覧取得 | 必須 |
| GET | `/api/v1/notifications/unread-count` | 未読通知数取得 | 必須 |
| PATCH | `/api/v1/notifications/:notificationId/read` | 通知を既読にする | 必須 |
| PATCH | `/api/v1/notifications/read-all` | 全通知を既読にする | 必須 |

#### 7.1.2 通知設定

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| GET | `/api/v1/me/notification-settings` | 自分の通知設定取得 | 必須 |
| PUT | `/api/v1/me/notification-settings` | 自分の通知設定更新 | 必須 |

#### 7.1.3 管理者用（将来実装）

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| GET | `/api/v1/admin/notification-channels` | チャネル一覧取得 | 管理者 |
| POST | `/api/v1/admin/notification-channels` | チャネル追加 | 管理者 |
| GET | `/api/v1/admin/notification-event-types` | イベントタイプ一覧取得 | 管理者 |
| POST | `/api/v1/admin/notification-event-types` | イベントタイプ追加 | 管理者 |

### 7.2 リクエスト・レスポンス例

#### 7.2.1 通知一覧取得

**リクエスト**:
```
GET /api/v1/notifications?page=1&page_size=20&unread_only=false
Authorization: Bearer <token>
```

**レスポンス**:
```json
{
  "items": [
    {
      "id": "uuid",
      "event_type_id": "tag_movie_added",
      "title": "タグに映画が追加されました",
      "message": "山田太郎さんが「アクション映画コレクション」に「ダークナイト」を追加しました",
      "link_url": "/tags/uuid",
      "is_read": false,
      "created_at": "2026-01-14T12:34:56Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total_count": 42
}
```

#### 7.2.2 未読通知数取得

**リクエスト**:
```
GET /api/v1/notifications/unread-count
Authorization: Bearer <token>
```

**レスポンス**:
```json
{
  "unread_count": 5
}
```

#### 7.2.3 通知設定取得

**リクエスト**:
```
GET /api/v1/me/notification-settings
Authorization: Bearer <token>
```

**レスポンス**:
```json
{
  "settings": {
    "tag_movie_added": {
      "in_app": true,
      "email": true,
      "push": false
    },
    "tag_followed": {
      "in_app": true,
      "email": false,
      "push": false
    },
    "user_followed": {
      "in_app": true,
      "email": true,
      "push": true
    }
  }
}
```

#### 7.2.4 通知設定更新

**リクエスト**:
```
PUT /api/v1/me/notification-settings
Authorization: Bearer <token>
Content-Type: application/json

{
  "settings": {
    "tag_movie_added": {
      "in_app": true,
      "email": false,
      "push": false
    }
  }
}
```

**レスポンス**:
```json
{
  "message": "通知設定を更新しました"
}
```

---

## 8. フロントエンド統合

### 8.1 UIコンポーネント

#### 8.1.1 通知ベルアイコン（ヘッダー）

**配置**: `apps/frontend/src/components/layout/NotificationBell.tsx`

- 未読通知数をバッジ表示
- クリックでドロップダウンに最新5件表示
- 「すべて見る」ボタンで通知ページへ遷移

#### 8.1.2 通知一覧ページ

**配置**: `apps/frontend/src/app/notifications/page.tsx`

- 全通知をページネーション表示
- 未読/既読フィルター
- 個別既読ボタン、全て既読ボタン

#### 8.1.3 通知設定ページ

**配置**: `apps/frontend/src/app/settings/notifications/page.tsx`

- イベント × チャネルのマトリックスUIでトグル設定
- 保存ボタンで一括更新

### 8.2 API関数

**配置**: `apps/frontend/src/lib/api/notifications/`

| ファイル | 関数 |
|---------|------|
| `list.ts` | `listNotifications(page, pageSize, unreadOnly)` |
| `getUnreadCount.ts` | `getUnreadCount()` |
| `markAsRead.ts` | `markAsRead(notificationId)` |
| `markAllAsRead.ts` | `markAllAsRead()` |
| `getSettings.ts` | `getNotificationSettings()` |
| `updateSettings.ts` | `updateNotificationSettings(settings)` |

### 8.3 Zodスキーマ

**配置**: `apps/frontend/src/lib/validation/notification.ts`

```typescript
export const NotificationSchema = z.object({
  id: z.string(),
  event_type_id: z.string(),
  title: z.string(),
  message: z.string(),
  link_url: z.string().nullable(),
  is_read: z.boolean(),
  created_at: z.string(),
});

export const NotificationListResponseSchema = z.object({
  items: z.array(NotificationSchema),
  page: z.number(),
  page_size: z.number(),
  total_count: z.number(),
});

export const NotificationUnreadCountSchema = z.object({
  unread_count: z.number(),
});

export const NotificationSettingsSchema = z.record(
  z.record(z.boolean())
);
```

### 8.4 React Query統合

- `useQuery` でポーリング（未読数取得: 30秒間隔）
- `useMutation` で既読操作、設定更新
- 楽観的更新（Optimistic Update）を活用

---

## 9. 実装フェーズ計画

### Phase 1: 基盤構築（優先度: 高）

**目標**: アプリ内通知の基本機能を実装

1. **データベース**
   - マイグレーション: 5つのテーブル作成
   - Seedデータ: チャネル（`in_app`, `email`, `push`）、イベントタイプ（4種類）登録

2. **バックエンド - Repository層**
   - 5つのRepositoryインターフェース＆実装作成

3. **バックエンド - Service層**
   - `NotificationService`, `NotificationSettingsService` 実装
   - `UserService.EnsureUser()` にデフォルト設定初期化を追加

4. **バックエンド - Handler層**
   - `NotificationHandler`, `NotificationSettingsHandler` 実装
   - ルーター登録

5. **イベント統合**
   - 既存の `TagService`, `UserFollowerService` に通知生成処理を追加
   - イベント: `tag_movie_added`, `tag_followed`, `user_followed`

6. **フロントエンド - API層**
   - Zodスキーマ定義
   - API関数作成

7. **フロントエンド - UI**
   - 通知ベルアイコン（ヘッダー）
   - 通知一覧ページ
   - 通知設定ページ

8. **テスト**
   - バックエンド: ユニットテスト、インテグレーションテスト
   - フロントエンド: 手動テスト

### Phase 2: メール通知（優先度: 中）

**目標**: メール配信機能を実装

1. **配信ワーカー実装**
   - `apps/backend/src/cmd/notification-worker/main.go`
   - 定期実行（Cron）または独立プロセス

2. **メール配信チャネル**
   - SMTP統合
   - HTMLテンプレート作成（`templates/email/`）

3. **配信履歴管理**
   - ステータス更新、リトライロジック

4. **環境変数追加**
   - `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`

### Phase 3: 拡張機能（優先度: 低）

**目標**: 運用性向上とUX改善

1. **管理者機能**
   - チャネル・イベントタイプの動的追加UI
   - 全ユーザー通知設定の一括変更

2. **リアルタイム通知**
   - WebSocket または SSE による即時配信

3. **通知グルーピング**
   - 同一タグの映画追加を要約表示

4. **Push通知（モバイル）**
   - FCM統合
   - デバイストークン管理

---

## 10. テスト戦略

### 10.1 バックエンド

#### ユニットテスト
- Repository層: モックDBでCRUDテスト
- Service層: モックRepositoryでビジネスロジックテスト

#### インテグレーションテスト
- エンドツーエンドでAPI呼び出しテスト
- トランザクション整合性確認

```bash
# ユニットテスト
go test ./src/internal/repository/... -v
go test ./src/internal/service/... -v

# インテグレーションテスト
docker compose up -d postgres-test
DATABASE_URL="postgres://postgres:postgres@localhost:5433/cinetag_test?sslmode=disable" \
  go test -tags=integration ./...
```

### 10.2 フロントエンド

- React Query DevToolsでAPI連携確認
- 手動テスト: 通知作成 → 取得 → 既読 → 設定変更の一連フロー
- Lint: `npm run lint`

---

## 11. セキュリティ考慮事項

1. **認可チェック**
   - 通知の取得・既読操作は必ず `user_id` でフィルタリング（他ユーザーの通知にアクセス不可）

2. **入力検証**
   - 通知設定更新時、存在するイベントタイプ・チャネルIDのみ受け付ける

3. **XSS対策**
   - 通知メッセージはサニタイズされたデータのみ使用（HTMLタグ禁止）

4. **DoS対策**
   - 通知作成APIはレート制限を設ける（例: 同一イベントを1分間に1回まで）

5. **メール配信**
   - SPF/DKIM設定を行い、スパム判定を回避
   - Unsubscribeリンクを含める（法令遵守）

---

## 12. パフォーマンス最適化

1. **インデックス戦略**
   - `notifications` テーブル: `(user_id, created_at)`, `(user_id, is_read)`
   - `notification_deliveries` テーブル: `(status)` (pending配信のみ)

2. **キャッシュ**
   - 未読通知数をRedisにキャッシュ（書き込み時に無効化）

3. **ページネーション**
   - 通知一覧は必ずページング（デフォルト20件）

4. **非同期配信**
   - メール・Push通知は本体処理をブロックしない

---

## 13. 運用・監視

### 13.1 メトリクス

- 配信成功率（チャネル別）
- 配信遅延時間
- リトライ回数
- 未読通知数（ユーザー別平均）

### 13.2 ログ

- 配信ワーカー: 配信成功/失敗ログ（JSON形式）
- エラー: 配信失敗時のスタックトレース

### 13.3 アラート

- 配信失敗率が閾値（例: 10%）を超えた場合
- 配信キューが滞留（pending件数が1000件超）

---

## 14. 参考資料

- [タグフォロー機能実装計画](./20260111_tag-follow-implementation-plan.md)
- [データベーススキーマ](../docs/data/database-schema.md)
- [バックエンドアーキテクチャ](../docs/architecture/backend-architecture.md)
- [API仕様書](../docs/api/api-spec.md)

---

## 15. 今後の拡張案

1. **通知の優先度設定**
   - 重要度（高/中/低）を設定し、UI表示を変える

2. **通知スケジューリング**
   - 夜間は配信しない「サイレント時間」設定

3. **通知ダイジェスト**
   - 1日1回、未読通知をまとめてメール配信

4. **多言語対応**
   - 通知メッセージのi18n対応

5. **Webhook統合**
   - Slack、Discord等への通知転送

---

## 変更履歴

| 日付 | バージョン | 変更内容 | 作成者 |
|------|-----------|---------|--------|
| 2026-01-14 | 1.0.0 | 初版作成 | GitHub Copilot |
