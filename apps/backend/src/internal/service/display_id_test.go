package service

import (
	"context"
	"testing"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

func TestGenerateUserDisplayID(t *testing.T) {
	t.Parallel()

	var calls int
	repo := &fakeUserRepo{
		FindByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
			calls++
			// 1回目は「既に存在する」扱いにして衝突を起こす
			if calls == 1 {
				return &model.User{ID: "u_exist"}, nil
			}
			return nil, gorm.ErrRecordNotFound
		},
	}

	got := GenerateUserDisplayID(context.Background(), repo)

	if calls < 2 {
		t.Fatalf("expected retry on collision, calls=%d", calls)
	}
	if len(got) != len(userDisplayIDPrefix)+userDisplayIDSuffixLen {
		t.Fatalf("unexpected length: %q", got)
	}
	if got[:len(userDisplayIDPrefix)] != userDisplayIDPrefix {
		t.Fatalf("expected prefix %q, got %q", userDisplayIDPrefix, got)
	}
}
