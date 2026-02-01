package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
)

type fakeMovieService struct {
	SearchMoviesFn     func(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error)
	EnsureMovieCacheFn func(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error)
}

func (f *fakeMovieService) SearchMovies(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error) {
	if f.SearchMoviesFn == nil {
		return []service.TMDBSearchResult{}, 0, nil
	}
	return f.SearchMoviesFn(ctx, query, page)
}

func (f *fakeMovieService) EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
	if f.EnsureMovieCacheFn == nil {
		return nil, nil
	}
	return f.EnsureMovieCacheFn(ctx, tmdbMovieID)
}

func newMovieHandlerRouter(t *testing.T, movieSvc service.MovieService) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	logger := testutil.NewTestLogger()
	h := NewMovieHandler(logger, movieSvc)

	api := r.Group("/api/v1")
	api.GET("/movies/search", h.SearchMovies)

	return r
}

func TestMovieHandler_SearchMovies(t *testing.T) {
	t.Parallel()

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		movieSvc := &fakeMovieService{
			SearchMoviesFn: func(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error) {
				return nil, 0, errors.New("tmdb error")
			},
		}

		r := newMovieHandlerRouter(t, movieSvc)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/movies/search?q=test", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotQuery string
		var gotPage int
		title := "Test Movie"
		posterPath := "/test.jpg"
		movieSvc := &fakeMovieService{
			SearchMoviesFn: func(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error) {
				gotQuery = query
				gotPage = page
				return []service.TMDBSearchResult{
					{TmdbMovieID: 123, Title: title, PosterPath: &posterPath},
				}, 1, nil
			},
		}

		r := newMovieHandlerRouter(t, movieSvc)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/movies/search?q=test&page=2", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotQuery != "test" {
			t.Fatalf("expected query=test, got %s", gotQuery)
		}
		if gotPage != 2 {
			t.Fatalf("expected page=2, got %d", gotPage)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
		if resp["page"] != float64(2) {
			t.Fatalf("expected page=2, got %v", resp["page"])
		}
	})

	t.Run("成功（空のクエリ）: 200", func(t *testing.T) {
		t.Parallel()

		var gotQuery string
		movieSvc := &fakeMovieService{
			SearchMoviesFn: func(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error) {
				gotQuery = query
				return []service.TMDBSearchResult{}, 0, nil
			},
		}

		r := newMovieHandlerRouter(t, movieSvc)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/movies/search?q=", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotQuery != "" {
			t.Fatalf("expected query=empty, got %s", gotQuery)
		}
	})

	t.Run("成功（デフォルトページ）: 200", func(t *testing.T) {
		t.Parallel()

		var gotPage int
		movieSvc := &fakeMovieService{
			SearchMoviesFn: func(ctx context.Context, query string, page int) ([]service.TMDBSearchResult, int, error) {
				gotPage = page
				return []service.TMDBSearchResult{}, 0, nil
			},
		}

		r := newMovieHandlerRouter(t, movieSvc)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/movies/search?q=test", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotPage != 1 {
			t.Fatalf("expected page=1, got %d", gotPage)
		}
	})
}
