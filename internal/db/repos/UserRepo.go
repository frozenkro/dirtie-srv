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
}

type userRepoImpl struct {
  tm *TxManager
}

func (r *userRepoImpl) GetUser(ctx context.Context, userId int32) (sqlc.User, error) {
  res, err := r.tm.WithTxRes(ctx, func (q *sqlc.Queries) (interface{}, error) { 
    return q.GetUser(ctx, userId)
  })

  return res.(sqlc.User), err
}

func (r *userRepoImpl) GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error) {
  res, err := r.tm.WithTxRes(ctx, func (q *sqlc.Queries) (interface{}, error) {
    return q.GetUserFromEmail(ctx, email)
  })
  return res.(sqlc.User), err
}

func (r *userRepoImpl) CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error) {
  res, err := r.tm.WithTxRes(ctx, func (q *sqlc.Queries) (interface{}, error) {
    params := sqlc.CreateUserParams{
      Email: email,
      PwHash: pwHash,
      Name: name,
    }
    return q.CreateUser(ctx, params)
  })
  return res.(sqlc.User), err
}

func (r *userRepoImpl) ChangePassword(ctx context.Context, userId int32, pwHash []byte) error {
  return r.tm.WithTx(ctx, func (q *sqlc.Queries) error {
    params := sqlc.ChangePasswordParams{
      UserID: userId,
      PwHash: pwHash,
    }
    return q.ChangePassword(ctx, params)
  })
}
