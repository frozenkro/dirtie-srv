package repos

import (
	"context"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepo interface {
  CreateSession(ctx context.Context, userId int32, token string, expiresAt time.Time) error
  GetSession(ctx context.Context, token string) (sqlc.Session, error)
  DeleteSession(ctx context.Context, token string) error
  DeleteUserSessions(ctx context.Context, userId int32) error 
}

type sessionRepoImpl struct {
  sr SqlRunner
}

func (r *sessionRepoImpl) CreateSession(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    params := sqlc.CreateSessionParams{
      UserID: userId,
      Token: token,
      ExpiresAt: pgtype.Timestamptz {
        Time: expiresAt,
      },
    }

    return q.CreateSession(ctx, params)
  })
}

func (r *sessionRepoImpl) GetSession(ctx context.Context, token string) (sqlc.Session, error) {
  res, err := r.sr.Query(ctx, func (q *sqlc.Queries) (interface{}, error) {
    return q.GetSession(ctx, token)
  })
  return res.(sqlc.Session), err
}

func (r *sessionRepoImpl) DeleteSession(ctx context.Context, token string) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    return q.DeleteSession(ctx, token)
  })
}

func (r *sessionRepoImpl) DeleteUserSessions(ctx context.Context, userId int32) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    return q.DeleteUserSessions(ctx, userId)
  })
}
