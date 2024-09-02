package db

import (
  "context"
  "fmt"
  "os"

  "github.com/jackc/pgx/v5"
)


func PgConnect() (*pgx.Conn, context.Context, error) {
  ctx := context.Background()

  connstr := fmt.Sprintf("postgres://%v:%v@%v:5432/%v", 
    os.Getenv("POSTGRES_USER"),
    os.Getenv("POSTGRES_PASSWORD"),
    os.Getenv("POSTGRES_SERVER"),
    os.Getenv("POSTGRES_DB"))

  conn, err := pgx.Connect(ctx, connstr)
  return conn, ctx, err
}
