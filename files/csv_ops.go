package files

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

// =======================
// ====== ЭКСПОРТ ========
// =======================

// CSVEncoder — стратегия кодирования в CSV.
type CSVEncoder struct{}

func (CSVEncoder) EncodeRows(rows []Row) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// заголовок
	if err := w.Write([]string{"type", "amount", "date", "category", "description"}); err != nil {
		return nil, err
	}

	// строки
	for _, r := range rows {
		rec := []string{
			strconv.Itoa(r.Type),
			r.Amount.StringFixed(2),
			r.Date.Format("2006-01-02"),
			r.Category,
			r.Description,
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Публичная точка — сигнатура НЕ менялась.
// Теперь внутри используем общий каркас ExportOperations + стратегию CSVEncoder.
func ExportOperationsCSV(
	ctx context.Context,
	ops *repo.PgOperationRepo,
	cats *repo.PgCategoryRepo,
	accID domain.AccountID,
	from, to time.Time,
	path string,
) error {
	return ExportOperations(ctx, ops, cats, accID, from, to, path, CSVEncoder{})
}

// =======================
// ====== ИМПОРТ =========
// =======================

// Template Method: CSVImporter реализует parse(), общий каркас — в BaseImporter.
type CSVImporter struct{}

func (CSVImporter) parse(data []byte) ([]Row, error) {
	r := csv.NewReader(bytes.NewReader(data))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, nil
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

// Публичная точка импорта — вызывает общий каркас.
func ImportOperationsCSV(path string) ([]Row, error) {
	base := BaseImporter{parser: CSVImporter{}}
	return base.Import(path)
}
