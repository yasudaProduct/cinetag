package service

import (
	"context"
	"math/rand"

	"cinetag-backend/src/internal/repository"
)

const (
	userDisplayIDPrefix    = "user-"
	userDisplayIDSuffixLen = 6
	userDisplayIDChars     = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// generateRandomString は指定長のランダム英数字を生成して返します。
func generateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = userDisplayIDChars[rand.Intn(len(userDisplayIDChars))]
	}
	return string(b)
}

// generateUniqueUserDisplayID は user display_id を生成します。
// 重複がないかチェックし、重複していたら再帰的に生成し直します。
func generateUniqueUserDisplayID(ctx context.Context, userRepo repository.UserRepository) string {
	displayID := userDisplayIDPrefix + generateRandomString(userDisplayIDSuffixLen)

	if IsValidUserDisplayID(ctx, userRepo, displayID) {
		return displayID
	}
	return generateUniqueUserDisplayID(ctx, userRepo)
}

// GenerateUserDisplayID は user display_id を生成します。
func GenerateUserDisplayID(ctx context.Context, userRepo repository.UserRepository) string {
	return generateUniqueUserDisplayID(ctx, userRepo)
}

func IsValidUserDisplayID(ctx context.Context, userRepo repository.UserRepository, displayID string) bool {
	if _, err := userRepo.FindByDisplayID(ctx, displayID); err == nil {
		return false
	}
	return true
}
