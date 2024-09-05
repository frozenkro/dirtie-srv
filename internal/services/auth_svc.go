package services

import (
	"context"
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

var (
  ErrInvalidToken = fmt.Errorf("Invalid session token")
  ErrExpiredToken = fmt.Errorf("Session token expired")
  ErrUserExists = fmt.Errorf("User Email already exists")
)

func NewAuthSvc(userRepo repos.UserRepo, sessionRepo repos.SessionRepo) *AuthSvc {
  return &AuthSvc{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *AuthSvc) CreateUser(ctx context.Context, email string, password string, name string) (*sqlc.User, error) {

  existingUser, err := s.userRepo.GetUserFromEmail(ctx, email)
  if err != nil {
    return nil, err
  }
  if existingUser.UserID > 0 {
    return nil, fmt.Errorf("Creating User with email %v: %w", email, ErrUserExists)
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

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*sqlc.User, error) {
  session, err := s.sessionRepo.GetSession(ctx, token)
  if err != nil {
    return nil, err
  }

  if session.UserID < 1 {
    return nil, fmt.Errorf("Validating token %v: %w", token, ErrInvalidToken)
  }

  if time.Now().After(session.ExpiresAt.Time) {
    return nil, fmt.Errorf("Validating token %v: %w", token, ErrExpiredToken)
  }
  
  user, err := s.userRepo.GetUser(ctx, session.UserID)
  if err != nil {
    return nil, err
  }

  return &user, nil
}

func (s *AuthSvc) Logout(ctx context.Context, token string) error {
  session, err := s.sessionRepo.GetSession(ctx, token)
  if err != nil {
    return err
  }

  err = s.sessionRepo.DeleteUserSessions(ctx, session.UserID)
  return err
}
