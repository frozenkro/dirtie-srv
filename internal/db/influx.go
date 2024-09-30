package db

import (
	"context"
  "fmt"
	"os"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
)

type DeviceDataPoint struct {
  value int64
  time time.Time
  key string
}

type DeviceDataRecorder interface {
  Record(ctx context.Context, deviceId int, measurementKey string, value int64) error
}

type DeviceDataRetriever interface {
  GetLatestValue(ctx context.Context, deviceId int, measurementKey string) (DeviceDataPoint, error)
  GetValuesRange(ctx context.Context, deviceId int, measurementKey string, start time.Time, end time.Time) ([]DeviceDataPoint, error)
}

type InfluxRepo struct {
  client *influxdb2.Client
}

func NewInfluxRepo() InfluxRepo {
  c := initIxClient()
  return InfluxRepo{ client: &c }
}

func initIxClient() influxdb2.Client {
	uri, ok := os.LookupEnv("INFLUX_URI")
	if !ok {
		uri = "localhost:8086"
	}
	return influxdb2.NewClient("http:"+uri, os.Getenv("INFLUX_TOKEN"))
}

func (r *InfluxRepo) Record(ctx context.Context, deviceId string, measurementKey string, value int64) error {
	c := *r.client
	writeAPI := c.WriteAPIBlocking(os.Getenv("INFLUX_ORG"), os.Getenv("INFLUX_DEFAULT_BUCKET"))

	p := influxdb2.NewPointWithMeasurement(measurementKey).
		AddTag("device", deviceId).
		AddField(measurementKey, value).
		SetTime(time.Now())
	err := writeAPI.WritePoint(ctx, p)

	return err
}

func (r *InfluxRepo) GetLatestValue(
  ctx context.Context, 
  deviceId int, 
  measurementKey string) (DeviceDataPoint, error) {
  c := *r.client
  queryAPI := c.QueryAPI(os.Getenv("INFLUX_ORG"))
  
  query := fmt.Sprintf(`
    from(bucket:"%v")
    |> filter(fn: (r) => r._measurement == "%v" and r._field == "%v")
    |> sort(columns: ["_time"], desc: true)
    |> limit(n:1)`, os.Getenv("INFLUX_DEFAULT_BUCKET"), measurementKey, measurementKey)

  qRes, err := queryAPI.Query(ctx, query)
  if err != nil {
    return DeviceDataPoint{}, fmt.Errorf("Error GetLatestValue -> Query: %w", err)
  }
  
  // TODO break me out to testable unit
  val := qRes.Record().Value()
  valInt, succ := val.(int64)
  if !succ {
    return DeviceDataPoint{}, fmt.Errorf("Error in GetLatestValue - failed to cast influx result. deviceId: '%v', measurementKey: '%v'", deviceId, measurementKey)
  }
  
  return DeviceDataPoint{
    value: valInt,
    time: qRes.Record().Time(),
    key: qRes.Record().Field(),
  }, nil
}

func (r *InfluxRepo) GetValuesRange(
  ctx context.Context,
  deviceId int, 
  measurementKey string, 
  start time.Time, 
  end time.Time) ([]DeviceDataPoint, error) {
  c := *r.client
  queryAPI := c.QueryAPI(os.Getenv("INFLUX_ORG"))

  query := fmt.Sprintf(`
    from(bucket:"%v")
    |> filter(fn: (r) => r._measurement == "%v" and r._field == "%v")
    |> range(start: "%v", stop: "%v")
  `, measurementKey, measurementKey, start.Format(time.RFC3339), end.Format(time.RFC3339))

  qRes, err := queryAPI.Query(ctx, query)
  if err != nil {
    return nil, fmt.Errorf("Error GetValuesRange -> Query: %w", err)
  }

  // TODO break me out to testable unit
  for qRes.Next() {
    // TODO.. linked list to array I guess.
  }
  //temp
  return nil, nil
}

func (r *InfluxRepo) Disconnect() {
	c := *r.client
	c.Close()
}
