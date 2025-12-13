package service

import (
	"context"
	"strings"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"
)

// TagListItem は公開タグ一覧で返す1件分の情報です。
type TagListItem struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Author        string    `json:"author"`
	CoverImageURL *string   `json:"cover_image_url,omitempty"`
	IsPublic      bool      `json:"is_public"`
	MovieCount    int       `json:"movie_count"`
	FollowerCount int       `json:"follower_count"`
	Images        []string  `json:"images"`
	CreatedAt     time.Time `json:"created_at"`
}

// TagService はタグに関するユースケースを表します。
type TagService interface {
	// ListPublicTags は公開タグを検索・ソート・ページングして返します。
	ListPublicTags(ctx context.Context, q, sort string, page, pageSize int) ([]TagListItem, int64, error)

	// CreateTag は新しいタグを作成して返します。
	CreateTag(ctx context.Context, in CreateTagInput) (*model.Tag, error)
}

type tagService struct {
	tagRepo      repository.TagRepository
	tagMovieRepo repository.TagMovieRepository
	movieService MovieService
	imageBaseURL string
}

// NewTagService は TagService の実装を生成します。
func NewTagService(
	tagRepo repository.TagRepository,
	tagMovieRepo repository.TagMovieRepository,
	movieService MovieService,
	imageBaseURL string,
) TagService {
	return &tagService{
		tagRepo:      tagRepo,
		tagMovieRepo: tagMovieRepo,
		movieService: movieService,
		imageBaseURL: strings.TrimRight(imageBaseURL, "/"),
	}
}

// CreateTagInput はタグ作成時の入力値を表します。
type CreateTagInput struct {
	UserID        string
	Title         string
	Description   *string
	CoverImageURL *string
	IsPublic      *bool
}

// CreateTag は新しいタグを作成します。
func (s *tagService) CreateTag(ctx context.Context, in CreateTagInput) (*model.Tag, error) {
	isPublic := true
	if in.IsPublic != nil {
		isPublic = *in.IsPublic
	}

	tag := model.Tag{
		UserID:        in.UserID,
		Title:         in.Title,
		Description:   in.Description,
		CoverImageURL: in.CoverImageURL,
		IsPublic:      isPublic,
	}

	if err := s.tagRepo.Create(ctx, &tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

func (s *tagService) ListPublicTags(ctx context.Context, q, sort string, page, pageSize int) ([]TagListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 検索条件
	query := strings.TrimSpace(q)
	offset := (page - 1) * pageSize

	rows, total, err := s.tagRepo.ListPublicTags(ctx, repository.TagListFilter{
		Query:  query,
		Sort:   sort,
		Offset: offset,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagListItem{}, 0, nil
	}

	// 映画ポスター画像の取得
	imagesByTag := make(map[string][]string, len(rows))
	if s.movieService != nil {
		for _, r := range rows {
			// 各タグごとに、最新の追加順で最大 4 件の映画を取得
			tagMovies, err := s.tagMovieRepo.ListRecentByTag(ctx, r.ID, 4)
			if err != nil {
				return nil, 0, err
			}

			for _, tm := range tagMovies {
				if len(imagesByTag[r.ID]) >= 4 {
					break
				}
				cache, err := s.movieService.EnsureMovieCache(ctx, tm.TmdbMovieID)
				if err != nil {
					// 画像の取得失敗はタグ一覧全体のエラーにはせずスキップする。
					continue
				}
				if cache.PosterPath == nil || *cache.PosterPath == "" {
					continue
				}
				poster := *cache.PosterPath
				if s.imageBaseURL != "" {
					poster = s.imageBaseURL + poster
				}
				imagesByTag[r.ID] = append(imagesByTag[r.ID], poster)
			}
		}
	}

	// 最終的なレスポンス構築
	items := make([]TagListItem, 0, len(rows))
	for _, r := range rows {
		item := TagListItem{
			ID:            r.ID,
			Title:         r.Title,
			Description:   r.Description,
			Author:        r.Author,
			CoverImageURL: r.CoverImageURL,
			IsPublic:      r.IsPublic,
			MovieCount:    r.MovieCount,
			FollowerCount: r.FollowerCount,
			Images:        imagesByTag[r.ID],
			CreatedAt:     r.CreatedAt,
		}
		items = append(items, item)
	}

	return items, total, nil
}
