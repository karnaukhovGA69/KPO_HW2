package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrEmptyOperationID     = errors.New("operation id is empty")
	ErrInvalidOperationType = errors.New("invalid operation type")
	ErrEmptyAccountRef      = errors.New("bank account id is empty")
	ErrEmptyCategoryRef     = errors.New("category id is empty")
	ErrNonPositiveOpAmt     = errors.New("operation amount must be > 0")
	ErrZeroDate             = errors.New("operation date is zero")
)

type Operation struct {
	ID          OperationID     `json:"id"              yaml:"id"`
	Type        OperationType   `json:"type"            yaml:"type"`
	BankAccount AccountID       `json:"bank_account_id" yaml:"bank_account_id"`
	Amount      decimal.Decimal `json:"amount"          yaml:"amount"`
	Date        time.Time       `json:"date"            yaml:"date"`
	Description string          `json:"description"     yaml:"description"`
	Category    CategoryID      `json:"category_id"     yaml:"category_id"`
}

func (o Operation) Validate() error {
	if strings.TrimSpace(string(o.ID)) == "" {
		return ErrEmptyOperationID
	}
	if o.Type != CatIncome && o.Type != CatExpense {
		return ErrInvalidOperationType
	}
	if strings.TrimSpace(string(o.BankAccount)) == "" {
		return ErrEmptyAccountRef
	}
	if strings.TrimSpace(string(o.Category)) == "" {
		return ErrEmptyCategoryRef
	}
	if o.Date.IsZero() {
		return ErrZeroDate
	}
	if !o.Amount.GreaterThan(decimal.Zero) {
		return ErrNonPositiveOpAmt
	}
	return nil
}

func (o Operation) IsIncome() bool  { return o.Type == CatIncome }
func (o Operation) IsExpense() bool { return o.Type == CatExpense }
func (o Operation) Sign() int {
	if o.IsExpense() {
		return -1
	}
	return 1
}
