package testutil

import (
	"context"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"
)

// FakeTagRepository は repository.TagRepository の手書き fake です。
// 必要なテストで Fn を差し替えて使います。
type FakeTagRepository struct {
	CreateFn              func(ctx context.Context, tag *model.Tag) error
	FindByIDFn            func(ctx context.Context, id string) (*model.Tag, error)
	FindDetailByIDFn      func(ctx context.Context, id string) (*repository.TagDetailRow, error)
	UpdateByIDFn          func(ctx context.Context, id string, patch repository.TagUpdatePatch) error
	IncrementMovieCountFn func(ctx context.Context, id string, delta int) error
	ListPublicTagsFn      func(ctx context.Context, filter repository.TagListFilter) ([]repository.TagSummary, int64, error)
	ListTagsByUserIDFn    func(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error)
}

func (f *FakeTagRepository) Create(ctx context.Context, tag *model.Tag) error {
	if f.CreateFn == nil {
		return nil
	}
	return f.CreateFn(ctx, tag)
}

func (f *FakeTagRepository) FindByID(ctx context.Context, id string) (*model.Tag, error) {
	if f.FindByIDFn == nil {
		return nil, nil
	}
	return f.FindByIDFn(ctx, id)
}

func (f *FakeTagRepository) FindDetailByID(ctx context.Context, id string) (*repository.TagDetailRow, error) {
	if f.FindDetailByIDFn == nil {
		return nil, nil
	}
	return f.FindDetailByIDFn(ctx, id)
}

func (f *FakeTagRepository) UpdateByID(ctx context.Context, id string, patch repository.TagUpdatePatch) error {
	if f.UpdateByIDFn == nil {
		return nil
	}
	return f.UpdateByIDFn(ctx, id, patch)
}

func (f *FakeTagRepository) IncrementMovieCount(ctx context.Context, id string, delta int) error {
	if f.IncrementMovieCountFn == nil {
		return nil
	}
	return f.IncrementMovieCountFn(ctx, id, delta)
}

func (f *FakeTagRepository) ListPublicTags(ctx context.Context, filter repository.TagListFilter) ([]repository.TagSummary, int64, error) {
	if f.ListPublicTagsFn == nil {
		return []repository.TagSummary{}, 0, nil
	}
	return f.ListPublicTagsFn(ctx, filter)
}

func (f *FakeTagRepository) ListTagsByUserID(ctx context.Context, filter repository.UserTagListFilter) ([]repository.TagSummary, int64, error) {
	if f.ListTagsByUserIDFn == nil {
		return []repository.TagSummary{}, 0, nil
	}
	return f.ListTagsByUserIDFn(ctx, filter)
}

// FakeTagMovieRepository は repository.TagMovieRepository の手書き fake です。
type FakeTagMovieRepository struct {
	ListRecentByTagFn       func(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error)
	ListByTagFn             func(ctx context.Context, tagID string, offset, limit int) ([]repository.TagMovieWithCache, int64, error)
	CreateFn                func(ctx context.Context, tagMovie *model.TagMovie) error
	FindByIDFn              func(ctx context.Context, tagMovieID string) (*model.TagMovie, error)
	DeleteFn                func(ctx context.Context, tagMovieID string) error
	ListContributorsByTagFn func(ctx context.Context, tagID string, ownerID string, limit int) ([]repository.TagContributor, int64, error)
}

func (f *FakeTagMovieRepository) ListContributorsByTag(ctx context.Context, tagID string, ownerID string, limit int) ([]repository.TagContributor, int64, error) {
	if f.ListContributorsByTagFn == nil {
		return []repository.TagContributor{}, 0, nil
	}
	return f.ListContributorsByTagFn(ctx, tagID, ownerID, limit)
}

func (f *FakeTagMovieRepository) ListRecentByTag(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error) {
	if f.ListRecentByTagFn == nil {
		return []model.TagMovie{}, nil
	}
	return f.ListRecentByTagFn(ctx, tagID, limit)
}

func (f *FakeTagMovieRepository) ListByTag(ctx context.Context, tagID string, offset, limit int) ([]repository.TagMovieWithCache, int64, error) {
	if f.ListByTagFn == nil {
		return []repository.TagMovieWithCache{}, 0, nil
	}
	return f.ListByTagFn(ctx, tagID, offset, limit)
}

func (f *FakeTagMovieRepository) Create(ctx context.Context, tagMovie *model.TagMovie) error {
	if f.CreateFn == nil {
		return nil
	}
	return f.CreateFn(ctx, tagMovie)
}

func (f *FakeTagMovieRepository) FindByID(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
	if f.FindByIDFn == nil {
		return nil, nil
	}
	return f.FindByIDFn(ctx, tagMovieID)
}

func (f *FakeTagMovieRepository) Delete(ctx context.Context, tagMovieID string) error {
	if f.DeleteFn == nil {
		return nil
	}
	return f.DeleteFn(ctx, tagMovieID)
}

// FakeTagFollowerRepository は repository.TagFollowerRepository の手書き fake です。
type FakeTagFollowerRepository struct {
	CreateFn            func(ctx context.Context, tagID, userID string) error
	DeleteFn            func(ctx context.Context, tagID, userID string) error
	IsFollowingFn       func(ctx context.Context, tagID, userID string) (bool, error)
	ListFollowersFn     func(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error)
	CountFollowersFn    func(ctx context.Context, tagID string) (int64, error)
	ListFollowingTagsFn func(ctx context.Context, userID string, page, pageSize int) ([]*model.Tag, int64, error)
}

func (f *FakeTagFollowerRepository) Create(ctx context.Context, tagID, userID string) error {
	if f.CreateFn == nil {
		return nil
	}
	return f.CreateFn(ctx, tagID, userID)
}

func (f *FakeTagFollowerRepository) Delete(ctx context.Context, tagID, userID string) error {
	if f.DeleteFn == nil {
		return nil
	}
	return f.DeleteFn(ctx, tagID, userID)
}

func (f *FakeTagFollowerRepository) IsFollowing(ctx context.Context, tagID, userID string) (bool, error) {
	if f.IsFollowingFn == nil {
		return false, nil
	}
	return f.IsFollowingFn(ctx, tagID, userID)
}

func (f *FakeTagFollowerRepository) ListFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListFollowersFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListFollowersFn(ctx, tagID, page, pageSize)
}

func (f *FakeTagFollowerRepository) CountFollowers(ctx context.Context, tagID string) (int64, error) {
	if f.CountFollowersFn == nil {
		return 0, nil
	}
	return f.CountFollowersFn(ctx, tagID)
}

func (f *FakeTagFollowerRepository) ListFollowingTags(ctx context.Context, userID string, page, pageSize int) ([]*model.Tag, int64, error) {
	if f.ListFollowingTagsFn == nil {
		return []*model.Tag{}, 0, nil
	}
	return f.ListFollowingTagsFn(ctx, userID, page, pageSize)
}
