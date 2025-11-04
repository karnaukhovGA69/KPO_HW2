package files

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

// DTO для JSON (строки для денег/даты)
type opRowJSON struct {
	Type        int    `json:"type"`        // -1/1
	Amount      string `json:"amount"`      // "123.45"
	Date        string `json:"date"`        // "YYYY-MM-DD"
	Category    string `json:"category"`    // имя категории
	Description string `json:"description"` // опционально
}

// ExportOperationsJSON — выгружает операции счета за период в JSON.
func ExportOperationsJSON(ctx context.Context, ops *repo.PgOperationRepo, cats *repo.PgCategoryRepo,
	accID domain.AccountID, from, to time.Time, path string) error {

	list, err := ops.ListByAccount(ctx, accID, from, to)
	if err != nil {
		return err
	}

	// маленький кэш id->name
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

	out := make([]opRowJSON, 0, len(list))
	for _, o := range list {
		t := 1
		if o.IsExpense() {
			t = -1
		}
		out = append(out, opRowJSON{
			Type:        t,
			Amount:      o.Amount.StringFixed(2),
			Date:        o.Date.Format("2006-01-02"),
			Category:    getCatName(o.Category),
			Description: o.Description,
		})
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// ImportOperationsJSON — читает JSON и возвращает универсальные Row (как из CSV).
func ImportOperationsJSON(path string) ([]Row, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var in []opRowJSON
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, err
	}

	out := make([]Row, 0, len(in))
	for _, r := range in {
		amt, err := decimal.NewFromString(r.Amount)
		if err != nil {
			continue
		}
		dt, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			continue
		}
		out = append(out, Row{
			Type:        r.Type,
			Amount:      amt.Round(2),
			Date:        dt,
			Category:    r.Category,
			Description: r.Description,
		})
	}
	return out, nil
}
