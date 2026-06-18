package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/frozenkro/dirtie-srv/internal/db"
)


type LogDumpPayload struct {
	MacAddr string `json:"macAddr"`
	Contract string `json:"Contract"`
	LogDump []string `json:"logdump"`
}

type LogPoster interface {
	PostLogs(payload db.LokiLogData) error
}

type LogDumpSvc struct {
	dg  DeviceGetter
	dpc DevicePrvCompleter
	lp  LogPoster
}

func NewLogDumpSvc(dg DeviceGetter, dpc DevicePrvCompleter, lp LogPoster) LogDumpSvc {
	return LogDumpSvc{
		dg:  dg,
		dpc: dpc,
		lp:  lp,
	}
}

func (s LogDumpSvc) DumpLogs(ctx context.Context, payload LogDumpPayload) error {
	dvc, err := s.dg.GetDeviceByMacAddress(ctx, payload.MacAddr)
	if err != nil {
		return fmt.Errorf("Error retrieving device in LogDumpSvc.DumpLogs: %w\n", err)
	}

	// Lazy provisioning
	if dvc.DeviceID <= 0 {
		dpp := DevicePrvPayload{MacAddr: payload.MacAddr, Contract: payload.Contract}
		ps, err := s.dpc.CompleteDeviceProvision(ctx, dpp)
		if err != nil {
			return fmt.Errorf("Error lazy provisioning in DumpLogs.GetProvisionStagingByContract: \n%w\n", err)
		}
		if ps.MacAddr.String == "" {
			// No device or provision staging record found for this contract / mac address
			return fmt.Errorf("Error in LogDumpSvc.DumpLogs (macAddr: %v): \n%w\n", payload.MacAddr, ErrNoDevice)
		}
	}

	devIdStr := strconv.Itoa(int(dvc.DeviceID))
	infoStream := db.LokiLogStream{
		Stream: db.LokiLogTags{
			MacAddr: payload.MacAddr,
			Contract: payload.Contract,
			DeviceId: devIdStr,
			Source: db.LogSource_Device,
			Level: db.LogLevel_Info,
		},
	}
	errStream := db.LokiLogStream{
		Stream: db.LokiLogTags{
			MacAddr: payload.MacAddr,
			Contract: payload.Contract,
			DeviceId: devIdStr,
			Source: db.LogSource_Device,
			Level: db.LogLevel_Error,
		},
	}
	unkStream := db.LokiLogStream{
		Stream: db.LokiLogTags{
			MacAddr: payload.MacAddr,
			Contract: payload.Contract,
			DeviceId: devIdStr,
			Source: db.LogSource_Device,
			Level: db.LogLevel_Unk,
		},
	}

	errAgg := make([]error, 0)
	for _, ld := range payload.LogDump {
		var entry [2]string

		// decode each log
		dec, err := Decode(ld);
		if err != nil {
			errAgg = append(errAgg, err)
			continue
		}

		// get timestamp
		ts_start := strings.Index(dec, "[")
		ts_end := strings.Index(dec, "]")
		if (ts_start < 0 || ts_end < 0) {
			errAgg = append(errAgg, fmt.Errorf("Malformed log, no timestamp found\n"))
		}
		ts := dec[ts_start+1:ts_end]

    ns := strconv.FormatUint(func(s string) uint64 {
        v, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					errAgg = append(errAgg, err)
				}
        return v * 1000000
    }(ts), 10)
		entry[0] = ns

		// Get log level
		lv_start := strings.Index(dec[ts_end:], "[")
		lv_end := strings.Index(dec[lv_start:], "]")

		var lv_str string
		if (lv_start < 0 || lv_end < 0) {
			lv_str = "UNK"
		} else {
			lv_str = dec[lv_start+1:lv_end]
		}

		var lv db.LogLevel
		if lv_str == "INFO" {
			lv = db.LogLevel_Info
		} else if lv_str == "ERR" {
			lv = db.LogLevel_Error
		} else {
			lv = db.LogLevel_Unk
		}

		log_msg := strings.Trim(dec[lv_end:], " ")
		entry[1] = log_msg

		switch lv {
		case db.LogLevel_Info:
			infoStream.Values = append(infoStream.Values, entry)
		case db.LogLevel_Error:
			errStream.Values = append(errStream.Values, entry)
		case db.LogLevel_Unk:
			unkStream.Values = append(unkStream.Values, entry)
		default:
			// unreachable
		}
	}
	if len(errAgg) > 0 {
		errStr := ""
		for _, e := range errAgg {
			errStr = fmt.Sprintf("%s\n%s", errStr, strings.Trim(e.Error(), "\n"))
		}
		return fmt.Errorf("Error(s) parsing logs in LogDumpSvc.DumpLogs:\n%s\n", errStr)
	}

	data := db.LokiLogData{}
	if len(infoStream.Values) > 0 {
		data.Streams = append(data.Streams, infoStream)
	}
	if len(errStream.Values) > 0 {
		data.Streams = append(data.Streams, errStream)
	}
	if len(unkStream.Values) > 0 {
		data.Streams = append(data.Streams, unkStream)
	}

	// post to loki api
	err = s.lp.PostLogs(data)
	if err != nil {
		return fmt.Errorf("Error posting logs to loki in LogDumpSvc.DumpLogs: %w\n", err)
	}

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
