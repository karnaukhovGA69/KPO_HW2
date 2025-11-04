package files

import (
	"context"
	"os"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

type opRowYAML struct {
	Type        int    `yaml:"type"`
	Amount      string `yaml:"amount"`
	Date        string `yaml:"date"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
}

// ExportOperationsYAML — выгружает операции счета за период в YAML.
func ExportOperationsYAML(ctx context.Context, ops *repo.PgOperationRepo, cats *repo.PgCategoryRepo,
	accID domain.AccountID, from, to time.Time, path string) error {

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

	out := make([]opRowYAML, 0, len(list))
	for _, o := range list {
		t := 1
		if o.IsExpense() {
			t = -1
		}
		out = append(out, opRowYAML{
			Type:        t,
			Amount:      o.Amount.StringFixed(2),
			Date:        o.Date.Format("2006-01-02"),
			Category:    getCatName(o.Category),
			Description: o.Description,
		})
	}

	b, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// ImportOperationsYAML — читает YAML и возвращает универсальные Row.
func ImportOperationsYAML(path string) ([]Row, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var in []opRowYAML
	if err := yaml.Unmarshal(b, &in); err != nil {
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
