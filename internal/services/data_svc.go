package services

import (
	"context"
	"fmt"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
)

type DataSvc struct {
	DataRetriever DeviceDataRetriever
}

func NewDataSvc(dataRet DeviceDataRetriever) DataSvc {
	return DataSvc{DataRetriever: dataRet}
}

func (s DataSvc) CapacitanceData(ctx context.Context, deviceId int, startTime string) ([]db.DeviceDataPoint, error) {
	return s.dataSince(ctx, deviceId, startTime, core.Capacitance)
}

func (s DataSvc) TemperatureData(ctx context.Context, deviceId int, startTime string) ([]db.DeviceDataPoint, error) {
	return s.dataSince(ctx, deviceId, startTime, core.Temperature)
}

func (s DataSvc) dataSince(ctx context.Context, deviceId int, startTime string, measurement string) ([]db.DeviceDataPoint, error) {
	startTimeT, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf(
			`Error parsing time in DataSvc - measurement '%v', startTime '%v: \nError:\n%w\n`,
			measurement, startTime, err)
	}

  endTimeT := time.Now()
	data, err := s.DataRetriever.GetValuesRange(ctx,
		deviceId,
		measurement,
		startTimeT,
		endTimeT)
	if err != nil {
		return nil, fmt.Errorf("Error in DataSvc -> GetValuesRange: \n%w\n", err)
	}
	return data, nil
}
