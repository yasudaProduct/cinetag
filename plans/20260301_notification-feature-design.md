# 通知機能 設計ドキュメント

**作成日**: 2026-03-01
**ステータス**: 検討中

---

## 1. 背景と目的

cinetagでは、タグフォロー・ユーザーフォロー機能が実装済みだが、フォロー先の変化をユーザーに伝える仕組みがない。API仕様書の「今後の拡張候補」にも通知・アクティビティが記載されており、ユーザーエンゲージメント向上のために通知機能を導入する。

### 1.1 解決したい課題

| 課題 | 現状 | 通知により |
|------|------|-----------|
| タグへの映画追加に気づけない | 手動でタグ詳細ページを確認する必要がある | フォロー中タグの更新を即座に把握 |
| フォローされたことがわからない | プロフィールのフォロワー数変化でしか判断できない | 新しいフォロワーを認識し交流につながる |
| フォロー中ユーザーの活動が追えない | ユーザーページを定期的に訪問する必要がある | 新しいタグ作成などの活動を自動通知 |

---

## 2. 通知内容の定義

### 2.1 通知タイプ一覧

| ID | 通知タイプ | トリガーアクション | 通知先 | 優先度 |
|----|-----------|-------------------|--------|--------|
| `tag_movie_added` | タグに映画が追加された | 他ユーザーがタグに映画を追加 | タグオーナー + タグフォロワー | 高 |
| `tag_followed` | タグがフォローされた | 他ユーザーがタグをフォロー | タグオーナー | 高 |
| `user_followed` | フォローされた | 他ユーザーが自分をフォロー | フォローされたユーザー | 高 |
| `following_user_created_tag` | フォロー中ユーザーが新タグ作成 | フォロー中ユーザーが公開タグを作成 | フォロワー全員 | 中 |

### 2.2 通知メッセージ例

| 通知タイプ | メッセージ例 |
|-----------|-------------|
| `tag_movie_added` | **{username}** があなたのタグ「{tag_title}」に「{movie_title}」を追加しました |
| `tag_followed` | **{username}** があなたのタグ「{tag_title}」をフォローしました |
| `user_followed` | **{username}** があなたをフォローしました |
| `following_user_created_tag` | **{username}** が新しいタグ「{tag_title}」を作成しました |

### 2.3 将来拡張候補（Phase 1では対象外）

- タグへのコメント機能が追加された場合の通知
- タグの映画が削除された場合の通知
- 人気タグのレコメンド通知
- 運営からのお知らせ通知

---

## 3. 通知方法の比較検討

### 3.1 方式比較

| 方式 | 即時性 | ユーザー体験 | 実装コスト | インフラ影響 | 到達率 |
|------|--------|-------------|-----------|-------------|--------|
| **アプリ内通知** | ポーリング間隔に依存 | 良い（アプリ内で完結） | 低 | 小 | アプリ利用時のみ |
| **ブラウザプッシュ通知** | 即時 | 良い（アプリ外でも届く） | 中 | 中 | ブラウザ許可依存 |
| **メール通知** | 数秒〜数分 | 普通（受信箱に埋もれうる） | 中〜高 | 外部サービス依存 | 高い |

### 3.2 推奨アプローチ：段階的導入

#### Phase 1: アプリ内通知（MVP）

最小コストで通知の基盤を構築する。

- **DB**: `notifications` テーブル追加
- **Backend**: 通知の生成・取得・既読管理API
- **Frontend**: サイドバーにベルアイコン + 未読バッジ、通知一覧ドロップダウン
- **データ取得**: ポーリング（60秒間隔）

**ポーリングを選択する理由**:
- Cloud Runはスケール0に対応しており、WebSocket/SSEの常時接続と相性が悪い
- Cloudflare Workers (Freeプラン) のCPU制限（10ms）を考慮すると、フロントエンド側でのSSE処理は重い
- ポーリング間隔60秒であれば、DB負荷・API負荷ともに許容範囲
- 通知の即時性が厳密に求められるユースケースではない

#### Phase 2: ブラウザプッシュ通知（オプション）

ユーザーの利用動向を見て導入を判断する。

- **Web Push API** + Service Worker
- **プッシュサーバー**: Cloud Run上にVAPIDベースの配信エンドポイント追加
- **購読管理**: `push_subscriptions` テーブルで購読情報を管理
- **配信タイミング**: アクション発生時に非同期で配信

#### Phase 3: メール通知（オプション）

高エンゲージメントユーザー向けのダイジェスト配信。

- **メールサービス**: Resend（Cloudflare統合が良い）or SendGrid
- **配信頻度**: 日次/週次ダイジェスト（リアルタイムではない）
- **テンプレート**: 未読通知のサマリーをHTML/テキストで送信
- **配信トリガー**: Cloud Schedulerによる定期バッチ or Cloudflare Cron Triggers
- **オプトイン**: ユーザー設定で通知頻度を選択可能

---

## 4. Phase 1: アプリ内通知 - 詳細設計

### 4.1 データベース設計

#### `notifications` テーブル

```sql
-- +goose Up
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 通知を受け取るユーザー
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- 通知を発生させたユーザー（システム通知の場合はNULL）
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    -- 通知タイプ
    notification_type TEXT NOT NULL,
    -- 関連リソースへの参照（型によって使い分け）
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    tag_movie_id UUID REFERENCES tag_movies(id) ON DELETE CASCADE,
    -- 既読管理
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    -- タイムスタンプ
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_notifications_recipient_created
    ON notifications (recipient_user_id, created_at DESC);
CREATE INDEX idx_notifications_recipient_unread
    ON notifications (recipient_user_id, is_read)
    WHERE is_read = false;

-- +goose Down
DROP TABLE IF EXISTS notifications;
```

**設計の意図**:

| 判断 | 理由 |
|------|------|
| `actor_user_id` をNULL許容 | 将来のシステム通知（運営からのお知らせ等）に対応 |
| `tag_id`, `tag_movie_id` を直接参照 | JSON blob よりも型安全かつクエリ・JOINが容易 |
| `ON DELETE CASCADE` | リソース削除時に関連通知も自動削除（孤立通知を防止） |
| `actor_user_id` は `ON DELETE SET NULL` | ユーザー削除後も通知履歴は残す（「退会済みユーザー」表示） |
| 部分インデックス (`WHERE is_read = false`) | 未読通知の高速取得（最も頻繁なクエリパターン） |

#### ER図の追加要素

```
users (1) --- (N) notifications (as recipient)
users (1) --- (N) notifications (as actor)
tags  (1) --- (N) notifications
tag_movies (1) --- (N) notifications
```

### 4.2 バックエンド設計

#### 4.2.1 モデル

**ファイル**: `apps/backend/src/internal/model/notification.go`

```go
type Notification struct {
    ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    RecipientUserID string     `gorm:"type:uuid;not null" json:"recipient_user_id"`
    ActorUserID     *string    `gorm:"type:uuid" json:"actor_user_id"`
    NotificationType string   `gorm:"type:text;not null" json:"notification_type"`
    TagID           *string    `gorm:"type:uuid" json:"tag_id"`
    TagMovieID      *string    `gorm:"type:uuid" json:"tag_movie_id"`
    IsRead          bool       `gorm:"not null;default:false" json:"is_read"`
    ReadAt          *time.Time `json:"read_at"`
    CreatedAt       time.Time  `gorm:"not null;default:now()" json:"created_at"`

    // リレーション（プリロード用）
    Recipient *User     `gorm:"foreignKey:RecipientUserID" json:"-"`
    Actor     *User     `gorm:"foreignKey:ActorUserID" json:"actor,omitempty"`
    Tag       *Tag      `gorm:"foreignKey:TagID" json:"tag,omitempty"`
    TagMovie  *TagMovie `gorm:"foreignKey:TagMovieID" json:"tag_movie,omitempty"`
}
```

**通知タイプ定数**:

```go
const (
    NotificationTypeTagMovieAdded         = "tag_movie_added"
    NotificationTypeTagFollowed           = "tag_followed"
    NotificationTypeUserFollowed          = "user_followed"
    NotificationTypeFollowingUserCreatedTag = "following_user_created_tag"
)
```

#### 4.2.2 Repository層

**ファイル**: `apps/backend/src/internal/repository/notification_repository.go`

```go
type NotificationRepository interface {
    // Create は通知を1件作成
    Create(ctx context.Context, notification *model.Notification) error
    // CreateBatch は通知を一括作成（フォロワー全員への通知等）
    CreateBatch(ctx context.Context, notifications []*model.Notification) error
    // ListByRecipient は指定ユーザーの通知一覧を返す（新しい順）
    ListByRecipient(ctx context.Context, userID string, page, pageSize int) ([]*model.Notification, int64, error)
    // CountUnread は未読通知数を返す
    CountUnread(ctx context.Context, userID string) (int64, error)
    // MarkAsRead は指定の通知を既読にする
    MarkAsRead(ctx context.Context, notificationID, userID string) error
    // MarkAllAsRead は全通知を既読にする
    MarkAllAsRead(ctx context.Context, userID string) error
}
```

#### 4.2.3 Service層

**ファイル**: `apps/backend/src/internal/service/notification_service.go`

```go
type NotificationService interface {
    // 通知の取得
    ListNotifications(ctx context.Context, userID string, page, pageSize int) ([]*NotificationItem, int64, error)
    GetUnreadCount(ctx context.Context, userID string) (int64, error)
    // 既読管理
    MarkAsRead(ctx context.Context, notificationID, userID string) error
    MarkAllAsRead(ctx context.Context, userID string) error
    // 通知の生成（各サービスから呼び出し）
    NotifyTagMovieAdded(ctx context.Context, tagID, tagMovieID, actorUserID string) error
    NotifyTagFollowed(ctx context.Context, tagID, actorUserID string) error
    NotifyUserFollowed(ctx context.Context, followeeUserID, actorUserID string) error
    NotifyFollowingUserCreatedTag(ctx context.Context, tagID, actorUserID string) error
}
```

**`NotificationItem`（レスポンスDTO）**:

```go
type NotificationItem struct {
    ID               string     `json:"id"`
    NotificationType string     `json:"notification_type"`
    IsRead           bool       `json:"is_read"`
    CreatedAt        time.Time  `json:"created_at"`
    // 展開されたリレーション
    Actor            *UserSummary `json:"actor"`
    Tag              *TagSummary  `json:"tag,omitempty"`
    MovieTitle       *string      `json:"movie_title,omitempty"`
}
```

**通知生成のフロー例（`NotifyTagMovieAdded`）**:

```
1. tag_id からタグ情報を取得
2. タグオーナーの user_id を取得
3. タグフォロワーの user_id 一覧を取得
4. 通知先 = (タグオーナー + フォロワー) - アクター自身
5. 通知先が空でなければ CreateBatch で一括作成
```

**自己通知の除外**: アクター自身には通知を送らない（自分で自分のタグに映画を追加した場合等）

#### 4.2.4 既存サービスへの通知呼び出し追加

通知生成は、既存のサービスメソッド内から `NotificationService` を呼び出す形で統合する。

| 既存メソッド | 追加する通知呼び出し |
|-------------|---------------------|
| `TagService.AddMoviesToTag()` | `NotifyTagMovieAdded()` |
| `TagService.FollowTag()` | `NotifyTagFollowed()` |
| `UserService.FollowUser()` | `NotifyUserFollowed()` |
| `TagService.CreateTag()` | `NotifyFollowingUserCreatedTag()` |

**通知生成の非同期化（Phase 1）**: goroutineで非同期実行し、メインのレスポンスをブロックしない。通知の生成失敗はログ出力のみで、元のアクションには影響させない。

```go
// TagService.AddMoviesToTag 内
go func() {
    if err := s.notificationService.NotifyTagMovieAdded(ctx, tagID, tagMovieID, actorUserID); err != nil {
        log.Printf("failed to send notification: %v", err)
    }
}()
```

### 4.3 APIエンドポイント設計

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| GET | `/api/v1/notifications` | 通知一覧取得 | 必須 |
| GET | `/api/v1/notifications/unread-count` | 未読通知数取得 | 必須 |
| PATCH | `/api/v1/notifications/:notificationId/read` | 個別既読 | 必須 |
| PATCH | `/api/v1/notifications/read-all` | 全件既読 | 必須 |

#### GET `/api/v1/notifications`

**クエリパラメータ**:

| パラメータ | 型 | デフォルト | 説明 |
|-----------|------|-----------|------|
| `page` | int | 1 | ページ番号 |
| `page_size` | int | 20 | 1ページあたりの件数（最大50） |

**レスポンス（200 OK）**:

```json
{
  "notifications": [
    {
      "id": "uuid",
      "notification_type": "tag_movie_added",
      "is_read": false,
      "created_at": "2026-03-01T10:00:00Z",
      "actor": {
        "id": "uuid",
        "display_id": "john",
        "display_name": "John",
        "avatar_url": "https://..."
      },
      "tag": {
        "id": "uuid",
        "title": "SF映画ベスト"
      },
      "movie_title": "インターステラー"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20
}
```

#### GET `/api/v1/notifications/unread-count`

**レスポンス（200 OK）**:

```json
{
  "unread_count": 5
}
```

#### PATCH `/api/v1/notifications/:notificationId/read`

**レスポンス**: 204 No Content

#### PATCH `/api/v1/notifications/read-all`

**レスポンス**: 204 No Content

### 4.4 フロントエンド設計

#### 4.4.1 API関数

**ファイル**: `apps/frontend/src/lib/api/notifications/`

| 関数 | 概要 |
|------|------|
| `listNotifications(page, pageSize, token)` | 通知一覧取得 |
| `getUnreadCount(token)` | 未読通知数取得 |
| `markAsRead(notificationId, token)` | 個別既読 |
| `markAllAsRead(token)` | 全件既読 |

#### 4.4.2 Zodスキーマ

**ファイル**: `apps/frontend/src/lib/validation/notification.api.ts`

```typescript
import { z } from "zod";

const notificationActorSchema = z.object({
  id: z.string(),
  display_id: z.string(),
  display_name: z.string(),
  avatar_url: z.string().nullable(),
});

const notificationTagSchema = z.object({
  id: z.string(),
  title: z.string(),
});

export const notificationItemSchema = z.object({
  id: z.string(),
  notification_type: z.enum([
    "tag_movie_added",
    "tag_followed",
    "user_followed",
    "following_user_created_tag",
  ]),
  is_read: z.boolean(),
  created_at: z.string(),
  actor: notificationActorSchema.nullable(),
  tag: notificationTagSchema.nullable(),
  movie_title: z.string().nullable(),
});

export const notificationListResponseSchema = z.object({
  notifications: z.array(notificationItemSchema),
  total: z.number(),
  page: z.number(),
  page_size: z.number(),
});
```

#### 4.4.3 UIコンポーネント設計

```
ヘッダー
├── NotificationBell         # ベルアイコン + 未読バッジ
│   └── NotificationDropdown # ドロップダウンパネル
│       ├── NotificationItem # 各通知カード
│       └── "すべて既読にする" ボタン
└── /notifications ページ    # 通知一覧フルページ（オプション）
```

**NotificationBell**:
- ヘッダー右側に配置（アバターアイコンの隣）
- 未読件数 > 0 の場合、赤いバッジを表示
- クリックでドロップダウンを開閉

**NotificationDropdown**:
- 最新20件の通知を表示
- 未読通知はハイライト背景
- クリックで該当リソースページに遷移 + 自動既読化
- 「すべて既読にする」リンク
- 遅延ロード（`dynamic import`）でバンドルサイズに影響させない

**ポーリング実装**:

```typescript
// React Query でポーリング
const { data: unreadCount } = useQuery({
  queryKey: ["notifications", "unread-count"],
  queryFn: () => getUnreadCount(token),
  refetchInterval: 60_000, // 60秒間隔
  enabled: isAuthenticated,
});
```

#### 4.4.4 通知クリック時の遷移先

| 通知タイプ | 遷移先 |
|-----------|--------|
| `tag_movie_added` | `/tags/{tag_id}` （タグ詳細ページ） |
| `tag_followed` | `/tags/{tag_id}` （タグ詳細ページ） |
| `user_followed` | `/users/{actor_display_id}` （フォロワーのプロフィール） |
| `following_user_created_tag` | `/tags/{tag_id}` （新しいタグのページ） |

---

## 5. Phase 2: ブラウザプッシュ通知

### 5.1 技術概要

**Web Push API** を利用し、ブラウザがバックグラウンドでもサーバーからの通知を受信できるようにする。

```
ユーザー操作（映画追加等）
  → Backend API
    → 通知レコード作成（Phase 1と同じ）
    → プッシュ配信キュー投入
      → 非同期Worker → Web Push送信
        → ブラウザ Service Worker → OS通知表示
```

### 5.2 追加DB設計

#### `push_subscriptions` テーブル

```sql
CREATE TABLE push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, endpoint)
);
```

### 5.3 実装要素

| 要素 | 技術 | 説明 |
|------|------|------|
| VAPID鍵ペア | Go `github.com/SherClockHolmes/webpush-go` | サーバー認証用の公開鍵/秘密鍵 |
| Service Worker | `apps/frontend/public/sw.js` | プッシュイベント受信 + 通知表示 |
| 購読フロー | `PushManager.subscribe()` | ユーザー許可 → 購読情報をAPIに送信 |
| 配信 | Cloud Run非同期処理 | goroutineまたはCloud Tasksで非同期配信 |

### 5.4 ユーザー設定

通知設定ページで、通知タイプごとにプッシュ通知のオン/オフを制御。

```json
{
  "push_enabled": true,
  "push_preferences": {
    "tag_movie_added": true,
    "tag_followed": true,
    "user_followed": true,
    "following_user_created_tag": false
  }
}
```

### 5.5 導入判断の基準

以下の条件を満たした場合にPhase 2の導入を検討する:

- Phase 1の通知機能が安定稼働している
- DAU（日次アクティブユーザー）が一定数に達している
- ユーザーからプッシュ通知の要望がある

---

## 6. Phase 3: メール通知

### 6.1 技術概要

未読通知をダイジェスト形式でメール配信する。リアルタイム通知ではなく、定期集約型。

```
Cloud Scheduler (毎朝9:00 JST)
  → Cloud Run エンドポイント呼び出し
    → 未読通知が溜まっているユーザーを抽出
      → メールテンプレートにレンダリング
        → Resend API でメール送信
```

### 6.2 メールサービスの比較

| サービス | 無料枠 | 特徴 | 月額（有料） |
|---------|--------|------|-------------|
| **Resend** | 100通/日 | モダンAPI、Cloudflare統合良好 | $20/月〜 |
| SendGrid | 100通/日 | 実績豊富、テンプレート機能充実 | $19.95/月〜 |
| Amazon SES | なし | 最安（$0.10/1000通）、設定やや複雑 | 従量課金 |

**推奨**: **Resend** — API設計がシンプルで、TypeScript/Go SDK対応、Cloudflare Workersとの相性が良い。

### 6.3 メールテンプレート

**ダイジェストメール例**:

```
件名: 【cinetag】未読通知のお知らせ（3件）

こんにちは {display_name} さん、

最近のアクティビティをお知らせします:

🎬 john さんがあなたのタグ「SF映画ベスト」に「インターステラー」を追加しました
👤 alice さんがあなたをフォローしました
🏷️ bob さんが新しいタグ「ホラー名作集」を作成しました

▶ cinetagで確認する: https://cinetag.app/notifications

---
このメールは cinetag の通知設定に基づいて送信されています。
通知設定の変更: https://cinetag.app/settings/notifications
```

### 6.4 ユーザー設定

| 設定 | 選択肢 | デフォルト |
|------|--------|-----------|
| メール通知 | オン / オフ | オフ |
| 配信頻度 | 毎日 / 週1回 / 月1回 | 週1回 |

### 6.5 導入判断の基準

- Phase 1が安定稼働している
- ユーザー登録にメールアドレスが紐づいている（Clerk経由で取得可能）
- メール通知への需要が確認できている

---

## 7. 通知設定テーブル（Phase 2以降で追加）

Phase 2またはPhase 3導入時に、ユーザーごとの通知設定を管理するテーブルを追加する。

```sql
CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    -- アプリ内通知（Phase 1では常にtrue、将来の設定画面用）
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    -- ブラウザプッシュ通知（Phase 2）
    push_enabled BOOLEAN NOT NULL DEFAULT false,
    push_tag_movie_added BOOLEAN NOT NULL DEFAULT true,
    push_tag_followed BOOLEAN NOT NULL DEFAULT true,
    push_user_followed BOOLEAN NOT NULL DEFAULT true,
    push_following_user_created_tag BOOLEAN NOT NULL DEFAULT true,
    -- メール通知（Phase 3）
    email_enabled BOOLEAN NOT NULL DEFAULT false,
    email_frequency TEXT NOT NULL DEFAULT 'weekly',  -- 'daily', 'weekly', 'monthly'
    -- タイムスタンプ
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 8. 非機能要件

### 8.1 パフォーマンス

| 項目 | 目標 |
|------|------|
| 未読通知数の取得 | 50ms以内（部分インデックス使用） |
| 通知一覧の取得 | 200ms以内（20件、リレーション含む） |
| 通知の生成 | メインリクエストをブロックしない（goroutine非同期） |
| ポーリング負荷 | 認証済みユーザー × 60秒間隔で許容範囲 |

### 8.2 データ管理

| 項目 | 方針 |
|------|------|
| 通知の保持期間 | 90日（それ以前は定期バッチで削除） |
| 通知の最大蓄積数 | 1ユーザーあたり1000件（超過分は古い順に削除） |
| 重複通知の防止 | 同一アクターによる同一リソースへの短時間連続操作は集約を検討 |

### 8.3 セキュリティ

| 項目 | 方針 |
|------|------|
| アクセス制御 | 自分の通知のみ取得・操作可能（`recipient_user_id` でフィルタ） |
| VAPID鍵管理（Phase 2） | 環境変数で管理、Secret Managerに保存 |
| メールヘッダ（Phase 3） | SPF/DKIM/DMARC設定 |

---

## 9. 実装スケジュール案

### Phase 1: アプリ内通知

| ステップ | 内容 | 依存 |
|---------|------|------|
| 1-1 | DBマイグレーション（`notifications` テーブル） | なし |
| 1-2 | `Notification` モデル定義 | 1-1 |
| 1-3 | `NotificationRepository` 実装 | 1-2 |
| 1-4 | `NotificationService` 実装 | 1-3 |
| 1-5 | 既存サービスに通知呼び出し追加 | 1-4 |
| 1-6 | `NotificationHandler` + ルーター登録 | 1-4 |
| 1-7 | フロントエンド API関数 + Zodスキーマ | 1-6 |
| 1-8 | NotificationBell + DropdownUI | 1-7 |
| 1-9 | ポーリング実装（React Query `refetchInterval`） | 1-8 |
| 1-10 | 通知クリック時の遷移・既読化 | 1-8 |

### Phase 2-3 は Phase 1 の運用状況を見て判断

---

## 10. 技術的リスクと対策

| リスク | 影響 | 対策 |
|--------|------|------|
| フォロワーが多いタグへの映画追加で大量通知が生成される | DB書き込み負荷 | `CreateBatch` で一括INSERT、バッチサイズ上限設定 |
| ポーリングによるAPI負荷増大 | Cloud Runインスタンス数増加 | `unread-count` は軽量クエリ（部分インデックス）、キャッシュヘッダ活用 |
| 通知テーブルの肥大化 | クエリ性能劣化 | 90日超過分の定期削除バッチ |
| Cloud Run コールドスタート時のレイテンシ | ポーリング応答遅延 | 最小インスタンス数1の設定を検討（コスト増） |
| goroutine内の通知生成失敗 | 通知が届かない | 構造化ログ出力 + 監視アラート設定 |
