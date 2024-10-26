package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/frozenkro/dirtie-srv/internal/api/middleware"
	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/di"
)

type capReader interface {
	CapacitanceData(context.Context, int, string) ([]db.DeviceDataPoint, error)
}

type tempReader interface {
	TemperatureData(context.Context, int, string) ([]db.DeviceDataPoint, error)
}

func SetupDatahanders(deps *di.Deps) {
	http.Handle("GET /data/capacitance", middleware.Adapt(
		getCapHandler(deps.DataSvc),
		middleware.LogTransaction(),
		middleware.Authorize(deps.AuthSvc),
	))
	http.Handle("GET /data/temperature", middleware.Adapt(
		getTempHandler(deps.DataSvc),
		middleware.LogTransaction(),
		middleware.Authorize(deps.AuthSvc),
	))
}

func getCapHandler(cr capReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		deviceIdStr := params.Get("deviceId")
		if deviceIdStr == "" {
			http.Error(w, core.GetMissingParamError("deviceId"), http.StatusBadRequest)
			return
		}
		deviceId, err := strconv.Atoi(deviceIdStr)
		if err != nil {
			http.Error(w, "parameter 'deviceId' must be a number", http.StatusBadRequest)
			return
		}

		startTime := params.Get("startTime")
		if startTime == "" {
			http.Error(w, core.GetMissingParamError("startTime"), http.StatusBadRequest)
			return
		}

		data, err := cr.CapacitanceData(r.Context(), deviceId, startTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		res, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write(res)
	})
}

func getTempHandler(tr tempReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		deviceIdStr := params.Get("deviceId")
		if deviceIdStr == "" {
			http.Error(w, core.GetMissingParamError("deviceId"), http.StatusBadRequest)
			return
		}
		deviceId, err := strconv.Atoi(deviceIdStr)
		if err != nil {
			http.Error(w, "parameter 'deviceId' must be a number", http.StatusBadRequest)
			return
		}

		startTime := params.Get("startTime")
		if startTime == "" {
			http.Error(w, core.GetMissingParamError("startTime"), http.StatusBadRequest)
			return
		}

		data, err := tr.TemperatureData(r.Context(), deviceId, startTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		res, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write(res)
	})
}
