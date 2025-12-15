package testutil

import (
	"context"

	"cinetag-backend/src/internal/model"
)

// FakeMovieService は service.MovieService の最小 fake（必要なメソッドのみ）です。
// 使うテストで必要になったら拡張します。
//
// NOTE: 現時点で TagService は MovieService を interface として受け取っているため、
// テストでは nil を渡して非同期処理（キャッシュ温め）を発火させない運用が基本です。
type FakeMovieService struct {
	EnsureMovieCacheFn func(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error)
}

func (f *FakeMovieService) EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
	if f.EnsureMovieCacheFn == nil {
		return &model.MovieCache{}, nil
	}
	return f.EnsureMovieCacheFn(ctx, tmdbMovieID)
}
