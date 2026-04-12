//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"cinetag-backend/src/internal/testutil"
)

// POST /api/v1/tags
// 認証済みユーザーがタグを作成し、201 と全フィールドが正しく返ることを確認する。
func TestCreateTag_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_tag1", "tag-user1", "TagUser1")

	body, _ := json.Marshal(map[string]any{
		"title":    "お気に入り映画",
		"is_public": true,
	})

	resp := env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	resp.AssertStatus(t, 201)

	data := resp.JSON(t)
	testutil.AssertHasKeys(t, data, "id", "created_at", "updated_at")
	testutil.AssertJSON(t, data, map[string]any{
		"title":            "お気に入り映画",
		"is_public":        true,
		"add_movie_policy": "everyone",
		"movie_count":      float64(0),
		"follower_count":   float64(0),
	})
	if data["description"] != nil {
		t.Errorf("expected description=nil, got %v", data["description"])
	}
	if data["cover_image_url"] != nil {
		t.Errorf("expected cover_image_url=nil, got %v", data["cover_image_url"])
	}
}

// POST /api/v1/tags
// 認証なしでタグ作成を試みると 401 と error メッセージが返ることを確認する。
func TestCreateTag_Unauthorized(t *testing.T) {
	env := setupTestEnv(t)

	body, _ := json.Marshal(map[string]any{
		"title": "テストタグ",
	})

	resp := env.request("POST", "/api/v1/tags", body, jsonHeaders())
	resp.AssertStatus(t, 401)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "unauthorized",
	})
}

// POST /api/v1/tags
// title が未指定のリクエストボディでタグ作成すると 400 と error メッセージが返ることを確認する。
func TestCreateTag_InvalidBody(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_tag2", "tag-user2", "TagUser2")

	resp := env.request("POST", "/api/v1/tags", []byte(`{}`), authHeaders(user.ID))
	resp.AssertStatus(t, 400)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "invalid request body",
	})
}

// GET /api/v1/tags
// タグが存在しない状態で公開タグ一覧を取得すると、200 と空の一覧が返ることを確認する。
func TestListPublicTags_Empty(t *testing.T) {
	env := setupTestEnv(t)

	resp := env.request("GET", "/api/v1/tags", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 0, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

// GET /api/v1/tags
// 公開タグ2件と非公開タグ1件を作成し、公開タグ一覧では公開分のみが返ることを確認する。
// items 内の各タグの全フィールドも検証する。
func TestListPublicTags_WithData(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_list1", "list-user1", "ListUser1")

	for i := 1; i <= 2; i++ {
		body, _ := json.Marshal(map[string]any{
			"title":    fmt.Sprintf("公開タグ%d", i),
			"is_public": true,
		})
		r := env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
		r.AssertStatus(t, 201)
	}

	body, _ := json.Marshal(map[string]any{
		"title":    "非公開タグ",
		"is_public": false,
	})
	r := env.request("POST", "/api/v1/tags", body, authHeaders(user.ID))
	r.AssertStatus(t, 201)

	resp := env.request("GET", "/api/v1/tags", nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertListResponse(t, data, 2, 1, 20)

	items := testutil.GetItems(t, data)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	for _, item := range items {
		testutil.AssertHasKeys(t, item,
			"id", "title", "author", "author_display_id",
			"is_public", "movie_count", "follower_count", "like_count",
			"images", "created_at",
		)
		testutil.AssertJSON(t, item, map[string]any{
			"is_public":        true,
			"author":           "ListUser1",
			"author_display_id": "list-user1",
			"movie_count":      float64(0),
			"follower_count":   float64(0),
			"like_count":       float64(0),
		})
	}
}

// GET /api/v1/tags/:tagId
// 存在しないタグ ID で詳細を取得すると 404 と error メッセージが返ることを確認する。
func TestGetTagDetail_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp := env.request("GET", "/api/v1/tags/00000000-0000-0000-0000-000000000000", nil, nil)
	resp.AssertStatus(t, 404)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "tag not found",
	})
}

// GET /api/v1/tags/:tagId
// タグを作成後、そのタグ ID で詳細を取得し、全フィールドが正しく返ることを確認する。
func TestGetTagDetail_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_detail1", "detail-user1", "DetailUser1")

	createBody, _ := json.Marshal(map[string]any{
		"title":       "詳細テスト",
		"description": "テスト用の説明",
		"is_public":   true,
	})
	createResp := env.request("POST", "/api/v1/tags", createBody, authHeaders(user.ID))
	createResp.AssertStatus(t, 201)
	created := createResp.JSON(t)
	tagID := created["id"].(string)

	resp := env.request("GET", "/api/v1/tags/"+tagID, nil, nil)
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertHasKeys(t, data,
		"id", "title", "description", "is_public", "add_movie_policy",
		"movie_count", "follower_count", "like_count", "is_liked",
		"owner", "can_edit", "can_add_movie",
		"participant_count", "participants",
		"created_at", "updated_at",
	)
	testutil.AssertJSON(t, data, map[string]any{
		"id":               tagID,
		"title":            "詳細テスト",
		"description":      "テスト用の説明",
		"is_public":        true,
		"add_movie_policy": "everyone",
		"movie_count":      float64(0),
		"follower_count":   float64(0),
		"like_count":       float64(0),
		"is_liked":         false,
		"can_edit":         false,
		"can_add_movie":    false,
		"participant_count": float64(0),
	})

	owner, ok := data["owner"].(map[string]any)
	if !ok {
		t.Fatalf("owner がオブジェクトではありません: %v", data["owner"])
	}
	testutil.AssertJSON(t, owner, map[string]any{
		"id":           user.ID,
		"display_id":   "detail-user1",
		"display_name": "DetailUser1",
	})
}

// PATCH /api/v1/tags/:tagId
// タグ作成者がタイトルを更新し、200 と更新後の全フィールドが返ることを確認する。
func TestUpdateTag_Success(t *testing.T) {
	env := setupTestEnv(t)
	user := env.createUser(t, "clerk_update1", "update-user1", "UpdateUser1")

	createBody, _ := json.Marshal(map[string]any{
		"title":    "更新前タイトル",
		"is_public": true,
	})
	createResp := env.request("POST", "/api/v1/tags", createBody, authHeaders(user.ID))
	createResp.AssertStatus(t, 201)
	tagID := createResp.JSON(t)["id"].(string)

	updateBody, _ := json.Marshal(map[string]any{
		"title": "更新後タイトル",
	})
	resp := env.request("PATCH", "/api/v1/tags/"+tagID, updateBody, authHeaders(user.ID))
	resp.AssertStatus(t, 200)

	data := resp.JSON(t)
	testutil.AssertHasKeys(t, data,
		"id", "title", "is_public", "add_movie_policy",
		"movie_count", "follower_count", "like_count", "is_liked",
		"owner", "can_edit", "can_add_movie",
		"participant_count", "participants",
		"created_at", "updated_at",
	)
	testutil.AssertJSON(t, data, map[string]any{
		"id":               tagID,
		"title":            "更新後タイトル",
		"is_public":        true,
		"add_movie_policy": "everyone",
		"movie_count":      float64(0),
		"follower_count":   float64(0),
		"like_count":       float64(0),
		"can_edit":         true,
	})
}

// PATCH /api/v1/tags/:tagId
// タグ作成者以外のユーザーがタグを更新しようとすると 403 と error メッセージが返ることを確認する。
func TestUpdateTag_Forbidden(t *testing.T) {
	env := setupTestEnv(t)
	owner := env.createUser(t, "clerk_owner1", "owner1", "Owner1")
	other := env.createUser(t, "clerk_other1", "other1", "Other1")

	createBody, _ := json.Marshal(map[string]any{
		"title":    "他人のタグ",
		"is_public": true,
	})
	createResp := env.request("POST", "/api/v1/tags", createBody, authHeaders(owner.ID))
	createResp.AssertStatus(t, 201)
	tagID := createResp.JSON(t)["id"].(string)

	updateBody, _ := json.Marshal(map[string]any{
		"title": "勝手に更新",
	})
	resp := env.request("PATCH", "/api/v1/tags/"+tagID, updateBody, authHeaders(other.ID))
	resp.AssertStatus(t, 403)

	data := resp.JSON(t)
	testutil.AssertJSON(t, data, map[string]any{
		"error": "forbidden",
	})
}
