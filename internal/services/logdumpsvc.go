package services

import (
	"context"
	"encoding/base64"
	"fmt"
)


type LogDumpPayload struct {
	MacAddr string `json:"macAddr"`
	Contract string `json:"Contract"`
	LogDump []string `json:"logdump"`
}

type LogDumpSvc struct {
	dpc DevicePrvCompleter
}

func NewLogDumpSvc(dpc DevicePrvCompleter) LogDumpSvc {
	return LogDumpSvc{
		dpc: dpc,
	}
}

func (s LogDumpSvc) DumpLogs(ctx context.Context, payload LogDumpPayload) error {
	// TODO
	// Logs will be base64 encoded 

	return nil
}

func Decode(encoded string) (string, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	n, err := base64.StdEncoding.Decode(dst, []byte(encoded))
	if err != nil {
		return "", fmt.Errorf("Decode error: %w\n", err)
	}

	dst = dst[:n]
	return string(dst), nil
}
