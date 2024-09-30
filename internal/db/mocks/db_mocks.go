package mocks

import (
	"context"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepo struct {
	mock.Mock
}

type MockSessionRepo struct {
	mock.Mock
}

type MockPwResetRepo struct {
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

func (m *MockPwResetRepo) CreatePwResetToken(ctx context.Context, userId int32, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userId, token, expiresAt)
	return args.Error(0)
}
func (m *MockPwResetRepo) GetPwResetToken(ctx context.Context, token string) (sqlc.PwResetToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(sqlc.PwResetToken), args.Error(1)
}
func (m *MockPwResetRepo) DeletePwResetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
func (m *MockPwResetRepo) DeleteUserPwResetTokens(ctx context.Context, userId int32) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}
