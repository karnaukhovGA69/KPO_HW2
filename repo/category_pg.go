package repo

import (
	"context"
	"errors"

	"main/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgCategoryRepo struct{ db *pgxpool.Pool }

func NewPgCategoryRepo(db *pgxpool.Pool) *PgCategoryRepo { return &PgCategoryRepo{db: db} }

func (r *PgCategoryRepo) UpdateName(ctx context.Context, id domain.CategoryID, name string) error {
	_, err := r.db.Exec(ctx, `UPDATE categories SET name=$2 WHERE id=$1`, id, name)
	return err
}

func (r *PgCategoryRepo) UpdateType(ctx context.Context, id domain.CategoryID, t domain.CategoryType) error {
	ct, err := r.db.Exec(ctx, `UPDATE categories SET type=$2 WHERE id=$1`, id, int(t))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("category not found")
	}
	return nil
}
func (r *PgCategoryRepo) Delete(ctx context.Context, id domain.CategoryID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}

func (r *PgCategoryRepo) HasOperations(ctx context.Context, id domain.CategoryID) (bool, error) {
	var n int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(1) FROM operations WHERE category_id=$1`, id).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}
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
	rows, err := r.db.Query(ctx, `
    SELECT DISTINCT ON (type, name) id, type, name
    FROM categories
    ORDER BY type, name, id
  `)
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
