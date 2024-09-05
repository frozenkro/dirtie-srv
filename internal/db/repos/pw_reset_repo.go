package repos

import (
  "context"
  "time"

  "github.com/frozenkro/dirtie-srv/internal/db/sqlc"
  "github.com/jackc/pgx/v5/pgtype"
)

type PwResetRepo interface {
  CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error
  GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error)
  DeletePwResetToken(ctx context.Context, token string) error
  DeleteUserPwResetTokens(ctx context.Context, userId int32) error
}

type pwResetRepoImpl struct {
  sr SqlRunner
}

func (r *pwResetRepoImpl) CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    params := sqlc.CreatePwResetTokenParams{
      UserID: userId,
      Token: token,
      ExpiresAt: pgtype.Timestamptz {
        Time: expiresAt,
      },
    }

    _, err := q.CreatePwResetToken(ctx, params)
    return err
  })
}

func (r *pwResetRepoImpl) GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error) {
  res, err := r.sr.Query(ctx, func (q *sqlc.Queries) (interface{}, error) {
    return q.GetPwResetToken(ctx, token)
  })
  return res.(sqlc.PwResetToken), err
}

func (r *pwResetRepoImpl) DeletePwResetToken(ctx context.Context, token string) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    return q.DeletePwResetToken(ctx, token)
  })
}

func (r *pwResetRepoImpl) DeleteUserPwResetTokens(ctx context.Context, userId int32) error {
  return r.sr.Execute(ctx, func (q *sqlc.Queries) error {
    return q.DeleteUserPwResetTokens(ctx, userId)
  })
}
