package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct {
  userRepo repos.UserRepo
  sessionRepo repos.SessionRepo
}


func NewAuthSvc(userRepo repos.UserRepo, sessionRepo repos.SessionRepo) *AuthSvc {
  return &AuthSvc{userRepo: userRepo, sessionRepo: sessionRepo}
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
  user, err := s.userRepo.GetUserFromEmail(ctx, email)
  if err != nil {
    return "", err
  }

  err = bcrypt.CompareHashAndPassword(user.PwHash, []byte(password))
  if err != nil {
    return "", err
  }

  err = s.userRepo.UpdateLastLoginTime(ctx, user.UserID)
  if err != nil {
    return "", err
  }

  token := uuid.NewString()
  dur, err := time.ParseDuration("2h")
  if err != nil {
    return "", err
  }
  expiresAt := time.Now().Add(dur)
  err = s.sessionRepo.CreateSession(ctx, user.UserID, token, expiresAt)
  if err != nil {
    return "", err
  }
  return token, nil
}

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*sqlc.User, bool) {
  session, err := s.sessionRepo.GetSession(ctx, token)
  if err != nil || session.UserID < 1 {
    return nil, false
  }
  
  // TODO 
  return &sqlc.User{}, true
}
