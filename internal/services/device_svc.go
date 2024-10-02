package services

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/db/repos"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
)

type DeviceRetriever interface {
  GetDeviceByMacAddr(ctx context.Context, macAddr string) (sqlc.Device, error)
}

type DeviceSvc struct {
	deviceRepo repos.DeviceRepo
	prvStgRepo repos.ProvisionStagingRepo
	userGetter utils.UserGetter
}

func NewDeviceSvc(deviceRepo repos.DeviceRepo,
	prvStgRepo repos.ProvisionStagingRepo,
	userGetter utils.UserGetter) *DeviceSvc {

	return &DeviceSvc{deviceRepo: deviceRepo,
		prvStgRepo: prvStgRepo,
		userGetter: userGetter,
	}
}

func (s DeviceSvc) GetUserDevices(ctx context.Context) ([]sqlc.Device, error) {
	user, err := s.userGetter.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error GetUserDevices -> GetUser: \n%w\n", err)
	}

	devices, err := s.deviceRepo.GetDevicesByUser(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error GetUserDevices -> GetDevicesByUser: \n%w\n", err)
	}
	return devices, nil
}

func (s DeviceSvc) GetDeviceByMacAddr(ctx context.Context, macAddr string) (sqlc.Device, error) {
	device, err := s.deviceRepo.GetDeviceByMacAddress(ctx, macAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error GetDeviceByMacAddr -> GetDeviceByMacAddress: \n%w\n", err)
	}
	return device, nil
}

// Called by user via rest api
func (s DeviceSvc) CreateDeviceProvision(ctx context.Context, displayName string) (string, error) {
	user, err := s.userGetter.GetUser(ctx)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> GetUser: \n%w\n", err)
	}

	device, err := s.deviceRepo.CreateDevice(ctx, user.UserID, displayName)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> CreateDevice: \n%w\n", err)
	}
	uuid := uuid.NewString()
	err = s.prvStgRepo.CreateProvisionStaging(ctx, device.DeviceID, uuid)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> CreateProvisionStaging: \n%w\n", err)
	}
	return uuid, nil
}

// Called by device via mqtt hub
func (s DeviceSvc) CompleteDeviceProvision(ctx context.Context, contract string, macAddr string) (sqlc.Device, error) {
	// lookup contract (uuid) from provision staging table
	prv, err := s.prvStgRepo.GetProvisionStagingByContract(ctx, contract)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> GetProvisionStagingByContract: \n%w\n", err)
	}

	// update mac address of device record
	err = s.deviceRepo.UpdateDeviceMacAddress(ctx, prv.DeviceID, macAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> UpdateDeviceMacAddress: \n%w\n", err)
	}
	// return device (ID will be used for influxdb entries)
	device, err := s.deviceRepo.GetDeviceByMacAddress(ctx, macAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> GetDeviceByMacAddress: \n%w\n", err)
	}
	return device, nil
}
