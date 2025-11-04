package di

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/dig"

	"main/db"
	"main/domain"
	"main/menu"
	"main/repo"
	"main/service"
)

type App struct {
	Menu menu.Menu
	Deps menu.Deps
	Pool *pgxpool.Pool
}

// Build собирает DI-контейнер и провайдеры.
func Build(ctx context.Context) (*dig.Container, error) {
	c := dig.New()

	// базовые провайдеры
	c.Provide(func() context.Context { return ctx })
	c.Provide(db.Connect) // func(ctx) (*pgxpool.Pool, error)
	c.Provide(func() domain.Factory { return domain.Factory{} })
	c.Provide(func(p *pgxpool.Pool) service.TxStarter { return p })
	// репозитории
	c.Provide(repo.NewPgAccountRepo)
	c.Provide(repo.NewPgCategoryRepo)
	c.Provide(repo.NewPgOperationRepo)

	// сервисы
	c.Provide(service.NewOperationService) // требует pool + Factory
	c.Provide(func(ops *repo.PgOperationRepo) *service.AnalyticsService {
		return service.NewAnalyticsService(ops)
	})

	// меню + сборка App
	c.Provide(loadMenu) // menu.Load("menu/menu.json")
	c.Provide(newApp)   // склеиваем всё в App

	return c, nil
}

func loadMenu() (menu.Menu, error) {
	return menu.Load("menu/menu.json")
}

// newApp: выбираем/создаём активный счёт, собираем Deps для меню и печатаем заголовок.
func newApp(
	ctx context.Context,
	pool *pgxpool.Pool,
	f domain.Factory,
	accounts *repo.PgAccountRepo,
	cats *repo.PgCategoryRepo,
	ops *repo.PgOperationRepo,
	m menu.Menu,
) (*App, error) {
	id, name, err := ensureActiveAccount(ctx, accounts, f)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Активный счёт: %s (%s)\n\n", name, id)

	deps := menu.Deps{
		Pool:      pool,
		Factory:   f,
		AccRepo:   accounts,
		CatRepo:   cats,
		OpsRepo:   ops,
		AccountID: id,
	}
	return &App{Menu: m, Deps: deps, Pool: pool}, nil
}
