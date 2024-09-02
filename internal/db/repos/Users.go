package repos

import (
  "github.com/frozenkro/dirtie-srv/internal/db"
  "github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

func GetUser(id int32) (sqlc.User, error) {
  conn, ctx, err := db.PgConnect()
  defer conn.Close(ctx)
  if err != nil {
    return sqlc.User{}, err
  }

  queries := sqlc.New(conn)

  return queries.GetUser(ctx, id)
}
