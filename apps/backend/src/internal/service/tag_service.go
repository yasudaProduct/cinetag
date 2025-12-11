package service

import (
	"context"
	"strings"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
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
	db           *gorm.DB
	movieService MovieService
	imageBaseURL string
}

// NewTagService は TagService の実装を生成します。
func NewTagService(db *gorm.DB, movieService MovieService, imageBaseURL string) TagService {
	return &tagService{
		db:           db,
		movieService: movieService,
		imageBaseURL: strings.TrimRight(imageBaseURL, "/"),
	}
}

// internal struct for DB scanning
type tagRow struct {
	ID            string    `gorm:"column:id"`
	Title         string    `gorm:"column:title"`
	Description   *string   `gorm:"column:description"`
	CoverImageURL *string   `gorm:"column:cover_image_url"`
	IsPublic      bool      `gorm:"column:is_public"`
	MovieCount    int       `gorm:"column:movie_count"`
	FollowerCount int       `gorm:"column:follower_count"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	Author        string    `gorm:"column:author"`
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

	if err := s.db.WithContext(ctx).Create(&tag).Error; err != nil {
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

	// ベースクエリ: 公開タグ + 作成者情報
	qb := s.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				t.movie_count, t.follower_count, t.created_at,
				u.username AS author`).
		Joins(`JOIN `+(model.User{}).TableName()+` AS u ON u.id = t.user_id`).
		Where("t.is_public = ?", true)

	// キーワード検索（タイトルのみ簡易実装）
	if strings.TrimSpace(q) != "" {
		keyword := "%" + strings.TrimSpace(q) + "%"
		qb = qb.Where("t.title ILIKE ?", keyword)
	}

	// 件数取得
	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagListItem{}, 0, nil
	}

	// ソート
	switch sort {
	case "recent":
		qb = qb.Order("t.created_at DESC")
	case "movie_count":
		qb = qb.Order("t.movie_count DESC")
	default:
		// デフォルトはフォロワー数順（人気）
		qb = qb.Order("t.follower_count DESC")
	}

	// ページング
	offset := (page - 1) * pageSize
	var rows []tagRow
	if err := qb.Limit(pageSize).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	// 映画ポスター画像の取得
	imagesByTag := make(map[string][]string, len(rows))
	if s.movieService != nil {
		for _, r := range rows {
			// 各タグごとに、最新の追加順で最大 4 件の映画を取得
			var tagMovies []model.TagMovie
			if err := s.db.WithContext(ctx).
				Where("tag_id = ?", r.ID).
				Order("created_at DESC").
				Limit(4).
				Find(&tagMovies).Error; err != nil {
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
