package facade

import (
	"context"
	"strings"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

type AccountFacade struct {
	F          domain.Factory
	Accounts   *repo.PgAccountRepo
	Operations *repo.PgOperationRepo
}

func (f AccountFacade) Create(ctx context.Context, name string) (domain.BankAccount, error) {
	acc, err := f.F.NewBankAccount(strings.TrimSpace(name))
	if err != nil {
		return domain.BankAccount{}, err
	}
	if err := f.Accounts.Create(ctx, acc); err != nil {
		return domain.BankAccount{}, err
	}
	return acc, nil
}

func (f AccountFacade) Rename(ctx context.Context, id domain.AccountID, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return domain.ErrEmptyAccountName
	}
	return f.Accounts.UpdateName(ctx, id, newName)
}

func (f AccountFacade) RecalculateBalance(ctx context.Context, id domain.AccountID) (decimal.Decimal, decimal.Decimal, error) {
	acc, err := f.Accounts.Get(ctx, id)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	from := time.Unix(0, 0)
	to := time.Now().AddDate(100, 0, 0)

	ops, err := f.Operations.ListByAccount(ctx, id, from, to)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	computed := decimal.Zero
	for _, op := range ops {
		if op.IsIncome() {
			computed = computed.Add(op.Amount)
		} else if op.IsExpense() {
			computed = computed.Sub(op.Amount)
		}
	}

	oldBal := acc.Balance
	delta := computed.Sub(acc.Balance)

	if delta.GreaterThan(decimal.Zero) {
		if err := acc.Credit(delta); err != nil {
			return oldBal, acc.Balance, err
		}
	} else if delta.LessThan(decimal.Zero) {
		if err := acc.Debit(delta.Abs()); err != nil {
			return oldBal, acc.Balance, err
		}
	}

	if err := f.Accounts.Update(ctx, acc); err != nil {
		return oldBal, acc.Balance, err
	}
	return oldBal, acc.Balance, nil
}
