package facade

import (
	"context"

	"main/domain"
)

type CategoryRepo interface {
	List(ctx context.Context) ([]domain.Category, error)
	Get(ctx context.Context, id domain.CategoryID) (domain.Category, error)

	Create(ctx context.Context, c domain.Category) error
	UpdateName(ctx context.Context, id domain.CategoryID, name string) error
	UpdateType(ctx context.Context, id domain.CategoryID, t domain.CategoryType) error
	Delete(ctx context.Context, id domain.CategoryID) error
	HasOperations(ctx context.Context, id domain.CategoryID) (bool, error)
}
