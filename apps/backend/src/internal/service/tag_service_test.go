package service

import (
	"context"
	"errors"
	"testing"

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
