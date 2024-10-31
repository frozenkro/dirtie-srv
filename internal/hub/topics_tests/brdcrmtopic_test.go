// Integration tests between mqtt handlers and databases
package topics_tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/int_tst"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/frozenkro/dirtie-srv/internal/hub/topics/brdcrmtopic"
	"github.com/frozenkro/dirtie-srv/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestBrdCrmInvokeTopic(t *testing.T) {
	ctx := int_tst.TestContext(t)
	db := int_tst.SetupTests(ctx, t)
	defer db.Close(ctx)

	deps := di.NewDeps(ctx)
	sut := brdcrmtopic.NewBrdCrmTopic(deps.BrdCrmSvc)

	t.Run("Success", func(t *testing.T) {
		data := services.BreadCrumb{
			MacAddr:     int_tst.TestDevice.MacAddr.String,
			Capacitance: 1234,
			Temperature: 69,
		}
		dBytes, err := json.Marshal(data)
		if err != nil {
			t.Errorf("Error encoding test breadcrumb: %v", err)
		}

		err = sut.InvokeTopic(ctx, dBytes)

		assert.Nil(t, err, fmt.Sprintf("InvokeTopic error: %v", err))

		capData, err := deps.InfluxRepo.GetLatestValue(ctx, int(int_tst.TestDevice.DeviceID), core.Capacitance)
		if err != nil {
			t.Errorf("Error retrieving capacitance data point: %v", err)
		}
		assert.Equal(t, data.Capacitance, capData.Value)

		tempData, err := deps.InfluxRepo.GetLatestValue(ctx, int(int_tst.TestDevice.DeviceID), core.Temperature)
		if err != nil {
			t.Errorf("Error retrieving temperature data point: %v", err)
		}
		assert.Equal(t, data.Temperature, tempData.Value)
	})
  t.Run("UnrecognizedDevice", func(t *testing.T) {
    data := services.BreadCrumb{
      MacAddr: "d035n0t3x1st",
      Capacitance: 420,
      Temperature: 69,
    }
    dBytes, err := json.Marshal(data)
    if err != nil {
			t.Errorf("Error encoding test breadcrumb: %v", err)
    }

    err = sut.InvokeTopic(ctx, dBytes)

    assert.NotNil(t, err)
    assert.ErrorIs(t, err, services.ErrNoDevice)
  })
}
