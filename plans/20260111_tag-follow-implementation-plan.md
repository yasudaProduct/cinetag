# タグフォロー機能 実装計画

**作成日**: 2026-01-11
**ステータス**: 全Phase完了

---

## 1. 現状分析

### 1.1 実装済み

| レイヤー | 項目 | 状態 |
|---------|------|------|
| DB | `tag_followers` テーブル定義 | ✅ 完了 |
| DB | `trigger_update_tag_follower_count` トリガー | ✅ 完了（ドキュメントに定義） |
| Backend | `TagFollower` モデル | ✅ 完了 ([tag_follower.go](apps/backend/src/internal/model/tag_follower.go)) |
| Backend | `Tag` モデルに `follower_count` フィールド | ✅ 完了 |
| Docs | API仕様（セクション7） | ✅ 定義済み（未実装マーク付き） |
| Docs | DBスキーマ | ✅ 定義済み |

### 1.2 未実装

| レイヤー | 項目 | 優先度 |
|---------|------|--------|
| Backend | `TagFollowerRepository` | 高 |
| Backend | `TagService` にフォロー関連メソッド追加 | 高 |
| Backend | `TagHandler` にフォロー関連ハンドラー追加 | 高 |
| Backend | ルーター登録（3エンドポイント） | 高 |
| Frontend | API関数（`followTag`, `unfollowTag`, `getTagFollowers`） | 高 |
| Frontend | タグ詳細ページにフォローボタンUI | 高 |
| Frontend | フォロー状態の取得・表示 | 中 |
| Frontend | タグフォロワー一覧表示 | 低 |

---

## 2. 実装対象のAPIエンドポイント

API仕様書（`docs/api-spec.md` セクション7）に基づく：

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| POST | `/api/v1/tags/:tagId/follow` | タグをフォロー | 必須 |
| DELETE | `/api/v1/tags/:tagId/follow` | タグフォロー解除 | 必須 |
| GET | `/api/v1/tags/:tagId/followers` | フォロワー一覧取得 | 任意 |

**追加で必要なエンドポイント**（ユーザーフォロー機能を参考）：

| メソッド | エンドポイント | 概要 | 認証 |
|---------|---------------|------|------|
| GET | `/api/v1/tags/:tagId/follow-status` | 自分のフォロー状態確認 | 必須 |

---

## 3. バックエンド実装計画

### 3.1 Repository層

**ファイル**: `apps/backend/src/internal/repository/tag_follower_repository.go`

`UserFollowerRepository` を参考に以下のインターフェースと実装を作成：

```go
type TagFollowerRepository interface {
    // Create はタグフォロー関係を作成
    Create(ctx context.Context, tagID, userID string) error
    // Delete はタグフォロー関係を削除
    Delete(ctx context.Context, tagID, userID string) error
    // IsFollowing はユーザーがタグをフォローしているかチェック
    IsFollowing(ctx context.Context, tagID, userID string) (bool, error)
    // ListFollowers はタグのフォロワー一覧を返す
    ListFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error)
    // CountFollowers はタグのフォロワー数を返す
    CountFollowers(ctx context.Context, tagID string) (int64, error)
}
```

### 3.2 Service層

**ファイル**: `apps/backend/src/internal/service/tag_service.go`（既存ファイルに追加）

`TagService` インターフェースに以下のメソッドを追加：

```go
// フォロー関連メソッド
FollowTag(ctx context.Context, tagID, userID string) error
UnfollowTag(ctx context.Context, tagID, userID string) error
IsFollowingTag(ctx context.Context, tagID, userID string) (bool, error)
ListTagFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error)
```

**ビジネスロジック**:
- フォロー時：タグの存在確認 → 重複チェック → 作成
- アンフォロー時：タグの存在確認 → レコード削除（存在しなくても204を返す）
- `follower_count` の更新はDBトリガーに任せる

### 3.3 Handler層

**ファイル**: `apps/backend/src/internal/handler/tag_handler.go`（既存ファイルに追加）

`TagHandler` に以下のメソッドを追加：

```go
// FollowTag POST /api/v1/tags/:tagId/follow
func (h *TagHandler) FollowTag(c *gin.Context)

// UnfollowTag DELETE /api/v1/tags/:tagId/follow
func (h *TagHandler) UnfollowTag(c *gin.Context)

// ListTagFollowers GET /api/v1/tags/:tagId/followers
func (h *TagHandler) ListTagFollowers(c *gin.Context)

// GetTagFollowStatus GET /api/v1/tags/:tagId/follow-status
func (h *TagHandler) GetTagFollowStatus(c *gin.Context)
```

### 3.4 Router登録

**ファイル**: `apps/backend/src/router/router.go`

```go
// タグフォロー関連（認証必須）
tagAuthGroup := v1.Group("/tags")
tagAuthGroup.Use(middleware.AuthMiddleware(clerkJWKSURL, clerkIssuer, clerkAudience))
{
    tagAuthGroup.POST("/:tagId/follow", tagHandler.FollowTag)
    tagAuthGroup.DELETE("/:tagId/follow", tagHandler.UnfollowTag)
    tagAuthGroup.GET("/:tagId/follow-status", tagHandler.GetTagFollowStatus)
}

// タグフォロワー一覧（認証任意）
tagOptionalAuthGroup := v1.Group("/tags")
tagOptionalAuthGroup.Use(middleware.OptionalAuthMiddleware(clerkJWKSURL, clerkIssuer, clerkAudience))
{
    tagOptionalAuthGroup.GET("/:tagId/followers", tagHandler.ListTagFollowers)
}
```

### 3.5 センチネルエラーの追加

**ファイル**: `apps/backend/src/internal/service/errors.go`（または既存のエラー定義ファイル）

```go
var (
    ErrAlreadyFollowingTag = errors.New("already following this tag")
    ErrNotFollowingTag     = errors.New("not following this tag")
)
```

---

## 4. フロントエンド実装計画

### 4.1 API関数

**ディレクトリ**: `apps/frontend/src/lib/api/tags/`

| ファイル | 関数 | 説明 |
|---------|------|------|
| `follow.ts` | `followTag(tagId: string)` | タグをフォロー |
| `unfollow.ts` | `unfollowTag(tagId: string)` | タグフォロー解除 |
| `getFollowStatus.ts` | `getTagFollowStatus(tagId: string)` | フォロー状態取得 |
| `listFollowers.ts` | `listTagFollowers(tagId: string, page?, pageSize?)` | フォロワー一覧取得 |

### 4.2 Zodスキーマ

**ファイル**: `apps/frontend/src/lib/validation/tag.ts`（既存ファイルに追加）

```typescript
export const TagFollowStatusSchema = z.object({
  is_following: z.boolean(),
});

export const TagFollowerSchema = z.object({
  id: z.string(),
  display_id: z.string(),
  display_name: z.string(),
  avatar_url: z.string().nullable(),
});

export const TagFollowersListSchema = z.object({
  items: z.array(TagFollowerSchema),
  page: z.number(),
  page_size: z.number(),
  total_count: z.number(),
});
```

### 4.3 UIコンポーネント

**タグ詳細ページ**: `apps/frontend/src/app/tags/[tagId]/page.tsx`

1. **フォローボタン追加**
   - 未ログイン：ボタン非表示またはログイン誘導
   - ログイン済み＆未フォロー：「フォロー」ボタン表示
   - ログイン済み＆フォロー中：「フォロー中」ボタン表示（クリックで解除）

2. **フォロワー数表示**
   - 既存の `follower_count` を表示

3. **React Query統合**
   - `useQuery` でフォロー状態を取得
   - `useMutation` でフォロー/アンフォロー実行
   - 楽観的更新（Optimistic Update）を検討

**新規コンポーネント**（オプション）:
- `components/tags/TagFollowButton.tsx` - フォローボタンコンポーネント

---

## 5. 実装順序

### Phase 1: バックエンド基盤（優先度: 高）

1. `TagFollowerRepository` の作成
2. `TagService` にフォロー関連メソッド追加
3. `TagHandler` にハンドラーメソッド追加
4. ルーター登録
5. バックエンドテスト作成・実行

### Phase 2: フロントエンドAPI層（優先度: 高）

1. Zodスキーマ追加
2. API関数作成（`follow.ts`, `unfollow.ts`, `getFollowStatus.ts`, `listFollowers.ts`）

### Phase 3: フロントエンドUI（優先度: 高）

1. タグ詳細ページにフォローボタン追加
2. フォロー状態の取得・表示
3. フォロー/アンフォローの操作実装

### Phase 4: 追加機能（優先度: 低）

1. フォロワー一覧表示（タブまたはモーダル）
2. ユーザーページに「フォロー中のタグ」タブ追加（API: `/api/v1/me/following-tags`）
3. ドキュメント更新（API仕様の「未実装」マーク削除）

---

## 6. テスト計画

### バックエンド

```bash
# ユニットテスト
go test ./src/internal/repository/... -v
go test ./src/internal/service/... -v

# インテグレーションテスト
docker compose up -d postgres-test
DATABASE_URL="postgres://postgres:postgres@localhost:5433/cinetag_test?sslmode=disable" \
  go test -tags=integration ./...
```

### フロントエンド

- 手動テスト（React Query DevTools）
- Lint: `npm run lint`

---

## 7. 参考実装

ユーザーフォロー機能の実装を参考にする：

| 項目 | 参照ファイル |
|------|-------------|
| Repository | [user_follower_repository.go](apps/backend/src/internal/repository/user_follower_repository.go) |
| Handler | [user_handler.go](apps/backend/src/internal/handler/user_handler.go) の `FollowUser`, `UnfollowUser` |
| Router | [router.go](apps/backend/src/router/router.go) のユーザーフォロールート |
| Frontend API | `apps/frontend/src/lib/api/users/followUser.ts` など |
| Frontend UI | `apps/frontend/src/app/[username]/page.tsx` のフォローボタン部分 |

---

## 8. 注意事項

- `follower_count` の更新はDBトリガー（`trigger_update_tag_follower_count`）に任せる
- 自分が作成したタグを自分でフォローすることは許可する（仕様確認が必要な場合あり）
- 非公開タグのフォローポリシーを確認（現時点では公開タグのみフォロー可能と想定）
