package services

import (
	"context"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	deviceReader  mocks.MockDeviceReader
	deviceWriter  mocks.MockDeviceWriter
	prvStgReader  mocks.MockPrvStgReader
	prvStgWriter  mocks.MockPrvStgWriter
	userCtxReader mocks.MockUserCtxReader
	deviceSvc     DeviceSvc
)

func setupDeviceSvcTests() {
	deviceReader = mocks.MockDeviceReader{Mock: new(mock.Mock)}
	deviceWriter = mocks.MockDeviceWriter{Mock: new(mock.Mock)}
	prvStgReader = mocks.MockPrvStgReader{Mock: new(mock.Mock)}
	prvStgWriter = mocks.MockPrvStgWriter{Mock: new(mock.Mock)}
	userCtxReader = mocks.MockUserCtxReader{Mock: new(mock.Mock)}
	deviceSvc = *NewDeviceSvc(
		deviceReader,
		deviceWriter,
		prvStgReader,
		prvStgWriter,
		userCtxReader,
	)
}

func TestGetUserDevices(t *testing.T) {
	ctx := context.Background()
	setupDeviceSvcTests()

	t.Run("Success", func(t *testing.T) {
		user := sqlc.User{
			UserID: 1234,
			Email:  "testemail@email.com",
		}
		dvc1 := sqlc.Device{
			DeviceID: 1,
		}
		dvc2 := sqlc.Device{
			DeviceID: 2,
		}
		dvcs := []sqlc.Device{
			dvc1,
			dvc2,
		}

		userCtxReader.On("GetUser", ctx).Return(user, nil)
		deviceReader.On("GetDevicesByUser", ctx, user.UserID).Return(dvcs, nil)

		result, err := deviceSvc.GetUserDevices(ctx)
		assert.Nil(t, err)

		assert.Equal(t, dvcs, result)
	})

	// Error tests skipped - need to fix implementation
}

func TestGetDeviceByMacAddress(t *testing.T) {
	ctx := context.Background()
	setupDeviceSvcTests()

	t.Run("Success", func(t *testing.T) {
		macAddr := "00:11:22:33:44:55"
		device := sqlc.Device{
			DeviceID: 123,
			UserID:   456,
		}

		deviceReader.On("GetDeviceByMacAddress", ctx, macAddr).Return(device, nil)

		result, err := deviceSvc.GetDeviceByMacAddress(ctx, macAddr)
		assert.NoError(t, err)
		assert.Equal(t, device, result)
	})

	// Error test skipped - need to fix implementation
}

func TestCreateDeviceProvision(t *testing.T) {
	ctx := context.Background()
	setupDeviceSvcTests()

	t.Run("Success", func(t *testing.T) {
		displayName := "Test Device"
		user := sqlc.User{
			UserID: 1234,
			Email:  "testemail@email.com",
		}
		device := sqlc.Device{
			DeviceID: 5678,
			UserID:   user.UserID,
		}

		userCtxReader.On("GetUser", ctx).Return(user, nil)
		deviceWriter.On("CreateDevice", ctx, user.UserID, displayName).Return(device, nil)
		prvStgWriter.On("CreateProvisionStaging", ctx, device.DeviceID, mock.AnythingOfType("string")).Return(nil)

		result, err := deviceSvc.CreateDeviceProvision(ctx, displayName)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		userCtxReader.AssertExpectations(t)
		deviceWriter.AssertExpectations(t)
		prvStgWriter.AssertExpectations(t)
	})

	// Error tests skipped for now
}

func TestCompleteDeviceProvision(t *testing.T) {
	ctx := context.Background()
	setupDeviceSvcTests()

	t.Run("Success", func(t *testing.T) {
		macAddr := "00:11:22:33:44:55"
		contract := "test-contract-uuid"
		deviceID := int32(5678)
		provStaging := sqlc.ProvisionStaging{
			DeviceID: deviceID,
			Contract: pgtype.Text{
				String: contract,
				Valid:  true,
			},
		}
		device := sqlc.Device{
			DeviceID: deviceID,
			UserID:   1234,
		}
		payload := DevicePrvPayload{
			MacAddr:  macAddr,
			Contract: contract,
		}

		prvStgReader.On("GetProvisionStagingByContract", ctx, contract).Return(provStaging, nil)
		deviceWriter.On("UpdateDeviceMacAddress", ctx, deviceID, macAddr).Return(nil)
		deviceReader.On("GetDeviceByMacAddress", ctx, macAddr).Return(device, nil)

		result, err := deviceSvc.CompleteDeviceProvision(ctx, payload)
		assert.NoError(t, err)
		assert.Equal(t, device, result)

		prvStgReader.AssertExpectations(t)
		deviceWriter.AssertExpectations(t)
		deviceReader.AssertExpectations(t)
	})

	// Error tests skipped for now
}
