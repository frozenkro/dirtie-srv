package services

import (
	"context"
  "fmt"
	"errors"

	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct {
  userRepo repos.UserRepo
}


func NewAuthSvc(userRepo repos.UserRepo) *AuthSvc {
  return &AuthSvc{userRepo: userRepo}
}

func (s *AuthSvc) CreateUser(ctx context.Context, email string, password string, name string) (*sqlc.User, error) {

  existingUser, err := s.userRepo.GetUserFromEmail(ctx, email)
  if err != nil {
    return nil, err
  }
  if existingUser.UserID > 0 {
    return nil, errors.New(fmt.Sprintf("User Email '%v' already exists", email))
  }

  pwHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
  if err != nil {
    return nil, err
  }
  
  newUser, err := s.userRepo.CreateUser(ctx, email, pwHash, name)
  if err != nil {
    return nil, err
  }
  return &newUser, err
}

func (s *AuthSvc) Login(ctx context.Context, email string, password string) (string, error) {



  return "", nil
}

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*sqlc.User, bool) {
  return &sqlc.User{}, true
}
