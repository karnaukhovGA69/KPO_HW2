package main

import (
	"context"
	"fmt"
	"os"

	"main/db"
	"main/domain"
	"main/menu"
	"main/repo"
)

func main() {
	ctx := context.Background()
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Println("ERROR: set DATABASE_URL")
		return
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	f := domain.Factory{}
	accRepo := repo.NewPgAccountRepo(pool)
	catRepo := repo.NewPgCategoryRepo(pool)
	opsRepo := repo.NewPgOperationRepo(pool)

	// Активный счёт: берём первый из БД или создаём "Основной"
	accID, accName, err := ensureActiveAccount(ctx, accRepo, f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Активный счёт: %s (%s)\n\n", accName, accID)

	// Грузим меню и запускаем цикл
	m, err := menu.Load("menu/menu.json")
	if err != nil {
		panic(err)
	}

	deps := menu.Deps{
		Pool:      pool,
		Factory:   f,
		AccRepo:   accRepo,
		CatRepo:   catRepo,
		OpsRepo:   opsRepo,
		AccountID: accID,
	}
	menu.Run(ctx, m, deps)
}

// ensureActiveAccount — берёт первый счёт из БД или создаёт "Основной".
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
