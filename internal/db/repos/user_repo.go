package repos

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

type UserRepo interface {
	GetUser(ctx context.Context, userId int32) (sqlc.User, error)
	GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error)
	CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error)
	ChangePassword(ctx context.Context, userId int32, pwHash []byte) error
	UpdateLastLoginTime(ctx context.Context, userId int32) error
}

type userRepoImpl struct {
	sr SqlRunner
}

func (r *userRepoImpl) GetUser(ctx context.Context, userId int32) (sqlc.User, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetUser(ctx, userId)
	})

	if err != nil || res == nil {
		return sqlc.User{}, err
	}
	return res.(sqlc.User), err
}

func (r *userRepoImpl) GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		return q.GetUserFromEmail(ctx, email)
	})

	if err != nil || res == nil {
		return sqlc.User{}, err
	}
	return res.(sqlc.User), err
}

func (r *userRepoImpl) CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error) {
	res, err := r.sr.Query(ctx, func(q *sqlc.Queries) (interface{}, error) {
		params := sqlc.CreateUserParams{
			Email:  email,
			PwHash: pwHash,
			Name:   name,
		}
		return q.CreateUser(ctx, params)
	})

	if err != nil || res == nil {
		return sqlc.User{}, err
	}
	return res.(sqlc.User), err
}

func (r *userRepoImpl) ChangePassword(ctx context.Context, userId int32, pwHash []byte) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		params := sqlc.ChangePasswordParams{
			UserID: userId,
			PwHash: pwHash,
		}
		return q.ChangePassword(ctx, params)
	})
}

func (r *userRepoImpl) UpdateLastLoginTime(ctx context.Context, userId int32) error {
	return r.sr.Execute(ctx, func(q *sqlc.Queries) error {
		return q.UpdateLastLoginTime(ctx, userId)
	})
}
