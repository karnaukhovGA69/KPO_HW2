package repo

import (
	"context"
	"main/domain"
	"sync"
)

type CachedCategoryRepo struct {
	inner *PgCategoryRepo
	mu    sync.RWMutex
	list  []domain.Category
	byID  map[domain.CategoryID]domain.Category
}

func NewCachedCategoryRepo(inner *PgCategoryRepo) *CachedCategoryRepo {
	return &CachedCategoryRepo{inner: inner, byID: map[domain.CategoryID]domain.Category{}}
}

func (r *CachedCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	r.mu.RLock()
	if r.list != nil {
		defer r.mu.RUnlock()
		return append([]domain.Category(nil), r.list...), nil
	}
	r.mu.RUnlock()

	cats, err := r.inner.List(ctx)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.list = append([]domain.Category(nil), cats...)
	r.byID = map[domain.CategoryID]domain.Category{}
	for _, c := range cats {
		r.byID[c.ID] = c
	}
	r.mu.Unlock()
	return cats, nil
}

func (r *CachedCategoryRepo) Get(ctx context.Context, id domain.CategoryID) (domain.Category, error) {
	r.mu.RLock()
	if c, ok := r.byID[id]; ok {
		r.mu.RUnlock()
		return c, nil
	}
	r.mu.RUnlock()
	c, err := r.inner.Get(ctx, id)
	if err != nil {
		return c, err
	}
	r.mu.Lock()
	r.byID[id] = c
	r.mu.Unlock()
	return c, nil
}

func (r *CachedCategoryRepo) Create(ctx context.Context, c domain.Category) error {
	if err := r.inner.Create(ctx, c); err != nil {
		return err
	}
	r.invalidate()
	return nil
}
func (r *CachedCategoryRepo) UpdateName(ctx context.Context, id domain.CategoryID, name string) error {
	if err := r.inner.UpdateName(ctx, id, name); err != nil {
		return err
	}
	r.invalidate()
	return nil
}
func (r *CachedCategoryRepo) UpdateType(ctx context.Context, id domain.CategoryID, t domain.CategoryType) error {
	if err := r.inner.UpdateType(ctx, id, t); err != nil {
		return err
	}
	r.invalidate()
	return nil
}
func (r *CachedCategoryRepo) Delete(ctx context.Context, id domain.CategoryID) error {
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate()
	return nil
}
func (r *CachedCategoryRepo) HasOperations(ctx context.Context, id domain.CategoryID) (bool, error) {
	return r.inner.HasOperations(ctx, id)
}
func (r *CachedCategoryRepo) invalidate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.list = nil
	r.byID = map[domain.CategoryID]domain.Category{}
}
