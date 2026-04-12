//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"cinetag-backend/src/internal/testutil"
)

// GET /api/v1/users/me
// 認証済みユーザーが自身のプロフィール情報を取得でき、全フィールドが正しく返ることを確認する。
func TestGetMe_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_me1", "me-user1", "MeUser1")

	resp := env.request("GET", "/api/v1/users/me", nil, authHeaders(user.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"id":           user.ID,
		"display_id":   "me-user1",
		"display_name": "MeUser1",
	})
	testutil.AssertHasKeys(t, data, "id", "display_id", "display_name")
}

// GET /api/v1/users/me
// 認証なしでアクセスすると 401 と error メッセージが返ることを確認する。
func TestGetMe_Unauthorized(t *testing.T) {
	env := setupTestEnv(t)

	resp := env.request("GET", "/api/v1/users/me", nil, nil)
	resp.AssertStatus(t, 401)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "unauthorized",
	})
}

// PATCH /api/v1/users/me
// 認証済みユーザーが表示名を更新でき、全フィールドが正しく返ることを確認する。
func TestUpdateMe_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_upd1", "upd-user1", "OldName")

	body, _ := json.Marshal(map[string]any{
		"display_name": "NewName",
	})
	resp := env.request("PATCH", "/api/v1/users/me", body, authHeaders(user.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"id":           user.ID,
		"display_id":   "upd-user1",
		"display_name": "NewName",
	})
}

// GET /api/v1/users/:displayId
// display_id でユーザーを取得し、全フィールドが正しく返ることを確認する。
func TestGetUserByDisplayID_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_disp1", "disp-user1", "DispUser1")

	resp := env.request("GET", "/api/v1/users/disp-user1", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"id":           user.ID,
		"display_id":   "disp-user1",
		"display_name": "DispUser1",
	})
	testutil.AssertHasKeys(t, data, "id", "display_id", "display_name")
}

// GET /api/v1/users/:displayId
// 存在しない display_id でアクセスすると 404 と error メッセージが返ることを確認する。
func TestGetUserByDisplayID_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp := env.request("GET", "/api/v1/users/nonexistent-user", nil, nil)
	resp.AssertStatus(t, 404)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "user not found",
	})
}

// GET /api/v1/users/:displayId/tags
// ユーザーの公開タグ一覧を取得し、非公開タグが含まれないこと、ページネーション・items の全フィールドを確認する。
func TestListUserTags_PublicOnly(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_ut1", "ut-user1", "UTUser1")

	body, _ := json.Marshal(map[string]any{"title": "公開タグ", "is_public": true})
	r := env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	r.AssertStatus(t, 201)

	body, _ = json.Marshal(map[string]any{"title": "非公開タグ", "is_public": false})
	r = env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	r.AssertStatus(t, 201)

	resp := env.request("GET", "/api/v1/users/ut-user1/tags", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 1, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	testutil.AssertJSON(t, items[0], map[string]any{
		"title":    "公開タグ",
		"is_public": true,
		"author":   "UTUser1",
		"author_display_id": "ut-user1",
	})
	testutil.AssertHasKeys(t, items[0],
		"id", "title", "author", "author_display_id",
		"is_public", "movie_count", "follower_count", "like_count",
		"images", "created_at",
	)
}

// GET /api/v1/users/:displayId/tags
// 本人がアクセスすると非公開タグを含む全タグが返ることを確認する。
func TestListUserTags_OwnerSeesAll(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_ut2", "ut-user2", "UTUser2")

	body, _ := json.Marshal(map[string]any{"title": "公開タグ", "is_public": true})
	r := env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	r.AssertStatus(t, 201)

	body, _ = json.Marshal(map[string]any{"title": "非公開タグ", "is_public": false})
	r = env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	r.AssertStatus(t, 201)

	resp := env.request("GET", "/api/v1/users/ut-user2/tags", nil, authHeaders(user.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 2, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

// POST /api/v1/users/:displayId/follow
// 認証済みユーザーが他のユーザーをフォローし、200 と成功メッセージが返ることを確認する。
func TestFollowUser_Success(t *testing.T) {
	env := setupTestEnv(t)
	follower := env.createUser(t, "clerk_fw1", "fw-user1", "Follower1")
	env.createUser(t, "clerk_fw2", "fw-user2", "Followee1")

	resp := env.request("POST", "/api/v1/users/fw-user2/follow", nil, authHeaders(follower.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"message": "successfully followed",
	})
}

// POST /api/v1/users/:displayId/follow
// 自分自身をフォローしようとすると 400 と error メッセージが返ることを確認する。
func TestFollowUser_Self(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_fws1", "fws-user1", "SelfFollower")

	resp := env.request("POST", "/api/v1/users/fws-user1/follow", nil, authHeaders(user.ID))
	resp.AssertStatus(t, 400)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "cannot follow yourself",
	})
}

// POST /api/v1/users/:displayId/follow
// 既にフォロー済みのユーザーを再フォローすると 409 と error メッセージが返ることを確認する。
func TestFollowUser_AlreadyFollowing(t *testing.T) {
	env := setupTestEnv(t)
	follower := env.createUser(t, "clerk_fwd1", "fwd-user1", "Follower2")
	env.createUser(t, "clerk_fwd2", "fwd-user2", "Followee2")

	r := env.request("POST", "/api/v1/users/fwd-user2/follow", nil, authHeaders(follower.ID))
	r.AssertStatus(t, 200)

	resp := env.request("POST", "/api/v1/users/fwd-user2/follow", nil, authHeaders(follower.ID))
	resp.AssertStatus(t, 409)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "already following",
	})
}

// DELETE /api/v1/users/:displayId/follow
// フォロー済みユーザーをアンフォローし、200 と成功メッセージが返ることを確認する。
func TestUnfollowUser_Success(t *testing.T) {
	env := setupTestEnv(t)
	follower := env.createUser(t, "clerk_uf1", "uf-user1", "Unfollower1")
	env.createUser(t, "clerk_uf2", "uf-user2", "Unfollowee1")

	r := env.request("POST", "/api/v1/users/uf-user2/follow", nil, authHeaders(follower.ID))
	r.AssertStatus(t, 200)

	resp := env.request("DELETE", "/api/v1/users/uf-user2/follow", nil, authHeaders(follower.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"message": "successfully unfollowed",
	})
}

// DELETE /api/v1/users/:displayId/follow
// フォローしていないユーザーをアンフォローすると 409 と error メッセージが返ることを確認する。
func TestUnfollowUser_NotFollowing(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_ufn1", "ufn-user1", "NotFollowing1")
	env.createUser(t, "clerk_ufn2", "ufn-user2", "NotFollowing2")

	resp := env.request("DELETE", "/api/v1/users/ufn-user2/follow", nil, authHeaders(user.ID))
	resp.AssertStatus(t, 409)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "not following",
	})
}

// GET /api/v1/users/:displayId/following
// ユーザーのフォロー中一覧を取得し、ページネーション・items 内のユーザー情報を全検証する。
func TestListFollowing_Success(t *testing.T) {
	env := setupTestEnv(t)
	follower := env.createUser(t, "clerk_lf1", "lf-user1", "LFUser1")
	followee := env.createUser(t, "clerk_lf2", "lf-user2", "LFUser2")

	r := env.request("POST", "/api/v1/users/lf-user2/follow", nil, authHeaders(follower.ID))
	r.AssertStatus(t, 200)

	resp := env.request("GET", "/api/v1/users/lf-user1/following", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 1, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	testutil.AssertJSON(t, items[0], map[string]any{
		"id":           followee.ID,
		"display_id":   "lf-user2",
		"display_name": "LFUser2",
	})
	testutil.AssertHasKeys(t, items[0], "id", "display_id", "display_name")
}

// GET /api/v1/users/:displayId/followers
// ユーザーのフォロワー一覧を取得し、ページネーション・items 内のユーザー情報を全検証する。
func TestListFollowers_Success(t *testing.T) {
	env := setupTestEnv(t)
	follower := env.createUser(t, "clerk_lr1", "lr-user1", "LRUser1")
	env.createUser(t, "clerk_lr2", "lr-user2", "LRUser2")

	r := env.request("POST", "/api/v1/users/lr-user2/follow", nil, authHeaders(follower.ID))
	r.AssertStatus(t, 200)

	resp := env.request("GET", "/api/v1/users/lr-user2/followers", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 1, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	testutil.AssertJSON(t, items[0], map[string]any{
		"id":           follower.ID,
		"display_id":   "lr-user1",
		"display_name": "LRUser1",
	})
	testutil.AssertHasKeys(t, items[0], "id", "display_id", "display_name")
}

// GET /api/v1/users/:displayId/follow-stats
// フォロー数・フォロワー数・is_following の全フィールドが正しく返ることを確認する。
func TestGetUserFollowStats_Success(t *testing.T) {
	env := setupTestEnv(t)
	userA := env.createUser(t, "clerk_fs1", "fs-user1", "FSUser1")
	env.createUser(t, "clerk_fs2", "fs-user2", "FSUser2")

	r := env.request("POST", "/api/v1/users/fs-user2/follow", nil, authHeaders(userA.ID))
	r.AssertStatus(t, 200)

	resp := env.request("GET", "/api/v1/users/fs-user2/follow-stats", nil, authHeaders(userA.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"following_count": float64(0),
		"followers_count": float64(1),
		"is_following":    true,
	})
}
