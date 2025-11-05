// di/di_run.go
package di

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/dig"

	"main/db"
	"main/domain"
	"main/facade"
	"main/menu"
	"main/repo"
	"main/service"
)

type App struct {
	Menu menu.Menu
	Deps menu.Deps
	Pool *pgxpool.Pool
}

func Build(ctx context.Context) (*App, error) {
	c := dig.New()

	// --- базовые зависимости
	if err := c.Provide(func() context.Context { return ctx }); err != nil {
		return nil, err
	}
	if err := c.Provide(db.Connect); err != nil { // *pgxpool.Pool
		return nil, err
	}
	if err := c.Provide(func() domain.Factory { return domain.Factory{} }); err != nil {
		return nil, err
	}

	// --- репозитории
	if err := c.Provide(repo.NewPgAccountRepo); err != nil {
		return nil, err
	}
	if err := c.Provide(repo.NewPgCategoryRepo); err != nil {
		return nil, err
	}
	if err := c.Provide(repo.NewPgOperationRepo); err != nil {
		return nil, err
	}

	// --- сервисы (если используются где-то ещё)
	if err := c.Provide(service.NewOperationService); err != nil {
		return nil, err
	}
	if err := c.Provide(service.NewAnalyticsService); err != nil {
		return nil, err
	}

	// --- путь к меню для menu.Load
	if err := c.Provide(func() string {
		if p := os.Getenv("MENU_PATH"); p != "" {
			return p
		}
		return "menu/menu.json"
	}); err != nil {
		return nil, err
	}

	// --- само меню
	if err := c.Provide(menu.Load); err != nil {
		return nil, err
	}

	// --- сборка App
	var app *App
	err := c.Invoke(func(
		ctx context.Context,
		pool *pgxpool.Pool,
		f domain.Factory,
		m menu.Menu,
		accounts *repo.PgAccountRepo,
		cats *repo.PgCategoryRepo,
		ops *repo.PgOperationRepo,
	) error {
		id, name, err := ensureActiveAccount(ctx, accounts, f)
		if err != nil {
			return err
		}
		fmt.Printf("Активный счёт: %s (%s)\n\n", name, id)

		// Proxy-кэш для категорий
		catsCached := repo.NewCachedCategoryRepo(cats)

		// Фасады используют кэш
		accFacade := facade.AccountFacade{
			F:          f,
			Accounts:   accounts,
			Operations: ops,
		}
		catFacade := facade.CategoryFacade{
			F:          f,
			Categories: catsCached,
		}
		opFacade := facade.OperationFacade{
			F:          f,
			Accounts:   accounts,
			Categories: catsCached,
			Operations: ops,
		}
		analytics := facade.AnalyticsFacade{
			Operations: ops,
			Categories: catsCached,
		}

		// В Deps кладём оригинальный PgCategoryRepo, чтобы не ломать типы
		deps := menu.Deps{
			Pool:      pool,
			Factory:   f,
			AccRepo:   accounts,
			CatRepo:   cats, // <-- ВАЖНО: здесь НЕ catsCached
			OpsRepo:   ops,
			AccountID: id,

			Acc: accFacade,
			Cat: catFacade,
			Op:  opFacade,
			Ana: analytics,
		}
		app = &App{Menu: m, Deps: deps, Pool: pool}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return app, nil
}
