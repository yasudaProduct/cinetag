package service

import (
	"context"

	"cinetag-backend/internal/model"
)

// TagService は映画タグに関するユースケースを表すインターフェースです。
type TagService interface {
	// ListTags は利用可能な映画タグ一覧を返します。
	ListTags(ctx context.Context) ([]model.Tag, error)
}

// mockTagService はインメモリのモック実装です。
type mockTagService struct {
	tags []model.Tag
}

// NewMockTagService はモックデータを持つ TagService 実装を返します。
func NewMockTagService() TagService {
	return &mockTagService{
		tags: []model.Tag{
			{ID: 1, Name: "アクション"},
			{ID: 2, Name: "コメディ"},
			{ID: 3, Name: "ドラマ"},
			{ID: 4, Name: "SF"},
			{ID: 5, Name: "ホラー"},
		},
	}
}

func (s *mockTagService) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.tags, nil
}


