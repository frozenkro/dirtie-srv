package services

import (
	"context"
	"fmt"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

type DeviceGetter interface {
	GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error)
}

type DeviceDataRecorder interface {
	Record(ctx context.Context, deviceId int, measurementKey string, value int64) error
}

type DeviceDataRetriever interface {
	GetLatestValue(ctx context.Context, deviceId int, measurementKey string) (db.DeviceDataPoint, error)
	GetValuesRange(ctx context.Context, deviceId int, measurementKey string, start time.Time, end time.Time) ([]db.DeviceDataPoint, error)
}

type DevicePrvCompleter interface {
	CompleteDeviceProvision(context.Context, DevicePrvPayload) (sqlc.Device, error)
}

type BrdCrmSvc struct {
	DataRecorder  DeviceDataRecorder
	DataRetriever DeviceDataRetriever
	DeviceGetter  DeviceGetter
	PrvCompleter  DevicePrvCompleter
}
type BreadCrumb struct {
	MacAddr     string `json:"macAddr"`
	Contract    string `json:"contract"`
	Capacitance int64  `json:"capacitance"`
	Temperature int64  `json:"temperature"`
}

func NewBrdCrmSvc(dataRec DeviceDataRecorder, 
	dataRet DeviceDataRetriever, 
	deviceGetter DeviceGetter, 
	prvCompleter DevicePrvCompleter,
) BrdCrmSvc {
	return BrdCrmSvc{
		DataRecorder: dataRec, 
		DataRetriever: dataRet, 
		DeviceGetter: deviceGetter,
		PrvCompleter: prvCompleter,
	}
}

var (
	ErrNoDevice = fmt.Errorf("Device not found")
)

func (s BrdCrmSvc) RecordBrdCrm(ctx context.Context, brdCrm BreadCrumb) error {
	dvc, err := s.DeviceGetter.GetDeviceByMacAddress(ctx, brdCrm.MacAddr)
	if err != nil {
		return fmt.Errorf("Error RecordBrdCrm -> GetDeviceByMacAddr: \n%w\n", err)
	}
	if dvc.DeviceID <= 0 {
		// TODO lazy provisioning
		payload := DevicePrvPayload{MacAddr: brdCrm.MacAddr, Contract: brdCrm.Contract}
		ps, err := s.PrvCompleter.CompleteDeviceProvision(ctx, payload)
		if err != nil {
			return fmt.Errorf("Error RecordBrdCrm -> GetProvisionStagingByContract: \n%w\n", err)
		}
		if ps.MacAddr.String == "" {
			// No device or provision staging record found for this contract / mac address
			return fmt.Errorf("Error in RecordBrdCrm (macAddr: %v): \n%w\n", brdCrm.MacAddr, ErrNoDevice)
		}
	}

	err = s.DataRecorder.Record(ctx, int(dvc.DeviceID), core.Capacitance, brdCrm.Capacitance)
	if err != nil {
		return fmt.Errorf("Error RecordBrdCrm -> Record capacitance: \n%w\n", err)
	}
	err = s.DataRecorder.Record(ctx, int(dvc.DeviceID), core.Temperature, brdCrm.Temperature)
	if err != nil {
		return fmt.Errorf("Error RecordBrdCrm -> Record temperature: \n%w\n", err)
	}
	return nil
}

func (s BrdCrmSvc) GetLatestBrdCrm(ctx context.Context, deviceId int) (*BreadCrumb, error) {
	cap, err := s.DataRetriever.GetLatestValue(ctx, deviceId, core.Capacitance)
	if err != nil {
		return nil, fmt.Errorf("Error GetLatestBrdCrm -> GetLatestValue (capacitance): \n%w\n", err)
	}

	temp, err := s.DataRetriever.GetLatestValue(ctx, deviceId, core.Temperature)
	if err != nil {
		return nil, fmt.Errorf("Error GetLatestBrdCrm -> GetLatestValue (temperature): \n%w\n", err)
	}

	return &BreadCrumb{
		Capacitance: cap.Value,
		Temperature: temp.Value,
	}, nil
}
