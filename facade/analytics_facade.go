package facade

import (
	"context"
	"sort"
	"time"

	"main/domain"
	"main/service"

	"github.com/shopspring/decimal"
)

type AnalyticsFacade struct {
	Svc *service.AnalyticsService
}

type FlowSummary struct {
	Income  decimal.Decimal
	Expense decimal.Decimal
	Net     decimal.Decimal
}

func (a AnalyticsFacade) Summary(ctx context.Context, acc domain.AccountID, from, to time.Time) (FlowSummary, error) {
	s, err := a.Svc.SummaryByPeriod(ctx, acc, from, to)
	if err != nil {
		return FlowSummary{}, err
	}
	return FlowSummary{
		Income:  s.Income,
		Expense: s.Expense,
		Net:     s.Net,
	}, nil
}

type Breakdown struct {
	Incomes  []CatSum
	Expenses []CatSum
}
type CatSum struct {
	Category string
	Amount   decimal.Decimal
}

func (a AnalyticsFacade) BreakdownByCategory(ctx context.Context, acc domain.AccountID, from, to time.Time) (Breakdown, error) {
	rows, err := a.Svc.ByCategory(ctx, acc, from, to)
	if err != nil {
		return Breakdown{}, err
	}
	var out Breakdown
	for _, r := range rows {
		switch r.Type {
		case domain.CatIncome:
			if !r.Income.IsZero() {
				out.Incomes = append(out.Incomes, CatSum{Category: r.Name, Amount: r.Income})
			}
		case domain.CatExpense:
			if !r.Expense.IsZero() {
				out.Expenses = append(out.Expenses, CatSum{Category: r.Name, Amount: r.Expense})
			}
		}
	}
	sort.Slice(out.Incomes, func(i, j int) bool { return out.Incomes[i].Amount.GreaterThan(out.Incomes[j].Amount) })
	sort.Slice(out.Expenses, func(i, j int) bool { return out.Expenses[i].Amount.GreaterThan(out.Expenses[j].Amount) })
	return out, nil
}
