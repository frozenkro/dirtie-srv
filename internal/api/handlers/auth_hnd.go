package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
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

type ChangePwData struct {
  Username     string
  Error        bool
  ErrorMessage string
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
  http.Handle("POST /logout", middleware.Adapt(
    logoutHandler(deps.AuthSvc),
    middleware.Authorize(deps.AuthSvc),
    middleware.LogTransaction(),
  ))
  http.Handle("POST /pw/reset", middleware.Adapt(
    resetPwHandler(deps.AuthSvc),
    middleware.LogTransaction(),
  ))
  http.Handle("/pw/change", middleware.Adapt(
    changePwHandler(deps.AuthSvc, deps.HtmlParser, deps.UserRepo),
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

func resetPwHandler(authSvc services.AuthSvc) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    email := r.URL.Query().Get("email")
    err := authSvc.ForgotPw(r.Context(), email)

    if err != nil {
      utils.LogErr(err.Error())
      if !errors.Is(err, services.ErrNoUser) {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    }

    w.WriteHeader(http.StatusOK)
  })
}

func changePwHandler(authSvc services.AuthSvc, htmlUtil utils.HtmlParser, userRepo repos.UserRepo) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO this does too much, refactor to service
    ctx := r.Context()

    token := r.URL.Query().Get("token")
    if token == "" {
      http.Error(w, core.GetMissingParamError("token"),http.StatusUnprocessableEntity)
      return
    }


    var changePwData ChangePwData
    userId, err := authSvc.ValidateForgotPwToken(ctx, token)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    if userId <= 0 {
      http.Error(w, "Unrecognized auth token", http.StatusUnauthorized)
      return
    }

    user, err := userRepo.GetUser(ctx, userId)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    changePwData.Username = user.Name
    changePwData.Error = false
    changePwData.ErrorMessage = ""

    if r.Method == http.MethodPost {
      newPw := r.FormValue("pw1")
      conf := r.FormValue("pw2")

      if newPw != conf {
        changePwData.Error = true
        changePwData.ErrorMessage = "Passwords do not match :("
      } else {
        err = authSvc.ChangePw(ctx, token, newPw)
        if err != nil {
          utils.LogErr(err.Error())
          changePwData.Error = true
          changePwData.ErrorMessage = "Something went wrong :("
        }
      }

    } else if r.Method != http.MethodGet {
      w.WriteHeader(http.StatusMethodNotAllowed)
      return
    }

    // serve page with data
    assetsDir := os.Getenv("ASSETS_DIR")
    tmpl, err := htmlUtil.ReadFile(ctx, fmt.Sprintf("%vchangePasswordPage.html", assetsDir))
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    err = htmlUtil.ReplaceAndWrite(ctx, changePwData, tmpl, w)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  })
}
