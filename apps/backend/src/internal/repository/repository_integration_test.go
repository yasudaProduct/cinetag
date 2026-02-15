//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// integration テスト用の DB を開きます。
func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL が未設定のため integration テストをスキップします")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("DB接続に失敗: %v", err)
	}

	// UUID のデフォルト（gen_random_uuid）を使うため
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`).Error; err != nil {
		t.Fatalf("pgcrypto extension の有効化に失敗: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Tag{}, &model.TagMovie{}, &model.TagFollower{}); err != nil {
		t.Fatalf("AutoMigrate に失敗: %v", err)
	}

	// 各テストの独立性を担保するため、対象テーブルをクリーンにする。
	// NOTE: integration テスト専用DBで実行すること（開発用DBでは実行しない）。
	if err := db.Exec(`TRUNCATE TABLE tag_followers, tag_movies, tags, users RESTART IDENTITY CASCADE;`).Error; err != nil {
		t.Fatalf("テスト用DBの初期化（TRUNCATE）に失敗: %v", err)
	}

	return db
}

// beginTx はトランザクションを開始します。
func beginTx(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("トランザクション開始に失敗: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})
	return tx
}

func createUser(t *testing.T, db *gorm.DB, clerkID, displayName string) *model.User {
	t.Helper()

	u := &model.User{
		ClerkUserID: clerkID,
		DisplayName: displayName,
		Email:       displayName + "@example.com",
	}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	if u.ID == "" {
		t.Fatalf("ユーザーIDが空です")
	}
	return u
}

func createTag(t *testing.T, db *gorm.DB, userID, title string, isPublic bool) *model.Tag {
	t.Helper()

	tag := &model.Tag{
		UserID:   userID,
		Title:    title,
		IsPublic: isPublic,
	}
	if err := db.Create(tag).Error; err != nil {
		t.Fatalf("タグ作成に失敗: %v", err)
	}
	if tag.ID == "" {
		t.Fatalf("タグIDが空です")
	}
	return tag
}

// 重複追加はエラーになる想定
func TestTagMovieRepository_Create_Duplicate(t *testing.T) {
	db := openIntegrationDB(t)
	tx := beginTx(t, db)

	u := createUser(t, tx, "clerk_u1", "user1")
	tag := createTag(t, tx, u.ID, "tag1", true)

	repo := NewTagMovieRepository(tx)
	ctx := context.Background()

	m1 := &model.TagMovie{
		TagID:       tag.ID,
		TmdbMovieID: 100,
		AddedByUser: u.ID,
		Position:    0,
	}
	if err := repo.Create(ctx, m1); err != nil {
		t.Fatalf("1回目の作成で失敗: %v", err)
	}

	m2 := &model.TagMovie{
		TagID:       tag.ID,
		TmdbMovieID: 100,
		AddedByUser: u.ID,
		Position:    0,
	}
	err := repo.Create(ctx, m2)
	if err == nil {
		t.Fatalf("重複追加はエラーになる想定")
	}
	if err != ErrTagMovieAlreadyExists {
		t.Fatalf("expected ErrTagMovieAlreadyExists, got %v", err)
	}

	var count int64
	if err := tx.Model(&model.TagMovie{}).Where("tag_id = ? AND tmdb_movie_id = ?", tag.ID, 100).Count(&count).Error; err != nil {
		t.Fatalf("件数取得に失敗: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestTagMovieRepository_ListRecentByTag_NewestFirst(t *testing.T) {
	db := openIntegrationDB(t)
	tx := beginTx(t, db)

	u := createUser(t, tx, "clerk_u1", "user1")
	tag := createTag(t, tx, u.ID, "tag1", true)

	now := time.Now().UTC()
	rows := []*model.TagMovie{
		{TagID: tag.ID, TmdbMovieID: 1, AddedByUser: u.ID, Position: 0, CreatedAt: now.Add(-3 * time.Minute)},
		{TagID: tag.ID, TmdbMovieID: 2, AddedByUser: u.ID, Position: 0, CreatedAt: now.Add(-2 * time.Minute)},
		{TagID: tag.ID, TmdbMovieID: 3, AddedByUser: u.ID, Position: 0, CreatedAt: now.Add(-1 * time.Minute)},
	}
	if err := tx.Create(&rows).Error; err != nil {
		t.Fatalf("事前データ作成に失敗: %v", err)
	}

	repo := NewTagMovieRepository(tx)
	ctx := context.Background()

	got, err := repo.ListRecentByTag(ctx, tag.ID, 2)
	if err != nil {
		t.Fatalf("ListRecentByTag に失敗: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].TmdbMovieID != 3 || got[1].TmdbMovieID != 2 {
		t.Fatalf("expected [3,2], got [%d,%d]", got[0].TmdbMovieID, got[1].TmdbMovieID)
	}
}

// follower_count の降順になるように設定
func TestTagRepository_ListPublicTags_FollowerCountDesc(t *testing.T) {
	db := openIntegrationDB(t)
	tx := beginTx(t, db)

	u1 := createUser(t, tx, "clerk_u1", "alice")
	u2 := createUser(t, tx, "clerk_u2", "bob")
	u3 := createUser(t, tx, "clerk_u3", "charlie")

	// 非公開タグは一覧に含まれない
	_ = createTag(t, tx, u1.ID, "非公開タグ", false)
	t1 := createTag(t, tx, u1.ID, "公開タグA", true)
	t2 := createTag(t, tx, u2.ID, "公開タグB", true)

	// t2 にフォロワーを2人追加、t1 にはフォロワー1人
	for _, uid := range []string{u1.ID, u3.ID} {
		if err := tx.Create(&model.TagFollower{TagID: t2.ID, UserID: uid}).Error; err != nil {
			t.Fatalf("tag_followers INSERT に失敗: %v", err)
		}
	}
	if err := tx.Create(&model.TagFollower{TagID: t1.ID, UserID: u2.ID}).Error; err != nil {
		t.Fatalf("tag_followers INSERT に失敗: %v", err)
	}

	repo := NewTagRepository(tx)
	ctx := context.Background()

	rows, total, err := repo.ListPublicTags(ctx, TagListFilter{Query: "", Sort: "", Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("ListPublicTags に失敗: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// デフォルトは follower_count DESC — t2(2) > t1(1)
	if rows[0].ID != t2.ID || rows[1].ID != t1.ID {
		t.Fatalf("expected order [B,A], got [%s,%s]", rows[0].ID, rows[1].ID)
	}
	if rows[0].Author != "bob" {
		t.Fatalf("expected author=bob, got %q", rows[0].Author)
	}
	if rows[0].FollowerCount != 2 {
		t.Fatalf("expected follower_count=2, got %d", rows[0].FollowerCount)
	}
	if rows[1].FollowerCount != 1 {
		t.Fatalf("expected follower_count=1, got %d", rows[1].FollowerCount)
	}
}

func TestTagRepository_ListPublicTags_TitleSearch(t *testing.T) {
	db := openIntegrationDB(t)
	tx := beginTx(t, db)

	u := createUser(t, tx, "clerk_u1", "alice")
	_ = createTag(t, tx, u.ID, "ドラマ特集", true)
	_ = createTag(t, tx, u.ID, "アクション特集", true)

	repo := NewTagRepository(tx)
	ctx := context.Background()

	rows, total, err := repo.ListPublicTags(ctx, TagListFilter{Query: "ドラマ", Sort: "recent", Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("ListPublicTags に失敗: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected 1 match, got total=%d len=%d", total, len(rows))
	}
	if rows[0].Title != "ドラマ特集" {
		t.Fatalf("expected title=ドラマ特集, got %q", rows[0].Title)
	}
}
