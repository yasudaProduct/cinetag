package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"

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

	// GetTagDetail は指定タグの詳細を返します。
	// viewerUserID は任意で、非公開タグの参照権限判定に利用します。
	GetTagDetail(ctx context.Context, tagID string, viewerUserID *string) (*TagDetail, error)

	// ListTagMovies は指定タグに含まれる映画一覧を返します。
	// viewerUserID は任意で、非公開タグの参照権限判定に利用します。
	ListTagMovies(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]TagMovieItem, int64, error)

	// CreateTag は新しいタグを作成して返します。
	CreateTag(ctx context.Context, in CreateTagInput) (*model.Tag, error)

	// AddMovieToTag はタグに映画を追加して返します（作成者のみ）。
	AddMovieToTag(ctx context.Context, in AddMovieToTagInput) (*model.TagMovie, error)

	// UpdateTag はタグのメタ情報を更新して返します（作成者のみ）。
	UpdateTag(ctx context.Context, tagID string, userID string, patch UpdateTagPatch) (*TagDetail, error)
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

// AddMovieToTagInput はタグへ映画を追加する際の入力値です。
type AddMovieToTagInput struct {
	TagID       string
	UserID      string
	TmdbMovieID int
	Note        *string
	Position    int
}

// TagDetail はタグ詳細API向けのレスポンスモデルです。
// フロントの zod schema が owner を許容しているため、owner を返す形に寄せます。
type TagDetail struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	CoverImageURL *string   `json:"cover_image_url,omitempty"`
	IsPublic      bool      `json:"is_public"`
	MovieCount    int       `json:"movie_count"`
	FollowerCount int       `json:"follower_count"`
	Owner         TagOwner  `json:"owner"`
	CanEdit       bool      `json:"can_edit"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// UI都合（既存画面の参加者表示）: 現時点では follower_count を参加者数として扱い、
	// 一覧は未実装のため空配列を返す。
	ParticipantCount int              `json:"participant_count"`
	Participants     []TagParticipant `json:"participants"`
}

type TagOwner struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type TagParticipant struct {
	Name string `json:"name"`
}

// TagMovieItem はタグ内映画一覧API向けのレスポンスモデルです。
type TagMovieItem struct {
	ID            string    `json:"id"`
	TagID         string    `json:"tag_id"`
	TmdbMovieID   int       `json:"tmdb_movie_id"`
	Note          *string   `json:"note,omitempty"`
	Position      int       `json:"position"`
	AddedByUserID string    `json:"added_by_user_id"`
	CreatedAt     time.Time `json:"created_at"`
	Movie         *MovieRef `json:"movie,omitempty"`
}

type MovieRef struct {
	Title         string   `json:"title"`
	OriginalTitle *string  `json:"original_title,omitempty"`
	PosterPath    *string  `json:"poster_path,omitempty"`
	ReleaseDate   *string  `json:"release_date,omitempty"`
	VoteAverage   *float64 `json:"vote_average,omitempty"`
}

// UpdateTagPatch はタグ更新の入力値です（部分更新）。
// nil のフィールドは更新しません。
type UpdateTagPatch struct {
	Title         *string
	Description   **string
	CoverImageURL **string
	IsPublic      *bool
}

var (
	ErrTagNotFound           = errors.New("tag not found")            // タグが存在しない
	ErrTagPermissionDenied   = errors.New("tag permission denied")    // タグの編集権限がない
	ErrTagMovieAlreadyExists = errors.New("tag movie already exists") // タグに既に映画が存在する
)

func (s *tagService) UpdateTag(ctx context.Context, tagID string, userID string, patch UpdateTagPatch) (*TagDetail, error) {

	// 必須バリデーション（tagID/userID）
	if strings.TrimSpace(tagID) == "" {
		return nil, fmt.Errorf("tag_id is required")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// タグの存在確認
	tag, err := s.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		// タグの取得に失敗した場合、エラーを返す
		return nil, err
	}
	// タグの作成者がユーザーIDと一致しない場合、エラーを返す
	if tag.UserID != userID {
		return nil, ErrTagPermissionDenied
	}

	// バリデーション（title/description）
	if patch.Title != nil {
		if l := len([]rune(strings.TrimSpace(*patch.Title))); l == 0 || l > 100 {
			return nil, fmt.Errorf("title must be between 1 and 100 characters")
		}
	}
	if patch.Description != nil && *patch.Description != nil {
		if l := len([]rune(**patch.Description)); l > 500 {
			return nil, fmt.Errorf("description must be 500 characters or less")
		}
	}

	// タグを更新する。
	err = s.tagRepo.UpdateByID(ctx, tagID, repository.TagUpdatePatch{
		Title:         patch.Title,
		Description:   patch.Description,
		CoverImageURL: patch.CoverImageURL,
		IsPublic:      patch.IsPublic,
	})
	if err != nil {
		return nil, err
	}

	// 更新後の詳細を返す（owner のため viewerUserID を userID にする）
	viewer := userID
	return s.GetTagDetail(ctx, tagID, &viewer)
}

func (s *tagService) GetTagDetail(ctx context.Context, tagID string, viewerUserID *string) (*TagDetail, error) {
	// 必須バリデーション（tagID）
	if strings.TrimSpace(tagID) == "" {
		return nil, fmt.Errorf("tag_id is required")
	}

	// タグの詳細を取得する。
	row, err := s.tagRepo.FindDetailByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	// 非公開タグの場合、ビューアーの権限をチェックする。
	if !row.IsPublic {
		if viewerUserID == nil || strings.TrimSpace(*viewerUserID) == "" || *viewerUserID != row.OwnerID {
			// ビューアーの権限がない場合、エラーを返す
			return nil, ErrTagPermissionDenied
		}
	}

	// ビューアーがタグの作成者の場合、編集可能なフラグを立てる。
	canEdit := viewerUserID != nil && strings.TrimSpace(*viewerUserID) != "" && *viewerUserID == row.OwnerID

	// タグの詳細を返す。
	return &TagDetail{
		ID:            row.ID,
		Title:         row.Title,
		Description:   row.Description,
		CoverImageURL: row.CoverImageURL,
		IsPublic:      row.IsPublic,
		MovieCount:    row.MovieCount,
		FollowerCount: row.FollowerCount,
		Owner: TagOwner{
			ID:          row.OwnerID,
			Username:    row.OwnerUsername,
			DisplayName: row.OwnerDisplayName,
			AvatarURL:   row.OwnerAvatarURL,
		},
		CanEdit:   canEdit,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		// UI互換
		ParticipantCount: row.FollowerCount,
		Participants:     []TagParticipant{},
	}, nil
}

func (s *tagService) ListTagMovies(ctx context.Context, tagID string, viewerUserID *string, page, pageSize int) ([]TagMovieItem, int64, error) {
	if strings.TrimSpace(tagID) == "" {
		return nil, 0, fmt.Errorf("tag_id is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tag, err := s.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrTagNotFound
		}
		return nil, 0, err
	}

	if !tag.IsPublic {
		if viewerUserID == nil || strings.TrimSpace(*viewerUserID) == "" || *viewerUserID != tag.UserID {
			return nil, 0, ErrTagPermissionDenied
		}
	}

	offset := (page - 1) * pageSize
	rows, total, err := s.tagMovieRepo.ListByTag(ctx, tagID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagMovieItem{}, 0, nil
	}

	items := make([]TagMovieItem, 0, len(rows))
	for _, r := range rows {
		var movie *MovieRef

		// movie_cache が join できている場合はそれを使う
		if r.MovieTitle != nil && strings.TrimSpace(*r.MovieTitle) != "" {
			title := strings.TrimSpace(*r.MovieTitle)
			var release *string
			if r.MovieReleaseDate != nil {
				s := r.MovieReleaseDate.Format("2006-01-02")
				release = &s
			}
			movie = &MovieRef{
				Title:         title,
				OriginalTitle: r.MovieOriginalTitle,
				PosterPath:    r.MoviePosterPath,
				ReleaseDate:   release,
				VoteAverage:   r.MovieVoteAverage,
			}
		} else if s.movieService != nil {
			// ベストエフォートでキャッシュを取得して埋める
			cache, err := s.movieService.EnsureMovieCache(ctx, r.TmdbMovieID)
			if err == nil && cache != nil {
				var release *string
				if cache.ReleaseDate != nil {
					s := cache.ReleaseDate.Format("2006-01-02")
					release = &s
				}
				movie = &MovieRef{
					Title:         cache.Title,
					OriginalTitle: cache.OriginalTitle,
					PosterPath:    cache.PosterPath,
					ReleaseDate:   release,
					VoteAverage:   cache.VoteAverage,
				}
			}
		}

		items = append(items, TagMovieItem{
			ID:            r.ID,
			TagID:         r.TagID,
			TmdbMovieID:   r.TmdbMovieID,
			Note:          r.Note,
			Position:      r.Position,
			AddedByUserID: r.AddedByUser,
			CreatedAt:     r.CreatedAt,
			Movie:         movie,
		})
	}

	return items, total, nil
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

// AddMovieToTag はタグに映画を追加します。
func (s *tagService) AddMovieToTag(ctx context.Context, in AddMovieToTagInput) (*model.TagMovie, error) {

	// バリデーション
	if in.TagID == "" {
		return nil, fmt.Errorf("tag_id is required")
	}
	if in.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if in.TmdbMovieID <= 0 {
		return nil, fmt.Errorf("invalid tmdb_movie_id: %d", in.TmdbMovieID)
	}
	if in.Position < 0 {
		return nil, fmt.Errorf("position must be 0 or greater")
	}

	// タグの存在確認
	_, err := s.tagRepo.FindByID(ctx, in.TagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	// TODO: タグに編集可能範囲のオプションを追加する。今は全ユーザー追加可能。
	// if tag.UserID != in.UserID {
	// 	return nil, ErrTagPermissionDenied
	// }

	// タグ映画を作成する。
	tm := model.TagMovie{
		TagID:       in.TagID,
		TmdbMovieID: in.TmdbMovieID,
		AddedByUser: in.UserID,
		Note:        in.Note,
		Position:    in.Position,
	}

	// タグ映画を作成する。
	if err := s.tagMovieRepo.Create(ctx, &tm); err != nil {
		if errors.Is(err, repository.ErrTagMovieAlreadyExists) {
			return nil, ErrTagMovieAlreadyExists
		}
		return nil, err
	}

	// タグの movie_count を加算（一覧の表示用）
	if err := s.tagRepo.IncrementMovieCount(ctx, in.TagID, 1); err != nil {
		return nil, err
	}

	// 可能であれば、作成時にベストエフォートでキャッシュを温める（失敗してもAPIは成功扱い）
	if s.movieService != nil {
		go func(movieID int) {
			ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := s.movieService.EnsureMovieCache(ctx2, movieID); err != nil {
				log.Printf("failed to warm movie cache: tmdb_movie_id=%d err=%v", movieID, err)
			}
		}(in.TmdbMovieID)
	}

	return &tm, nil
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
			fmt.Println("tagMovies", tagMovies)

			for _, tm := range tagMovies {
				if len(imagesByTag[r.ID]) >= 4 {
					break
				}
				fmt.Println("tm", tm)
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
