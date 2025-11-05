package files

import (
	"context"
	"os"
	"time"

	"main/domain"
	"main/repo"
)

// Encoder — стратегия кодирования набора строк в байты (CSV/JSON/YAML и т.д.)
type Encoder interface {
	EncodeRows(rows []Row) ([]byte, error)
}

// ExportOperations — общий каркас экспорта (Strategy).
// 1) достаём операции из БД
// 2) преобразуем в []Row с категориями по именам
// 3) кодируем стратегией enc
// 4) пишем на диск
func ExportOperations(
	ctx context.Context,
	ops *repo.PgOperationRepo,
	cats *repo.PgCategoryRepo,
	accID domain.AccountID,
	from, to time.Time,
	path string,
	enc Encoder,
) error {
	list, err := ops.ListByAccount(ctx, accID, from, to)
	if err != nil {
		return err
	}

	// кэш имён категорий
	cmap := map[domain.CategoryID]string{}
	getCatName := func(id domain.CategoryID) string {
		if n, ok := cmap[id]; ok {
			return n
		}
		c, err := cats.Get(ctx, id)
		if err != nil {
			return ""
		}
		cmap[id] = c.Name
		return c.Name
	}

	// доменные операции -> переносимые строки
	rows := make([]Row, 0, len(list))
	for _, o := range list {
		t := 1
		if o.IsExpense() {
			t = -1
		}
		rows = append(rows, Row{
			Type:        t,
			Amount:      o.Amount, // форматирование берёт на себя Encoder
			Date:        o.Date,   // форматирование берёт на себя Encoder
			Category:    getCatName(o.Category),
			Description: o.Description,
		})
	}

	b, err := enc.EncodeRows(rows)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
