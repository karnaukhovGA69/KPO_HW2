package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"main/domain"
)

type PgOperationRepo struct{ db *pgxpool.Pool }

func NewPgOperationRepo(db *pgxpool.Pool) *PgOperationRepo { return &PgOperationRepo{db: db} }

// Create — вставка операции. FK требуют, чтобы account_id и category_id существовали.
func (r *PgOperationRepo) Create(ctx context.Context, o domain.Operation) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO operations(id,type,bank_account_id,amount,"date",description,category_id)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		o.ID, int(o.Type), o.BankAccount, o.Amount.StringFixed(2), o.Date, o.Description, o.Category,
	)
	return err
}

// ListByAccount — список операций по счёту и периоду (включительно).
func (r *PgOperationRepo) ListByAccount(ctx context.Context, accID domain.AccountID, from, to time.Time) ([]domain.Operation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,type,bank_account_id,amount,"date",description,category_id
		   FROM operations
		  WHERE bank_account_id=$1 AND "date" BETWEEN $2 AND $3
		  ORDER BY "date", id`,
		accID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Operation
	for rows.Next() {
		var o domain.Operation
		var amt string
		if err := rows.Scan(&o.ID, &o.Type, &o.BankAccount, &amt, &o.Date, &o.Description, &o.Category); err != nil {
			return nil, err
		}
		dec, err := decimal.NewFromString(amt)
		if err != nil {
			return nil, err
		}
		o.Amount = dec
		out = append(out, o)
	}
	return out, rows.Err()
}
