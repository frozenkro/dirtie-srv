package monitor

import (
	"context"
	"os"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
)

var (
	client *influxdb2.Client
)

func RecordCapacitance(deviceOid string, capacitance int64) error {
	c := *client
	writeAPI := c.WriteAPIBlocking(os.Getenv("INFLUX_ORG"), os.Getenv("INFLUX_DEFAULT_BUCKET"))

	p := influxdb2.NewPointWithMeasurement("capacitance").
		AddTag("device", deviceOid).
		AddField("capacitance", capacitance).
		SetTime(time.Now())
	err := writeAPI.WritePoint(context.Background(), p)

	return err
}

func Connect() *influxdb2.Client {
	uri, ok := os.LookupEnv("INFLUX_URI")
	if !ok {
		uri = "localhost:8086"
	}
	c := influxdb2.NewClient("http:"+uri, os.Getenv("INFLUX_TOKEN"))
	client = &c

	return client
}

func Disconnect() {
	c := *client
	c.Close()
}
