package services

import (
	"context"
	"testing"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/db"
	"github.com/frozenkro/dirtie-srv/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
var(
  dataRet mocks.MockDeviceDataRetriever
  deviceGet mocks.MockDeviceGetter
  dataSvc DataSvc
)

func setupDataSvcTests() {
  dataRet = mocks.MockDeviceDataRetriever{ Mock: new(mock.Mock) }
  deviceGet = mocks.MockDeviceGetter{ Mock: new(mock.Mock) }
  dataSvc = NewDataSvc(dataRet, deviceGet)
}

func testCapacitanceData(t *testing.T) {
  ctx := context.Background()
  setupDataSvcTests()

  t.Run("Success", func(t *testing.T) {
    deviceId := 123
    now := time.Now()
    startTime := now.Add(-3 * time.Hour).Format(time.RFC3339)

    expData := []db.DeviceDataPoint {
      db.DeviceDataPoint {
        Value: 1234,
        Time: now.Add(-1 * time.Hour),
        Key: core.Capacitance,
      },
      db.DeviceDataPoint {
        Value: 1233,
        Time: now.Add(-2 * time.Hour),
        Key: core.Capacitance,
      },
    }
    
    dataRet.On("GetValuesRange", 
      ctx, deviceId, core.Capacitance, 
      startTime, mock.AnythingOfType("time.Time"),
      ).Return(expData, nil)

    result, err := dataSvc.CapacitanceData(ctx,
      deviceId,
      startTime)
    if err != nil {
      t.Fatalf("Error in data_svc.CapacitanceData: %v", err)
    }

    assert.Len(t, result, 2)
    assert.Equal(t, expData[0], result[0])
    assert.Equal(t, expData[1], result[1])
  })
}
