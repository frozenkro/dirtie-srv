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
	userReader    mocks.MockUserReader
	userWriter    mocks.MockUserWriter
	sessionReader mocks.MockSessionReader
	sessionWriter mocks.MockSessionWriter
	pwResetReader mocks.MockPwResetReader
	pwResetWriter mocks.MockPwResetWriter
	emailSender   mocks.MockEmailSender
	htmlParser    mocks.MockHtmlParser
	authSvc       AuthSvc
)

func setupAuthSvcTests() {
	userReader = mocks.MockUserReader{Mock: new(mock.Mock)}
	userWriter = mocks.MockUserWriter{Mock: new(mock.Mock)}
	sessionReader = mocks.MockSessionReader{Mock: new(mock.Mock)}
	sessionWriter = mocks.MockSessionWriter{Mock: new(mock.Mock)}
	pwResetReader = mocks.MockPwResetReader{Mock: new(mock.Mock)}
	pwResetWriter = mocks.MockPwResetWriter{Mock: new(mock.Mock)}
	htmlParser = mocks.MockHtmlParser{Mock: new(mock.Mock)}
	emailSender = mocks.MockEmailSender{Mock: new(mock.Mock)}

	authSvc = NewAuthSvc(userReader,
		userWriter,
		sessionReader,
		sessionWriter,
		pwResetReader,
		pwResetWriter,
		htmlParser,
		emailSender)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	setupAuthSvcTests()

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
		userReader.AssertExpectations(t)
		userWriter.AssertExpectations(t)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		email := "existing@example.com"
		password := "password123"
		name := "Existing User"

		userReader.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1}, nil)

		user, err := authSvc.CreateUser(ctx, email, password, name)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrUserExists))
		userReader.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	setupAuthSvcTests()

	t.Run("Success", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

		userReader.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1, Email: email, PwHash: hashedPassword}, nil)
		userWriter.On("UpdateLastLoginTime", ctx, int32(1)).Return(nil)
		sessionWriter.On("CreateSession", ctx, int32(1), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
		sessionWriter.On("DeleteUserSessions", ctx, int32(1)).Return(nil)

		token, err := authSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		userReader.AssertExpectations(t)
		userWriter.AssertExpectations(t)
		sessionWriter.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		email := "test@example.com"
		password := "wrongpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 10)
		userReader.On("GetUserFromEmail", ctx, email).Return(sqlc.User{UserID: 1, Email: email, PwHash: hashedPassword}, nil)

		token, err := authSvc.Login(ctx, email, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		userReader.AssertExpectations(t)
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	setupAuthSvcTests()

	t.Run("ValidToken", func(t *testing.T) {
		token := uuid.New().String()
		userID := int32(1)
		expiresAt := time.Now().Add(time.Hour)

		sessionReader.On("GetSession", ctx, token).Return(sqlc.Session{UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: expiresAt}}, nil)
		userReader.On("GetUser", ctx, userID).Return(sqlc.User{UserID: userID, Email: "test@example.com"}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.UserID)
		sessionReader.AssertExpectations(t)
		userReader.AssertExpectations(t)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		token := uuid.New().String()
		userID := int32(1)
		expiresAt := time.Now().Add(-time.Hour) // Expired

		sessionReader.On("GetSession", ctx, token).Return(sqlc.Session{UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: expiresAt}}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrExpiredToken))
		sessionReader.AssertExpectations(t)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		token := "invalidtoken"

		sessionReader.On("GetSession", ctx, token).Return(sqlc.Session{}, nil)

		user, err := authSvc.ValidateToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, ErrInvalidToken))
		sessionReader.AssertExpectations(t)
	})
}

func TestLogout(t *testing.T) {
  ctx := context.Background()
	setupAuthSvcTests()

  t.Run("Success", func(t *testing.T) {
    token := "test_token"
    session := sqlc.Session{ UserID: 42069 }

    sessionReader.On("GetSession", ctx, token).Return(session, nil)
    sessionWriter.On("DeleteUserSessions", ctx, session.UserID).Return(nil)

    err := authSvc.Logout(ctx, token)
    assert.Nil(t, err)
    sessionReader.AssertExpectations(t)
    sessionWriter.AssertExpectations(t)
  })
}
