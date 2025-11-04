package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"main/domain"
)

type OperationService struct {
	db *pgxpool.Pool
	f  domain.Factory
}

func NewOperationService(db *pgxpool.Pool, f domain.Factory) *OperationService {
	return &OperationService{db: db, f: f}
}

func (s *OperationService) ApplyOperation(
	ctx context.Context,
	t domain.OperationType,
	accountID domain.AccountID, // счёт
	amount decimal.Decimal, // > 0
	when time.Time, // дата операции
	categoryID domain.CategoryID,
	desc string,
) (domain.Operation, error) {

	// 0) Сконструировать операцию (валидация суммы/дат/ID)
	op, err := s.f.NewOperation(t, accountID, amount, when, categoryID, desc)
	if err != nil {
		return domain.Operation{}, err
	}

	// 1) Транзакция
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Operation{}, err
	}
	defer tx.Rollback(ctx)

	// 2) Прочитать текущий баланс с блокировкой строки
	var balStr string
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, accountID).Scan(&balStr); err != nil {
		return domain.Operation{}, err
	}
	curBal, err := decimal.NewFromString(balStr)
	if err != nil {
		return domain.Operation{}, err
	}

	// 3) Применить доменную логику к балансу (жёсткий запрет на овердрафт)
	acc := domain.BankAccount{ID: accountID, Balance: curBal}
	if t == domain.OpIncome {
		if err := acc.Credit(op.Amount); err != nil {
			return domain.Operation{}, err
		}
	} else {
		if err := acc.Debit(op.Amount); err != nil { // вернёт ErrInsufficientFunds, если денег мало
			return domain.Operation{}, err
		}
	}

	// 4) Вставить операцию
	if _, err := tx.Exec(ctx,
		`INSERT INTO operations(id,type,bank_account_id,amount,"date",description,category_id)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		op.ID, int(op.Type), op.BankAccount, op.Amount.StringFixed(2), op.Date, op.Description, op.Category,
	); err != nil {
		return domain.Operation{}, err
	}

	// 5) Обновить баланс счёта
	if _, err := tx.Exec(ctx,
		`UPDATE accounts SET balance=$2 WHERE id=$1`,
		accountID, acc.Balance.StringFixed(2),
	); err != nil {
		return domain.Operation{}, err
	}

	// 6) Коммит
	if err := tx.Commit(ctx); err != nil {
		return domain.Operation{}, err
	}
	return op, nil
}
