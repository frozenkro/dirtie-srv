package services

import (
	"context"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	dataRec mocks.MockDeviceDataRecorder
	devGet  mocks.MockDeviceGetter
	sut     BrdCrmSvc
)

func setupBrdCrmSvcTests() {
	dataRec = mocks.MockDeviceDataRecorder{Mock: new(mock.Mock)}
	dataRet = mocks.MockDeviceDataRetriever{Mock: new(mock.Mock)}
	devGet = mocks.MockDeviceGetter{Mock: new(mock.Mock)}

	sut = NewBrdCrmSvc(dataRec, dataRet, devGet)
}

func TestRecordBrdCrm(t *testing.T) {
	ctx := context.Background()
	setupBrdCrmSvcTests()

	t.Run("Success", func(t *testing.T) {
		brdCrm := BreadCrumb{
			MacAddr:     "TestMacAddr",
			Capacitance: 420,
			Temperature: 69,
		}
		dvc := sqlc.Device{
			DeviceID: 111,
			UserID:   222,
			MacAddr: pgtype.Text{
				String: brdCrm.MacAddr,
				Valid:  true,
			},
			DisplayName: pgtype.Text{
				String: "Testie",
				Valid:  true,
			},
		}

		devGet.On("GetDeviceByMacAddress", ctx, brdCrm.MacAddr).Return(dvc)
		dataRec.On("Record", ctx, int(dvc.DeviceID), core.Capacitance, brdCrm.Capacitance).Return(nil)
		dataRec.On("Record", ctx, int(dvc.DeviceID), core.Temperature, brdCrm.Temperature).Return(nil)

		err := sut.RecordBrdCrm(ctx, brdCrm)
		assert.Nil(t, err)

		devGet.AssertCalled(t, "GetDeviceByMacAddress", ctx, brdCrm.MacAddr)
		dataRec.AssertCalled(t, "Record", ctx, int(dvc.DeviceID), core.Capacitance, brdCrm.Capacitance)
		dataRec.AssertCalled(t, "Record", ctx, int(dvc.DeviceID), core.Temperature, brdCrm.Temperature)
	})
}
