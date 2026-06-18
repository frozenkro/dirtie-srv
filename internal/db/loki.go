package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
)

type LokiClient struct{
	client *http.Client
}

func NewLokiClient() *LokiClient {
	return &LokiClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type LogSource string
const (
	LogSource_Device LogSource = "Device"
	LogSource_Api    LogSource = "Api"
	LogSource_Mobile LogSource = "Mobile"
)

type LogLevel string
const (
	LogLevel_Info LogLevel = "INFO"
	LogLevel_Error LogLevel = "ERROR"
	LogLevel_Unk LogLevel = "UNKNOWN"
)

type LokiLogTags struct {
	MacAddr string `json:"mac_addr,omitempty"`
	Contract string `json:"contract,omitempty"`
	DeviceId string `json:"DeviceId,omitempty"`
	Source LogSource `json:"source"`
	Level LogLevel `json:"level"`
}

type LokiLogStream struct {
	Stream LokiLogTags `json:"stream"`
	Values [][2]string `json:"values"`
}

type LokiLogData struct {
	Streams []LokiLogStream `json:"streams"`
}

func (c *LokiClient) PostLogs(data LokiLogData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error marshaling payload in LokiClient.PostLogs: %w\n", err)
	}
	bReader := bytes.NewReader(b)

	lokiUri, err := url.JoinPath(strings.Trim(core.LOKI_URL, "/"), "loki", "api", "v1", "push")
	if err != nil {
		return fmt.Errorf("Error creating URI in LokiClient.PostLogs: %w\n", err)
	}

	req, err := http.NewRequest(http.MethodPost, lokiUri, bReader)
	if err != nil {
		return fmt.Errorf("Error creating http Request in LokiClient.PostLogs: %w\n", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error pushing logs to Loki in LokiClient.PostLogs: %w\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		rescon, err := io.ReadAll(res.Body)
		if err != nil {
			rescon = []byte(fmt.Sprintf("Error reading response body in LokiClient.PostLogs: %v\n", err.Error()))
		}

		return fmt.Errorf("Response code '%v' in LokiClient.PostLogs. Message from server: \n%v\n", res.StatusCode, rescon)
	}

	return nil
}
