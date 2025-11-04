package files

import (
	"time"

	"github.com/shopspring/decimal"
)

// Row — универсальная запись операции для импорт/экспорт.
type Row struct {
	Type        int             `json:"type" yaml:"type"`         // -1 расход, 1 доход
	Amount      decimal.Decimal `json:"amount" yaml:"amount"`     // > 0, 2 знака
	Date        time.Time       `json:"date" yaml:"date"`         // YYYY-MM-DD
	Category    string          `json:"category" yaml:"category"` // имя категории
	Description string          `json:"description" yaml:"description"`
}
