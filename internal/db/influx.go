package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type DeviceDataPoint struct {
	Value int64     `json:"value"`
	Time  time.Time `json:"time"`
	Key   string    `json:"key"`
}

type InfluxRepo struct {
	client *influxdb2.Client
}

func NewInfluxRepo() InfluxRepo {
	c := initIxClient()
	return InfluxRepo{client: &c}
}

func initIxClient() influxdb2.Client {
	uri := core.INFLUX_URI
	return influxdb2.NewClient("http://"+uri, core.INFLUX_TOKEN)
}

func (r InfluxRepo) Record(ctx context.Context, deviceId int, measurementKey string, value int64) error {
	c := *r.client
	writeAPI := c.WriteAPIBlocking(core.INFLUX_ORG, core.INFLUX_DEFAULT_BUCKET)

	p := influxdb2.NewPointWithMeasurement(measurementKey).
		AddTag("device", strconv.Itoa(deviceId)).
		AddField(measurementKey, value).
		SetTime(time.Now())
	err := writeAPI.WritePoint(ctx, p)

	return err
}

func (r InfluxRepo) GetLatestValue(
	ctx context.Context,
	deviceId int,
	measurementKey string) (DeviceDataPoint, error) {
	c := *r.client
	queryAPI := c.QueryAPI(core.INFLUX_ORG)

	query := fmt.Sprintf(`
    from(bucket:"%v")
    |> range(start: -1w)
    |> filter(fn: (r) => r._measurement == "%v" and r._field == "%v")
    |> last()`, core.INFLUX_DEFAULT_BUCKET, measurementKey, measurementKey)

	qRes, err := queryAPI.Query(ctx, query)
	if err != nil {
		return DeviceDataPoint{}, fmt.Errorf("Error GetLatestValue -> Query: %w", err)
	}

	if !qRes.Next() {
		return DeviceDataPoint{}, nil
	}

	return newDeviceDataPoint(qRes)
}

func (r InfluxRepo) GetValuesRange(
	ctx context.Context,
	deviceId int,
	measurementKey string,
	start time.Time,
	end time.Time) ([]DeviceDataPoint, error) {
	c := *r.client
	queryAPI := c.QueryAPI(core.INFLUX_ORG)

	query := fmt.Sprintf(`
    from(bucket:"%v")
    |> range(start: "%v", stop: "%v")
    |> filter(fn: (r) => r._measurement == "%v" and r._field == "%v")
  `, core.INFLUX_DEFAULT_BUCKET, measurementKey, measurementKey, start.Format(time.RFC3339), end.Format(time.RFC3339))

	qRes, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Error GetValuesRange -> Query: %w", err)
	}

	return llToSlice(qRes)
}

func llToSlice(r *api.QueryTableResult) ([]DeviceDataPoint, error) {
	var d []DeviceDataPoint
	for r.Next() {
		p, err := newDeviceDataPoint(r)
		if err != nil {
			return nil, fmt.Errorf(
				"Error GetValuesRange -> newDeviceDataPoint: %w",
				err)
		}
		d = append(d, p)
	}
	return d, nil
}

func newDeviceDataPoint(r *api.QueryTableResult) (DeviceDataPoint, error) {
	val := r.Record().Value()
	if r == nil || r.Record() == nil {
		return DeviceDataPoint{}, fmt.Errorf(
			`Error in newDeviceDataPoint - no influx result`,
		)
	}
	valInt, succ := val.(int64)
	if !succ {
		return DeviceDataPoint{}, fmt.Errorf(
			`Error in newDeviceDataPoint - failed to cast influx result. 
      deviceId: '%v', measurementKey: '%v'`,
			r.Record().ValueByKey("device"),
			r.Record().Measurement(),
		)
	}
	return DeviceDataPoint{
		Value: valInt,
		Time:  r.Record().Time(),
		Key:   r.Record().Field(),
	}, nil
}

func (r *InfluxRepo) Disconnect() {
	c := *r.client
	c.Close()
}
