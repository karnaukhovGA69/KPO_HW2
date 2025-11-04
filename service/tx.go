package service

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TxStarter interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}
