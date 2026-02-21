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
	ListPublicTagsFn     func(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error)
	ListTagsByUserIDFn   func(ctx context.Context, userID string, publicOnly bool, page, pageSize int) ([]service.TagListItem, int64, error)
	GetTagDetailFn       func(ctx context.Context, tagID string, viewerUserID *string) (*service.TagDetail, error)
	ListTagMoviesFn      func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error)
	CreateTagFn          func(ctx context.Context, in service.CreateTagInput) (*model.Tag, error)
	AddMoviesToTagFn     func(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error)
	UpdateTagFn          func(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error)
	RemoveMovieFromTagFn func(ctx context.Context, tagMovieID string, userID string) error
	FollowTagFn          func(ctx context.Context, tagID, userID string) error
	UnfollowTagFn        func(ctx context.Context, tagID, userID string) error
	IsFollowingTagFn     func(ctx context.Context, tagID, userID string) (bool, error)
	ListTagFollowersFn   func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error)
	ListFollowingTagsFn  func(ctx context.Context, userID string, page, pageSize int) ([]service.TagListItem, int64, error)
}

func (f *fakeTagService) ListPublicTags(ctx context.Context, q, sort string, page, pageSize int) ([]service.TagListItem, int64, error) {
	if f.ListPublicTagsFn == nil {
		return []service.TagListItem{}, 0, nil
	}
	return f.ListPublicTagsFn(ctx, q, sort, page, pageSize)
}

func (f *fakeTagService) ListTagsByUserID(ctx context.Context, userID string, publicOnly bool, page, pageSize int) ([]service.TagListItem, int64, error) {
	if f.ListTagsByUserIDFn == nil {
		return []service.TagListItem{}, 0, nil
	}
	return f.ListTagsByUserIDFn(ctx, userID, publicOnly, page, pageSize)
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

func (f *fakeTagService) AddMoviesToTag(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error) {
	if f.AddMoviesToTagFn == nil {
		return &service.AddMoviesResult{}, nil
	}
	return f.AddMoviesToTagFn(ctx, in)
}

func (f *fakeTagService) UpdateTag(ctx context.Context, tagID string, userID string, patch service.UpdateTagPatch) (*service.TagDetail, error) {
	if f.UpdateTagFn == nil {
		return &service.TagDetail{}, nil
	}
	return f.UpdateTagFn(ctx, tagID, userID, patch)
}

func (f *fakeTagService) RemoveMovieFromTag(ctx context.Context, tagMovieID string, userID string) error {
	if f.RemoveMovieFromTagFn == nil {
		return nil
	}
	return f.RemoveMovieFromTagFn(ctx, tagMovieID, userID)
}

func (f *fakeTagService) FollowTag(ctx context.Context, tagID, userID string) error {
	if f.FollowTagFn == nil {
		return nil
	}
	return f.FollowTagFn(ctx, tagID, userID)
}

func (f *fakeTagService) UnfollowTag(ctx context.Context, tagID, userID string) error {
	if f.UnfollowTagFn == nil {
		return nil
	}
	return f.UnfollowTagFn(ctx, tagID, userID)
}

func (f *fakeTagService) IsFollowingTag(ctx context.Context, tagID, userID string) (bool, error) {
	if f.IsFollowingTagFn == nil {
		return false, nil
	}
	return f.IsFollowingTagFn(ctx, tagID, userID)
}

func (f *fakeTagService) ListTagFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListTagFollowersFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListTagFollowersFn(ctx, tagID, page, pageSize)
}

func (f *fakeTagService) ListFollowingTags(ctx context.Context, userID string, page, pageSize int) ([]service.TagListItem, int64, error) {
	if f.ListFollowingTagsFn == nil {
		return []service.TagListItem{}, 0, nil
	}
	return f.ListFollowingTagsFn(ctx, userID, page, pageSize)
}

// newTagHandlerRouter は TagHandler のテスト用ルーターを生成します。
func newTagHandlerRouter(t *testing.T, tagSvc service.TagService, user *model.User) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	logger := testutil.NewTestLogger()
	h := NewTagHandler(logger, tagSvc)

	api := r.Group("/api/v1")
	api.GET("/tags", h.ListPublicTags)

	// Optional Auth (認証なしでもアクセス可能)
	optionalAuth := api.Group("/")
	if user != nil {
		optionalAuth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	optionalAuth.GET("/tags/:tagId", h.GetTagDetail)
	optionalAuth.GET("/tags/:tagId/movies", h.ListTagMovies)
	optionalAuth.GET("/tags/:tagId/followers", h.ListTagFollowers)

	// 認証が必要なエンドポイント
	auth := api.Group("/")
	if user != nil {
		auth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	auth.POST("/tags", h.CreateTag)
	auth.PATCH("/tags/:tagId", h.UpdateTag)
	auth.POST("/tags/:tagId/movies", h.AddMoviesToTag)
	auth.DELETE("/tags/:tagId/movies/:tagMovieId", h.RemoveMovieFromTag)
	auth.POST("/tags/:tagId/follow", h.FollowTag)
	auth.DELETE("/tags/:tagId/follow", h.UnfollowTag)
	auth.GET("/tags/:tagId/follow-status", h.GetTagFollowStatus)
	auth.GET("/me/following-tags", h.ListFollowingTags)

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

	t.Run("user が無効な型: 500", func(t *testing.T) {
		t.Parallel()

		r := testutil.NewTestRouter()
		logger := testutil.NewTestLogger()
		h := NewTagHandler(logger, &fakeTagService{})

		r.Use(func(c *gin.Context) {
			// 無効な型をセット
			c.Set("user", "invalid-user-type")
			c.Next()
		})
		r.POST("/api/v1/tags", h.CreateTag)

		body := testutil.MustMarshalJSON(t, map[string]any{"title": "t"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags", body, map[string]string{
			"Content-Type": "application/json",
		})
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
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

func TestTagHandler_AddMoviesToTag(t *testing.T) {
	t.Parallel()

	url := "/api/v1/tags/t1/movies"
	jsonHeader := map[string]string{"Content-Type": "application/json"}

	t.Run("不正なJSON: 400", func(t *testing.T) {
		t.Parallel()
		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, url, []byte("{"), jsonHeader)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("moviesが空: 400", func(t *testing.T) {
		t.Parallel()
		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"movies": []any{}})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("movies上限超過(51件): 400", func(t *testing.T) {
		t.Parallel()
		movies := make([]any, 51)
		for i := range movies {
			movies[i] = map[string]any{"tmdb_movie_id": i + 1, "position": 0}
		}
		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{"movies": movies})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("tmdb_movie_idが正でない: 400", func(t *testing.T) {
		t.Parallel()
		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 0, "position": 0},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("positionが負: 400", func(t *testing.T) {
		t.Parallel()
		r := newTagHandlerRouter(t, &fakeTagService{}, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 1, "position": -1},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()
		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 1, "position": 0},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグが存在しない: 404", func(t *testing.T) {
		t.Parallel()
		svc := &fakeTagService{
			AddMoviesToTagFn: func(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error) {
				return nil, service.ErrTagNotFound
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 1, "position": 0},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし: 403", func(t *testing.T) {
		t.Parallel()
		svc := &fakeTagService{
			AddMoviesToTagFn: func(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error) {
				return nil, service.ErrTagPermissionDenied
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 1, "position": 0},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})

	t.Run("全件成功: 201", func(t *testing.T) {
		t.Parallel()
		var got service.AddMoviesToTagInput
		svc := &fakeTagService{
			AddMoviesToTagFn: func(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error) {
				got = in
				return &service.AddMoviesResult{
					Results: []service.MovieResult{
						{TmdbMovieID: 10, Status: "created"},
						{TmdbMovieID: 20, Status: "created"},
					},
					Summary: service.AddMoviesSummary{Created: 2},
				}, nil
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 10, "position": 0},
				map[string]any{"tmdb_movie_id": 20, "position": 1},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rw.Code)
		}
		if got.TagID != "t1" || got.UserID != "u1" || len(got.Movies) != 2 {
			t.Fatalf("unexpected input: %+v", got)
		}
		if got.Movies[0].TmdbMovieID != 10 || got.Movies[1].TmdbMovieID != 20 {
			t.Fatalf("unexpected movie ids: %+v", got.Movies)
		}
	})

	t.Run("部分成功(重複含む): 207", func(t *testing.T) {
		t.Parallel()
		svc := &fakeTagService{
			AddMoviesToTagFn: func(ctx context.Context, in service.AddMoviesToTagInput) (*service.AddMoviesResult, error) {
				return &service.AddMoviesResult{
					Results: []service.MovieResult{
						{TmdbMovieID: 10, Status: "created"},
						{TmdbMovieID: 20, Status: "already_exists", Error: "movie already added to tag"},
					},
					Summary: service.AddMoviesSummary{Created: 1, AlreadyExists: 1},
				}, nil
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		body := testutil.MustMarshalJSON(t, map[string]any{
			"movies": []any{
				map[string]any{"tmdb_movie_id": 10, "position": 0},
				map[string]any{"tmdb_movie_id": 20, "position": 0},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, url, body, jsonHeader)
		if rw.Code != http.StatusMultiStatus {
			t.Fatalf("expected 207, got %d", rw.Code)
		}
		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		summary := resp["summary"].(map[string]any)
		if summary["created"] != float64(1) {
			t.Fatalf("expected created=1, got %v", summary["created"])
		}
		if summary["already_exists"] != float64(1) {
			t.Fatalf("expected already_exists=1, got %v", summary["already_exists"])
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

func TestTagHandler_RemoveMovieFromTag(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/movies/tm1", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグ映画が存在しない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			RemoveMovieFromTagFn: func(ctx context.Context, tagMovieID string, userID string) error {
				return service.ErrTagMovieNotFound
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/movies/tm1", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし: 403", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			RemoveMovieFromTagFn: func(ctx context.Context, tagMovieID string, userID string) error {
				return service.ErrTagPermissionDenied
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/movies/tm1", nil, nil)
		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("boom")
		svc := &fakeTagService{
			RemoveMovieFromTagFn: func(ctx context.Context, tagMovieID string, userID string) error {
				return expected
			},
		}
		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/movies/tm1", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 204 No Content", func(t *testing.T) {
		t.Parallel()

		var gotTagMovieID, gotUserID string
		svc := &fakeTagService{
			RemoveMovieFromTagFn: func(ctx context.Context, tagMovieID string, userID string) error {
				gotTagMovieID = tagMovieID
				gotUserID = userID
				return nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/movies/tm1", nil, nil)
		if rw.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rw.Code)
		}
		if gotTagMovieID != "tm1" || gotUserID != "u1" {
			t.Fatalf("unexpected input: tagMovieID=%s userID=%s", gotTagMovieID, gotUserID)
		}
	})
}

func TestTagHandler_ListTagMovies(t *testing.T) {
	t.Parallel()

	t.Run("タグが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListTagMoviesFn: func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error) {
				return nil, 0, service.ErrTagNotFound
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/movies", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし: 403", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListTagMoviesFn: func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error) {
				return nil, 0, service.ErrTagPermissionDenied
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/movies", nil, nil)
		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListTagMoviesFn: func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/movies", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotTagID string
		var gotPage, gotPageSize int
		svc := &fakeTagService{
			ListTagMoviesFn: func(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]service.TagMovieItem, int64, error) {
				gotTagID = tagID
				gotPage = page
				gotPageSize = pageSize
				return []service.TagMovieItem{
					{ID: "tm1", TmdbMovieID: 123, TagID: "t1"},
				}, 1, nil
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/movies?page=2&page_size=10", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotTagID != "t1" {
			t.Fatalf("expected tagID=t1, got %s", gotTagID)
		}
		if gotPage != 2 {
			t.Fatalf("expected page=2, got %d", gotPage)
		}
		if gotPageSize != 10 {
			t.Fatalf("expected pageSize=10, got %d", gotPageSize)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
	})
}

func TestTagHandler_FollowTag(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			FollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return service.ErrTagNotFound
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("権限なし(非公開タグ): 403", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			FollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return service.ErrTagPermissionDenied
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rw.Code)
		}
	})

	t.Run("既にフォロー済み: 409", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			FollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return service.ErrAlreadyFollowingTag
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			FollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotTagID, gotUserID string
		svc := &fakeTagService{
			FollowTagFn: func(ctx context.Context, tagID, userID string) error {
				gotTagID = tagID
				gotUserID = userID
				return nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotTagID != "t1" || gotUserID != "u1" {
			t.Fatalf("unexpected args: tagID=%s userID=%s", gotTagID, gotUserID)
		}
	})
}

func TestTagHandler_UnfollowTag(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("タグが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			UnfollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return service.ErrTagNotFound
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("フォローしていない: 409", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			UnfollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return service.ErrNotFollowingTag
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			UnfollowTagFn: func(ctx context.Context, tagID, userID string) error {
				return errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotTagID, gotUserID string
		svc := &fakeTagService{
			UnfollowTagFn: func(ctx context.Context, tagID, userID string) error {
				gotTagID = tagID
				gotUserID = userID
				return nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/tags/t1/follow", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotTagID != "t1" || gotUserID != "u1" {
			t.Fatalf("unexpected args: tagID=%s userID=%s", gotTagID, gotUserID)
		}
	})
}

func TestTagHandler_GetTagFollowStatus(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/follow-status", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			IsFollowingTagFn: func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/follow-status", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功(フォロー中): 200", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			IsFollowingTagFn: func(ctx context.Context, tagID, userID string) (bool, error) {
				return true, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/follow-status", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["is_following"] != true {
			t.Fatalf("expected is_following=true, got %v", resp["is_following"])
		}
	})

	t.Run("成功(フォローしていない): 200", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			IsFollowingTagFn: func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/follow-status", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["is_following"] != false {
			t.Fatalf("expected is_following=false, got %v", resp["is_following"])
		}
	})
}

func TestTagHandler_ListTagFollowers(t *testing.T) {
	t.Parallel()

	t.Run("タグが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListTagFollowersFn: func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				return nil, 0, service.ErrTagNotFound
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/followers", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListTagFollowersFn: func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/followers", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		avatarURL := "https://example.com/avatar.png"
		svc := &fakeTagService{
			ListTagFollowersFn: func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				return []*model.User{
					{ID: "u1", DisplayID: "user1", DisplayName: "User 1", AvatarURL: &avatarURL},
				}, 1, nil
			},
		}

		r := newTagHandlerRouter(t, svc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/tags/t1/followers", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
	})
}

func TestTagHandler_ListFollowingTags(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newTagHandlerRouter(t, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/me/following-tags", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		svc := &fakeTagService{
			ListFollowingTagsFn: func(ctx context.Context, userID string, page, pageSize int) ([]service.TagListItem, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/me/following-tags", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotUserID string
		var gotPage, gotPageSize int
		svc := &fakeTagService{
			ListFollowingTagsFn: func(ctx context.Context, userID string, page, pageSize int) ([]service.TagListItem, int64, error) {
				gotUserID = userID
				gotPage = page
				gotPageSize = pageSize
				return []service.TagListItem{
					{ID: "t1", Title: "Tag 1", Author: "User 1", AuthorDisplayID: "user1"},
				}, 1, nil
			},
		}

		r := newTagHandlerRouter(t, svc, &model.User{ID: "u1"})
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/me/following-tags?page=2&page_size=10", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotUserID != "u1" {
			t.Fatalf("expected userID=u1, got %s", gotUserID)
		}
		if gotPage != 2 {
			t.Fatalf("expected page=2, got %d", gotPage)
		}
		if gotPageSize != 10 {
			t.Fatalf("expected pageSize=10, got %d", gotPageSize)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
	})
}
