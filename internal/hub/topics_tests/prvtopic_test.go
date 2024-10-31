package topics_tests

import (
	"encoding/json"
	"testing"

	"github.com/frozenkro/dirtie-srv/internal/core/int_tst"
	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/di"
	"github.com/frozenkro/dirtie-srv/internal/hub/topics/prvtopic"
	"github.com/frozenkro/dirtie-srv/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestPrvInvokeTopic(t *testing.T) {
  ctx := int_tst.TestContext(t)
  db := int_tst.SetupTests(ctx, t)
  defer db.Close(ctx)

  deps := di.NewDeps(ctx)
  sut := prvtopic.NewProvisionTopic(deps.DeviceSvc)

  t.Run("Success", func(t *testing.T) {
    data := services.DevicePrvPayload {
      MacAddr: "s@mp13m4c@ddr355",
      Contract: int_tst.TestProvStg.Contract.String,
    }
    dBytes, err := json.Marshal(data)
    if err != nil {
      t.Fatalf("Error marshaling struct for provision completion test: \n%v\n", err)
    }

    err = sut.InvokeTopic(ctx, dBytes)
    assert.Nil(t, err)

    row, err := db.Query(
      ctx,
      "SELECT device_id, user_id, mac_addr, display_name FROM devices WHERE mac_addr = $1",
      data.MacAddr)

    assert.Nil(t, err, err)

    if !row.Next() {
      t.Fatalf("New device not saved to devices table")
    }

    device := sqlc.Device{}
    err = row.Scan(&device.DeviceID, &device.UserID, &device.MacAddr, &device.DisplayName)
    if err != nil {
      t.Fatalf("Error converting inserted row to device struct: \n%v\n", err)
    }

    assert.False(t, row.Next())
    assert.Equal(t, data.MacAddr, device.MacAddr.String)
  })

  //t.Run("OverwriteDeviceMacAddr", func(t *testing.T) {
  //}
}
