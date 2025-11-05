package facade

import (
	"context"
	"errors"
	"strings"

	"main/domain"
)

// CategoryFacade инкапсулирует сценарии для категорий.
type CategoryFacade struct {
	F          domain.Factory
	Categories CategoryRepo // <-- интерфейс вместо *repo.PgCategoryRepo
}

// Создать категорию (если с таким именем уже есть — вернём ошибку)
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

// Переименовать категорию
func (f CategoryFacade) Rename(ctx context.Context, id domain.CategoryID, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return domain.ErrEmptyCategoryName
	}
	return f.Categories.UpdateName(ctx, id, newName)
}

// Сменить тип категории
func (f CategoryFacade) ChangeType(ctx context.Context, id domain.CategoryID, t domain.CategoryType) error {
	if t != domain.CatIncome && t != domain.CatExpense {
		return domain.ErrInvalidCategoryType
	}
	return f.Categories.UpdateType(ctx, id, t)
}

// Список категорий
func (f CategoryFacade) List(ctx context.Context) ([]domain.Category, error) {
	return f.Categories.List(ctx)
}
