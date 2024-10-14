// Integration tests between mqtt handlers and databases
package topics_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/frozenkro/dirtie-srv/internal/core/int_tst"
	"github.com/frozenkro/dirtie-srv/internal/di"
	brdcrm_topic "github.com/frozenkro/dirtie-srv/internal/hub/topics/brdcrmtopic"
	"github.com/frozenkro/dirtie-srv/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestBrdCrmInvokeTopic(t *testing.T) {
  ctx := context.Background()
  db := int_tst.SetupTests()
  defer db.Close(ctx)

  deps := di.NewDeps()
  sut := brdcrm_topic.NewBrdCrmTopic(deps.BrdCrmSvc)

  t.Run("Success", func(t *testing.T) {
    data := services.BreadCrumb{
      MacAddr: int_tst.TestDevice.MacAddr.String,
      Capacitance: 1234,
      Temperature: 69,
    }
    dBytes, err := json.Marshal(data)
    if err != nil {
      t.Errorf("Error encoding test breadcrumb: %v", err)
    }

    err = sut.InvokeTopic(ctx, dBytes)

    assert.Nil(t, err, fmt.Sprintf("InvokeTopic error: %v", err.Error()))

    capData, err := deps.DeviceDataRetriever.GetLatestValue(ctx, int(int_tst.TestDevice.DeviceID), core.Capacitance)
    if err != nil {
      t.Errorf("Error retrieving capacitance data point: %v", err)
    }
    assert.Equal(t, data.Capacitance, capData.Value)

    tempData, err := deps.DeviceDataRetriever.GetLatestValue(ctx, int(int_tst.TestDevice.DeviceID), core.Temperature)
    if err != nil {
      t.Errorf("Error retrieving temperature data point: %v", err)
    }
    assert.Equal(t, data.Temperature, tempData.Value)
  })
}
