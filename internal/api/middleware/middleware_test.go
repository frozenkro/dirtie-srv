package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(ctx context.Context, token string) (*sqlc.User, error) {
	args := m.Called(ctx, token)
	if user, ok := args.Get(0).(*sqlc.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockTokenValidator)
		setupRequest   func(*http.Request)
		expectedStatus int
		checkContext   bool
	}{
		{
			name: "Valid token",
			setupMock: func(m *MockTokenValidator) {
				m.On("ValidateToken", mock.Anything, "valid-token").Return(&sqlc.User{UserID: 1}, nil)
			},
			setupRequest: func(r *http.Request) {
				cookie := &http.Cookie{
					Name:  "dirtie.auth",
					Value: "valid-token",
				}
				r.AddCookie(cookie)
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name: "No cookie",
			setupMock: func(m *MockTokenValidator) {
				// No mock setup needed
			},
			setupRequest: func(r *http.Request) {
				// Don't add any cookie
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Expired token",
			setupMock: func(m *MockTokenValidator) {
				m.On("ValidateToken", mock.Anything, "expired-token").Return(nil, services.ErrExpiredToken)
			},
			setupRequest: func(r *http.Request) {
				cookie := &http.Cookie{
					Name:  "dirtie.auth",
					Value: "expired-token",
				}
				r.AddCookie(cookie)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Invalid token",
			setupMock: func(m *MockTokenValidator) {
				m.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, services.ErrInvalidToken)
			},
			setupRequest: func(r *http.Request) {
				cookie := &http.Cookie{
					Name:  "dirtie.auth",
					Value: "invalid-token",
				}
				r.AddCookie(cookie)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := &MockTokenValidator{}
			tt.setupMock(mockValidator)

			var contextUser *sqlc.User
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					contextUser = r.Context().Value("user").(*sqlc.User)
				}
			})

			handler := Authorize(mockValidator)(mockHandler)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(r)

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkContext {
				assert.NotNil(t, contextUser)
				assert.Equal(t, int32(1), contextUser.UserID)
			}
			mockValidator.AssertExpectations(t)
		})
	}
}
