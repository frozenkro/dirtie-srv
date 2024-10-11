package db

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PgConnect() (*pgxpool.Pool, error) {
	ctx := context.Background()

	connstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		core.POSTGRES_DB)

	config, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	return pool, err
}
