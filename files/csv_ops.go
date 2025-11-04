package files

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

// ExportOperationsCSV — выгружает операции счета за период в CSV.
// Формат: type,amount,date,category,description
func ExportOperationsCSV(
	ctx context.Context,
	ops *repo.PgOperationRepo,
	cats *repo.PgCategoryRepo,
	accID domain.AccountID,
	from, to time.Time,
	path string,
) error {
	list, err := ops.ListByAccount(ctx, accID, from, to)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// хедер
	if err := w.Write([]string{"type", "amount", "date", "category", "description"}); err != nil {
		return err
	}

	// кэш id->name категорий
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

	for _, o := range list {
		t := 1
		if o.IsExpense() {
			t = -1
		}
		rec := []string{
			strconv.Itoa(t),
			o.Amount.StringFixed(2),
			o.Date.Format("2006-01-02"),
			getCatName(o.Category),
			o.Description,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	return w.Error()
}

// ImportOperationsCSV — читает CSV в универсальные строки Row.
// Преобразование Row → доменные операции делается в menu.Execute.
func ImportOperationsCSV(path string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, nil // только хедер
	}

	out := make([]Row, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		rec := rows[i]
		if len(rec) < 5 {
			continue
		}

		t, err := strconv.Atoi(rec[0])
		if err != nil {
			continue
		}
		amt, err := decimal.NewFromString(rec[1])
		if err != nil {
			continue
		}
		dt, err := time.Parse("2006-01-02", rec[2])
		if err != nil {
			continue
		}
		out = append(out, Row{
			Type:        t,
			Amount:      amt.Round(2),
			Date:        dt,
			Category:    rec[3],
			Description: rec[4],
		})
	}
	return out, nil
}
