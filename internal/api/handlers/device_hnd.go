package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

func SetupDeviceHandlers(deps *core.Deps) {
  http.Handle("/devices", middleware.Adapt(
    getUserDevicesHandler(deps.DeviceSvc),
    middleware.LogTransaction(),
    middleware.Authorize(deps.AuthSvc),
  ))
}

func getUserDevicesHandler(deviceSvc services.DeviceSvc) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    devices, err := deviceSvc.GetUserDevices(r.Context())
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    resBody, err := json.Marshal(devices)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    w.Write(resBody)
  })
}

func getDeviceHandler(deviceSvc services.DeviceSvc) http.Handler {
  return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    // device, err := deviceSvc.GetUserDevice(r.Context(), r.URL.Parse
    return
  })
}
