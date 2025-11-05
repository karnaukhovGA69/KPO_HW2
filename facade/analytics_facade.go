package facade

import (
	"context"
	"sort"
	"time"

	"main/domain"
	"main/repo"

	"github.com/shopspring/decimal"
)

type AnalyticsFacade struct {
	Operations *repo.PgOperationRepo
	Categories CategoryRepo // <-- интерфейс
}

// Суммарные потоки
type FlowSummary struct {
	Income  decimal.Decimal
	Expense decimal.Decimal
	Net     decimal.Decimal
}

func (a AnalyticsFacade) Summary(ctx context.Context, acc domain.AccountID, from, to time.Time) (FlowSummary, error) {
	ops, err := a.Operations.ListByAccount(ctx, acc, from, to)
	if err != nil {
		return FlowSummary{}, err
	}
	var income, expense decimal.Decimal
	for _, op := range ops {
		if op.IsIncome() {
			income = income.Add(op.Amount)
		} else if op.IsExpense() {
			expense = expense.Sub(decimal.Zero).Add(op.Amount) // просто суммируем расход
		}
	}
	return FlowSummary{
		Income:  income,
		Expense: expense,
		Net:     income.Sub(expense),
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
	ops, err := a.Operations.ListByAccount(ctx, acc, from, to)
	if err != nil {
		return Breakdown{}, err
	}
	cats, err := a.Categories.List(ctx)
	if err != nil {
		return Breakdown{}, err
	}
	names := make(map[domain.CategoryID]string, len(cats))
	for _, c := range cats {
		names[c.ID] = c.Name
	}

	inc := map[string]decimal.Decimal{}
	exp := map[string]decimal.Decimal{}
	for _, o := range ops {
		name := names[o.Category]
		if name == "" {
			name = "(unknown)"
		}
		if o.IsIncome() {
			inc[name] = inc[name].Add(o.Amount)
		} else if o.IsExpense() {
			exp[name] = exp[name].Add(o.Amount)
		}
	}
	out := Breakdown{}
	for k, v := range inc {
		out.Incomes = append(out.Incomes, CatSum{Category: k, Amount: v})
	}
	for k, v := range exp {
		out.Expenses = append(out.Expenses, CatSum{Category: k, Amount: v})
	}
	sort.Slice(out.Incomes, func(i, j int) bool { return out.Incomes[i].Amount.GreaterThan(out.Incomes[j].Amount) })
	sort.Slice(out.Expenses, func(i, j int) bool { return out.Expenses[i].Amount.GreaterThan(out.Expenses[j].Amount) })
	return out, nil
}
