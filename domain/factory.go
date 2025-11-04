package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Factory struct{}

func (_ Factory) NewBankAccount(name string) (BankAccount, error) {
	a := BankAccount{
		ID:      AccountID(uuid.NewString()),
		Name:    strings.TrimSpace(name),
		Balance: decimal.Zero,
	}
	return a, a.Validate()
}

func (_ Factory) NewCategory(name string, t CategoryType) (Category, error) {
	c := Category{
		ID:   CategoryID(uuid.NewString()),
		Type: t,
		Name: strings.TrimSpace(name),
	}
	return c, c.Validate()
}

func (_ Factory) NewOperation(
	t OperationType,
	accountID AccountID,
	amount decimal.Decimal,
	when time.Time,
	categoryID CategoryID,
	desc string,
) (Operation, error) {
	op := Operation{
		ID:          OperationID(uuid.NewString()),
		Type:        t,
		BankAccount: accountID,
		Amount:      amount.Round(2),
		Date:        when,
		Description: strings.TrimSpace(desc),
		Category:    categoryID,
	}
	return op, op.Validate()
}
