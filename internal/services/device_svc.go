package services

import (
	"context"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/google/uuid"
)

type DeviceReader interface {
	GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error)
  GetDevicesByUser(ctx context.Context, userId int32) ([]sqlc.Device, error)
}

type ProvisionStagingReader interface {
	GetProvisionStagingByContract(ctx context.Context, contract string) (sqlc.ProvisionStaging, error)
}
type ProvisionStagingWriter interface {
	CreateProvisionStaging(ctx context.Context, deviceId int32, contract string) error
	DeleteProvisionStaging(ctx context.Context, deviceId int32) error
}


type DeviceWriter interface {
	CreateDevice(ctx context.Context, userId int32, displayName string) (sqlc.Device, error)
	RenameDevice(ctx context.Context, deviceId int32, displayName string) error
	UpdateDeviceMacAddress(ctx context.Context, deviceId int32, macAddr string) error
}

type UserCtxReader interface {
	GetUser(ctx context.Context) (sqlc.User, error)
}


type DeviceSvc struct {
	deviceReader DeviceReader
	deviceWriter DeviceWriter
	prvStgReader ProvisionStagingReader
	prvStgWriter ProvisionStagingWriter
	userCtxReader UserCtxReader
}

type DevicePrvPayload struct {
	MacAddr  string
	Contract string
}

func NewDeviceSvc(deviceReader DeviceReader,
  deviceWriter DeviceWriter,
	prvStgReader ProvisionStagingReader,
	prvStgWriter ProvisionStagingWriter,
	userCtxReader UserCtxReader) *DeviceSvc {

  return &DeviceSvc{
    deviceReader: deviceReader,
    deviceWriter: deviceWriter,
		prvStgReader: prvStgReader,
		prvStgWriter: prvStgWriter,
		userCtxReader: userCtxReader,
	}
}

func (s DeviceSvc) GetUserDevices(ctx context.Context) ([]sqlc.Device, error) {
	user, err := s.userCtxReader.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error GetUserDevices -> GetUser: \n%w\n", err)
	}

	devices, err := s.deviceReader.GetDevicesByUser(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error GetUserDevices -> GetDevicesByUser: \n%w\n", err)
	}
	return devices, nil
}

func (s DeviceSvc) GetDeviceByMacAddress(ctx context.Context, macAddr string) (sqlc.Device, error) {
	device, err := s.deviceReader.GetDeviceByMacAddress(ctx, macAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error GetDeviceByMacAddr -> GetDeviceByMacAddress: \n%w\n", err)
	}
	return device, nil
}

// Called by user via rest api
func (s DeviceSvc) CreateDeviceProvision(ctx context.Context, displayName string) (string, error) {
	user, err := s.userCtxReader.GetUser(ctx)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> GetUser: \n%w\n", err)
	}

	device, err := s.deviceWriter.CreateDevice(ctx, user.UserID, displayName)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> CreateDevice: \n%w\n", err)
	}
	uuid := uuid.NewString()
	err = s.prvStgWriter.CreateProvisionStaging(ctx, device.DeviceID, uuid)
	if err != nil {
		return "", fmt.Errorf("Error CreateDeviceProvision -> CreateProvisionStaging: \n%w\n", err)
	}
	return uuid, nil
}

// Called by device via mqtt hub
func (s DeviceSvc) CompleteDeviceProvision(ctx context.Context, data DevicePrvPayload) (sqlc.Device, error) {
	// lookup contract (uuid) from provision staging table
	prv, err := s.prvStgReader.GetProvisionStagingByContract(ctx, data.Contract)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> GetProvisionStagingByContract: \n%w\n", err)
	}

	// update mac address of device record
	err = s.deviceWriter.UpdateDeviceMacAddress(ctx, prv.DeviceID, data.MacAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> UpdateDeviceMacAddress: \n%w\n", err)
	}
	// return device (ID will be used for influxdb entries)
	device, err := s.deviceReader.GetDeviceByMacAddress(ctx, data.MacAddr)
	if err != nil {
		return sqlc.Device{}, fmt.Errorf("Error CompleteDeviceProvision -> GetDeviceByMacAddress: \n%w\n", err)
	}
	return device, nil
}
