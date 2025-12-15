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
	CreateTagFn      func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error)
	AddMovieToTagFn  func(ctx context.Context, in service.AddMovieToTagInput) (*model.TagMovie, error)
}

func (f *fakeTagService) ListPublicTags(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
	if f.ListPublicTagsFn == nil {
		return []service.TagListItem{}, 0, nil
	}
	return f.ListPublicTagsFn(ctx, q, sort, page, pageSize)
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

func newTagHandlerRouter(t *testing.T, tagSvc service.TagService, user *model.User) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	h := NewTagHandler(tagSvc)

	api := r.Group("/api/v1")
	api.GET("/tags", h.ListPublicTags)

	auth := api.Group("/")
	if user != nil {
		auth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	auth.POST("/tags", h.CreateTag)
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
