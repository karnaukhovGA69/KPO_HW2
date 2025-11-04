package domain

type AccountID string
type CategoryID string
type OperationID string

type CategoryType int

const (
	CatExpense CategoryType = -1
	CatIncome  CategoryType = 1
)

type OperationType = CategoryType

const (
	OpExpense OperationType = CatExpense
	OpIncome  OperationType = CatIncome
)
