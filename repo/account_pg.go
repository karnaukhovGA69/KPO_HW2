package repo

import (
	"context"
	"errors"

	"main/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PgAccountRepo struct{ db *pgxpool.Pool }

func NewPgAccountRepo(db *pgxpool.Pool) *PgAccountRepo { return &PgAccountRepo{db: db} }

func (r *PgAccountRepo) Create(ctx context.Context, a domain.BankAccount) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO accounts(id,name,balance) VALUES($1,$2,$3)`,
		a.ID, a.Name, a.Balance.StringFixed(2),
	)
	return err
}

func (r *PgAccountRepo) Get(ctx context.Context, id domain.AccountID) (domain.BankAccount, error) {
	var a domain.BankAccount
	var bal string
	err := r.db.QueryRow(ctx,
		`SELECT id, name, balance FROM accounts WHERE id=$1`, id,
	).Scan(&a.ID, &a.Name, &bal)
	if err != nil {
		return domain.BankAccount{}, err
	}
	dec, err := decimal.NewFromString(bal)
	if err != nil {
		return domain.BankAccount{}, err
	}
	a.Balance = dec
	return a, nil
}

func (r *PgAccountRepo) Update(ctx context.Context, a domain.BankAccount) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE accounts SET name=$2, balance=$3 WHERE id=$1`,
		a.ID, a.Name, a.Balance.StringFixed(2),
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("account not found")
	}
	return nil
}
func (r *PgAccountRepo) List(ctx context.Context) ([]domain.BankAccount, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, balance FROM accounts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.BankAccount
	for rows.Next() {
		var a domain.BankAccount
		var bal string
		if err := rows.Scan(&a.ID, &a.Name, &bal); err != nil {
			return nil, err
		}
		dec, err := decimal.NewFromString(bal)
		if err != nil {
			return nil, err
		}
		a.Balance = dec
		out = append(out, a)
	}
	return out, rows.Err()
}
