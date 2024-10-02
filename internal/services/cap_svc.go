package services

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
)

type CapSvc struct {
  DataRecorder db.DeviceDataRecorder
  DataRetriever db.DeviceDataRetriever
  DeviceRetriever DeviceRetriever
}

func NewCapSvc(dataRec db.DeviceDataRecorder, dataRet db.DeviceDataRetriever, devRet DeviceRetriever) CapSvc {
  return CapSvc{ DataRecorder: dataRec, DataRetriever: dataRet, DeviceRetriever: devRet }
}

var (
  ErrNoDevice = fmt.Errorf("Device not found")
)

func (s CapSvc) RecordCapacitance(ctx context.Context, macAddr string, value int64) error {
  dvc, err := s.DeviceRetriever.GetDeviceByMacAddr(ctx, macAddr)   
  if err != nil {
    return fmt.Errorf("Error RecordCapacitance -> GetDeviceByMacAddr: \n%w\n", err)
  }
  if dvc.DeviceID <= 0 {
    return fmt.Errorf("Error in RecordCapacitance (macAddr: %v): \n%w\n", macAddr, ErrNoDevice)
  }

  err = s.DataRecorder.Record(ctx, int(dvc.DeviceID), core.Capacitance, value)
  if err != nil {
    return fmt.Errorf("Error RecordCapacitance -> Record: \n%w\n", err)
  }
  return nil
}

