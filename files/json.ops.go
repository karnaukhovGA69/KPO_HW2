package files

import (
	"context"
	"encoding/json"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

type opRowJSON struct {
	Type        int    `json:"type"`        // -1/1
	Amount      string `json:"amount"`      // "123.45"
	Date        string `json:"date"`        // "YYYY-MM-DD"
	Category    string `json:"category"`    // имя категории
	Description string `json:"description"` // опционально
}

type JSONEncoder struct{}

func (JSONEncoder) EncodeRows(rows []Row) ([]byte, error) {
	out := make([]opRowJSON, 0, len(rows))
	for _, r := range rows {
		out = append(out, opRowJSON{
			Type:        r.Type,
			Amount:      r.Amount.StringFixed(2),
			Date:        r.Date.Format("2006-01-02"),
			Category:    r.Category,
			Description: r.Description,
		})
	}
	return json.MarshalIndent(out, "", "  ")
}

func ExportOperationsJSON(
	ctx context.Context,
	ops *repo.PgOperationRepo,
	cats *repo.PgCategoryRepo,
	accID domain.AccountID,
	from, to time.Time,
	path string,
) error {
	return ExportOperations(ctx, ops, cats, accID, from, to, path, JSONEncoder{})
}

type JSONImporter struct{}

func (JSONImporter) parse(data []byte) ([]Row, error) {
	var in []opRowJSON
	if err := json.Unmarshal(data, &in); err != nil {
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

func ImportOperationsJSON(path string) ([]Row, error) {
	base := BaseImporter{parser: JSONImporter{}}
	return base.Import(path)
}
