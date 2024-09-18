package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlRunner interface {
	Query(ctx context.Context, fn func(*sqlc.Queries) (interface{}, error)) (interface{}, error)
	Execute(ctx context.Context, fn func(*sqlc.Queries) error) error
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (tm *TxManager) Query(ctx context.Context, fn func(*sqlc.Queries) (interface{}, error)) (interface{}, error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if res, err := fn(sqlc.New(tx)); err == nil {
		if err = tx.Commit(ctx); err == nil {
			return res, nil
		}
	}
	return nil, err
}

func (tm *TxManager) Execute(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err = fn(sqlc.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
