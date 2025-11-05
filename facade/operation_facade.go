package facade

import (
	"context"
	"errors"
	"strings"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

// Входы для сценариев
type AddOpInput struct {
	AccountID    domain.AccountID
	Amount       decimal.Decimal
	When         time.Time
	CategoryName string
	Description  string
}
type EditOpInput struct {
	OperationID domain.OperationID
	NewAmount   *decimal.Decimal
	NewWhen     *time.Time
	NewCategory *string // имя категории
	NewDesc     *string
	// Если меняем тип операции, фасаду подсказывают ожидаемый тип категории
	ForcedType *domain.CategoryType
}

type OperationFacade struct {
	F          domain.Factory
	Accounts   *repo.PgAccountRepo
	Categories CategoryRepo
	Operations *repo.PgOperationRepo
}

func (f OperationFacade) AddIncome(ctx context.Context, in AddOpInput) (domain.Operation, error) {
	return f.add(ctx, domain.OpIncome, in)
}
func (f OperationFacade) AddExpense(ctx context.Context, in AddOpInput) (domain.Operation, error) {
	return f.add(ctx, domain.OpExpense, in)
}

func (f OperationFacade) add(ctx context.Context, t domain.OperationType, in AddOpInput) (domain.Operation, error) {
	if strings.TrimSpace(in.CategoryName) == "" {
		return domain.Operation{}, errors.New("category is required")
	}
	// найти / создать категорию нужного типа по имени
	cats, err := f.Categories.List(ctx)
	if err != nil {
		return domain.Operation{}, err
	}
	var catID domain.CategoryID
	for _, c := range cats {
		if strings.EqualFold(c.Name, in.CategoryName) {
			catID = c.ID
			break
		}
	}
	if catID == "" {
		var ct domain.CategoryType
		if t == domain.OpIncome {
			ct = domain.CatIncome
		} else {
			ct = domain.CatExpense
		}
		cat, err := f.F.NewCategory(in.CategoryName, ct)
		if err != nil {
			return domain.Operation{}, err
		}
		if err := f.Categories.Create(ctx, cat); err != nil {
			return domain.Operation{}, err
		}
		catID = cat.ID
	}

	op, err := f.F.NewOperation(t, in.AccountID, in.Amount, in.When, catID, in.Description)
	if err != nil {
		return domain.Operation{}, err
	}

	acc, err := f.Accounts.Get(ctx, in.AccountID)
	if err != nil {
		return domain.Operation{}, err
	}
	switch t {
	case domain.OpIncome:
		if err := acc.Credit(in.Amount); err != nil {
			return domain.Operation{}, err
		}
	case domain.OpExpense:
		if err := acc.Debit(in.Amount); err != nil {
			return domain.Operation{}, err
		}
	default:
		return domain.Operation{}, errors.New("unknown operation type")
	}
	if err := f.Operations.Create(ctx, op); err != nil {
		return domain.Operation{}, err
	}
	if err := f.Accounts.Update(ctx, acc); err != nil {
		return domain.Operation{}, err
	}
	return op, nil
}

func (f OperationFacade) Edit(ctx context.Context, in EditOpInput) (domain.Operation, error) {
	old, err := f.Operations.Get(ctx, in.OperationID)
	if err != nil {
		return domain.Operation{}, err
	}

	newOp := old

	// Сумма
	if in.NewAmount != nil {
		newOp.Amount = *in.NewAmount
	}
	// Дата
	if in.NewWhen != nil {
		newOp.Date = *in.NewWhen
	}
	// Категория по имени
	if in.NewCategory != nil && strings.TrimSpace(*in.NewCategory) != "" {
		cats, err := f.Categories.List(ctx)
		if err != nil {
			return domain.Operation{}, err
		}
		var foundID domain.CategoryID
		for _, c := range cats {
			if strings.EqualFold(c.Name, *in.NewCategory) {
				// если навязывают тип — проверим соответствие
				if in.ForcedType == nil || c.Type == *in.ForcedType {
					foundID = c.ID
					break
				}
			}
		}
		if foundID == "" {
			// если ForcedType не задан — определяем по типу СТАРОЙ операции
			ct := domain.CatExpense
			if old.IsIncome() {
				ct = domain.CatIncome
			}
			if in.ForcedType != nil {
				ct = *in.ForcedType
			}
			cat, err := f.F.NewCategory(*in.NewCategory, ct)
			if err != nil {
				return domain.Operation{}, err
			}
			if err := f.Categories.Create(ctx, cat); err != nil {
				return domain.Operation{}, err
			}
			foundID = cat.ID
		}
		newOp.Category = foundID
	}
	// Описание
	if in.NewDesc != nil {
		newOp.Description = *in.NewDesc
	}

	// Баланс при изменении суммы
	if !newOp.Amount.Equal(old.Amount) {
		acc, err := f.Accounts.Get(ctx, old.BankAccount)
		if err != nil {
			return domain.Operation{}, err
		}
		diff := newOp.Amount.Sub(old.Amount)
		if old.IsIncome() {
			// доход стал больше/меньше
			if diff.GreaterThan(decimal.Zero) {
				if err := acc.Credit(diff); err != nil {
					return domain.Operation{}, err
				}
			} else if diff.LessThan(decimal.Zero) {
				if err := acc.Debit(diff.Abs()); err != nil {
					return domain.Operation{}, err
				}
			}
		} else { // расход
			if diff.GreaterThan(decimal.Zero) {
				if err := acc.Debit(diff); err != nil {
					return domain.Operation{}, err
				}
			} else if diff.LessThan(decimal.Zero) {
				if err := acc.Credit(diff.Abs()); err != nil {
					return domain.Operation{}, err
				}
			}
		}
		if err := f.Accounts.Update(ctx, acc); err != nil {
			return domain.Operation{}, err
		}
	}

	// NB: если у тебя есть Update в PgOperationRepo — тут его вызывай.
	// Иначе оставляем newOp как результат без сохранения.
	return newOp, nil
}

func (f OperationFacade) Delete(ctx context.Context, id domain.OperationID) error {
	// Реализация зависит от методов твоего репозитория.
	// Если в PgOperationRepo есть Delete(id) — здесь:
	// 1) r.Operations.Get(ctx, id)
	// 2) откатить баланс (обратно)
	// 3) r.Operations.Delete(ctx, id)
	return errors.New("Delete not implemented in OperationFacade (requires repo method)")
}
