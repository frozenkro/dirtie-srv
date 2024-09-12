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

type LoginArgs struct {
  email string
  password string
}

func SetupAuthHandlers(deps *core.Deps) {
  http.Handle("POST /users", middleware.Adapt(
    createUserHandler(deps.AuthSvc),
    middleware.LogTransaction(),
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

func loginHandler(authSvc services.AuthSvc) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r  *http.Request) {
    var args LoginArgs

    err := json.NewDecoder(r.Body).Decode(&args)
    if err != nil {
      http.Error(w, core.RequestParseError, http.StatusBadRequest)
    }

    token, err := authSvc.Login(r.Context(), args.email, args.password)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    cookie := http.Cookie {
      Name: "dirtie.auth",
      Value: token,
    }
    http.SetCookie(w, &cookie)
  })
}

func logoutHandler(authSvc services.AuthSvc) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("dirtie.auth")
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
    }

    err = authSvc.Logout(r.Context(), cookie.Value)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    w.WriteHeader(http.StatusOK)
  })
}
