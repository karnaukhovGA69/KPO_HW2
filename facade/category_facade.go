package facade

import (
	"context"
	"errors"
	"strings"

	"main/domain"
)

type CategoryFacade struct {
	F          domain.Factory
	Categories CategoryRepo
}

func (f CategoryFacade) Create(ctx context.Context, name string, t domain.CategoryType) (domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Category{}, domain.ErrEmptyCategoryName
	}

	all, err := f.Categories.List(ctx)
	if err != nil {
		return domain.Category{}, err
	}
	for _, c := range all {
		if strings.EqualFold(c.Name, name) {
			return domain.Category{}, errors.New("category with this name already exists")
		}
	}

	c, err := f.F.NewCategory(name, t)
	if err != nil {
		return domain.Category{}, err
	}
	if err := f.Categories.Create(ctx, c); err != nil {
		return domain.Category{}, err
	}
	return c, nil
}

func (f CategoryFacade) Rename(ctx context.Context, id domain.CategoryID, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return domain.ErrEmptyCategoryName
	}
	return f.Categories.UpdateName(ctx, id, newName)
}

func (f CategoryFacade) ChangeType(ctx context.Context, id domain.CategoryID, t domain.CategoryType) error {
	if t != domain.CatIncome && t != domain.CatExpense {
		return domain.ErrInvalidCategoryType
	}
	return f.Categories.UpdateType(ctx, id, t)
}

func (f CategoryFacade) List(ctx context.Context) ([]domain.Category, error) {
	return f.Categories.List(ctx)
}
