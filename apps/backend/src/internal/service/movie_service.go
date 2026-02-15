package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MovieService interface {
	// 指定した TMDB 映画 ID に対応する movie_cache レコードの存在と有効期限を保証する。
	// - 有効なキャッシュがあればそれを返す
	// - キャッシュが無い、または期限切れの場合は TMDB から取得してキャッシュを更新する
	EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error)

	// TMDB の検索APIで映画を検索し、候補一覧を返す。
	SearchMovies(ctx context.Context, query string, page int) ([]TMDBSearchResult, int, error)

	// 指定した TMDB 映画 ID の詳細情報を取得する。
	GetMovieDetail(ctx context.Context, tmdbMovieID int) (*MovieDetailResponse, error)

	// 指定した TMDB 映画 ID が含まれるタグの一覧を取得する。
	GetMovieRelatedTags(ctx context.Context, tmdbMovieID int, limit int) ([]MovieRelatedTagItem, error)
}

// TMDB 連携に必要な設定値。
type TMDBConfig struct {
	APIKey          string
	BaseURL         string
	DefaultLanguage string
}

// MovieService の実装です。
type movieService struct {
	logger     *slog.Logger
	db         *gorm.DB
	httpClient *http.Client
	cfg        TMDBConfig
}

// MovieService を生成する。
func NewMovieService(logger *slog.Logger, db *gorm.DB) MovieService {
	cfg := TMDBConfig{
		APIKey:          os.Getenv("TMDB_API_KEY"),
		BaseURL:         os.Getenv("TMDB_BASE_URL"),
		DefaultLanguage: os.Getenv("TMDB_DEFAULT_LANGUAGE"),
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.themoviedb.org/3"
	}
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = "ja-JP"
	}

	return &movieService{
		logger: logger,
		db:     db,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cfg: cfg,
	}
}

// テストや将来の拡張用に、設定と HTTP クライアントを外部から注入するためのコンストラクタ。
func NewMovieServiceWithConfig(db *gorm.DB, cfg TMDBConfig, client *http.Client) MovieService {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.themoviedb.org/3"
	}
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = "ja-JP"
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &movieService{
		db:         db,
		httpClient: client,
		cfg:        cfg,
	}
}

// TMDB の /movie/{movie_id} レスポンスのうち、必要なフィールドのみを表す構造体。
type tmdbMovieResponse struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title"`
	PosterPath    *string  `json:"poster_path"`
	BackdropPath  *string  `json:"backdrop_path"`
	ReleaseDate   string   `json:"release_date"`
	VoteAverage   *float64 `json:"vote_average"`
	Overview      *string  `json:"overview"`
	Genres        []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Runtime             *int `json:"runtime"`
	ProductionCountries []struct {
		ISO31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	Credits *tmdbCreditsResponse `json:"credits,omitempty"`
}

// TMDB の credits レスポンスを表す構造体。
type tmdbCreditsResponse struct {
	Cast []struct {
		Name      string `json:"name"`
		Character string `json:"character"`
		Order     int    `json:"order"`
	} `json:"cast"`
	Crew []struct {
		Name string `json:"name"`
		Job  string `json:"job"`
	} `json:"crew"`
}

// TMDB の /search/movie の必要最小限のレスポンスを表す構造体。
type tmdbSearchResponse struct {
	Page         int `json:"page"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
	Results      []struct {
		ID            int      `json:"id"`
		Title         string   `json:"title"`
		OriginalTitle string   `json:"original_title"`
		PosterPath    *string  `json:"poster_path"`
		ReleaseDate   string   `json:"release_date"`
		VoteAverage   *float64 `json:"vote_average"`
	} `json:"results"`
}

// フロントに返す検索候補を表す構造体。
type TMDBSearchResult struct {
	TmdbMovieID   int      `json:"tmdb_movie_id"`
	Title         string   `json:"title"`
	OriginalTitle *string  `json:"original_title,omitempty"`
	PosterPath    *string  `json:"poster_path,omitempty"`
	ReleaseDate   *string  `json:"release_date,omitempty"`
	VoteAverage   *float64 `json:"vote_average,omitempty"`
}

// TMDB の検索APIで映画を検索し、候補一覧を返す。
func (s *movieService) SearchMovies(ctx context.Context, query string, page int) ([]TMDBSearchResult, int, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []TMDBSearchResult{}, 0, nil
	}
	if page <= 0 {
		page = 1
	}
	if s.cfg.APIKey == "" {
		return nil, 0, errors.New("TMDB_API_KEY is not set")
	}

	base, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid TMDB_BASE_URL: %w", err)
	}
	base.Path = path.Join(base.Path, "search", "movie")

	params := base.Query()
	params.Set("query", q)
	params.Set("page", strconv.Itoa(page))
	if s.cfg.DefaultLanguage != "" {
		params.Set("language", s.cfg.DefaultLanguage)
	}
	base.RawQuery = params.Encode()

	// リクエストを作成する。
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create TMDB request: %w", err)
	}

	token := strings.TrimSpace(s.cfg.APIKey)
	if token != "" {
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			req.Header.Set("Authorization", token)
		} else {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	req.Header.Set("Accept", "application/json")

	// TMDB にリクエストを送信。
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to call TMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("tmdb request failed: status=%d", resp.StatusCode)
	}

	var body tmdbSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, 0, fmt.Errorf("failed to decode TMDB response: %w", err)
	}

	// TMDBSearchResult に変換。
	out := make([]TMDBSearchResult, 0, len(body.Results))
	for _, r := range body.Results {
		var release *string
		if strings.TrimSpace(r.ReleaseDate) != "" {
			s := strings.TrimSpace(r.ReleaseDate)
			release = &s
		}
		var original *string
		if strings.TrimSpace(r.OriginalTitle) != "" {
			s := strings.TrimSpace(r.OriginalTitle)
			original = &s
		}
		out = append(out, TMDBSearchResult{
			TmdbMovieID:   r.ID,
			Title:         r.Title,
			OriginalTitle: original,
			PosterPath:    r.PosterPath,
			ReleaseDate:   release,
			VoteAverage:   r.VoteAverage,
		})
	}

	return out, body.TotalResults, nil
}

// 指定した TMDB 映画 ID に対応する movie_cache レコードの存在と有効期限を保証する。
// - 有効なキャッシュがあればそれを返す
// - キャッシュが無い、または期限切れの場合は TMDB から取得してキャッシュを更新する
func (s *movieService) EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error) {
	if tmdbMovieID <= 0 {
		return nil, fmt.Errorf("invalid tmdb movie id: %d", tmdbMovieID)
	}
	if s.cfg.APIKey == "" {
		return nil, errors.New("TMDB_API_KEY is not set")
	}

	now := time.Now()

	var cache model.MovieCache
	err := s.db.WithContext(ctx).
		Where("tmdb_movie_id = ?", tmdbMovieID).
		First(&cache).
		Error

	switch {
	case err == nil && cache.ExpiresAt.After(now):
		// 有効なキャッシュがある場合はそのまま返す
		// デバッグログ（DEBUG）
		s.logger.Debug("service.EnsureMovieCache cache hit",
			slog.Int("tmdb_movie_id", tmdbMovieID),
		)
		return &cache, nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		// それ以外の DB エラーはそのまま返す
		// エラーログ（ERROR）
		s.logger.Error("service.EnsureMovieCache failed",
			slog.Int("tmdb_movie_id", tmdbMovieID),
			slog.Any("error", err),
		)
		return nil, err
	}

	// キャッシュが存在しない、または期限切れの場合は TMDB から取得する
	tmdbMovie, err := s.fetchMovieFromTMDB(ctx, tmdbMovieID)
	if err != nil {
		return nil, err
	}

	cache, err = s.buildMovieCacheFromTMDB(tmdbMovie, now)
	if err != nil {
		return nil, err
	}

	if err := s.upsertMovieCache(ctx, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// TMDB の /movie/{movie_id} エンドポイントから映画情報を取得する。
func (s *movieService) fetchMovieFromTMDB(ctx context.Context, tmdbMovieID int) (*tmdbMovieResponse, error) {
	base, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB_BASE_URL: %w", err)
	}

	base.Path = path.Join(base.Path, "movie", strconv.Itoa(tmdbMovieID))

	q := base.Query()
	if s.cfg.DefaultLanguage != "" {
		q.Set("language", s.cfg.DefaultLanguage)
	}
	q.Set("append_to_response", "credits")
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create TMDB request: %w", err)
	}
	// TMDB は v4 認証として Authorization: Bearer をサポートする。
	// このリポジトリでは TMDB_API_KEY を Bearer トークンとして送る前提にする（クエリには付けない）。
	token := strings.TrimSpace(s.cfg.APIKey)
	if token != "" {
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			req.Header.Set("Authorization", token)
		} else {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	req.Header.Set("Accept", "application/json")

	// デバッグログ（DEBUG）
	s.logger.Debug("service.fetchMovieFromTMDB request",
		slog.Int("tmdb_movie_id", tmdbMovieID),
		slog.String("url", base.String()),
	)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// エラーログ（ERROR）
		s.logger.Error("service.fetchMovieFromTMDB request failed",
			slog.Int("tmdb_movie_id", tmdbMovieID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("failed to call TMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// デバッグログ（DEBUG）
		s.logger.Debug("service.fetchMovieFromTMDB not found",
			slog.Int("tmdb_movie_id", tmdbMovieID),
			slog.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("tmdb movie not found: %d", tmdbMovieID)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// エラーログ（ERROR）
		s.logger.Error("service.fetchMovieFromTMDB request failed",
			slog.Int("tmdb_movie_id", tmdbMovieID),
			slog.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("tmdb request failed: status=%d", resp.StatusCode)
	}

	var body tmdbMovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		// エラーログ（ERROR）
		s.logger.Error("service.fetchMovieFromTMDB decode failed",
			slog.Int("tmdb_movie_id", tmdbMovieID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("failed to decode TMDB response: %w", err)
	}

	// デバッグログ（DEBUG）
	s.logger.Debug("service.fetchMovieFromTMDB success",
		slog.Int("tmdb_movie_id", tmdbMovieID),
		slog.String("title", body.Title),
	)
	return &body, nil
}

// TMDB レスポンスから movie_cache レコードを構築する。
func (s *movieService) buildMovieCacheFromTMDB(movie *tmdbMovieResponse, now time.Time) (model.MovieCache, error) {
	cache := model.MovieCache{
		TmdbMovieID: movie.ID,
		Title:       movie.Title,
		CachedAt:    now,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
	}

	if movie.OriginalTitle != "" {
		cache.OriginalTitle = &movie.OriginalTitle
	}
	cache.PosterPath = movie.PosterPath
	cache.BackdropPath = movie.BackdropPath

	if movie.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", movie.ReleaseDate); err == nil {
			cache.ReleaseDate = &t
		}
	}

	cache.VoteAverage = movie.VoteAverage
	cache.Overview = movie.Overview

	if len(movie.Genres) > 0 {
		b, err := json.Marshal(movie.Genres)
		if err != nil {
			return model.MovieCache{}, fmt.Errorf("failed to marshal genres: %w", err)
		}
		cache.Genres = datatypes.JSON(b)
	}

	cache.Runtime = movie.Runtime

	if len(movie.ProductionCountries) > 0 {
		b, err := json.Marshal(movie.ProductionCountries)
		if err != nil {
			return model.MovieCache{}, fmt.Errorf("failed to marshal production_countries: %w", err)
		}
		cache.ProductionCountries = datatypes.JSON(b)
	}

	if movie.Credits != nil {
		b, err := json.Marshal(movie.Credits)
		if err != nil {
			return model.MovieCache{}, fmt.Errorf("failed to marshal credits: %w", err)
		}
		cache.Credits = datatypes.JSON(b)
	}

	return cache, nil
}

// 映画詳細レスポンスの型定義。
type MovieDetailResponse struct {
	TmdbMovieID         int                 `json:"tmdb_movie_id"`
	Title               string              `json:"title"`
	OriginalTitle       *string             `json:"original_title,omitempty"`
	PosterPath          *string             `json:"poster_path,omitempty"`
	ReleaseDate         *string             `json:"release_date,omitempty"`
	VoteAverage         *float64            `json:"vote_average,omitempty"`
	Overview            *string             `json:"overview,omitempty"`
	Genres              []GenreItem         `json:"genres"`
	Runtime             *int                `json:"runtime,omitempty"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	Directors           []string            `json:"directors"`
	Cast                []CastMember        `json:"cast"`
}

type GenreItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductionCountry struct {
	ISO31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

type CastMember struct {
	Name      string `json:"name"`
	Character string `json:"character"`
}

// この映画が含まれる公開タグの情報。
type MovieRelatedTagItem struct {
	TagID         string `json:"tag_id"`
	Title         string `json:"title"`
	FollowerCount int    `json:"follower_count"`
	MovieCount    int    `json:"movie_count"`
}

// 指定した TMDB 映画 ID の詳細情報を取得する。
func (s *movieService) GetMovieDetail(ctx context.Context, tmdbMovieID int) (*MovieDetailResponse, error) {
	cache, err := s.EnsureMovieCache(ctx, tmdbMovieID)
	if err != nil {
		return nil, err
	}

	resp := &MovieDetailResponse{
		TmdbMovieID:   cache.TmdbMovieID,
		Title:         cache.Title,
		OriginalTitle: cache.OriginalTitle,
		PosterPath:    cache.PosterPath,
		VoteAverage:   cache.VoteAverage,
		Overview:      cache.Overview,
		Runtime:       cache.Runtime,
	}

	if cache.ReleaseDate != nil {
		s := cache.ReleaseDate.Format("2006-01-02")
		resp.ReleaseDate = &s
	}

	// Genres の復元
	if len(cache.Genres) > 0 {
		var genres []GenreItem
		if err := json.Unmarshal(cache.Genres, &genres); err == nil {
			resp.Genres = genres
		}
	}
	if resp.Genres == nil {
		resp.Genres = []GenreItem{}
	}

	// ProductionCountries の復元
	if len(cache.ProductionCountries) > 0 {
		var countries []ProductionCountry
		if err := json.Unmarshal(cache.ProductionCountries, &countries); err == nil {
			resp.ProductionCountries = countries
		}
	}
	if resp.ProductionCountries == nil {
		resp.ProductionCountries = []ProductionCountry{}
	}

	// Credits の復元（監督・キャスト抽出）
	if len(cache.Credits) > 0 {
		var credits tmdbCreditsResponse
		if err := json.Unmarshal(cache.Credits, &credits); err == nil {
			// 監督を抽出
			for _, c := range credits.Crew {
				if c.Job == "Director" {
					resp.Directors = append(resp.Directors, c.Name)
				}
			}
			// キャストを order 順で上位10名
			limit := 10
			if len(credits.Cast) < limit {
				limit = len(credits.Cast)
			}
			for i := 0; i < limit; i++ {
				resp.Cast = append(resp.Cast, CastMember{
					Name:      credits.Cast[i].Name,
					Character: credits.Cast[i].Character,
				})
			}
		}
	}
	if resp.Directors == nil {
		resp.Directors = []string{}
	}
	if resp.Cast == nil {
		resp.Cast = []CastMember{}
	}

	return resp, nil
}

// 指定した TMDB 映画 ID が含まれる公開タグの一覧を取得する。
func (s *movieService) GetMovieRelatedTags(ctx context.Context, tmdbMovieID int, limit int) ([]MovieRelatedTagItem, error) {
	if limit <= 0 {
		limit = 10
	}

	var results []MovieRelatedTagItem
	err := s.db.WithContext(ctx).
		Table("tag_movies tm").
		Select(`t.id AS tag_id, t.title,
			(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
			(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count`).
		Joins("JOIN tags t ON t.id = tm.tag_id AND t.is_public = true").
		Where("tm.tmdb_movie_id = ?", tmdbMovieID).
		Order("follower_count DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get related tags: %w", err)
	}
	if results == nil {
		results = []MovieRelatedTagItem{}
	}
	return results, nil
}

// movie_cache テーブルに対して UPSERT を行う。
func (s *movieService) upsertMovieCache(ctx context.Context, cache *model.MovieCache) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tmdb_movie_id"}},
		UpdateAll: true,
	}).Create(cache).Error
}
