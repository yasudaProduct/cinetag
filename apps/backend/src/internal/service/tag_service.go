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
	db *gorm.DB
}

// NewTagService は TagService の実装を生成します。
func NewTagService(db *gorm.DB) TagService {
	return &tagService{db: db}
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

type tagMoviePosterRow struct {
	TagID      string  `gorm:"column:tag_id"`
	PosterPath *string `gorm:"column:poster_path"`
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

	// 画像取得用に tag_id の一覧を集める
	tagIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		tagIDs = append(tagIDs, r.ID)
	}

	imagesByTag := make(map[string][]string, len(tagIDs))
	if len(tagIDs) > 0 {
		var posterRows []tagMoviePosterRow
		// 各タグごとに最大4件のポスターを取得するため、まずは全体を created_at DESC で取得し、
		// 後続のループでタグごとに先頭4件に絞り込む。
		if err := s.db.WithContext(ctx).
			Table((model.TagMovie{}).TableName()+" AS tm").
			Select("tm.tag_id, mc.poster_path").
			Joins("JOIN "+(model.MovieCache{}).TableName()+" AS mc ON mc.tmdb_movie_id = tm.tmdb_movie_id").
			Where("tm.tag_id IN ?", tagIDs).
			Order("tm.created_at DESC").
			Scan(&posterRows).Error; err != nil {
			return nil, 0, err
		}

		for _, pr := range posterRows {
			if pr.PosterPath == nil || *pr.PosterPath == "" {
				continue
			}
			list := imagesByTag[pr.TagID]
			if len(list) >= 4 {
				continue
			}
			// ここでは TMDb のベースURLまでは付けず、poster_path をそのまま返す。
			// 必要に応じて環境変数からベースURLを読み取り、ここで連結することもできる。
			list = append(list, *pr.PosterPath)
			imagesByTag[pr.TagID] = list
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
