package domain

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	ErrEmptyAccountID         = errors.New("account id is empty")
	ErrEmptyAccountName       = errors.New("account name is empty")
	ErrNonPositiveAmt         = errors.New("amount must be > 0")
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrNegativeInitialBalance = errors.New("initial balance must be >= 0")
)

type BankAccount struct {
	ID      AccountID         `json:"id"   yaml:"id"`
	Name    string          `json:"name" yaml:"name"`
	Balance decimal.Decimal `json:"balance" yaml:"balance"`
}

func (a BankAccount) Validate() error {
	if strings.TrimSpace(string(a.ID)) == "" {
		return ErrEmptyAccountID
	}
	if strings.TrimSpace(a.Name) == "" {
		return ErrEmptyAccountName
	}
	if a.Balance.IsNegative() {
		return ErrNegativeInitialBalance
	}
	return nil
}

func (a BankAccount) CanDebit(amount decimal.Decimal) bool {
	amt, ok := normalizeMoney(amount)
	if !ok {
		return false
	}
	return a.Balance.GreaterThanOrEqual(amt)
}

func (a *BankAccount) Rename(name string) error {
	if a == nil {
		return errors.New("nil receiver: BankAccount")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyAccountName
	}
	a.Name = name
	return nil
}

func (a *BankAccount) Credit(amount decimal.Decimal) error {
	if a == nil {
		return errors.New("nil receiver: BankAccount")
	}
	amt, ok := normalizeMoney(amount)
	if !ok {
		return ErrNonPositiveAmt
	}
	a.Balance = a.Balance.Add(amt).Round(2)
	return nil
}

func (a *BankAccount) Debit(amount decimal.Decimal) error {
	if a == nil {
		return errors.New("nil receiver: BankAccount")
	}
	amt, ok := normalizeMoney(amount)
	if !ok {
		return ErrNonPositiveAmt
	}
	if a.Balance.LessThan(amt) {
		return ErrInsufficientFunds
	}
	a.Balance = a.Balance.Sub(amt).Round(2)
	return nil
}

func normalizeMoney(v decimal.Decimal) (decimal.Decimal, bool) {
	if !v.GreaterThan(decimal.Zero) {
		return decimal.Zero, false
	}
	return v.Round(2), true
}
