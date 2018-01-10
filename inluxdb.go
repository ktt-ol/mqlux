package mqlux

import (
	"net/url"
	"time"

	"github.com/influxdb/influxdb/client"
)

type InfluxDBClient struct {
	client   *client.Client
	database string
}

func NewInfluxDBClient(conf Config) (*InfluxDBClient, error) {
	clientCfg := client.NewConfig()
	u, err := url.Parse(conf.InfluxDB.URL)
	if err != nil {
		return nil, err
	}
	clientCfg.URL = *u
	clientCfg.Username = conf.InfluxDB.Username
	clientCfg.Password = conf.InfluxDB.Password

	c, err := client.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return &InfluxDBClient{
		client:   c,
		database: conf.InfluxDB.Database,
	}, nil
}

func (i *InfluxDBClient) WriteStatus(s SpaceStatus) error {
	if i == nil {
		return nil
	}
	return i.writePoints(statusPoints(s))
}

func (i *InfluxDBClient) WriteDevices(d Devices) error {
	if i == nil {
		return nil
	}
	return i.writePoints(devicesPoints(d))
}

func (i *InfluxDBClient) WriteSensor(s SensorConfig, v float64) error {
	if i == nil {
		return nil
	}
	return i.writePoints(sensorPoints(s, v))
}

func (i *InfluxDBClient) writePoints(pts []client.Point) error {
	bps := client.BatchPoints{
		Points:          pts,
		Database:        i.database,
		RetentionPolicy: "default",
	}

	_, err := i.client.Write(bps)
	if err != nil {
		return err
	}
	return nil
}

func statusPoints(s SpaceStatus) []client.Point {
	now := time.Now()
	open := 0.0
	if s.Open {
		open = 1.0
	}
	if s.Closing {
		open = 0.5
	}
	return []client.Point{
		{
			Measurement: "space_open",
			Fields: map[string]interface{}{
				"value": open,
			},
			Time: now,
		},
	}
}

func devicesPoints(d Devices) []client.Point {
	now := time.Now()
	return []client.Point{
		{
			Measurement: "devices_total",
			Fields: map[string]interface{}{
				"value": float64(d.Devices),
			},
			Time: now,
		},
		{
			Measurement: "devices_unknown",
			Fields: map[string]interface{}{
				"value": float64(d.Unknown),
			},
			Time: now,
		},
		{
			Measurement: "people",
			Fields: map[string]interface{}{
				"value": float64(d.People),
			},
			Time: now,
		},
	}
}

func sensorPoints(s SensorConfig, v float64) []client.Point {
	now := time.Now()
	return []client.Point{
		{
			Measurement: s.Measurement,
			Fields: map[string]interface{}{
				"value": v,
			},
			Tags: s.Tags,
			Time: now,
		},
	}
}
