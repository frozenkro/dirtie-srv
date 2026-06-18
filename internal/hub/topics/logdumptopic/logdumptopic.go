package logdumptopic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frozenkro/dirtie-srv/internal/services"
)

type LogDumper interface {
	DumpLogs(context.Context, services.LogDumpPayload) error
}

type LogDumpTopic struct {
	ld LogDumper
}

func NewLogDumpTopic(ld LogDumper) *LogDumpTopic {
	return &LogDumpTopic{
		ld: ld,
	}
}

func (t *LogDumpTopic) InvokeTopic(ctx context.Context, payload []byte) error {
	data := services.LogDumpPayload{}
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return fmt.Errorf("Error LogDumpTopic InvokeTopic -> Unmarshal: %w", err)
	}

	err = t.ld.DumpLogs(ctx, data)
	if err != nil {
		return fmt.Errorf("Error LogDumpTopic InvokeTopic -> DumpLogs: %w", err)
	}

	return nil
}
