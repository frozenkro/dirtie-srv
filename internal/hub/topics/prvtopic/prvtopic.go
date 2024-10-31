package prvtopic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/db/sqlc"
	"github.com/frozenkro/dirtie-srv/internal/services"
)

type DevicePrvCompleter interface {
	CompleteDeviceProvision(context.Context, services.DevicePrvPayload) (sqlc.Device, error)
}

type ProvisionTopic struct {
	dpc DevicePrvCompleter
}

func NewProvisionTopic(service DevicePrvCompleter) *ProvisionTopic {
	return &ProvisionTopic{dpc: service}
}

func (t *ProvisionTopic) InvokeTopic(ctx context.Context, payload []byte) error {
	var data services.DevicePrvPayload
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return fmt.Errorf("Error ProvisionTopic InvokeTopic -> Unmarshal: %w", err)
	}

	_, err = t.dpc.CompleteDeviceProvision(ctx, data)
	if err != nil {
		return fmt.Errorf("Error ProvisionTopic InvokeTopic -> CompleteDeviceProvision: %w", err)
	}

	return nil
}
