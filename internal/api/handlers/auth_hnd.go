package handlers

import (
	"encoding/json"
  "net/http"

	"github.com/frozenkro/dirtie-srv/internal/core"
  "github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type CreateUserArgs struct {
  email string
  password string
  name string
}

func SetupAuthHandlers(deps *core.Deps) {
  http.Handle("POST /users", middleware.Adapt(
    createUserHandler(deps.AuthSvc),
    middleware.LogTransaction(),
    middleware.Authorize(deps.AuthSvc),
  ))
}

func createUserHandler(authSvc services.AuthSvc) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    var args CreateUserArgs
    
    err := json.NewDecoder(r.Body).Decode(&args)
    if err != nil {
      http.Error(w, core.RequestParseError, http.StatusBadRequest)
    }

    user, err := authSvc.CreateUser(r.Context(), args.email, args.password, args.name)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    res, err := json.Marshal(user)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    w.Write(res)
  })
}

func forgotPasswordHandler() {}

func resetPasswordHandler() {}
