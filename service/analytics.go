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

// CategorySummary — агрегат по категории.
type CategorySummary struct {
	Name    string
	Type    domain.CategoryType
	Income  decimal.Decimal
	Expense decimal.Decimal
	Net     decimal.Decimal // Income - Expense
}

// ByCategory — агрегирует операции счета по категориям за период [from; to].
func (s *AnalyticsService) ByCategory(ctx context.Context, accID domain.AccountID, from, to time.Time) ([]CategorySummary, error) {
	// Один SQL с JOIN на categories
	rows, err := s.ops.Db().Query(ctx, `
		SELECT c.name, c.type,
		       SUM(CASE WHEN o.type = 1  THEN o.amount ELSE 0 END) AS income,
		       SUM(CASE WHEN o.type = -1 THEN o.amount ELSE 0 END) AS expense
		  FROM operations o
		  JOIN categories c ON c.id = o.category_id
		 WHERE o.bank_account_id = $1 AND o."date" BETWEEN $2 AND $3
		 GROUP BY c.name, c.type
		 ORDER BY
		   -- крупные расходы вверх, затем по доходам
		   SUM(CASE WHEN o.type = -1 THEN o.amount ELSE 0 END) DESC,
		   SUM(CASE WHEN o.type =  1 THEN o.amount ELSE 0 END) DESC`,
		accID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CategorySummary{}
	for rows.Next() {
		var name string
		var t domain.CategoryType
		var incomeStr, expenseStr string
		if err := rows.Scan(&name, &t, &incomeStr, &expenseStr); err != nil {
			return nil, err
		}
		inc, err := decimal.NewFromString(incomeStr)
		if err != nil {
			return nil, err
		}
		exp, err := decimal.NewFromString(expenseStr)
		if err != nil {
			return nil, err
		}
		out = append(out, CategorySummary{
			Name:    name,
			Type:    t,
			Income:  inc.Round(2),
			Expense: exp.Round(2),
			Net:     inc.Sub(exp).Round(2),
		})
	}
	return out, rows.Err()
}
