package repos

import (
	"context"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type PwResetRepo struct {
	sr SqlRunner
}

func (r PwResetRepo) CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.CreatePwResetTokenParams{
			UserID: userId,
			Token:  token,
			ExpiresAt: pgtype.Timestamptz{
				Time:  expiresAt,
				Valid: true,
			},
		}

		_, err := q.CreatePwResetToken(ctx, params)
		return err
	})
}

func (r PwResetRepo) GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetPwResetToken(ctx, token)
	})

	if err != nil || res == nil {
		return sqlc.PwResetToken{}, err
	}
	return res.(sqlc.PwResetToken), err
}

func (r PwResetRepo) DeletePwResetToken(ctx context.Context, token string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeletePwResetToken(ctx, token)
	})
}

func (r PwResetRepo) DeleteUserPwResetTokens(ctx context.Context, userId int32) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeleteUserPwResetTokens(ctx, userId)
	})
}
