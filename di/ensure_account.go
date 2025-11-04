package di

import (
	"context"

	"main/domain"
	"main/repo"
)

func ensureActiveAccount(ctx context.Context, accRepo *repo.PgAccountRepo, f domain.Factory) (domain.AccountID, string, error) {
	accs, err := accRepo.List(ctx)
	if err == nil && len(accs) > 0 {
		return accs[0].ID, accs[0].Name, nil
	}
	acc, err := f.NewBankAccount("Основной")
	if err != nil {
		return "", "", err
	}
	if err := accRepo.Create(ctx, acc); err != nil {
		return "", "", err
	}
	return acc.ID, acc.Name, nil
}
