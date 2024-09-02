package services

import (
	"context"

  //"golang.org/x/crypto/bcrypt"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
)

type AuthService struct {
  userRepo repos.UserRepo
}

func NewAuthService(userRepo repos.UserRepo) *AuthService {
  return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
  
  return "", nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*sqlc.User, bool) {
  return &sqlc.User{}, true
}
