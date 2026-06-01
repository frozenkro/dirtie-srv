package db

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
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
	if err != nil {
		return nil, err
	}

	if !core.IS_TEST {
		if initErr := initSchema(ctx, pool); initErr != nil {
			utils.LogErr(fmt.Sprintf("schema init: %v", initErr))
		}
	}

	return pool, nil
}

func initSchema(ctx context.Context, pool *pgxpool.Pool) error {
	var tableExists int
	row := pool.QueryRow(ctx, `
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'users'`)
	err := row.Scan(&tableExists)
	if err == nil {
		// init schema already run
		return nil
	}

	utils.LogInfo("schema not found, applying schema.sql")
	_, err = pool.Exec(ctx, string(SchemaSql))
	if err != nil {
		return err
	}
	utils.LogInfo("schema applied")
	return nil
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
