package utils

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

type UserGetter interface {
	GetUser(ctx context.Context) (sqlc.User, error)
}

type CtxUtil struct{}

var ErrNoUserCtx = fmt.Errorf("Error: No user object in context")

func (s *CtxUtil) GetUser(ctx context.Context) (sqlc.User, error) {
	val := ctx.Value("user")
	if val == nil {
		return sqlc.User{}, ErrNoUserCtx
	}

	user, valid := val.(sqlc.User)
	if !valid {
		return sqlc.User{}, ErrNoUserCtx
	}
	return user, nil
}
