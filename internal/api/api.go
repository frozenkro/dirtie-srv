package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type Adapter func(http.Handler) http.Handler

func Init(deps *core.Deps) {
  logger := log.New(os.Stdout, "server: ", log.Lshortfile)
  
  rootHandler := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "request sent to root /\n")
  })

  http.Handle("/", adapt(rootHandler, 
    LogTransaction(logger), 
    Authorize(deps.AuthSvc),
  ))
  
  fmt.Println("Starting web server on :8080")

  if err := http.ListenAndServe(":8080", nil); err != nil {
    fmt.Printf("Web server error: %v\n", err)
  }
}

func adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

func LogTransaction(logger *log.Logger) Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      now := time.Now().UTC()
      rIP := r.RemoteAddr
      verb := r.Method
      url := r.URL
      logger.Printf("[%v] IP: %v, Method: %v %v\n", now, rIP, verb, url)
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
