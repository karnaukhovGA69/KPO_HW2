package files

import (
	"bytes"
	"context"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// =======================
// ====== ЭКСПОРТ ========
// =======================

type opRowYAML struct {
	Type        int    `yaml:"type"`
	Amount      string `yaml:"amount"`
	Date        string `yaml:"date"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
}

// YAMLEncoder — стратегия кодирования в YAML.
type YAMLEncoder struct{}

func (YAMLEncoder) EncodeRows(rows []Row) ([]byte, error) {
	out := make([]opRowYAML, 0, len(rows))
	for _, r := range rows {
		out = append(out, opRowYAML{
			Type:        r.Type,
			Amount:      r.Amount.StringFixed(2),
			Date:        r.Date.Format("2006-01-02"),
			Category:    r.Category,
			Description: r.Description,
		})
	}
	return yaml.Marshal(out)
}

// Публичная точка — сигнатура НЕ менялась.
func ExportOperationsYAML(
	ctx context.Context,
	ops *repo.PgOperationRepo,
	cats *repo.PgCategoryRepo,
	accID domain.AccountID,
	from, to time.Time,
	path string,
) error {
	return ExportOperations(ctx, ops, cats, accID, from, to, path, YAMLEncoder{})
}

// =======================
// ====== ИМПОРТ =========
// =======================

type YAMLImporter struct{}

func (YAMLImporter) parse(data []byte) ([]Row, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var in []opRowYAML
	if err := dec.Decode(&in); err != nil {
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

func ImportOperationsYAML(path string) ([]Row, error) {
	base := BaseImporter{parser: YAMLImporter{}}
	return base.Import(path)
}
