# タグ削除 & オーナー譲渡 設計ドキュメント

**作成日**: 2026-04-12
**ステータス**: Phase 1 未着手

---

## 1. 背景と目的

現在 cinetag にはタグを削除する手段がなく、不要になったタグが残り続ける。API 仕様書（`docs/api/api-spec.md` セクション 5.5）でも `DELETE /api/v1/tags/:tagId` は **[未実装]** と記載されている。

また、cinetag はタグへの映画追加を他ユーザーにも開放しているため、タグは「共有資産」としての側面を持つ。オーナーがタグを手放したい場合に、削除以外の選択肢（オーナー譲渡）を提供することで、他ユーザーの貢献データを保護できる。

### 1.1 フェーズ構成

| フェーズ | 内容 | 優先度 |
|---------|------|--------|
| **Phase 1** | タグの物理削除 | 高 |
| **Phase 2** | タグオーナーの譲渡 | 中（ユーザー要望に応じて） |

---

## 2. 設計方針の決定事項

| 項目 | 決定 | 理由 |
|------|------|------|
| 削除方式 | **物理削除** | 復元ニーズが現時点でなく、論理削除はクエリ全体への影響が大きい |
| 復元機能 | 不要（将来検討の余地あり） | MVP として最小スコープを優先 |
| 削除権限 | タグオーナー（`tags.user_id`）のみ | 既存の更新権限と同じ考え方 |
| 削除済み URL | 404 +「このタグは存在しません」 | 物理削除のため存在/削除済みの区別は不可 |
| アーカイブ機能 | 不要 | 公開設定（`is_public`）で代替可能 |
| 子テーブルの削除 | アプリケーション層でトランザクション内削除 | FK 制約が未定義のため CASCADE に依存できない |

---

## 3. Phase 1: タグの物理削除

### 3.1 API 設計

#### DELETE `/api/v1/tags/:tagId`

| 項目 | 内容 |
|------|------|
| 認証 | 必須 |
| 権限 | タグオーナー（`tags.user_id == リクエストユーザー`）のみ |
| 成功レスポンス | `204 No Content` |

**エラーレスポンス**:

| ステータス | 条件 | レスポンス例 |
|-----------|------|-------------|
| 401 | 未認証 | `{"error": "unauthorized"}` |
| 403 | オーナー以外 | `{"error": "tag permission denied"}` |
| 404 | タグが存在しない | `{"error": "tag not found"}` |

### 3.2 削除対象と順序

タグ削除時に、以下の子テーブルのレコードを **1 トランザクション内で** 削除する。

```
BEGIN
  1. DELETE FROM notifications   WHERE tag_id = :tagId
  2. DELETE FROM tag_likes       WHERE tag_id = :tagId
  3. DELETE FROM tag_followers    WHERE tag_id = :tagId
  4. DELETE FROM tag_movies       WHERE tag_id = :tagId
  5. DELETE FROM tags             WHERE id = :tagId
COMMIT
```

**順序の理由**: `notifications` テーブルは `tag_movies(id)` への FK 参照（`ON DELETE CASCADE`）を持つため、`tag_movies` より先に削除する。他のテーブルは FK 未定義だが、将来の FK 追加に備えて子→親の順序を守る。

### 3.3 バックエンド実装

#### 3.3.1 Repository 層

**ファイル**: `apps/backend/src/internal/repository/tag_repository.go`

`TagRepository` インターフェースにメソッドを追加:

```go
type TagRepository interface {
    // ...既存メソッド...
    DeleteByID(ctx context.Context, id string) error
}
```

実装では `gorm.DB` のトランザクション内で子テーブル → 親テーブルの順に削除する。

```go
func (r *tagRepository) DeleteByID(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("tag_id = ?", id).Delete(&model.Notification{}).Error; err != nil {
            return err
        }
        if err := tx.Where("tag_id = ?", id).Delete(&model.TagLike{}).Error; err != nil {
            return err
        }
        if err := tx.Where("tag_id = ?", id).Delete(&model.TagFollower{}).Error; err != nil {
            return err
        }
        if err := tx.Where("tag_id = ?", id).Delete(&model.TagMovie{}).Error; err != nil {
            return err
        }
        return tx.Where("id = ?", id).Delete(&model.Tag{}).Error
    })
}
```

#### 3.3.2 Service 層

**ファイル**: `apps/backend/src/internal/service/tag_service.go`

`TagService` インターフェースにメソッドを追加:

```go
type TagService interface {
    // ...既存メソッド...
    DeleteTag(ctx context.Context, tagID string, userID string) error
}
```

**ビジネスロジック**:

1. `tagRepo.FindByID(tagID)` でタグの存在確認
2. `tag.UserID != userID` なら `ErrTagPermissionDenied` を返す
3. `tagRepo.DeleteByID(tagID)` で物理削除

```go
func (s *tagService) DeleteTag(ctx context.Context, tagID string, userID string) error {
    tag, err := s.tagRepo.FindByID(ctx, tagID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrTagNotFound
        }
        return fmt.Errorf("find tag: %w", err)
    }
    if tag.UserID != userID {
        return ErrTagPermissionDenied
    }
    if err := s.tagRepo.DeleteByID(ctx, tagID); err != nil {
        return fmt.Errorf("delete tag: %w", err)
    }
    return nil
}
```

#### 3.3.3 Handler 層

**ファイル**: `apps/backend/src/internal/handler/tag_handler.go`

```go
// DeleteTag は DELETE /api/v1/tags/:tagId を処理する。
func (h *TagHandler) DeleteTag(c *gin.Context) {
    tagID := c.Param("tagId")
    userID := c.GetString("userID")

    if err := h.tagService.DeleteTag(c.Request.Context(), tagID, userID); err != nil {
        switch {
        case errors.Is(err, service.ErrTagNotFound):
            c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
        case errors.Is(err, service.ErrTagPermissionDenied):
            c.JSON(http.StatusForbidden, gin.H{"error": "tag permission denied"})
        default:
            slog.ErrorContext(c.Request.Context(), "failed to delete tag", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        }
        return
    }
    c.Status(http.StatusNoContent)
}
```

#### 3.3.4 Router 登録

**ファイル**: `apps/backend/src/router/router.go`

```go
authGroup.DELETE("/tags/:tagId", deps.TagHandler.DeleteTag)
```

### 3.4 フロントエンド実装

#### 3.4.1 API 関数

**ファイル**: `apps/frontend/src/lib/api/tags/deleteTag.ts`

```typescript
export async function deleteTag(params: { tagId: string; token: string }) {
  const res = await fetch(
    `${process.env.NEXT_PUBLIC_BACKEND_API_BASE}/api/v1/tags/${params.tagId}`,
    {
      method: "DELETE",
      headers: { Authorization: `Bearer ${params.token}` },
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "タグの削除に失敗しました");
  }
}
```

#### 3.4.2 削除確認ダイアログ

**ファイル**: タグ詳細ページ（`apps/frontend/src/app/tags/[tagId]/_components/`）に削除ボタンと確認モーダルを追加。

**表示条件**: 閲覧者がタグオーナーの場合のみ削除ボタンを表示。

**確認ダイアログの内容**:

- **通常時**: 「このタグを削除しますか？この操作は取り消せません。」
- **他ユーザーが参加している場合（映画追加 or フォロワーが存在）**: 警告を強調表示
  - 「このタグには **{n}人のフォロワー** と **他のユーザーが追加した{m}件の映画** があります。削除するとこれらのデータもすべて失われます。本当に削除しますか？」

**ダイアログに必要な情報**:

| 情報 | 取得元 |
|------|--------|
| フォロワー数 | 既存の `GET /api/v1/tags/:tagId` レスポンスの `follower_count` |
| 他ユーザーが追加した映画数 | 後述（3.5 参照） |

**削除成功後の遷移先**: マイページ（`/users/{display_id}`）にリダイレクト。

#### 3.4.3 削除後のキャッシュ無効化

React Query のキャッシュから削除したタグに関連するクエリを無効化する。

```typescript
queryClient.invalidateQueries({ queryKey: ["tags"] });
queryClient.invalidateQueries({ queryKey: ["user-tags"] });
queryClient.removeQueries({ queryKey: ["tag", tagId] });
```

### 3.5 「他ユーザーが追加した映画数」の取得

現在の `GET /api/v1/tags/:tagId` のレスポンスには `movie_count` はあるが、「他ユーザーが追加した映画数」は含まれていない。

**選択肢**:

- **A. 既存レスポンスに `other_users_movie_count` を追加**: タグ詳細取得時に常に計算する（クエリコスト増）
- **B. フロント側で `tag_movies` の `added_by_user_id` を集計**: 既存の映画一覧 API から計算可能だが、全件取得が必要
- **C. 削除確認時のみ取得する専用エンドポイント**: 過剰設計

**推奨**: **A**。`movie_count` を取得している既存クエリに `COUNT CASE WHEN added_by_user_id != :ownerUserId` を追加するだけで済む。タグ詳細ページでは常にオーナー情報を持っているため、フロント側での判定も容易。

### 3.6 テスト計画

#### バックエンド

| テスト | 内容 |
|--------|------|
| Service ユニットテスト | オーナーによる削除成功 / 非オーナーの 403 / 存在しないタグの 404 |
| Handler ユニットテスト | 各レスポンスコードの検証 |
| Repository ユニットテスト | 子テーブルを含む削除の検証 |
| インテグレーションテスト | 実 DB での削除 + 子テーブルのクリーンアップ確認 |

#### フロントエンド

| テスト | 内容 |
|--------|------|
| 手動テスト | 削除ボタン表示条件 / 確認ダイアログの文言 / 削除後のリダイレクト |
| 手動テスト | 他ユーザー参加時の警告表示 |

### 3.7 実装ステップ

| ステップ | 内容 | 依存 |
|---------|------|------|
| 1-1 | `TagRepository.DeleteByID` 実装 | なし |
| 1-2 | `TagService.DeleteTag` 実装 | 1-1 |
| 1-3 | `TagHandler.DeleteTag` 実装 + Router 登録 | 1-2 |
| 1-4 | バックエンドテスト（ユニット + インテグレーション） | 1-3 |
| 1-5 | タグ詳細レスポンスに `other_users_movie_count` 追加（任意） | なし |
| 1-6 | フロントエンド API 関数 `deleteTag` 作成 | 1-3 |
| 1-7 | 削除ボタン + 確認ダイアログ UI 実装 | 1-6 |
| 1-8 | 削除後のリダイレクト + キャッシュ無効化 | 1-7 |
| 1-9 | `docs/api/api-spec.md` の [未実装] マーク削除 + レスポンス仕様更新 | 1-4 |

---

## 4. Phase 2: タグオーナーの譲渡

### 4.1 概要

タグオーナーが別のユーザーにオーナー権限を譲渡する機能。タグを削除したくないが管理も手放したいケースや、共同編集タグの管理者交代に対応する。

**導入タイミング**: Phase 1 の運用後、ユーザーからの要望に応じて判断。

### 4.2 譲渡フロー

**即時譲渡方式**を採用する（承認制は実装コストが高く、MVP では過剰）。

```
オーナーが譲渡先ユーザーを指定
  → API で tags.user_id を更新
    → 譲渡完了通知を双方に送信
```

**将来的に承認制が必要になった場合**: `tag_transfer_requests` テーブルを追加し、pending → accepted/rejected のステートマシンで管理する。

### 4.3 API 設計

#### PATCH `/api/v1/tags/:tagId/owner`

| 項目 | 内容 |
|------|------|
| 認証 | 必須 |
| 権限 | 現オーナーのみ |

**リクエストボディ**:

```json
{
  "new_owner_user_id": "uuid"
}
```

**レスポンス（200 OK）**: 更新後のタグ詳細（`TagDetail` と同じ構造）

**エラーレスポンス**:

| ステータス | 条件 |
|-----------|------|
| 400 | `new_owner_user_id` が未指定 / 自分自身を指定 |
| 401 | 未認証 |
| 403 | オーナー以外 |
| 404 | タグまたは譲渡先ユーザーが存在しない |

### 4.4 譲渡先の条件

- 譲渡先は **cinetag に登録済みの任意のユーザー**（フォロワー限定にはしない）
- 自分自身への譲渡は 400 エラー

### 4.5 通知

| 通知タイプ | 通知先 | メッセージ例 |
|-----------|--------|-------------|
| `tag_ownership_received` | 新オーナー | **{username}** からタグ「{tag_title}」のオーナー権限が譲渡されました |
| `tag_ownership_transferred` | 旧オーナー | タグ「{tag_title}」のオーナー権限を **{username}** に譲渡しました |

### 4.6 バックエンド実装概要

#### Repository 層

```go
type TagRepository interface {
    // ...既存メソッド...
    UpdateOwner(ctx context.Context, tagID string, newOwnerUserID string) error
}
```

#### Service 層

```go
type TagService interface {
    // ...既存メソッド...
    TransferTagOwnership(ctx context.Context, tagID, currentUserID, newOwnerUserID string) (*TagDetail, error)
}
```

**ビジネスロジック**:

1. タグの存在確認
2. 現オーナーの権限確認
3. 譲渡先ユーザーの存在確認
4. 自分自身への譲渡を拒否
5. `tags.user_id` を更新
6. 双方に通知を送信

#### Handler 層

```go
func (h *TagHandler) TransferTagOwnership(c *gin.Context)
```

#### Router 登録

```go
authGroup.PATCH("/tags/:tagId/owner", deps.TagHandler.TransferTagOwnership)
```

### 4.7 フロントエンド実装概要

- タグ設定画面（またはタグ詳細ページの設定セクション）に「オーナーを譲渡」ボタンを追加
- ユーザー検索 UI で譲渡先を選択
- 確認ダイアログ:「タグ「{title}」のオーナー権限を **{display_name}** に譲渡しますか？この操作後、あなたはこのタグの編集・削除ができなくなります。」

### 4.8 削除との連携

Phase 2 実装後は、削除確認ダイアログに以下の導線を追加する:

- 「他ユーザーが参加しています」の警告表示時に、「オーナーを譲渡する」リンクを併記
- ユーザーが削除ではなく譲渡を選択できるようにする

### 4.9 実装ステップ

| ステップ | 内容 | 依存 |
|---------|------|------|
| 2-1 | `TagRepository.UpdateOwner` 実装 | なし |
| 2-2 | `TagService.TransferTagOwnership` 実装 | 2-1 |
| 2-3 | `TagHandler.TransferTagOwnership` 実装 + Router 登録 | 2-2 |
| 2-4 | 通知タイプ追加（`tag_ownership_received`, `tag_ownership_transferred`） | 2-3 |
| 2-5 | バックエンドテスト | 2-3 |
| 2-6 | フロントエンド API 関数 + UI 実装 | 2-3 |
| 2-7 | 削除確認ダイアログに譲渡導線を追加 | Phase 1 完了 + 2-6 |
| 2-8 | `docs/api/api-spec.md` 更新 | 2-5 |

---

## 5. 技術的リスクと対策

| リスク | 影響 | 対策 |
|--------|------|------|
| FK 未定義による子テーブルの削除漏れ | 孤立レコードが残る | トランザクション内で明示的に全子テーブルを削除。将来の FK 追加マイグレーションで整合性を DB 層でも保証 |
| 大量の子レコードがあるタグの削除が遅い | タイムアウト | 現時点のデータ規模では問題にならない。将来的にバッチ削除を検討 |
| 削除直後に同じ URL にアクセスされる | キャッシュされた古いデータが表示される可能性 | フロント側でキャッシュ無効化を徹底。CDN キャッシュは現時点で未使用のため問題なし |
| 譲渡先ユーザーが退会済み | 無効なオーナーが設定される | `users.deleted_at IS NULL` の存在チェックを譲渡時に実施 |

---

## 6. 参考

| 項目 | 参照先 |
|------|--------|
| API 仕様（タグ削除） | `docs/api/api-spec.md` セクション 5.5 |
| 既存の削除実装（映画削除） | `tag_handler.go` の `RemoveMovieFromTag` |
| 既存の権限チェックパターン | `tag_service.go` の `UpdateTag` |
| 改善バックログ | `docs/plans/20260406_cinetag-improvement-backlog.md` |
