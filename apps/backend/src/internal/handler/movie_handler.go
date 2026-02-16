package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// 映画検索等のHTTPハンドラです。
type MovieHandler struct {
	logger       *slog.Logger
	movieService service.MovieService
}

func NewMovieHandler(logger *slog.Logger, movieService service.MovieService) *MovieHandler {
	return &MovieHandler{
		logger:       logger,
		movieService: movieService,
	}
}

// TMDB 検索結果を返します。
// GET /api/v1/movies/search?q={query}&page={page}
func (h *MovieHandler) SearchMovies(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page := parseIntDefault(c.Query("page"), 1)

	items, total, err := h.movieService.SearchMovies(c.Request.Context(), q, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"total_count": total,
	})
}

// 映画詳細を返します。
// GET /api/v1/movies/:tmdbMovieId
func (h *MovieHandler) GetMovieDetail(c *gin.Context) {
	tmdbMovieID, err := strconv.Atoi(c.Param("tmdbMovieId"))
	if err != nil || tmdbMovieID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tmdb_movie_id"})
		return
	}

	detail, err := h.movieService.GetMovieDetail(c.Request.Context(), tmdbMovieID)
	if err != nil {
		h.logger.Error("handler.GetMovieDetail failed",
			"tmdb_movie_id", tmdbMovieID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie detail"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// この映画が含まれるタグ一覧を返します。
// GET /api/v1/movies/:tmdbMovieId/tags
func (h *MovieHandler) GetMovieTags(c *gin.Context) {
	tmdbMovieID, err := strconv.Atoi(c.Param("tmdbMovieId"))
	if err != nil || tmdbMovieID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tmdb_movie_id"})
		return
	}

	limit := parseIntDefault(c.Query("limit"), 10)

	tags, err := h.movieService.GetMovieRelatedTags(c.Request.Context(), tmdbMovieID, limit)
	if err != nil {
		h.logger.Error("handler.GetMovieTags failed",
			"tmdb_movie_id", tmdbMovieID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": tags})
}
