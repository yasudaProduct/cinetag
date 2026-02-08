package seed

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func isDevelop() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv("ENV"))) == "develop"
}

type devSeedData struct {
	Users        []devSeedUser        `json:"users"`
	Tags         []devSeedTag         `json:"tags"`
	TagMovies    []devSeedTagMovie    `json:"tag_movies"`
	TagFollowers []devSeedTagFollower `json:"tag_followers"`
}

type devSeedUser struct {
	Ref         string  `json:"ref"`
	ClerkUserID string  `json:"clerk_user_id"`
	DisplayID   string  `json:"display_id"`
	DisplayName string  `json:"display_name"`
	Email       string  `json:"email"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
}

type devSeedTag struct {
	ID             string  `json:"id"`
	UserRef        string  `json:"user_ref"`
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	CoverImageURL  *string `json:"cover_image_url"`
	IsPublic       bool    `json:"is_public"`
	AddMoviePolicy string  `json:"add_movie_policy"`
}

type devSeedTagMovie struct {
	TagID          string  `json:"tag_id"`
	TmdbMovieID    int     `json:"tmdb_movie_id"`
	Position       int     `json:"position"`
	Note           *string `json:"note"`
	AddedByUserRef string  `json:"added_by_user_ref"`
}

type devSeedTagFollower struct {
	TagID   string `json:"tag_id"`
	UserRef string `json:"user_ref"`
}

func loadDevSeedData() (*devSeedData, string, error) {
	candidates, err := devSeedJSONCandidatePaths()
	if err != nil {
		return nil, "", err
	}

	var errs []error
	for _, p := range candidates {
		fi, statErr := os.Stat(p)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			errs = append(errs, fmt.Errorf("stat %s: %w", p, statErr))
			continue
		}

		var (
			d *devSeedData
		)
		if fi.IsDir() {
			d, err = loadDevSeedDataFromDir(p)
		} else {
			d, err = loadDevSeedDataFromFile(p)
		}
		if err != nil {
			return nil, "", err
		}
		if err := validateDevSeedData(d); err != nil {
			return nil, "", fmt.Errorf("validate %s: %w", p, err)
		}
		return d, p, nil
	}

	if len(errs) > 0 {
		return nil, "", errors.Join(errs...)
	}
	return nil, "", fmt.Errorf("dev seed data not found. tried: %s", strings.Join(candidates, ", "))
}

func devSeedJSONCandidatePaths() ([]string, error) {
	if p := strings.TrimSpace(os.Getenv("SEED_DEV_JSON_PATH")); p != "" {
		return []string{p}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}

	relDirFromBackend := filepath.Join("src", "internal", "seed", "data", "dev_seed")
	relDirFromRepoRoot := filepath.Join("apps", "backend", "src", "internal", "seed", "data", "dev_seed")
	// 互換性のため、単一ファイル形式も候補に残す
	relFileFromBackend := filepath.Join("src", "internal", "seed", "data", "dev_seed.json")
	relFileFromRepoRoot := filepath.Join("apps", "backend", "src", "internal", "seed", "data", "dev_seed.json")

	var out []string
	out = append(out, filepath.Join(cwd, relDirFromBackend))
	out = append(out, filepath.Join(cwd, relDirFromRepoRoot))
	out = append(out, filepath.Join(cwd, relFileFromBackend))
	out = append(out, filepath.Join(cwd, relFileFromRepoRoot))

	// cwd から数階層だけ上へ辿って探索（壊れにくさ優先・無限探索はしない）
	const maxUp = 6
	dir := cwd
	for i := 0; i < maxUp; i++ {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
		out = append(out, filepath.Join(dir, relDirFromBackend))
		out = append(out, filepath.Join(dir, relDirFromRepoRoot))
		out = append(out, filepath.Join(dir, relFileFromBackend))
		out = append(out, filepath.Join(dir, relFileFromRepoRoot))
	}

	return uniqueStrings(out), nil
}

func loadDevSeedDataFromFile(path string) (*devSeedData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var d devSeedData
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &d, nil
}

func loadDevSeedDataFromDir(dir string) (*devSeedData, error) {
	usersPath := filepath.Join(dir, "users.json")
	tagsPath := filepath.Join(dir, "tags.json")
	tagMoviesPath := filepath.Join(dir, "tag_movies.json")
	tagFollowersPath := filepath.Join(dir, "tag_followers.json")

	var d devSeedData
	if err := readJSONArray(usersPath, &d.Users); err != nil {
		return nil, err
	}
	if err := readJSONArray(tagsPath, &d.Tags); err != nil {
		return nil, err
	}
	if err := readJSONArray(tagMoviesPath, &d.TagMovies); err != nil {
		return nil, err
	}
	if err := readJSONArray(tagFollowersPath, &d.TagFollowers); err != nil {
		return nil, err
	}
	return &d, nil
}

func readJSONArray(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func validateDevSeedData(d *devSeedData) error {
	if d == nil {
		return errors.New("seed data is nil")
	}

	userIDs := map[string]struct{}{}
	userRefs := map[string]struct{}{}
	for i, u := range d.Users {
		if strings.TrimSpace(u.Ref) == "" {
			return fmt.Errorf("users[%d].ref is required", i)
		}
		if _, ok := userRefs[u.Ref]; ok {
			return fmt.Errorf("users[%d].ref duplicated: %s", i, u.Ref)
		}
		userRefs[u.Ref] = struct{}{}

		if strings.TrimSpace(u.ClerkUserID) == "" {
			return fmt.Errorf("users[%d].clerk_user_id is required", i)
		}
		if strings.TrimSpace(u.DisplayID) == "" {
			return fmt.Errorf("users[%d].display_id is required", i)
		}
		if strings.TrimSpace(u.DisplayName) == "" {
			return fmt.Errorf("users[%d].display_name is required", i)
		}
		if strings.TrimSpace(u.Email) == "" {
			return fmt.Errorf("users[%d].email is required", i)
		}
		_ = userIDs // reserved for future constraints
	}

	tagIDs := map[string]struct{}{}
	for i, t := range d.Tags {
		if strings.TrimSpace(t.ID) == "" {
			return fmt.Errorf("tags[%d].id is required", i)
		}
		if _, ok := tagIDs[t.ID]; ok {
			return fmt.Errorf("tags[%d].id duplicated: %s", i, t.ID)
		}
		tagIDs[t.ID] = struct{}{}

		if strings.TrimSpace(t.UserRef) == "" {
			return fmt.Errorf("tags[%d].user_ref is required", i)
		}
		if _, ok := userRefs[t.UserRef]; !ok {
			return fmt.Errorf("tags[%d].user_ref not found: %s", i, t.UserRef)
		}
		if strings.TrimSpace(t.Title) == "" {
			return fmt.Errorf("tags[%d].title is required", i)
		}
	}

	type movieKey struct {
		tagID string
		tmdb  int
	}
	type posKey struct {
		tagID    string
		position int
	}
	movieKeys := map[movieKey]struct{}{}
	posKeys := map[posKey]struct{}{}
	for i, tm := range d.TagMovies {
		if strings.TrimSpace(tm.TagID) == "" {
			return fmt.Errorf("tag_movies[%d].tag_id is required", i)
		}
		if _, ok := tagIDs[tm.TagID]; !ok {
			return fmt.Errorf("tag_movies[%d].tag_id not found: %s", i, tm.TagID)
		}
		if tm.TmdbMovieID <= 0 {
			return fmt.Errorf("tag_movies[%d].tmdb_movie_id must be > 0", i)
		}
		if tm.Position < 0 {
			return fmt.Errorf("tag_movies[%d].position must be >= 0", i)
		}
		if strings.TrimSpace(tm.AddedByUserRef) == "" {
			return fmt.Errorf("tag_movies[%d].added_by_user_ref is required", i)
		}
		if _, ok := userRefs[tm.AddedByUserRef]; !ok {
			return fmt.Errorf("tag_movies[%d].added_by_user_ref not found: %s", i, tm.AddedByUserRef)
		}

		mk := movieKey{tagID: tm.TagID, tmdb: tm.TmdbMovieID}
		if _, ok := movieKeys[mk]; ok {
			return fmt.Errorf("tag_movies[%d] duplicated (tag_id, tmdb_movie_id): (%s, %d)", i, tm.TagID, tm.TmdbMovieID)
		}
		movieKeys[mk] = struct{}{}

		pk := posKey{tagID: tm.TagID, position: tm.Position}
		if _, ok := posKeys[pk]; ok {
			return fmt.Errorf("tag_movies[%d] duplicated position within tag: (tag_id=%s, position=%d)", i, tm.TagID, tm.Position)
		}
		posKeys[pk] = struct{}{}
	}

	type followerKey struct {
		tagID string
		ref   string
	}
	followKeys := map[followerKey]struct{}{}
	for i, f := range d.TagFollowers {
		if strings.TrimSpace(f.TagID) == "" {
			return fmt.Errorf("tag_followers[%d].tag_id is required", i)
		}
		if _, ok := tagIDs[f.TagID]; !ok {
			return fmt.Errorf("tag_followers[%d].tag_id not found: %s", i, f.TagID)
		}
		if strings.TrimSpace(f.UserRef) == "" {
			return fmt.Errorf("tag_followers[%d].user_ref is required", i)
		}
		if _, ok := userRefs[f.UserRef]; !ok {
			return fmt.Errorf("tag_followers[%d].user_ref not found: %s", i, f.UserRef)
		}

		fk := followerKey{tagID: f.TagID, ref: f.UserRef}
		if _, ok := followKeys[fk]; ok {
			return fmt.Errorf("tag_followers[%d] duplicated (tag_id, user_ref): (%s, %s)", i, f.TagID, f.UserRef)
		}
		followKeys[fk] = struct{}{}
	}

	return nil
}

// 開発環境（ENV=develop）のときだけ、開発用のseedデータを投入する。
//
// 何度実行しても整合性が壊れないよう、ユニーク制約を利用した upsert / do-nothing で冪等化する。
func SeedDevelop(db *gorm.DB) error {
	if !isDevelop() {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		d, path, err := loadDevSeedData()
		if err != nil {
			return fmt.Errorf("load dev seed data: %w", err)
		}

		refToUserID := make(map[string]string, len(d.Users))
		for _, u := range d.Users {
			id, err := upsertUser(tx, model.User{
				ClerkUserID: u.ClerkUserID,
				DisplayID:   u.DisplayID,
				DisplayName: u.DisplayName,
				Email:       u.Email,
				AvatarURL:   u.AvatarURL,
				Bio:         u.Bio,
			})
			if err != nil {
				return fmt.Errorf("upsert seed user (ref=%s, src=%s): %w", u.Ref, path, err)
			}
			refToUserID[u.Ref] = id
		}

		for _, t := range d.Tags {
			ownerID := refToUserID[t.UserRef]
			addMoviePolicy := strings.TrimSpace(t.AddMoviePolicy)
			if addMoviePolicy == "" {
				addMoviePolicy = "everyone"
			}

			tag := model.Tag{
				ID:             t.ID,
				UserID:         ownerID,
				Title:          t.Title,
				Description:    t.Description,
				CoverImageURL:  t.CoverImageURL,
				IsPublic:       t.IsPublic,
				AddMoviePolicy: addMoviePolicy,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"user_id",
					"title",
					"description",
					"cover_image_url",
					"is_public",
					"add_movie_policy",
				}),
			}).Create(&tag).Error; err != nil {
				return fmt.Errorf("upsert seed tag (id=%s, src=%s): %w", t.ID, path, err)
			}
		}

		for _, m := range d.TagMovies {
			addedBy := refToUserID[m.AddedByUserRef]
			tm := model.TagMovie{
				TagID:       m.TagID,
				TmdbMovieID: m.TmdbMovieID,
				AddedByUser: addedBy,
				Note:        m.Note,
				Position:    m.Position,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "tag_id"}, {Name: "tmdb_movie_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"added_by_user_id",
					"note",
					"position",
				}),
			}).Create(&tm).Error; err != nil {
				return fmt.Errorf("upsert seed tag_movies (tag_id=%s, tmdb=%d, src=%s): %w", m.TagID, m.TmdbMovieID, path, err)
			}
		}

		for _, f := range d.TagFollowers {
			userID := refToUserID[f.UserRef]
			follow := model.TagFollower{
				TagID:  f.TagID,
				UserID: userID,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tag_id"}, {Name: "user_id"}},
				DoNothing: true,
			}).Create(&follow).Error; err != nil {
				return fmt.Errorf("insert seed tag_follower (tag_id=%s, src=%s): %w", f.TagID, path, err)
			}
		}

		return nil
	})
}

func upsertUser(tx *gorm.DB, u model.User) (string, error) {
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "clerk_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_id",
			"display_name",
			"email",
			"avatar_url",
			"bio",
		}),
	}).Create(&u).Error; err != nil {
		return "", err
	}

	var out model.User
	if err := tx.Where("clerk_user_id = ?", u.ClerkUserID).First(&out).Error; err != nil {
		return "", err
	}
	return out.ID, nil
}
