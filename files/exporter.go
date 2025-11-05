package files

import (
	"context"
	"os"
	"time"

	"main/domain"
	"main/repo"
)

type Encoder interface {
	EncodeRows(rows []Row) ([]byte, error)
}

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

	rows := make([]Row, 0, len(list))
	for _, o := range list {
		t := 1
		if o.IsExpense() {
			t = -1
		}
		rows = append(rows, Row{
			Type:        t,
			Amount:      o.Amount,
			Date:        o.Date,
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
