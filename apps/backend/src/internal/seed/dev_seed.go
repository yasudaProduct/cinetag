package seed

import (
	"fmt"
	"os"
	"strings"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	seedTagID = "154c44ab-3ba9-4c26-9642-49d11d3d1ff5"
)

func isDevelop() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv("ENV"))) == "develop"
}

// 開発環境（ENV=develop）のときだけ、開発用のseedデータを投入する。
//
// 何度実行しても整合性が壊れないよう、ユニーク制約を利用した upsert / do-nothing で冪等化する。
func SeedDevelop(db *gorm.DB) error {
	if !isDevelop() {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		creatorID, err := upsertUser(tx, model.User{
			ClerkUserID: "user_376ohwfbOYMZnEyFXrIpZpRzU0h", // clerkに登録済みのユーザーID
			DisplayID:   "demo01",
			DisplayName: "デモ01",
			Email:       "demo01@example.com",
		})
		if err != nil {
			return fmt.Errorf("upsert seed creator: %w", err)
		}

		followerID, err := upsertUser(tx, model.User{
			ClerkUserID: "user_376ozoBYiBZiFipgkEerZGXCjMm",
			DisplayID:   "demo02",
			DisplayName: "デモ02",
			Email:       "demo02@example.com",
		})
		if err != nil {
			return fmt.Errorf("upsert seed follower: %w", err)
		}

		tagDescription := "開発用seedタグです。migrate実行で自動投入されます。"
		seedTag := model.Tag{
			ID:          seedTagID,
			UserID:      creatorID,
			Title:       "Seed Tag",
			Description: &tagDescription,
			IsPublic:    true,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_id",
				"title",
				"description",
				"cover_image_url",
				"is_public",
			}),
		}).Create(&seedTag).Error; err != nil {
			return fmt.Errorf("upsert seed tag: %w", err)
		}

		movies := []struct {
			tmdbID   int
			position int
			note     *string
		}{
			{tmdbID: 550, position: 0, note: strPtr("Fight Club")},
			{tmdbID: 603, position: 1, note: strPtr("The Matrix")},
			{tmdbID: 155, position: 2, note: strPtr("The Dark Knight")},
			{tmdbID: 680, position: 3, note: strPtr("Pulp Fiction")},
			{tmdbID: 27205, position: 4, note: strPtr("Inception")},
		}

		for _, m := range movies {
			tm := model.TagMovie{
				TagID:       seedTagID,
				TmdbMovieID: m.tmdbID,
				AddedByUser: creatorID,
				Note:        m.note,
				Position:    m.position,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "tag_id"}, {Name: "tmdb_movie_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"added_by_user_id",
					"note",
					"position",
				}),
			}).Create(&tm).Error; err != nil {
				return fmt.Errorf("upsert seed tag_movies (tmdb=%d): %w", m.tmdbID, err)
			}
		}

		follow := model.TagFollower{
			TagID:  seedTagID,
			UserID: followerID,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tag_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&follow).Error; err != nil {
			return fmt.Errorf("insert seed tag_follower: %w", err)
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

func strPtr(s string) *string {
	return &s
}
