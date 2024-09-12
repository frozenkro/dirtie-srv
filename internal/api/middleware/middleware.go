package middleware

import (
  "context"
  "errors"
  "fmt"
  "net/http"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services"
  "github.com/frozenkro/dirtie-srv/internal/core/utils"
)

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

func LogTransaction() Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      rIP := r.RemoteAddr
      verb := r.Method
      url := r.URL
      logmsg := fmt.Sprintf("IP: %v, Method: %v %v\n", rIP, verb, url)
      utils.LogInfo(logmsg)
      h.ServeHTTP(w, r)  
    })
  }
}

func Authorize(authSvc services.AuthSvc) Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      cookie, err := r.Cookie("dirtie.auth")
      if errors.Is(err, http.ErrNoCookie) {
        w.WriteHeader(http.StatusUnauthorized)
      }

      user, err := authSvc.ValidateToken(r.Context(), cookie.Value)
      if isUnauthorized(user, err) {
        w.WriteHeader(http.StatusUnauthorized)
      }

      ctx := context.WithValue(r.Context(), "user", user)
      newRqst := r.WithContext(ctx)

      h.ServeHTTP(w, newRqst)
    })
  }
}

func isUnauthorized(user *sqlc.User, err error) bool {
  return errors.Is(err, services.ErrExpiredToken) ||
    errors.Is(err, services.ErrInvalidToken) ||
    user == nil ||
    user.UserID < 1
}