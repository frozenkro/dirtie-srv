package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type CreateProvisionResponse struct {
  Contract string `json:"contract"`
}

func SetupDeviceHandlers(deps *di.Deps) {
	http.Handle("GET /devices", middleware.Adapt(
		getUserDevicesHandler(deps.DeviceSvc),
		middleware.LogTransaction(),
		middleware.Authorize(deps.AuthSvc),
	))

	http.Handle("POST /devices/createProvision", middleware.Adapt(
		createDeviceProvisionHandler(deps.DeviceSvc),
		middleware.LogTransaction(),
		middleware.Authorize(deps.AuthSvc),
	))
}

func getUserDevicesHandler(deviceSvc services.DeviceSvc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		devices, err := deviceSvc.GetUserDevices(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(devices)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(res)
	})
}

func createDeviceProvisionHandler(deviceSvc services.DeviceSvc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		displayName := params.Get("displayName")
		if displayName == "" {
			http.Error(w, core.GetMissingParamError("displayName"), http.StatusBadRequest)
			return
		}

		contract, err := deviceSvc.CreateDeviceProvision(r.Context(), displayName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
    res := CreateProvisionResponse{ Contract: contract }
    res_b, err := json.Marshal(res)
    if err != nil {
      // todo log stuff like this 
      http.Error(w, "An error has occurred", http.StatusInternalServerError)
    }

		w.Write(res_b)
	})
}
