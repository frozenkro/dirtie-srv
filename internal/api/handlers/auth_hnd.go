package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type CreateUserArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SetupAuthHandlers(deps *core.Deps) {
	http.Handle("POST /users", middleware.Adapt(
		createUserHandler(deps.AuthSvc),
		middleware.LogTransaction(),
	))
	http.Handle("POST /login", middleware.Adapt(
		loginHandler(deps.AuthSvc),
		middleware.LogTransaction(),
	))
}

func createUserHandler(authSvc services.AuthSvc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var args CreateUserArgs

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, core.RequestParseError, http.StatusBadRequest)
			return
		}

		user, err := authSvc.CreateUser(r.Context(), args.Email, args.Password, args.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(res)
	})
}

func loginHandler(authSvc services.AuthSvc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var args LoginArgs

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, core.RequestParseError, http.StatusBadRequest)
			return
		}

		token, err := authSvc.Login(r.Context(), args.Email, args.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:  "dirtie.auth",
			Value: token,
		}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
	})
}

func logoutHandler(authSvc services.AuthSvc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("dirtie.auth")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = authSvc.Logout(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
