package db

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PgConnect(ctx context.Context) (*pgxpool.Pool, error) {
  dbName := getDbName(ctx)

	connstr := fmt.Sprintf("postgres://%v:%v@%v/%v",
		core.POSTGRES_USER,
		core.POSTGRES_PASSWORD,
		core.POSTGRES_SERVER,
		dbName)

	config, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	return pool, err
}

func getDbName(ctx context.Context) string {
  if !core.IS_TEST {
    return core.POSTGRES_DB
  }

  dbNameAny := ctx.Value("testdb")
  if dbNameAny == nil || dbNameAny == "" {
    panic("DB Name must be set on context while app runs in test mode")
  }
  dbName, succ := dbNameAny.(string)
  if !succ {
    panic("DB Name on context not castable to string")
  }

  return dbName
}
