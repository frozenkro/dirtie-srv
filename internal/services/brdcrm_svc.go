package services

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
)

type BrdCrmSvc struct {
	DataRecorder    db.DeviceDataRecorder
	DataRetriever   db.DeviceDataRetriever
	DeviceRetriever DeviceRetriever
}
type BreadCrumb struct {
	Capacitance int64
	Temperature int64
}

func NewBrdCrmSvc(dataRec db.DeviceDataRecorder, dataRet db.DeviceDataRetriever, devRet DeviceRetriever) BrdCrmSvc {
	return BrdCrmSvc{DataRecorder: dataRec, DataRetriever: dataRet, DeviceRetriever: devRet}
}

var (
	ErrNoDevice = fmt.Errorf("Device not found")
)

func (s BrdCrmSvc) RecordBrdCrm(ctx context.Context, macAddr string, brdCrm BreadCrumb) error {
	dvc, err := s.DeviceRetriever.GetDeviceByMacAddr(ctx, macAddr)
	if err != nil {
		return fmt.Errorf("Error RecordCapacitance -> GetDeviceByMacAddr: \n%w\n", err)
	}
	if dvc.DeviceID <= 0 {
		return fmt.Errorf("Error in RecordCapacitance (macAddr: %v): \n%w\n", macAddr, ErrNoDevice)
	}

	err = s.DataRecorder.Record(ctx, int(dvc.DeviceID), core.Capacitance, brdCrm.Capacitance)
	err = s.DataRecorder.Record(ctx, int(dvc.DeviceID), core.Temperature, brdCrm.Temperature)
	if err != nil {
		return fmt.Errorf("Error RecordCapacitance -> Record: \n%w\n", err)
	}
	return nil
}
