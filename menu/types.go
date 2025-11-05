package menu

import (
	"main/domain"
	"main/facade"
	"main/repo"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	Key   string `json:"key"`   // строковый ключ действия
	Field string `json:"field"` // текст для вывода
}

type Menu struct {
	Items []Item
}
type Deps struct {
	Pool      *pgxpool.Pool
	Factory   domain.Factory
	AccRepo   *repo.PgAccountRepo
	CatRepo   *repo.PgCategoryRepo
	OpsRepo   *repo.PgOperationRepo
	AccountID domain.AccountID

	Op  facade.OperationFacade
	Acc facade.AccountFacade
	Cat facade.CategoryFacade
	Ana facade.AnalyticsFacade
}
