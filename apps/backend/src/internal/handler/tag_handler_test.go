package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
)

type fakeTagService struct {
	ListPublicTagsFn func(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error)
	GetTagDetailFn   func(ctx context.Context, tagID string, viewerUserID *string) (*service.TagDetail, error)
	ListTagMoviesFn  func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error)
	CreateTagFn      func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error)
	AddMovieToTagFn  func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error)
	UpdateTagFn      func(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error)
}

func (f *fakeTagService) ListPublicTags(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
	if f.ListPublicTagsFn == nil {
		return []service.TagListItem{}, 0, nil
	}
	return f.ListPublicTagsFn(ctx, q, sort, page, pageSize)
}

func (f *fakeTagService) GetTagDetail(ctx context.Context, tagID string, viewerUserID *string) (*service.TagDetail, error) {
	if f.GetTagDetailFn == nil {
		return &service.TagDetail{}, nil
	}
	return f.GetTagDetailFn(ctx, tagID, viewerUserID)
}

func (f *fakeTagService) ListTagMovies(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error) {
	if f.ListTagMoviesFn == nil {
		return []service.TagMovieItem{}, 0, nil
	}
	return f.ListTagMoviesFn(ctx, tagID, viewerUserID, page, pageSize)
}

func (f *fakeTagService) CreateTag(ctx context.Context, in service.CreateTagInput) (*model.Tag, error) {
	if f.CreateTagFn == nil {
		return &model.Tag{}, nil
	}
	return f.CreateTagFn(ctx, in)
}

func (f *fakeTagService) AddMovieToTag(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
	if f.AddMovieToTagFn == nil {
		return &model.TagMovie{}, nil
	}
	return f.AddMovieToTagFn(ctx, in)
}

func (f *fakeTagService) UpdateTag(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error) {
	if f.UpdateTagFn == nil {
		return &service.TagDetail{}, nil
	}
	return f.UpdateTagFn(ctx, tagID, userID, patch)
}

// newTagHandlerRouter は TagHandler のテスト用ルーターを生成します。
func newTagHandlerRouter(t *testing.T, tagSvc service.TagService, user *model.User) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	h := NewTagHandler(tagSvc)

	api := r.Group("/api/v1")
	api.GET("/tags", h.ListPublicTags)

	// GetTagDetailはOptionalAuthなので、userが設定されている場合のみ設定
	getTagDetailGroup := api.Group("/")
	if user != nil {
		getTagDetailGroup.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	getTagDetailGroup.GET("/tags/:tagId", h.GetTagDetail)

	auth := api.Group("/")
	if user != nil {
		auth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	auth.POST("/tags", h.CreateTag)
	auth.PATCH("/tags/:tagId", h.UpdateTag)
	auth.POST("/tags/:tagId/movies", h.AddMovieToTag)

	return r
}

func TestTagHandler_CreateTag(t *testing.T) {
	t.Parallel()

	t.Run("不正なJSON: 400", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", []byte("{"), map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("必須項目不足(title無し): 400", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		body := testutil.MustMarshalJSON(t, map[string]any{"title": "t"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タイトル長が100を超える: 400", func(t *testing.T) {
		t.Parallel()

		title := ""
		for i := 0; i < 101; i++ {
			title += "a"
		}

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"title": title})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("説明文長が500を超える: 400", func(t *testing.T) {
		t.Parallel()

		desc := ""
		for i := 0; i < 501; i++ {
			desc += "a"
		}

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"title": "t", "description": desc})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("boom")
		svc := &fakeTagService{
			CreateTagFn: func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error) {
				return nil, expected
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"title": "t"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 201 かつ service への入力が正しい", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		var got service.CreateTagInput
		svc := &fakeTagService{
			CreateTagFn: func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error) {
				got = in
				return &model.Tag{
					ID:            "tag1",
					UserID:        in.UserID,
					Title:         in.Title,
					Description:   in.Description,
					CoverImageURL: in.CoverImageURL,
					IsPublic:      true,
					MovieCount:    0,
					FollowerCount: 0,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newTagHandlerRouter(t, svc, u)

		desc := "d"
		cover := "https://example.com/x.png"
		isPublic := true
		body := testutil.MustMarshalJSON(t, map[string]any{
			"title":           "t",
			"description":     desc,
			"cover_image_url": cover,
			"is_public":       isPublic,
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rw.Code)
		}

		if got.UserID != "u1" || got.Title != "t" {
			t.Fatalf("unexpected input: %+v", got)
		}
		if got.Description == nil || *got.Description != desc {
			t.Fatalf("expected description to be passed")
		}
		if got.CoverImageURL == nil || *got.CoverImageURL != cover {
			t.Fatalf("expected cover_image_url to be passed")
		}
		if got.IsPublic == nil || *got.IsPublic != isPublic {
			t.Fatalf("expected is_public to be passed")
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["id"] != "tag1" {
			t.Fatalf("expected id=tag1, got %v", resp["id"])
		}
		if resp["title"] != "t" {
			t.Fatalf("expected title=t, got %v", resp["title"])
		}
		if resp["created_at"] == "" || resp["updated_at"] == "" {
			t.Fatalf("expected created_at/updated_at")
		}
	})

	t.Run("成功: add_movie_policy が指定された場合、正しく渡される", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		var got service.CreateTagInput
		svc := &fakeTagService{
			CreateTagFn: func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error) {
				got = in
				return &model.Tag{
					ID:             "tag1",
					UserID:         in.UserID,
					Title:          in.Title,
					AddMoviePolicy: "owner_only",
					IsPublic:       true,
					MovieCount:     0,
					FollowerCount:  0,
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newTagHandlerRouter(t, svc, u)

		body := testutil.MustMarshalJSON(t, map[string]any{
			"title":            "t",
			"add_movie_policy": "owner_only",
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rw.Code)
		}

		if got.AddMoviePolicy == nil || *got.AddMoviePolicy != "owner_only" {
			t.Fatalf("expected add_movie_policy=owner_only, got %v", got.AddMoviePolicy)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["add_movie_policy"] != "owner_only" {
			t.Fatalf("expected add_movie_policy=owner_only in response, got %v", resp["add_movie_policy"])
		}
	})
}

func TestTagHandler_AddMovieToTag(t *testing.T) {
	t.Parallel()

	t.Run("不正なJSON: 400", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", []byte("{"), map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("tmdb_movie_id が正でない: 400", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 0, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("position が負: 400", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": -1})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグが存在しない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			AddMovieToTagFn: func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
				return nil, service.ErrTagNotFound
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし: 403", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			AddMovieToTagFn: func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
				return nil, service.ErrTagPermissionDenied
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})

	t.Run("既に追加済み: 409", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			AddMovieToTagFn: func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
				return nil, service.ErrTagMovieAlreadyExists
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("boom")
		svc := &fakeTagService{
			AddMovieToTagFn: func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
				return nil, expected
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 1, "position": 0})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 201", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		var got service.AddMovieToTagInput
		svc := &fakeTagService{
			AddMovieToTagFn: func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error) {
				got = in
				note := "n"
				return &model.TagMovie{
					ID:          "tm1",
					TagID:       in.TagID,
					TmdbMovieID: in.TmdbMovieID,
					AddedByUser: in.UserID,
					Note:        &note,
					Position:    in.Position,
					CreatedAt:   now,
				}, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"tmdb_movie_id": 10, "position": 2})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/movies", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rw.Code)
		}
		if got.TagID != "t1" || got.UserID != "u1" || got.TmdbMovieID != 10 || got.Position != 2 {
			t.Fatalf("unexpected input: %+v", got)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["tag_id"] != "t1" {
			t.Fatalf("expected tag_id=t1, got %v", resp["tag_id"])
		}
		if resp["tmdb_movie_id"] == nil {
			t.Fatalf("expected tmdb_movie_id")
		}
		if resp["added_by_user_id"] != "u1" {
			t.Fatalf("expected added_by_user_id=u1, got %v", resp["added_by_user_id"])
		}
	})
}

func TestTagHandler_ListPublicTags(t *testing.T) {
	t.Parallel()

	t.Run("デフォルト(page=1,page_size=20)でサービスが呼ばれる: 200", func(t *testing.T) {
		t.Parallel()

		var gotQ, gotSort string
		var gotPage, gotPageSize int
		svc := &fakeTagService{
			ListPublicTagsFn: func(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
				gotQ, gotSort, gotPage, gotPageSize = q, sort, page, pageSize
				return []service.TagListItem{}, 0, nil
			},
		}
		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotQ != "" || gotSort != "" || gotPage != 1 || gotPageSize != 20 {
			t.Fatalf("unexpected args: q=%q sort=%q page=%d pageSize=%d", gotQ, gotSort, gotPage, gotPageSize)
		}
	})

	t.Run("不正なpage/page_sizeはデフォルトにフォールバック: 200", func(t *testing.T) {
		t.Parallel()

		var gotPage, gotPageSize int
		svc := &fakeTagService{
			ListPublicTagsFn: func(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
				gotPage, gotPageSize = page, pageSize
				return []service.TagListItem{}, 0, nil
			},
		}
		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags?page=x&page_size=y", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotPage != 1 || gotPageSize != 20 {
			t.Fatalf("expected defaults (1,20), got (%d,%d)", gotPage, gotPageSize)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("boom")
		svc := &fakeTagService{
			ListPublicTagsFn: func(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
				return nil, 0, expected
			},
		}
		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})
}

func TestTagHandler_GetTagDetail(t *testing.T) {
	t.Parallel()

	t.Run("成功: can_add_movie が正しく返される", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			GetTagDetailFn: func(ctx context.Context, tagID string, viewerUserID *string) (*service.TagDetail, error) {
				return &service.TagDetail{
					ID:             "t1",
					Title:          "Test Tag",
					AddMoviePolicy: "everyone",
					CanAddMovie:    true,
					CanEdit:        false,
					Owner: service.TagOwner{
						ID:          "u1",
						Username:    "user1",
						DisplayName: "User1",
					},
				}, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1", nil, nil)

		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["can_add_movie"] != true {
			t.Fatalf("expected can_add_movie=true, got %v", resp["can_add_movie"])
		}
		if resp["add_movie_policy"] != "everyone" {
			t.Fatalf("expected add_movie_policy=everyone, got %v", resp["add_movie_policy"])
		}
	})

	t.Run("未認証ユーザー: can_add_movie=false が返される", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			GetTagDetailFn: func(ctx context.Context, tagID string, viewerUserID *string) (*service.TagDetail, error) {
				return &service.TagDetail{
					ID:             "t1",
					Title:          "Test Tag",
					AddMoviePolicy: "everyone",
					CanAddMovie:    false,
					CanEdit:        false,
					Owner: service.TagOwner{
						ID:          "u1",
						Username:    "user1",
						DisplayName: "User1",
					},
				}, nil
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1", nil, nil)

		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["can_add_movie"] != false {
			t.Fatalf("expected can_add_movie=false, got %v", resp["can_add_movie"])
		}
	})
}

func TestTagHandler_UpdateTag(t *testing.T) {
	t.Parallel()

	t.Run("成功: add_movie_policy が更新される", func(t *testing.T) {
		t.Parallel()

		var gotPatch service.UpdateTagPatch
		svc := &fakeTagService{
			UpdateTagFn: func(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error) {
				gotPatch = patch
				return &service.TagDetail{
					ID:             tagID,
					Title:          "Updated Tag",
					AddMoviePolicy: "owner_only",
					CanEdit:        true,
					CanAddMovie:    true,
					Owner: service.TagOwner{
						ID:          userID,
						Username:    "user1",
						DisplayName: "User1",
					},
				}, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"add_movie_policy": "owner_only",
		})
		rw := testutil.PerformRequest(r, http.MethodPatch, "/api/v1/tags/t1", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		if gotPatch.AddMoviePolicy == nil || *gotPatch.AddMoviePolicy != "owner_only" {
			t.Fatalf("expected AddMoviePolicy=owner_only, got %v", gotPatch.AddMoviePolicy)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["add_movie_policy"] != "owner_only" {
			t.Fatalf("expected add_movie_policy=owner_only in response, got %v", resp["add_movie_policy"])
		}
	})

	t.Run("未認証: 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		body := testutil.MustMarshalJSON(t, map[string]any{
			"title": "Updated",
		})
		rw := testutil.PerformRequest(r, http.MethodPatch, "/api/v1/tags/t1", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグが存在しない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			UpdateTagFn: func(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error) {
				return nil, service.ErrTagNotFound
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"title": "Updated",
		})
		rw := testutil.PerformRequest(r, http.MethodPatch, "/api/v1/tags/t1", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし: 403", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			UpdateTagFn: func(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error) {
				return nil, service.ErrTagPermissionDenied
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"title": "Updated",
		})
		rw := testutil.PerformRequest(r, http.MethodPatch, "/api/v1/tags/t1", body, map[string]string{
			"Content-Type": "application/json",
		})

		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})
}
