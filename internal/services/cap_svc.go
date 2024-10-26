package services

import (
	"context"
	"time"
)

type CapSvc struct {
	DataRetriever   DeviceDataRetriever
	DeviceGetter DeviceGetter
}

func NewCapSvc(dataRet DeviceDataRetriever, devGet DeviceGetter) CapSvc {
	return CapSvc{DataRetriever: dataRet, DeviceGetter: devGet}
}

func GetSince(ctx context.Context, deviceId int, startTime time.Time) ([]int, error) {
	// TODO
	return nil, nil
}
