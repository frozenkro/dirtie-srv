package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock repositories
type MockUserRepo struct {
	mock.Mock
}

type MockSessionRepo struct {
	mock.Mock
}

// Implement UserRepo interface methods for MockUserRepo
func (m *MockUserRepo) GetUserFromEmail(ctx context.Context, email string) (sqlc.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(ctx context.Context, email string, pwHash []byte, name string) (sqlc.User, error) {
	args := m.Called(ctx, email, pwHash, name)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockUserRepo) UpdateLastLoginTime(ctx context.Context, userID int32) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepo) GetUser(ctx context.Context, userID int32) (sqlc.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockUserRepo) ChangePassword(ctx context.Context, userId int32, pwHash []byte) error {
  args := m.Called(ctx, userId, pwHash)
  return args.Error(0)
}

// Implement SessionRepo interface methods for MockSessionRepo
func (m *MockSessionRepo) CreateSession(ctx context.Context, userID int32, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockSessionRepo) GetSession(ctx context.Context, token string) (sqlc.Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(sqlc.Session), args.Error(1)
}

func (m *MockSessionRepo) DeleteSession(ctx context.Context, token string) error {
  args := m.Called(ctx, token)
  return args.Error(0)
}

func (m *MockSessionRepo) DeleteUserSessions(ctx context.Context, userId int32) error {
  args := m.Called(ctx, userId)
  return args.Error(0)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepo)
	mockSessionRepo := new(MockSessionRepo)
	authSvc := NewAuthSvc(mockUserRepo, mockSessionRepo)

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		name := "Test User"

		mockUserRepo.On("GetUserFromEmail", ctx, email).Return(sqlc.User{}, nil)
		mockUserRepo.On("CreateUser", ctx, email, mock.AnythingOfType("[]uint8"), name).Return(sqlc.User{UserID: 1, Email: email, Name: name}, nil)

		user, err := authSvc.CreateUser(ctx, email, password, name)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Name)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		email := "existing@example.com"
		password := "password123"
		name := "Existing User"

		mockUserRepo.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1}, nil)

		user, err := authSvc.CreateUser(ctx, email, password, name)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrUserExists))
		mockUserRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepo)
	mockSessionRepo := new(MockSessionRepo)
	authSvc := NewAuthSvc(mockUserRepo, mockSessionRepo)

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

		mockUserRepo.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1, Email: email, PwHash: hashedPassword}, nil)
		mockUserRepo.On("UpdateLastLoginTime", ctx, int32(1)).Return(nil)
		mockSessionRepo.On("CreateSession", ctx, int32(1), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

		token, err := authSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		email := "test@example.com"
		password := "wrongpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 10)

		mockUserRepo.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1, Email: email, PwHash: hashedPassword}, nil)

		token, err := authSvc.Login(ctx, email, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepo)
	mockSessionRepo := new(MockSessionRepo)
	authSvc := NewAuthSvc(mockUserRepo, mockSessionRepo)

	t.Run("ValidToken", func(t *testing.T) {
		token := uuid.New().String()
		userID := int32(1)
		expiresAt := time.Now().Add(time.Hour)

		mockSessionRepo.On("GetSession", ctx, token).Return(sqlc.Session{UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: expiresAt}}, nil)
		mockUserRepo.On("GetUser", ctx, userID).Return(sqlc.User{UserID: userID, Email: "test@example.com"}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.UserID)
		mockSessionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		token := uuid.New().String()
		userID := int32(1)
		expiresAt := time.Now().Add(-time.Hour) // Expired

		mockSessionRepo.On("GetSession", ctx, token).Return(sqlc.Session{UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: expiresAt}}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrExpiredToken))
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		token := "invalidtoken"

		mockSessionRepo.On("GetSession", ctx, token).Return(sqlc.Session{}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrInvalidToken))
		mockSessionRepo.AssertExpectations(t)
	})
}