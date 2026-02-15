package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/testutil"

	"gorm.io/gorm"
)

// fakeMovieService は MovieService の fake 実装です。
type fakeMovieService struct {
	EnsureMovieCacheFn func(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error)
	SearchMoviesFn     func(ctx context.Context, query string, page int) ([]TMDBSearchResult, int, error)
}

func (f *fakeMovieService) EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
	if f.EnsureMovieCacheFn == nil {
		return &model.MovieCache{}, nil
	}
	return f.EnsureMovieCacheFn(ctx, tmdbMovieID)
}

func (f *fakeMovieService) SearchMovies(ctx context.Context, query string, page int) ([]TMDBSearchResult, int, error) {
	if f.SearchMoviesFn == nil {
		return []TMDBSearchResult{}, 0, nil
	}
	return f.SearchMoviesFn(ctx, query, page)
}

type deps struct {
	tagRepo         *testutil.FakeTagRepository
	tagMovieRepo    *testutil.FakeTagMovieRepository
	tagFollowerRepo *testutil.FakeTagFollowerRepository
	movieService    MovieService
	imageBaseURL    string
}

func newTagService(t *testing.T, opt func(*deps)) TagService {
	t.Helper()

	logger := testutil.NewTestLogger()
	d := &deps{
		tagRepo:         &testutil.FakeTagRepository{},
		tagMovieRepo:    &testutil.FakeTagMovieRepository{},
		tagFollowerRepo: &testutil.FakeTagFollowerRepository{},
		movieService:    nil,
		imageBaseURL:    "",
	}
	if opt != nil {
		opt(d)
	}
	return NewTagService(logger, d.tagRepo, d.tagMovieRepo, d.tagFollowerRepo, d.movieService, d.imageBaseURL)
}

func TestTagService_AddMovieToTag(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "",
			TmdbMovieID: 1,
			Position:    0,
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: tmdb_movie_id は正の整数", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 0,
			Position:    0,
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: position は 0 以上", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    -1,
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("タグが見つからない: gorm.ErrRecordNotFound は ErrTagNotFound に変換される", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if !errors.Is(err, ErrTagNotFound) {
			t.Fatalf("expected ErrTagNotFound, got: %v", err)
		}
	})

	t.Run("タグ検索で失敗: FindByID のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("db down")
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, expected
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("重複追加: repository.ErrTagMovieAlreadyExists は ErrTagMovieAlreadyExists に変換される", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				return repository.ErrTagMovieAlreadyExists
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if !errors.Is(err, ErrTagMovieAlreadyExists) {
			t.Fatalf("expected ErrTagMovieAlreadyExists, got: %v", err)
		}
	})

	t.Run("タグ映画の作成で失敗: TagMovieRepository.Create のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("insert failed")
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				return expected
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("成功: tag_movie を作成する", func(t *testing.T) {
		t.Parallel()

		var created *model.TagMovie

		tagRepo := &testutil.FakeTagRepository{
			FindByIDFn: func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			},
		}
		tagMovieRepo := &testutil.FakeTagMovieRepository{
			CreateFn: func(ctx context.Context, tagMovie *model.TagMovie) error {
				created = &model.TagMovie{
					TagID:       tagMovie.TagID,
					TmdbMovieID: tagMovie.TmdbMovieID,
					AddedByUser: tagMovie.AddedByUser,
					Note:        tagMovie.Note,
					Position:    tagMovie.Position,
				}
				return nil
			},
		}

		logger := testutil.NewTestLogger()
		svc := NewTagService(logger, tagRepo, tagMovieRepo, &testutil.FakeTagFollowerRepository{}, nil, "")

		// Act
		note := "hello"
		out, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 123,
			Note:        &note,
			Position:    2,
		})

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out == nil {
			t.Fatalf("expected output")
		}
		if created == nil {
			t.Fatalf("expected TagMovieRepository.Create to be called")
		}

		if created.TagID != "t1" || created.AddedByUser != "u1" || created.TmdbMovieID != 123 || created.Position != 2 {
			t.Fatalf("unexpected created tag movie: %+v", created)
		}
		if created.Note == nil || *created.Note != note {
			t.Fatalf("expected note to be set")
		}
	})

	t.Run("権限チェック: add_movie_policy=owner_only の場合、作成者以外は ErrTagPermissionDenied", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "owner_only",
				}, nil
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "other_user",
			TmdbMovieID: 1,
			Position:    0,
		})
		if !errors.Is(err, ErrTagPermissionDenied) {
			t.Fatalf("expected ErrTagPermissionDenied, got: %v", err)
		}
	})

	t.Run("権限チェック: add_movie_policy=owner_only の場合、作成者は成功", func(t *testing.T) {
		t.Parallel()

		var created *model.TagMovie
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "owner_only",
				}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				created = tagMovie
				return nil
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "owner1",
			TmdbMovieID: 1,
			Position:    0,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if created == nil {
			t.Fatalf("expected TagMovieRepository.Create to be called")
		}
	})

	t.Run("権限チェック: add_movie_policy=everyone の場合、誰でも追加可能", func(t *testing.T) {
		t.Parallel()

		var created *model.TagMovie
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "everyone",
				}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				created = tagMovie
				return nil
			}
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "other_user",
			TmdbMovieID: 1,
			Position:    0,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if created == nil {
			t.Fatalf("expected TagMovieRepository.Create to be called")
		}
	})

	t.Run("非同期処理: movieService がある場合、キャッシュウォームが実行される", func(t *testing.T) {
		t.Parallel()

		// goroutine が呼ばれたことを検証するためのチャネル
		done := make(chan int, 1)

		movieSvc := &fakeMovieService{
			EnsureMovieCacheFn: func(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
				done <- tmdbMovieID
				return &model.MovieCache{}, nil
			},
		}

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				return nil
			}
			d.movieService = movieSvc
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 123,
			Position:    0,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// goroutine の完了を待つ（タイムアウト付き）
		select {
		case movieID := <-done:
			if movieID != 123 {
				t.Fatalf("expected movieID=123, got %d", movieID)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("expected EnsureMovieCache to be called within 1 second")
		}
	})

	t.Run("非同期処理: movieService のエラーは無視される（APIは成功）", func(t *testing.T) {
		t.Parallel()

		done := make(chan bool, 1)

		movieSvc := &fakeMovieService{
			EnsureMovieCacheFn: func(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
				done <- true
				return nil, errors.New("cache warm failed")
			},
		}

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagMovieRepo.CreateFn = func(ctx context.Context, tagMovie *model.TagMovie) error {
				return nil
			}
			d.movieService = movieSvc
		})

		_, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 123,
			Position:    0,
		})
		// APIは成功する
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// goroutine が実行されたことを確認
		select {
		case <-done:
			// OK
		case <-time.After(1 * time.Second):
			t.Fatalf("expected EnsureMovieCache to be called within 1 second")
		}
	})
}

func TestTagService_CreateTag(t *testing.T) {
	t.Parallel()

	t.Run("デフォルト: IsPublic 未指定なら true で作成される", func(t *testing.T) {
		t.Parallel()

		var created *model.Tag
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.CreateFn = func(ctx context.Context, tag *model.Tag) error {
				created = tag
				tag.ID = "tag1"
				return nil
			}
		})

		desc := "desc"
		cover := "https://example.com/cover.png"

		// Act
		out, err := svc.CreateTag(context.Background(), CreateTagInput{
			UserID:        "u1",
			Title:         "title",
			Description:   &desc,
			CoverImageURL: &cover,
			IsPublic:      nil,
		})

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out == nil {
			t.Fatalf("expected tag")
		}
		if created == nil {
			t.Fatalf("expected TagRepository.Create to be called")
		}
		if created.UserID != "u1" || created.Title != "title" {
			t.Fatalf("unexpected created tag: %+v", created)
		}
		if created.Description == nil || *created.Description != desc {
			t.Fatalf("expected description to be set")
		}
		if created.CoverImageURL == nil || *created.CoverImageURL != cover {
			t.Fatalf("expected cover_image_url to be set")
		}
		if created.IsPublic != true {
			t.Fatalf("expected IsPublic=true, got %v", created.IsPublic)
		}
		if created.AddMoviePolicy != "everyone" {
			t.Fatalf("expected AddMoviePolicy=everyone, got %v", created.AddMoviePolicy)
		}
		if out.ID != "tag1" {
			t.Fatalf("expected out.ID=tag1, got %q", out.ID)
		}
	})

	t.Run("明示指定: IsPublic=false を指定すると false で作成される", func(t *testing.T) {
		t.Parallel()

		var created *model.Tag
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.CreateFn = func(ctx context.Context, tag *model.Tag) error {
				created = tag
				tag.ID = "tag2"
				return nil
			}
		})

		isPublic := false
		out, err := svc.CreateTag(context.Background(), CreateTagInput{
			UserID:   "u1",
			Title:    "title",
			IsPublic: &isPublic,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if created == nil {
			t.Fatalf("expected TagRepository.Create to be called")
		}
		if created.IsPublic != false {
			t.Fatalf("expected IsPublic=false, got %v", created.IsPublic)
		}
		if out.ID != "tag2" {
			t.Fatalf("expected out.ID=tag2, got %q", out.ID)
		}
	})

	t.Run("映画追加権限の明示指定: AddMoviePolicy=owner_only を指定すると owner_only で作成される", func(t *testing.T) {
		t.Parallel()

		var created *model.Tag

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.CreateFn = func(ctx context.Context, tag *model.Tag) error {
				created = tag
				tag.ID = "tag3"
				return nil
			}
		})

		addMoviePolicy := "owner_only"
		// Act
		_, err := svc.CreateTag(context.Background(), CreateTagInput{
			UserID:         "u1",
			Title:          "title",
			AddMoviePolicy: &addMoviePolicy,
		})

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if created == nil {
			t.Fatalf("expected TagRepository.Create to be called")
		}
		if created.AddMoviePolicy != "owner_only" {
			t.Fatalf("expected AddMoviePolicy=owner_only, got %v", created.AddMoviePolicy)
		}
	})
}

func TestTagService_ListPublicTags(t *testing.T) {
	t.Parallel()

	t.Run("ページングの正規化とクエリtrimが反映される", func(t *testing.T) {
		t.Parallel()

		var gotFilter repository.TagListFilter
		now := time.Now()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListPublicTagsFn = func(ctx context.Context, filter repository.TagListFilter) ([]repository.TagSummary, int64, error) {
				gotFilter = filter
				return []repository.TagSummary{
					{
						ID:            "t1",
						Title:         "公開A",
						Description:   nil,
						CoverImageURL: nil,
						IsPublic:      true,
						MovieCount:    1,
						FollowerCount: 2,
						CreatedAt:     now,
						Author:        "alice",
					},
				}, 1, nil
			}
		})

		items, total, err := svc.ListPublicTags(context.Background(), "  キーワード  ", "", 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 1 || len(items) != 1 {
			t.Fatalf("expected total=1 len=1, got total=%d len=%d", total, len(items))
		}
		if gotFilter.Query != "キーワード" {
			t.Fatalf("expected Query=キーワード, got %q", gotFilter.Query)
		}
		if gotFilter.Offset != 0 || gotFilter.Limit != 20 {
			t.Fatalf("expected Offset=0 Limit=20, got Offset=%d Limit=%d", gotFilter.Offset, gotFilter.Limit)
		}
		if items[0].ID != "t1" || items[0].Author != "alice" {
			t.Fatalf("unexpected item: %+v", items[0])
		}
		// movieService=nil のため images は空（nil）になる
		if len(items[0].Images) != 0 {
			t.Fatalf("expected images empty, got %v", items[0].Images)
		}
	})

	t.Run("page_size 上限: 100 を超える場合は 100 に丸められる", func(t *testing.T) {
		t.Parallel()

		var gotFilter repository.TagListFilter
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListPublicTagsFn = func(ctx context.Context, filter repository.TagListFilter) ([]repository.TagSummary, int64, error) {
				gotFilter = filter
				return []repository.TagSummary{}, 0, nil
			}
		})

		_, _, err := svc.ListPublicTags(context.Background(), "", "recent", 2, 1000)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotFilter.Limit != 100 {
			t.Fatalf("expected Limit=100, got %d", gotFilter.Limit)
		}
		if gotFilter.Offset != 100 {
			t.Fatalf("expected Offset=100, got %d", gotFilter.Offset)
		}
	})

	t.Run("total=0 の場合は空配列を返す", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListPublicTagsFn = func(ctx context.Context, filter repository.TagListFilter) ([]repository.TagSummary, int64, error) {
				return []repository.TagSummary{}, 0, nil
			}
		})

		items, total, err := svc.ListPublicTags(context.Background(), strings.Repeat(" ", 3), "", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 0 {
			t.Fatalf("expected total=0, got %d", total)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("expected empty slice, got %#v", items)
		}
	})
}

func TestTagService_GetTagDetail(t *testing.T) {
	t.Parallel()

	t.Run("can_add_movie: add_movie_policy=everyone の場合、認証済みユーザーは追加可能", func(t *testing.T) {
		t.Parallel()

		viewerID := "viewer1"
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindDetailByIDFn = func(ctx context.Context, id string) (*repository.TagDetailRow, error) {
				return &repository.TagDetailRow{
					ID:               id,
					Title:            "Test Tag",
					IsPublic:         true,
					AddMoviePolicy:   "everyone",
					OwnerID:          "owner1",
					OwnerDisplayName: "Owner",
				}, nil
			}
		})

		out, err := svc.GetTagDetail(context.Background(), "t1", &viewerID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out == nil {
			t.Fatalf("expected output")
		}
		if !out.CanAddMovie {
			t.Fatalf("expected CanAddMovie=true, got %v", out.CanAddMovie)
		}
		if out.AddMoviePolicy != "everyone" {
			t.Fatalf("expected AddMoviePolicy=everyone, got %v", out.AddMoviePolicy)
		}
	})

	t.Run("can_add_movie: add_movie_policy=owner_only の場合、作成者のみ追加可能", func(t *testing.T) {
		t.Parallel()

		ownerID := "owner1"
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindDetailByIDFn = func(ctx context.Context, id string) (*repository.TagDetailRow, error) {
				return &repository.TagDetailRow{
					ID:               id,
					Title:            "Test Tag",
					IsPublic:         true,
					AddMoviePolicy:   "owner_only",
					OwnerID:          ownerID,
					OwnerDisplayName: "Owner",
				}, nil
			}
		})

		// 作成者の場合
		out, err := svc.GetTagDetail(context.Background(), "t1", &ownerID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !out.CanAddMovie {
			t.Fatalf("expected CanAddMovie=true for owner, got %v", out.CanAddMovie)
		}

		// 作成者以外の場合
		otherID := "other1"
		out2, err := svc.GetTagDetail(context.Background(), "t1", &otherID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out2.CanAddMovie {
			t.Fatalf("expected CanAddMovie=false for non-owner, got %v", out2.CanAddMovie)
		}
	})

	t.Run("can_add_movie: 未認証ユーザーの場合、追加不可", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindDetailByIDFn = func(ctx context.Context, id string) (*repository.TagDetailRow, error) {
				return &repository.TagDetailRow{
					ID:               id,
					Title:            "Test Tag",
					IsPublic:         true,
					AddMoviePolicy:   "everyone",
					OwnerID:          "owner1",
					OwnerDisplayName: "Owner",
				}, nil
			}
		})

		out, err := svc.GetTagDetail(context.Background(), "t1", nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out.CanAddMovie {
			t.Fatalf("expected CanAddMovie=false for unauthenticated user, got %v", out.CanAddMovie)
		}
	})
}

func TestTagService_UpdateTag(t *testing.T) {
	t.Parallel()

	t.Run("add_movie_policy の更新: owner_only に変更可能", func(t *testing.T) {
		t.Parallel()

		var gotPatch repository.TagUpdatePatch
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:     id,
					UserID: "u1",
				}, nil
			}
			d.tagRepo.UpdateByIDFn = func(ctx context.Context, id string, patch repository.TagUpdatePatch) error {
				gotPatch = patch
				return nil
			}
			d.tagRepo.FindDetailByIDFn = func(ctx context.Context, id string) (*repository.TagDetailRow, error) {
				return &repository.TagDetailRow{
					ID:               id,
					Title:            "Test Tag",
					IsPublic:         true,
					AddMoviePolicy:   "owner_only",
					OwnerID:          "u1",
					OwnerDisplayName: "User1",
				}, nil
			}
		})

		newPolicy := "owner_only"
		_, err := svc.UpdateTag(context.Background(), "t1", "u1", UpdateTagPatch{
			AddMoviePolicy: &newPolicy,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPatch.AddMoviePolicy == nil || *gotPatch.AddMoviePolicy != "owner_only" {
			t.Fatalf("expected AddMoviePolicy=owner_only, got %v", gotPatch.AddMoviePolicy)
		}
	})

	t.Run("add_movie_policy のバリデーション: 不正な値はエラー", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:     id,
					UserID: "u1",
				}, nil
			}
		})

		invalidPolicy := "invalid"
		_, err := svc.UpdateTag(context.Background(), "t1", "u1", UpdateTagPatch{
			AddMoviePolicy: &invalidPolicy,
		})
		if err == nil {
			t.Fatalf("expected error for invalid add_movie_policy")
		}
		if !strings.Contains(err.Error(), "add_movie_policy must be") {
			t.Fatalf("expected error message about add_movie_policy, got: %v", err)
		}
	})
}

func TestTagService_RemoveMovieFromTag(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_movie_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		err := svc.RemoveMovieFromTag(context.Background(), "", "u1")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("タグ映画が見つからない: gorm.ErrRecordNotFound は ErrTagMovieNotFound に変換される", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "u1")
		if !errors.Is(err, ErrTagMovieNotFound) {
			t.Fatalf("expected ErrTagMovieNotFound, got: %v", err)
		}
	})

	t.Run("タグ映画検索で失敗: FindByID のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("db down")
		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return nil, expected
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "u1")
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("タグが見つからない: gorm.ErrRecordNotFound は ErrTagNotFound に変換される", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "u1"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "u1")
		if !errors.Is(err, ErrTagNotFound) {
			t.Fatalf("expected ErrTagNotFound, got: %v", err)
		}
	})

	t.Run("権限チェック: owner_only タグで作成者以外は削除不可", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "other_user"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "owner_only",
				}, nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "other_user")
		if !errors.Is(err, ErrTagPermissionDenied) {
			t.Fatalf("expected ErrTagPermissionDenied, got: %v", err)
		}
	})

	t.Run("権限チェック: owner_only タグで作成者は削除可能", func(t *testing.T) {
		t.Parallel()

		var deletedID string
		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "other_user"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "owner_only",
				}, nil
			}
			d.tagMovieRepo.DeleteFn = func(ctx context.Context, tagMovieID string) error {
				deletedID = tagMovieID
				return nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "owner1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if deletedID != "tm1" {
			t.Fatalf("expected delete to be called with tm1, got %s", deletedID)
		}
	})

	t.Run("権限チェック: タグ作成者は全ての映画を削除可能", func(t *testing.T) {
		t.Parallel()

		var deletedID string
		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "other_user"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "everyone",
				}, nil
			}
			d.tagMovieRepo.DeleteFn = func(ctx context.Context, tagMovieID string) error {
				deletedID = tagMovieID
				return nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "owner1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if deletedID != "tm1" {
			t.Fatalf("expected delete to be called with tm1, got %s", deletedID)
		}
	})

	t.Run("権限チェック: 他ユーザーは自分が追加した映画のみ削除可能", func(t *testing.T) {
		t.Parallel()

		var deletedID string
		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "user2"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "everyone",
				}, nil
			}
			d.tagMovieRepo.DeleteFn = func(ctx context.Context, tagMovieID string) error {
				deletedID = tagMovieID
				return nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "user2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if deletedID != "tm1" {
			t.Fatalf("expected delete to be called with tm1, got %s", deletedID)
		}
	})

	t.Run("権限チェック: 他ユーザーが追加した映画は削除不可", func(t *testing.T) {
		t.Parallel()

		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "user2"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "everyone",
				}, nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "user3")
		if !errors.Is(err, ErrTagPermissionDenied) {
			t.Fatalf("expected ErrTagPermissionDenied, got: %v", err)
		}
	})

	t.Run("削除に成功", func(t *testing.T) {
		t.Parallel()

		var deletedID string

		svc := newTagService(t, func(d *deps) {
			d.tagMovieRepo.FindByIDFn = func(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
				return &model.TagMovie{ID: tagMovieID, TagID: "t1", AddedByUser: "u1"}, nil
			}
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{
					ID:             id,
					UserID:         "owner1",
					AddMoviePolicy: "everyone",
				}, nil
			}
			d.tagMovieRepo.DeleteFn = func(ctx context.Context, tagMovieID string) error {
				deletedID = tagMovieID
				return nil
			}
		})

		err := svc.RemoveMovieFromTag(context.Background(), "tm1", "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if deletedID != "tm1" {
			t.Fatalf("expected delete to be called with tm1, got %s", deletedID)
		}
	})
}

func TestTagService_ListTagMovies_CanDelete(t *testing.T) {
	t.Parallel()

	makeSvc := func(tag *model.Tag, rows []repository.TagMovieWithCache) TagService {
		return newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return tag, nil
			}
			d.tagMovieRepo.ListByTagFn = func(ctx context.Context, tagID string, offset, limit int) ([]repository.TagMovieWithCache, int64, error) {
				return rows, int64(len(rows)), nil
			}
		})
	}

	rows := []repository.TagMovieWithCache{
		{ID: "tm1", TagID: "t1", TmdbMovieID: 101, AddedByUser: "userA", Position: 0, CreatedAt: time.Now()},
		{ID: "tm2", TagID: "t1", TmdbMovieID: 102, AddedByUser: "userB", Position: 1, CreatedAt: time.Now()},
	}

	t.Run("未認証(viewerUserID=nil): can_delete は全て false", func(t *testing.T) {
		t.Parallel()

		svc := makeSvc(&model.Tag{ID: "t1", UserID: "owner1", IsPublic: true, AddMoviePolicy: "everyone"}, rows)
		out, _, err := svc.ListTagMovies(context.Background(), "t1", nil, 1, 50)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("expected 2 items, got %d", len(out))
		}
		if out[0].CanDelete || out[1].CanDelete {
			t.Fatalf("expected all can_delete=false, got: %+v", out)
		}
	})

	t.Run("タグ作成者: can_delete は全て true", func(t *testing.T) {
		t.Parallel()

		viewer := "owner1"
		svc := makeSvc(&model.Tag{ID: "t1", UserID: "owner1", IsPublic: true, AddMoviePolicy: "everyone"}, rows)
		out, _, err := svc.ListTagMovies(context.Background(), "t1", &viewer, 1, 50)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("expected 2 items, got %d", len(out))
		}
		if !out[0].CanDelete || !out[1].CanDelete {
			t.Fatalf("expected all can_delete=true, got: %+v", out)
		}
	})

	t.Run("owner_only タグ: 作成者以外は can_delete=false（追加者でも不可）", func(t *testing.T) {
		t.Parallel()

		viewer := "userA"
		svc := makeSvc(&model.Tag{ID: "t1", UserID: "owner1", IsPublic: true, AddMoviePolicy: "owner_only"}, rows)
		out, _, err := svc.ListTagMovies(context.Background(), "t1", &viewer, 1, 50)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("expected 2 items, got %d", len(out))
		}
		if out[0].CanDelete || out[1].CanDelete {
			t.Fatalf("expected all can_delete=false, got: %+v", out)
		}
	})

	t.Run("everyone タグ: 追加者のみ can_delete=true", func(t *testing.T) {
		t.Parallel()

		viewer := "userA"
		svc := makeSvc(&model.Tag{ID: "t1", UserID: "owner1", IsPublic: true, AddMoviePolicy: "everyone"}, rows)
		out, _, err := svc.ListTagMovies(context.Background(), "t1", &viewer, 1, 50)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(out) != 2 {
			t.Fatalf("expected 2 items, got %d", len(out))
		}
		if out[0].ID != "tm1" || out[1].ID != "tm2" {
			t.Fatalf("unexpected item ids: %+v", out)
		}
		if out[0].CanDelete != true || out[1].CanDelete != false {
			t.Fatalf("expected [true,false], got: [%v,%v]", out[0].CanDelete, out[1].CanDelete)
		}
	})
}

func TestTagService_FollowTag(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		err := svc.FollowTag(context.Background(), "", "u1")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		err := svc.FollowTag(context.Background(), "t1", "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("タグが存在しない: ErrTagNotFound", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		err := svc.FollowTag(context.Background(), "t1", "u1")
		if !errors.Is(err, ErrTagNotFound) {
			t.Fatalf("expected ErrTagNotFound, got: %v", err)
		}
	})

	t.Run("非公開タグを作成者以外がフォロー: ErrTagPermissionDenied", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id, UserID: "owner1", IsPublic: false}, nil
			}
		})

		err := svc.FollowTag(context.Background(), "t1", "u1")
		if !errors.Is(err, ErrTagPermissionDenied) {
			t.Fatalf("expected ErrTagPermissionDenied, got: %v", err)
		}
	})

	t.Run("非公開タグを作成者がフォロー: 成功", func(t *testing.T) {
		t.Parallel()
		var created bool
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id, UserID: "owner1", IsPublic: false}, nil
			}
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, nil
			}
			d.tagFollowerRepo.CreateFn = func(ctx context.Context, tagID, userID string) error {
				created = true
				return nil
			}
		})

		err := svc.FollowTag(context.Background(), "t1", "owner1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !created {
			t.Fatalf("expected Create to be called")
		}
	})

	t.Run("既にフォロー済み: ErrAlreadyFollowingTag", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id, UserID: "owner1", IsPublic: true}, nil
			}
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return true, nil
			}
		})

		err := svc.FollowTag(context.Background(), "t1", "u1")
		if !errors.Is(err, ErrAlreadyFollowingTag) {
			t.Fatalf("expected ErrAlreadyFollowingTag, got: %v", err)
		}
	})

	t.Run("成功: フォローが作成される", func(t *testing.T) {
		t.Parallel()
		var gotTagID, gotUserID string
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id, UserID: "owner1", IsPublic: true}, nil
			}
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, nil
			}
			d.tagFollowerRepo.CreateFn = func(ctx context.Context, tagID, userID string) error {
				gotTagID = tagID
				gotUserID = userID
				return nil
			}
		})

		err := svc.FollowTag(context.Background(), "t1", "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotTagID != "t1" || gotUserID != "u1" {
			t.Fatalf("unexpected args: tagID=%s userID=%s", gotTagID, gotUserID)
		}
	})
}

func TestTagService_UnfollowTag(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		err := svc.UnfollowTag(context.Background(), "", "u1")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		err := svc.UnfollowTag(context.Background(), "t1", "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("タグが存在しない: ErrTagNotFound", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		err := svc.UnfollowTag(context.Background(), "t1", "u1")
		if !errors.Is(err, ErrTagNotFound) {
			t.Fatalf("expected ErrTagNotFound, got: %v", err)
		}
	})

	t.Run("フォローしていない: ErrNotFollowingTag", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, nil
			}
		})

		err := svc.UnfollowTag(context.Background(), "t1", "u1")
		if !errors.Is(err, ErrNotFollowingTag) {
			t.Fatalf("expected ErrNotFollowingTag, got: %v", err)
		}
	})

	t.Run("成功: フォローが削除される", func(t *testing.T) {
		t.Parallel()
		var gotTagID, gotUserID string
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return true, nil
			}
			d.tagFollowerRepo.DeleteFn = func(ctx context.Context, tagID, userID string) error {
				gotTagID = tagID
				gotUserID = userID
				return nil
			}
		})

		err := svc.UnfollowTag(context.Background(), "t1", "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotTagID != "t1" || gotUserID != "u1" {
			t.Fatalf("unexpected args: tagID=%s userID=%s", gotTagID, gotUserID)
		}
	})
}

func TestTagService_IsFollowingTag(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		_, err := svc.IsFollowingTag(context.Background(), "", "u1")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("user_id が空の場合: false を返す", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		result, err := svc.IsFollowingTag(context.Background(), "t1", "")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Fatalf("expected false")
		}
	})

	t.Run("フォローしている: true を返す", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return true, nil
			}
		})

		result, err := svc.IsFollowingTag(context.Background(), "t1", "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !result {
			t.Fatalf("expected true")
		}
	})

	t.Run("フォローしていない: false を返す", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagFollowerRepo.IsFollowingFn = func(ctx context.Context, tagID, userID string) (bool, error) {
				return false, nil
			}
		})

		result, err := svc.IsFollowingTag(context.Background(), "t1", "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Fatalf("expected false")
		}
	})
}

func TestTagService_ListTagFollowers(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: tag_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		_, _, err := svc.ListTagFollowers(context.Background(), "", 1, 20)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("タグが存在しない: ErrTagNotFound", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return nil, gorm.ErrRecordNotFound
			}
		})

		_, _, err := svc.ListTagFollowers(context.Background(), "t1", 1, 20)
		if !errors.Is(err, ErrTagNotFound) {
			t.Fatalf("expected ErrTagNotFound, got: %v", err)
		}
	})

	t.Run("ページング正規化: page < 1 はデフォルト 1", func(t *testing.T) {
		t.Parallel()
		var gotPage int
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagFollowerRepo.ListFollowersFn = func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				gotPage = page
				return []*model.User{}, 0, nil
			}
		})

		_, _, err := svc.ListTagFollowers(context.Background(), "t1", 0, 10)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPage != 1 {
			t.Fatalf("expected page=1, got %d", gotPage)
		}
	})

	t.Run("ページング正規化: pageSize > 100 は 100", func(t *testing.T) {
		t.Parallel()
		var gotPageSize int
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagFollowerRepo.ListFollowersFn = func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				gotPageSize = pageSize
				return []*model.User{}, 0, nil
			}
		})

		_, _, err := svc.ListTagFollowers(context.Background(), "t1", 2, 1000)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPageSize != 100 {
			t.Fatalf("expected pageSize=100, got %d", gotPageSize)
		}
	})

	t.Run("成功: フォロワー一覧を返す", func(t *testing.T) {
		t.Parallel()
		expected := []*model.User{
			{ID: "u1", DisplayName: "User1"},
			{ID: "u2", DisplayName: "User2"},
		}
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagFollowerRepo.ListFollowersFn = func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
				return expected, 2, nil
			}
		})

		users, total, err := svc.ListTagFollowers(context.Background(), "t1", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 2 || len(users) != 2 {
			t.Fatalf("unexpected result: total=%d len=%d", total, len(users))
		}
	})
}

func TestTagService_ListFollowingTags(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		_, _, err := svc.ListFollowingTags(context.Background(), "", 1, 20)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ページング正規化が反映される", func(t *testing.T) {
		t.Parallel()
		var gotPage, gotPageSize int
		svc := newTagService(t, func(d *deps) {
			d.tagFollowerRepo.ListFollowingTagsFn = func(ctx context.Context, userID string, page, pageSize int) ([]repository.TagSummary, int64, error) {
				gotPage = page
				gotPageSize = pageSize
				return []repository.TagSummary{}, 0, nil
			}
		})

		_, _, err := svc.ListFollowingTags(context.Background(), "u1", 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPage != 1 || gotPageSize != 20 {
			t.Fatalf("expected (1,20), got (%d,%d)", gotPage, gotPageSize)
		}
	})

	t.Run("total=0 の場合は空配列を返す", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagFollowerRepo.ListFollowingTagsFn = func(ctx context.Context, userID string, page, pageSize int) ([]repository.TagSummary, int64, error) {
				return []repository.TagSummary{}, 0, nil
			}
		})

		items, total, err := svc.ListFollowingTags(context.Background(), "u1", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 0 {
			t.Fatalf("expected total=0, got %d", total)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("expected empty slice, got %#v", items)
		}
	})

	t.Run("成功: フォロー中のタグ一覧を返す", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		svc := newTagService(t, func(d *deps) {
			d.tagFollowerRepo.ListFollowingTagsFn = func(ctx context.Context, userID string, page, pageSize int) ([]repository.TagSummary, int64, error) {
				return []repository.TagSummary{
					{
						ID:              "t1",
						Title:           "Tag1",
						IsPublic:        true,
						MovieCount:      5,
						FollowerCount:   10,
						CreatedAt:       now,
						Author:          "owner1",
						AuthorDisplayID: "owner1_id",
					},
				}, 1, nil
			}
		})

		items, total, err := svc.ListFollowingTags(context.Background(), "u1", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 1 || len(items) != 1 {
			t.Fatalf("unexpected result: total=%d len=%d", total, len(items))
		}
		if items[0].ID != "t1" || items[0].Author != "owner1" {
			t.Fatalf("unexpected item: %+v", items[0])
		}
	})
}

func TestTagService_ListTagsByUserID(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, nil)
		_, _, err := svc.ListTagsByUserID(context.Background(), "", true, 1, 20)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ページング正規化が反映される", func(t *testing.T) {
		t.Parallel()
		var gotFilter repository.UserTagListFilter
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListTagsByUserIDFn = func(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error) {
				gotFilter = filter
				return []repository.TagSummary{}, 0, nil
			}
		})

		_, _, err := svc.ListTagsByUserID(context.Background(), "u1", true, 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotFilter.Offset != 0 || gotFilter.Limit != 20 {
			t.Fatalf("expected Offset=0 Limit=20, got Offset=%d Limit=%d", gotFilter.Offset, gotFilter.Limit)
		}
	})

	t.Run("publicOnly フラグが正しく渡される", func(t *testing.T) {
		t.Parallel()
		var gotFilter repository.UserTagListFilter
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListTagsByUserIDFn = func(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error) {
				gotFilter = filter
				return []repository.TagSummary{}, 0, nil
			}
		})

		_, _, err := svc.ListTagsByUserID(context.Background(), "u1", true, 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !gotFilter.IncludePublic {
			t.Fatalf("expected IncludePublic=true, got %v", gotFilter.IncludePublic)
		}
	})

	t.Run("total=0 の場合は空配列を返す", func(t *testing.T) {
		t.Parallel()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListTagsByUserIDFn = func(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error) {
				return []repository.TagSummary{}, 0, nil
			}
		})

		items, total, err := svc.ListTagsByUserID(context.Background(), "u1", false, 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 0 {
			t.Fatalf("expected total=0, got %d", total)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("expected empty slice, got %#v", items)
		}
	})

	t.Run("成功: ユーザーのタグ一覧を返す", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.ListTagsByUserIDFn = func(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error) {
				return []repository.TagSummary{
					{
						ID:            "t1",
						Title:         "MyTag",
						IsPublic:      true,
						MovieCount:    3,
						FollowerCount: 7,
						CreatedAt:     now,
						Author:        "user1",
					},
				}, 1, nil
			}
		})

		items, total, err := svc.ListTagsByUserID(context.Background(), "u1", false, 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 1 || len(items) != 1 {
			t.Fatalf("unexpected result: total=%d len=%d", total, len(items))
		}
		if items[0].ID != "t1" || items[0].Author != "user1" {
			t.Fatalf("unexpected item: %+v", items[0])
		}
	})
}
