package services

import (
	"context"
	"time"
)

type CapSvc struct {
  DataRetriever DeviceDataRetriever
  DeviceRetriever DeviceRetriever
}

func NewCapSvc(dataRet DeviceDataRetriever, devRet DeviceRetriever) CapSvc {
  return CapSvc{DataRetriever: dataRet, DeviceRetriever: devRet}
}

func GetSince(ctx context.Context, deviceId int, startTime time.Time) ([]int, error) {
  // TODO
  return nil, nil
}
