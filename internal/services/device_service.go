package services

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
)

type DeviceSvc struct {
  deviceRepo repos.DeviceRepo
  prvStgRepo repos.ProvisionStagingRepo
}

func NewDeviceSvc(deviceRepo repos.DeviceRepo, prvStgRepo repos.ProvisionStagingRepo) *DeviceSvc {
  return &DeviceSvc{deviceRepo: deviceRepo, prvStgRepo: prvStgRepo}
}

func (s *DeviceSvc) GetUserDevices(ctx context.Context) ([]sqlc.Device, error){
  fmt.Println("GetUserDevices - todo")
  // TODO 
  // get user from context
  // get all devices from db for user
  return nil, nil
}

func (s *DeviceSvc) GetUserDevice(ctx context.Context, deviceId int) (sqlc.Device, error) {
  fmt.Println("GetUserDevice - todo")
  // TODO 
  // get user from context
  // get device from deviceRepo
  // confirm device belongs to user
  return sqlc.Device{}, nil
}

// Called by user via rest api
func (s *DeviceSvc) CreateDeviceProvision(ctx context.Context, displayName string) (string, error) {
  // TODO
  // Save new device record to db
  // create uuid contract
  // Save to provision staging table 
  // return uuid to client (will be fwded to device)
  return "", nil
}


// Called by device via mqtt hub
func (s *DeviceSvc) CompleteDeviceProvision(ctx context.Context, contract string, macAddr string) (sqlc.Device, error) {
  // TODO 
  // lookup contract (uuid) from provision staging table
  // update mac address of device record
  // return device (ID will be used for influxdb entries)
  return sqlc.Device{}, nil
}
