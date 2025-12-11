package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MovieService は TMDB API と movie_cache テーブルを利用して
// 映画情報の取得・キャッシュ更新を行うユースケースを表します。
type MovieService interface {
	// EnsureMovieCache は指定した TMDB 映画 ID に対応する movie_cache レコードの
	// 存在と有効期限を保証します。
	// - 有効なキャッシュがあればそれを返す
	// - キャッシュが無い、または期限切れの場合は TMDB から取得してキャッシュを更新する
	EnsureMovieCache(ctx context.Context, tmdbMovieID int) (*model.MovieCache, error)
}

// TMDBConfig は TMDB 連携に必要な設定値を表します。
type TMDBConfig struct {
	APIKey          string
	BaseURL         string
	DefaultLanguage string
}

type movieService struct {
	db         *gorm.DB
	httpClient *http.Client
	cfg        TMDBConfig
}

// NewMovieService は MovieService の実装を生成します。
// 環境変数から TMDB の設定値を読み込みます。
func NewMovieService(db *gorm.DB) MovieService {
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
		db: db,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cfg: cfg,
	}
}

// NewMovieServiceWithConfig はテストや将来の拡張用に、設定と HTTP クライアントを外部から注入するためのコンストラクタです。
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

// tmdbMovieResponse は TMDB の /movie/{movie_id} レスポンスのうち、必要なフィールドのみを表します。
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
	Runtime *int `json:"runtime"`
}

// EnsureMovieCache は指定した TMDB 映画 ID に対応する movie_cache レコードの
// 存在と有効期限を保証します。
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
		return &cache, nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		// それ以外の DB エラーはそのまま返す
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

// fetchMovieFromTMDB は TMDB の /movie/{movie_id} エンドポイントから映画情報を取得します。
func (s *movieService) fetchMovieFromTMDB(ctx context.Context, tmdbMovieID int) (*tmdbMovieResponse, error) {
	base, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB_BASE_URL: %w", err)
	}

	base.Path = path.Join(base.Path, "movie", strconv.Itoa(tmdbMovieID))

	q := base.Query()
	q.Set("api_key", s.cfg.APIKey)
	if s.cfg.DefaultLanguage != "" {
		q.Set("language", s.cfg.DefaultLanguage)
	}
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create TMDB request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call TMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tmdb movie not found: %d", tmdbMovieID)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb request failed: status=%d", resp.StatusCode)
	}

	var body tmdbMovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode TMDB response: %w", err)
	}

	return &body, nil
}

// buildMovieCacheFromTMDB は TMDB レスポンスから MovieCache エンティティを構築します。
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

	return cache, nil
}

// upsertMovieCache は movie_cache テーブルに対して UPSERT を行います。
func (s *movieService) upsertMovieCache(ctx context.Context, cache *model.MovieCache) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tmdb_movie_id"}},
		UpdateAll: true,
	}).Create(cache).Error
}


