package api

import (
	"fmt"
	"net/http"

	"github.com/frozenkro/dirtie-srv/internal/api/handlers"
	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
)

const PORT = 8080

func Init(deps *core.Deps) {
  rootHandler := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "request sent to root /\n")
  })

  http.Handle("/", middleware.Adapt(rootHandler, 
    middleware.LogTransaction(), 
    middleware.Authorize(deps.AuthSvc),
  ))

  handlers.SetupDeviceHandlers(deps)
  
  utils.LogInfo(fmt.Sprintf("Starting web server on port %v", PORT))

  portStr := fmt.Sprintf(":%v", PORT)
  if err := http.ListenAndServe(portStr, nil); err != nil {
    utils.LogErr(fmt.Sprintf("Web server error: %v\n", err))
  }
}

