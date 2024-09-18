package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func PgConnect() (*pgxpool.Pool, error) {
	ctx := context.Background()

	connstr := fmt.Sprintf("postgres://%v:%v@%v:5432/%v",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_SERVER"),
		os.Getenv("POSTGRES_DB"))

	config, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	return pool, err
}
