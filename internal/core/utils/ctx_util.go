package utils

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

type CtxUtil struct{}

var ErrNoUserCtx = fmt.Errorf("Error: No user object in context")
var ErrInvalidUserCtx = fmt.Errorf("Error: Unable to parse user object from context")

func (s *CtxUtil) GetUser(ctx context.Context) (sqlc.User, error) {
	val := ctx.Value("user")
	if val == nil {
		return sqlc.User{}, ErrNoUserCtx
	}

	user, valid := val.(*sqlc.User)
	if !valid {
		return sqlc.User{}, ErrInvalidUserCtx
	}
	return *user, nil
}
