package services

import (
	"context"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services/mocks"
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
}
