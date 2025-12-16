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

type deps struct {
	tagRepo      *testutil.FakeTagRepository
	tagMovieRepo *testutil.FakeTagMovieRepository
	movieService MovieService
	imageBaseURL string
}

func newTagService(t *testing.T, opt func(*deps)) TagService {
	t.Helper()

	d := &deps{
		tagRepo:      &testutil.FakeTagRepository{},
		tagMovieRepo: &testutil.FakeTagMovieRepository{},
		movieService: nil,
		imageBaseURL: "",
	}
	if opt != nil {
		opt(d)
	}
	return NewTagService(d.tagRepo, d.tagMovieRepo, d.movieService, d.imageBaseURL)
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

	t.Run("タグの movie_count 更新で失敗: IncrementMovieCount のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("increment failed")
		svc := newTagService(t, func(d *deps) {
			d.tagRepo.FindByIDFn = func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			}
			d.tagRepo.IncrementMovieCountFn = func(ctx context.Context, id string, delta int) error {
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

	t.Run("成功: tag_movie を作成し movie_count を +1 する", func(t *testing.T) {
		t.Parallel()

		var gotTagID string
		var gotDelta int
		var created *model.TagMovie

		tagRepo := &testutil.FakeTagRepository{
			FindByIDFn: func(ctx context.Context, id string) (*model.Tag, error) {
				return &model.Tag{ID: id}, nil
			},
			IncrementMovieCountFn: func(ctx context.Context, id string, delta int) error {
				gotTagID = id
				gotDelta = delta
				return nil
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
		svc := NewTagService(tagRepo, tagMovieRepo, nil, "")

		note := "hello"
		out, err := svc.AddMovieToTag(context.Background(), AddMovieToTagInput{
			TagID:       "t1",
			UserID:      "u1",
			TmdbMovieID: 123,
			Note:        &note,
			Position:    2,
		})
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

		if gotTagID != "t1" || gotDelta != 1 {
			t.Fatalf("expected IncrementMovieCount(t1, 1), got (%s, %d)", gotTagID, gotDelta)
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
		out, err := svc.CreateTag(context.Background(), CreateTagInput{
			UserID:        "u1",
			Title:         "title",
			Description:   &desc,
			CoverImageURL: &cover,
			IsPublic:      nil,
		})
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
