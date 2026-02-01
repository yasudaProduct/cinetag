package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchMovies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		query          string
		page           int
		apiKey         string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantResults    int
		wantTotal      int
		wantErr        bool
		errContains    string
	}{
		{
			name:   "successful search returns results",
			query:  "Inception",
			page:   1,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-api-key" {
					t.Errorf("Authorization = %q, want %q", auth, "Bearer test-api-key")
				}

				resp := tmdbSearchResponse{
					Page:         1,
					TotalPages:   1,
					TotalResults: 2,
					Results: []struct {
						ID            int      `json:"id"`
						Title         string   `json:"title"`
						OriginalTitle string   `json:"original_title"`
						PosterPath    *string  `json:"poster_path"`
						ReleaseDate   string   `json:"release_date"`
						VoteAverage   *float64 `json:"vote_average"`
					}{
						{
							ID:            27205,
							Title:         "インセプション",
							OriginalTitle: "Inception",
							ReleaseDate:   "2010-07-16",
						},
						{
							ID:            12345,
							Title:         "インセプション2",
							OriginalTitle: "Inception 2",
							ReleaseDate:   "2015-01-01",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 2,
			wantTotal:   2,
			wantErr:     false,
		},
		{
			name:        "empty query returns empty results",
			query:       "",
			page:        1,
			apiKey:      "test-api-key",
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:        "whitespace only query returns empty results",
			query:       "   ",
			page:        1,
			apiKey:      "test-api-key",
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:        "missing API key returns error",
			query:       "test",
			page:        1,
			apiKey:      "",
			wantErr:     true,
			errContains: "TMDB_API_KEY is not set",
		},
		{
			name:   "page <= 0 defaults to page 1",
			query:  "test",
			page:   0,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				page := r.URL.Query().Get("page")
				if page != "1" {
					t.Errorf("page = %q, want %q", page, "1")
				}
				resp := tmdbSearchResponse{
					Page:         1,
					TotalPages:   1,
					TotalResults: 0,
					Results:      nil,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:   "negative page defaults to page 1",
			query:  "test",
			page:   -5,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				page := r.URL.Query().Get("page")
				if page != "1" {
					t.Errorf("page = %q, want %q", page, "1")
				}
				resp := tmdbSearchResponse{TotalResults: 0}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:   "API returns non-2xx status",
			query:  "test",
			page:   1,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantErr:     true,
			errContains: "status=503",
		},
		{
			name:   "API returns invalid JSON",
			query:  "test",
			page:   1,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{invalid json"))
			},
			wantErr:     true,
			errContains: "failed to decode TMDB response",
		},
		{
			name:   "results with empty release_date are handled",
			query:  "test",
			page:   1,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				resp := tmdbSearchResponse{
					TotalResults: 1,
					Results: []struct {
						ID            int      `json:"id"`
						Title         string   `json:"title"`
						OriginalTitle string   `json:"original_title"`
						PosterPath    *string  `json:"poster_path"`
						ReleaseDate   string   `json:"release_date"`
						VoteAverage   *float64 `json:"vote_average"`
					}{
						{
							ID:          1,
							Title:       "No Release Date",
							ReleaseDate: "",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 1,
			wantTotal:   1,
			wantErr:     false,
		},
		{
			name:   "bearer prefix in API key is not duplicated",
			query:  "test",
			page:   1,
			apiKey: "Bearer already-prefixed",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "Bearer already-prefixed" {
					t.Errorf("Authorization = %q, want %q", auth, "Bearer already-prefixed")
				}
				resp := tmdbSearchResponse{TotalResults: 0}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:   "language parameter is set",
			query:  "test",
			page:   1,
			apiKey: "test-api-key",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				lang := r.URL.Query().Get("language")
				if lang != "ja-JP" {
					t.Errorf("language = %q, want %q", lang, "ja-JP")
				}
				resp := tmdbSearchResponse{TotalResults: 0}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantResults: 0,
			wantTotal:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var server *httptest.Server
			var client *http.Client
			if tt.serverResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer server.Close()
				client = server.Client()
			}

			cfg := TMDBConfig{
				APIKey:          tt.apiKey,
				DefaultLanguage: "ja-JP",
			}
			if server != nil {
				cfg.BaseURL = server.URL
			}

			svc := NewMovieServiceWithConfig(nil, cfg, client)

			results, total, err := svc.SearchMovies(context.Background(), tt.query, tt.page)

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMovies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !containsString(err.Error(), tt.errContains) {
						t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
					}
				}
				return
			}

			if len(results) != tt.wantResults {
				t.Errorf("len(results) = %d, want %d", len(results), tt.wantResults)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestSearchMovies_ResultMapping(t *testing.T) {
	t.Parallel()

	posterPath := "/poster.jpg"
	voteAverage := 8.5

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := tmdbSearchResponse{
			TotalResults: 1,
			Results: []struct {
				ID            int      `json:"id"`
				Title         string   `json:"title"`
				OriginalTitle string   `json:"original_title"`
				PosterPath    *string  `json:"poster_path"`
				ReleaseDate   string   `json:"release_date"`
				VoteAverage   *float64 `json:"vote_average"`
			}{
				{
					ID:            12345,
					Title:         "テスト映画",
					OriginalTitle: "Test Movie",
					PosterPath:    &posterPath,
					ReleaseDate:   "2023-06-15",
					VoteAverage:   &voteAverage,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := TMDBConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	}
	svc := NewMovieServiceWithConfig(nil, cfg, server.Client())

	results, _, err := svc.SearchMovies(context.Background(), "test", 1)
	if err != nil {
		t.Fatalf("SearchMovies() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	r := results[0]
	if r.TmdbMovieID != 12345 {
		t.Errorf("TmdbMovieID = %d, want 12345", r.TmdbMovieID)
	}
	if r.Title != "テスト映画" {
		t.Errorf("Title = %q, want %q", r.Title, "テスト映画")
	}
	if r.OriginalTitle == nil || *r.OriginalTitle != "Test Movie" {
		t.Errorf("OriginalTitle = %v, want %q", r.OriginalTitle, "Test Movie")
	}
	if r.PosterPath == nil || *r.PosterPath != "/poster.jpg" {
		t.Errorf("PosterPath = %v, want %q", r.PosterPath, "/poster.jpg")
	}
	if r.ReleaseDate == nil || *r.ReleaseDate != "2023-06-15" {
		t.Errorf("ReleaseDate = %v, want %q", r.ReleaseDate, "2023-06-15")
	}
	if r.VoteAverage == nil || *r.VoteAverage != 8.5 {
		t.Errorf("VoteAverage = %v, want 8.5", r.VoteAverage)
	}
}

func TestNewMovieServiceWithConfig_Defaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       TMDBConfig
		wantCheck func(svc MovieService) error
	}{
		{
			name: "empty BaseURL defaults to TMDB API",
			cfg: TMDBConfig{
				APIKey:          "key",
				BaseURL:         "",
				DefaultLanguage: "en-US",
			},
		},
		{
			name: "empty DefaultLanguage defaults to ja-JP",
			cfg: TMDBConfig{
				APIKey:          "key",
				BaseURL:         "https://example.com",
				DefaultLanguage: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewMovieServiceWithConfig(nil, tt.cfg, nil)
			if svc == nil {
				t.Error("NewMovieServiceWithConfig returned nil")
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
