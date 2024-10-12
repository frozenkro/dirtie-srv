package prv_topic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/services"
)

type ProvisionTopic struct {
  deviceSvc services.DeviceSvc
}

func NewProvisionTopic(deviceSvc services.DeviceSvc) *ProvisionTopic {
  return &ProvisionTopic{deviceSvc: deviceSvc}
}

func(t *ProvisionTopic) InvokeTopic(ctx context.Context, payload []byte) error {
  var data services.DevicePrvPayload
  err := json.Unmarshal(payload, data)
  if err != nil {
    return fmt.Errorf("Error ProvisionTopic InvokeTopic -> Unmarshal: %w", err)
  }

  _, err = t.deviceSvc.CompleteDeviceProvision(ctx, data)
  if err != nil {
    return fmt.Errorf("Error ProvisionTopic InvokeTopic -> Unmarshal: %w", err)
  }

  return nil
}
