package service

import (
	"context"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

type AnalyticsService struct {
	ops *repo.PgOperationRepo
}

func NewAnalyticsService(ops *repo.PgOperationRepo) *AnalyticsService {
	return &AnalyticsService{ops: ops}
}

type Summary struct {
	Income  decimal.Decimal
	Expense decimal.Decimal
	Net     decimal.Decimal // Income - Expense
}

// SummaryByPeriod — агрегирует операции счета за период [from; to].
func (s *AnalyticsService) SummaryByPeriod(ctx context.Context, accID domain.AccountID, from, to time.Time) (Summary, error) {
	list, err := s.ops.ListByAccount(ctx, accID, from, to)
	if err != nil {
		return Summary{}, err
	}
	income := decimal.Zero
	expense := decimal.Zero
	for _, o := range list {
		amt := o.Amount.Round(2)
		if o.IsIncome() {
			income = income.Add(amt)
		} else {
			expense = expense.Add(amt)
		}
	}
	return Summary{
		Income:  income.Round(2),
		Expense: expense.Round(2),
		Net:     income.Sub(expense).Round(2),
	}, nil
}
