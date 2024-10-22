package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var (
  userReader  mocks.MockUserReader
  userWriter  mocks.MockUserWriter
	sessionRepo mocks.MockSessionRepo
	pwResetRepo mocks.MockPwResetRepo
	emailSender mocks.MockEmailSender
	htmlParser  mocks.MockHtmlParser
	authSvc     *AuthSvc
)

func setup() {
  userReader  = *new(mocks.MockUserReader)
  userWriter  = *new(mocks.MockUserWriter)
	sessionRepo = *new(mocks.MockSessionRepo)
	pwResetRepo = *new(mocks.MockPwResetRepo)
	emailSender = *new(mocks.MockEmailSender)
	htmlParser  = *new(mocks.MockHtmlParser)

	authSvc = NewAuthSvc(userReader,
		userWriter,
    &sessionRepo,
    &pwResetRepo,
    emailSender,
    htmlParser)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	setup()

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		name := "Test User"

		userReader.On("GetUserFromEmail", ctx, email).Return(sqlc.User{}, nil)
		userWriter.On("CreateUser", ctx, email, mock.AnythingOfType("[]uint8"), name).Return(sqlc.User{UserID: 1, Email: email, Name: name}, nil)

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
	setup()

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

		mockUserRepo.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1, Email: email, PwHash: hashedPassword}, nil)
		mockUserRepo.On("UpdateLastLoginTime", ctx, int32(1)).Return(nil)
		mockSessionRepo.On("CreateSession", ctx, int32(1), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
		mockSessionRepo.On("DeleteUserSessions", ctx, int32(1)).Return(nil)

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
	setup()

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
