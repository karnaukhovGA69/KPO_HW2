package menu

import (
	"main/domain"
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
	Pool    *pgxpool.Pool  // подключение к PG (pool)
	Factory domain.Factory // доменная фабрика

	AccRepo *repo.PgAccountRepo
	CatRepo *repo.PgCategoryRepo
	OpsRepo *repo.PgOperationRepo

	AccountID domain.AccountID // активный счёт
}

// KeyAt — вернуть ключ по номеру пункта (1..N), "" если мимо.
func (m Menu) KeyAt(index int) string {
	if index < 1 || index > len(m.Items) {
		return ""
	}
	return m.Items[index-1].Key
}
