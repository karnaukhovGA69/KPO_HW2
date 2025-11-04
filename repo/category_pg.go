package repo

import (
	"context"

	"main/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgCategoryRepo struct{ db *pgxpool.Pool }

func NewPgCategoryRepo(db *pgxpool.Pool) *PgCategoryRepo { return &PgCategoryRepo{db: db} }

func (r *PgCategoryRepo) Create(ctx context.Context, c domain.Category) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO categories(id, type, name) VALUES ($1, $2, $3)`,
		c.ID, int(c.Type), c.Name,
	)
	return err
}

func (r *PgCategoryRepo) Get(ctx context.Context, id domain.CategoryID) (domain.Category, error) {
	var c domain.Category
	err := r.db.QueryRow(ctx,
		`SELECT id, type, name FROM categories WHERE id=$1`, id,
	).Scan(&c.ID, &c.Type, &c.Name)
	return c, err
}

func (r *PgCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, type, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Type, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
