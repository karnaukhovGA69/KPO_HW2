package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"main/domain"
)

type OperationService struct {
	db TxStarter // <--- было *pgxpool.Pool
	f  domain.Factory
}

func NewOperationService(db TxStarter, f domain.Factory) *OperationService { // <--- принимали *pgxpool.Pool
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
func (s *OperationService) RemoveOperation(ctx context.Context, opID domain.OperationID) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1) прочитать операцию
	var t int
	var accID domain.AccountID
	var amtStr string
	err = tx.QueryRow(ctx,
		`SELECT type, bank_account_id, amount FROM operations WHERE id=$1`, opID).
		Scan(&t, &accID, &amtStr)
	if err != nil {
		return err
	}
	amt, err := decimal.NewFromString(amtStr)
	if err != nil {
		return err
	}

	// 2) заблокировать строку счёта и взять баланс
	var balStr string
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, accID).Scan(&balStr); err != nil {
		return err
	}
	curBal, err := decimal.NewFromString(balStr)
	if err != nil {
		return err
	}

	acc := domain.BankAccount{ID: accID, Balance: curBal}

	// 3) применить "обратный" эффект
	if domain.OperationType(t) == domain.OpIncome {
		// удаляем доход -> вычесть деньги
		if err := acc.Debit(amt); err != nil {
			return err
		} // запретить минус
	} else {
		// удаляем расход -> добавить деньги
		if err := acc.Credit(amt); err != nil {
			return err
		}
	}

	// 4) удалить операцию
	if _, err := tx.Exec(ctx, `DELETE FROM operations WHERE id=$1`, opID); err != nil {
		return err
	}

	// 5) обновить баланс счёта
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$2 WHERE id=$1`, accID, acc.Balance.StringFixed(2)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
func (s *OperationService) UpdateOperation(
	ctx context.Context,
	opID domain.OperationID,
	newType domain.OperationType,
	newAmount decimal.Decimal,
	newDate time.Time,
	newCategory domain.CategoryID,
	newDesc string,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1) прочитать старую операцию
	var oldType int
	var accID domain.AccountID
	var oldAmtStr string
	if err := tx.QueryRow(ctx,
		`SELECT type, bank_account_id, amount FROM operations WHERE id=$1`, opID,
	).Scan(&oldType, &accID, &oldAmtStr); err != nil {
		return err
	}
	oldAmt, err := decimal.NewFromString(oldAmtStr)
	if err != nil {
		return err
	}

	// 2) проверить, что категория совпадает по типу с новой операцией
	var catType int
	if err := tx.QueryRow(ctx, `SELECT type FROM categories WHERE id=$1`, newCategory).Scan(&catType); err != nil {
		return err
	}
	if catType != int(newType) {
		return fmt.Errorf("тип категории (%d) не совпадает с типом операции (%d)", catType, int(newType))
	}

	// 3) заблокировать строку счёта и взять баланс
	var balStr string
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, accID).Scan(&balStr); err != nil {
		return err
	}
	curBal, err := decimal.NewFromString(balStr)
	if err != nil {
		return err
	}
	acc := domain.BankAccount{ID: accID, Balance: curBal}

	// 4) снять эффект старой операции
	if domain.OperationType(oldType) == domain.OpIncome {
		if err := acc.Debit(oldAmt); err != nil {
			return err
		}
	} else {
		if err := acc.Credit(oldAmt); err != nil {
			return err
		}
	}

	// 5) применить новую операцию
	newAmount = newAmount.Round(2)
	if newType == domain.OpIncome {
		if err := acc.Credit(newAmount); err != nil {
			return err
		}
	} else {
		if err := acc.Debit(newAmount); err != nil {
			return err
		}
	}

	// 6) обновить операцию
	if _, err := tx.Exec(ctx,
		`UPDATE operations
		    SET type=$2, amount=$3, "date"=$4, description=$5, category_id=$6
		  WHERE id=$1`,
		opID, int(newType), newAmount.StringFixed(2), newDate, newDesc, newCategory,
	); err != nil {
		return err
	}

	// 7) записать новый баланс
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$2 WHERE id=$1`,
		accID, acc.Balance.StringFixed(2),
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
