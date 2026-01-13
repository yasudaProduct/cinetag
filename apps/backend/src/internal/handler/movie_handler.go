package handler

import (
	"log/slog"
	"net/http"
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
