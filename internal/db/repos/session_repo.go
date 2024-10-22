package repos

import (
	"context"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepo struct {
	sr SqlRunner
}

func (r SessionRepo) CreateSession(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		expiresAtTz := pgtype.Timestamptz{Time: expiresAt, Valid: true}
		params := sqlc.CreateSessionParams{
			UserID:    userId,
			Token:     token,
			ExpiresAt: expiresAtTz,
		}

		return q.CreateSession(ctx, params)
	})
}

func (r SessionRepo) GetSession(ctx context.Context, token string) (sqlc.Session, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetSession(ctx, token)
	})
	if err != nil || res == nil {
		return sqlc.Session{}, err
	}
	return res.(sqlc.Session), err
}

func (r SessionRepo) DeleteSession(ctx context.Context, token string) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeleteSession(ctx, token)
	})
}

func (r SessionRepo) DeleteUserSessions(ctx context.Context, userId int32) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.DeleteUserSessions(ctx, userId)
	})
}
