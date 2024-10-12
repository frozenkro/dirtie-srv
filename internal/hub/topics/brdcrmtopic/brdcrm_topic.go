package brdcrm_topic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/services"
)

type BrdCrmTopic struct {
	brdCrmSvc services.BrdCrmSvc
}

func NewBrdCrmTopic(brdCrmSvc services.BrdCrmSvc) *BrdCrmTopic {
	return &BrdCrmTopic{brdCrmSvc: brdCrmSvc}
}

func (t *BrdCrmTopic) InvokeTopic(ctx context.Context, payload []byte) error {
	var data services.BreadCrumb
	err := json.Unmarshal(payload, data)
	if err != nil {
		return fmt.Errorf("Error BrdCrmTopic InvokeTopic -> Unmarshal: %w", err)
	}

	err = t.brdCrmSvc.RecordBrdCrm(ctx, data)
	if err != nil {
		return fmt.Errorf("Error BrdCrmTopic InvokeTopic -> RecordBrdCrm: %w", err)
	}

	return nil
}
